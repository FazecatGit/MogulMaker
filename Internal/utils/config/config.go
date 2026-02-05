package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Global struct {
		MarketHours struct {
			RegularOpen    string `yaml:"regular_open"`
			RegularClose   string `yaml:"regular_close"`
			PremarketOpen  string `yaml:"premarket_open"`
			AfterhourClose string `yaml:"afterhours_close"`
			Timezone       string `yaml:"timezone"`
		} `yaml:"market_hours"`
		LiquidityMinimumUSD int `yaml:"liquidity_minimum_usd"`
	} `yaml:"global"`

	Notifications struct {
		Channels struct {
			Console bool `yaml:"console"`
			FileLog bool `yaml:"file_log"`
			Discord bool `yaml:"discord"`
		} `yaml:"channels"`
		BatchDigestTime string `yaml:"batch_digest_time"`
	} `yaml:"notifications"`

	Archive struct {
		DaysBeforeArchive    int `yaml:"days_before_archive"`
		RecheckSkipAfterDays int `yaml:"recheck_skip_after_days"`
	} `yaml:"archive"`

	Profiles map[string]ProfileConfig `yaml:"profiles"`

	Features struct {
		CryptoSupport      bool   `yaml:"crypto_support"`
		EnableShortSignals bool   `yaml:"enable_short_signals"`
		AssetType          string `yaml:"asset_type"`
	} `yaml:"features"`
}

type ProfileConfig struct {
	Threshold        float64         `yaml:"threshold"`
	ScanIntervalDays int             `yaml:"scan_interval_days"`
	Indicators       IndicatorConfig `yaml:"indicators"`
	SignalWeights    SignalWeights   `yaml:"signal_weights"`
}

type IndicatorConfig struct {
	RSI    RSIConfig    `yaml:"rsi"`
	ATR    ATRConfig    `yaml:"atr"`
	Volume VolumeConfig `yaml:"volume"`
}

type RSIConfig struct {
	MinOversold   float64 `yaml:"min_oversold"`
	MaxOverbought float64 `yaml:"max_overbought"`
}

type ATRConfig struct {
	MinVolatility float64 `yaml:"min_volatility"`
}

type VolumeConfig struct {
	MinRatio float64 `yaml:"min_ratio"`
}

type SignalWeights struct {
	RSIWeight           float64 `yaml:"rsi_weight"`
	ATRWeight           float64 `yaml:"atr_weight"`
	VolumeWeight        float64 `yaml:"volume_weight"`
	NewsSentimentWeight float64 `yaml:"news_sentiment_weight"`
	WhaleActivityWeight float64 `yaml:"whale_activity_weight"`
}

func LoadConfig() (*Config, error) {
	// Resolve path relative to this file first
	_, filePath, _, ok := runtime.Caller(0)
	var basePath string
	if ok {
		basePath = filepath.Dir(filePath)
	}

	// Get current working directory as fallback
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Try multiple paths to find config.yaml
	possiblePaths := []string{}
	if basePath != "" {
		possiblePaths = append(possiblePaths, filepath.Join(basePath, "config.yaml"))
	}
	possiblePaths = append(possiblePaths,
		filepath.Join(cwd, "Internal", "utils", "config", "config.yaml"),
		"Internal/utils/config/config.yaml",
		"config.yaml",
		filepath.Join("Internal", "utils", "config", "config.yaml"),
	)

	var data []byte
	var foundPath string

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// Log which path was used for debugging
	if foundPath != "" {
		// Optional: log.Printf("Config loaded from: %s", foundPath)
	}

	return &cfg, nil
}

func (c *Config) GetScreenerCriteria(profileName string) map[string]interface{} {
	if profile, exists := c.Profiles[profileName]; exists {
		return map[string]interface{}{
			"MinOversoldRSI": profile.Indicators.RSI.MinOversold,
			"MaxRSI":         profile.Indicators.RSI.MaxOverbought,
			"MinATR":         profile.Indicators.ATR.MinVolatility,
			"MinVolumeRatio": profile.Indicators.Volume.MinRatio,
			"SignalWeights":  profile.SignalWeights,
		}
	}
	return nil
}

func (c *Config) GetProfile(profileName string) *ProfileConfig {
	if profile, exists := c.Profiles[profileName]; exists {
		return &profile
	}
	return nil
}

func SaveConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile("Internal/utils/config/config.yaml", data, 0644)
	if err != nil {
		return err
	}
	return nil
}
