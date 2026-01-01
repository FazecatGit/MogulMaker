package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ConfigureInteractive allows users to interactively configure the system
func ConfigureInteractive(cfg *Config) error {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n⚙️  Configuration Menu:")
		fmt.Println("1. View Current Configuration")
		fmt.Println("2. Configure Profile Thresholds")
		fmt.Println("3. Configure Signal Weights")
		fmt.Println("4. Configure Features")
		fmt.Println("5. Save & Exit")
		fmt.Print("Select option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			DisplayConfiguration(cfg)
		case "2":
			configureProfileThresholds(cfg, reader)
		case "3":
			configureSignalWeights(cfg, reader)
		case "4":
			configureFeatures(cfg, reader)
		case "5":
			err := SaveConfig(cfg)
			if err != nil {
				fmt.Printf("❌ Error saving config: %v\n", err)
				continue
			}
			fmt.Println("✅ Configuration saved successfully!")
			return nil
		default:
			fmt.Println("❌ Invalid option")
		}
	}
}

// DisplayConfiguration shows current configuration
func DisplayConfiguration(cfg *Config) {
	fmt.Println("\n📋 Current Configuration:")
	fmt.Println("\n=== Profiles ===")
	for name, profile := range cfg.Profiles {
		fmt.Printf("\n%s:\n", strings.ToUpper(name))
		fmt.Printf("  • Threshold: %.1f\n", profile.Threshold)
		fmt.Printf("  • Scan Interval: %d days\n", profile.ScanIntervalDays)
		fmt.Printf("  • RSI Min Oversold: %.0f\n", profile.Indicators.RSI.MinOversold)
		fmt.Printf("  • RSI Max Overbought: %.0f\n", profile.Indicators.RSI.MaxOverbought)
		fmt.Printf("  • ATR Min Volatility: %.2f\n", profile.Indicators.ATR.MinVolatility)
		fmt.Printf("  • Volume Min Ratio: %.2f\n", profile.Indicators.Volume.MinRatio)
		fmt.Printf("  • Signal Weights:\n")
		fmt.Printf("    - RSI: %.2f\n", profile.SignalWeights.RSIWeight)
		fmt.Printf("    - ATR: %.2f\n", profile.SignalWeights.ATRWeight)
		fmt.Printf("    - Volume: %.2f\n", profile.SignalWeights.VolumeWeight)
		fmt.Printf("    - News Sentiment: %.2f\n", profile.SignalWeights.NewsSentimentWeight)
		fmt.Printf("    - Whale Activity: %.2f\n", profile.SignalWeights.WhaleActivityWeight)
	}

	fmt.Println("\n=== Features ===")
	fmt.Printf("Crypto Support: %v\n", enabledStr(cfg.Features.CryptoSupport))
	fmt.Printf("Short Signals: %v\n", enabledStr(cfg.Features.EnableShortSignals))

	fmt.Println("\n=== Market Hours ===")
	fmt.Printf("Regular Open: %s\n", cfg.Global.MarketHours.RegularOpen)
	fmt.Printf("Regular Close: %s\n", cfg.Global.MarketHours.RegularClose)
	fmt.Printf("Timezone: %s\n", cfg.Global.MarketHours.Timezone)
}

func configureProfileThresholds(cfg *Config, reader *bufio.Reader) {
	fmt.Println("\n📊 Configure Profile Thresholds:")
	profiles := make([]string, 0)
	for name := range cfg.Profiles {
		profiles = append(profiles, name)
	}

	for i, name := range profiles {
		fmt.Printf("%d. %s\n", i+1, name)
	}
	fmt.Print("Select profile (number): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(profiles) {
		fmt.Println("❌ Invalid selection")
		return
	}

	profileName := profiles[idx-1]
	profile := cfg.Profiles[profileName]

	fmt.Printf("\n✏️  Configuring %s profile:\n", profileName)

	// Threshold
	fmt.Printf("Current threshold: %.1f\n", profile.Threshold)
	fmt.Print("New threshold (0-10): ")
	input, _ := reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		profile.Threshold = val
	}

	// Scan interval
	fmt.Printf("Current scan interval: %d days\n", profile.ScanIntervalDays)
	fmt.Print("New scan interval (days): ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.Atoi(strings.TrimSpace(input)); err == nil {
		profile.ScanIntervalDays = val
	}

	// RSI settings
	fmt.Printf("Current RSI min oversold: %.0f\n", profile.Indicators.RSI.MinOversold)
	fmt.Print("New RSI min oversold (0-100): ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		profile.Indicators.RSI.MinOversold = val
	}

	fmt.Printf("Current RSI max overbought: %.0f\n", profile.Indicators.RSI.MaxOverbought)
	fmt.Print("New RSI max overbought (0-100): ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		profile.Indicators.RSI.MaxOverbought = val
	}

	// ATR settings
	fmt.Printf("Current ATR min volatility: %.2f\n", profile.Indicators.ATR.MinVolatility)
	fmt.Print("New ATR min volatility: ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		profile.Indicators.ATR.MinVolatility = val
	}

	// Volume settings
	fmt.Printf("Current volume min ratio: %.2f\n", profile.Indicators.Volume.MinRatio)
	fmt.Print("New volume min ratio: ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		profile.Indicators.Volume.MinRatio = val
	}

	cfg.Profiles[profileName] = profile
	fmt.Println("✅ Profile updated")
}

func configureSignalWeights(cfg *Config, reader *bufio.Reader) {
	fmt.Println("\n⚖️  Configure Signal Weights:")
	profiles := make([]string, 0)
	for name := range cfg.Profiles {
		profiles = append(profiles, name)
	}

	for i, name := range profiles {
		fmt.Printf("%d. %s\n", i+1, name)
	}
	fmt.Print("Select profile (number): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(profiles) {
		fmt.Println("❌ Invalid selection")
		return
	}

	profileName := profiles[idx-1]
	profile := cfg.Profiles[profileName]
	weights := profile.SignalWeights

	fmt.Printf("\n✏️  Configuring weights for %s:\n", profileName)
	fmt.Println("(Weights should sum to ~1.0 for balanced scoring)")

	fmt.Printf("Current RSI weight: %.2f\n", weights.RSIWeight)
	fmt.Print("New RSI weight: ")
	input, _ := reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		weights.RSIWeight = val
	}

	fmt.Printf("Current ATR weight: %.2f\n", weights.ATRWeight)
	fmt.Print("New ATR weight: ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		weights.ATRWeight = val
	}

	fmt.Printf("Current Volume weight: %.2f\n", weights.VolumeWeight)
	fmt.Print("New Volume weight: ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		weights.VolumeWeight = val
	}

	fmt.Printf("Current News Sentiment weight: %.2f\n", weights.NewsSentimentWeight)
	fmt.Print("New News Sentiment weight: ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		weights.NewsSentimentWeight = val
	}

	fmt.Printf("Current Whale Activity weight: %.2f\n", weights.WhaleActivityWeight)
	fmt.Print("New Whale Activity weight: ")
	input, _ = reader.ReadString('\n')
	if val, err := strconv.ParseFloat(strings.TrimSpace(input), 64); err == nil {
		weights.WhaleActivityWeight = val
	}

	profile.SignalWeights = weights
	cfg.Profiles[profileName] = profile

	// Calculate sum
	sum := weights.RSIWeight + weights.ATRWeight + weights.VolumeWeight + weights.NewsSentimentWeight + weights.WhaleActivityWeight
	fmt.Printf("✅ Weights updated (Sum: %.2f)\n", sum)
	if sum != 1.0 {
		fmt.Printf("⚠️  Note: Weights sum to %.2f (ideally 1.0 for balanced scoring)\n", sum)
	}
}

func configureFeatures(cfg *Config, reader *bufio.Reader) {
	fmt.Println("\n🚀 Configure Features:")
	fmt.Printf("1. Crypto Support: %s\n", enabledStr(cfg.Features.CryptoSupport))
	fmt.Printf("2. Short Signals: %s\n", enabledStr(cfg.Features.EnableShortSignals))
	fmt.Print("Select feature to toggle (1-2) or press Enter to skip: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		cfg.Features.CryptoSupport = !cfg.Features.CryptoSupport
		fmt.Printf("✅ Crypto Support: %s\n", enabledStr(cfg.Features.CryptoSupport))
	case "2":
		cfg.Features.EnableShortSignals = !cfg.Features.EnableShortSignals
		fmt.Printf("✅ Short Signals: %s\n", enabledStr(cfg.Features.EnableShortSignals))
	default:
		fmt.Println("No changes made")
	}
}

func enabledStr(enabled bool) string {
	if enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}
