package scanner

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"

	datafeed "github.com/fazecat/mogulmaker/Internal/database"
	. "github.com/fazecat/mogulmaker/Internal/news_scraping"
	"github.com/fazecat/mogulmaker/Internal/strategy/detection"
	"github.com/fazecat/mogulmaker/Internal/strategy/indicators"
	signalsPkg "github.com/fazecat/mogulmaker/Internal/strategy/signals"
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils"
)

type TradeSignal struct {
	Direction  string
	Confidence float64
	Reasoning  string
}

type ScreenerCriteria struct {
	MinOversoldRSI float64
	MaxRSI         float64
	MinATR         float64
	MinVolumeRatio float64
}

type StockScore struct {
	Symbol         string
	Score          float64
	Signals        []string
	RSI            *float64
	ATR            *float64
	NewsSentiment  SentimentScore
	NewsImpact     float64
	FinalSignal    signalsPkg.CombinedSignal
	Recommendation string
	LongSignal     *TradeSignal
	ShortSignal    *TradeSignal
	SRValidation   *signalsPkg.SignalValidationWithSR // S/R analysis
}

func DefaultScreenerCriteria() ScreenerCriteria {
	return ScreenerCriteria{
		MinOversoldRSI: 35,
		MaxRSI:         75,
		MinATR:         0.1,
		MinVolumeRatio: 1.0,
	}
}

func ScreenStocksWithType(symbols []string, timeframe string, numBars int, criteria ScreenerCriteria, newsStorage *NewsStorage, assetType string) ([]StockScore, error) {
	var results []StockScore

	for _, symbol := range symbols {
		score, signals, rsi, atr, longSignal, shortSignal, srValidation, err := scoreStockWithType(symbol, timeframe, numBars, criteria, newsStorage, assetType)
		if err != nil {
			log.Printf("Error screening %s: %v", symbol, err)
			continue
		}
		if score == 0 && len(signals) == 0 && rsi == nil && atr == nil {
			log.Printf("Skipping %s: no data available", symbol)
			continue
		}
		results = append(results, StockScore{
			Symbol:       symbol,
			Score:        score,
			Signals:      signals,
			RSI:          rsi,
			ATR:          atr,
			LongSignal:   longSignal,
			ShortSignal:  shortSignal,
			SRValidation: srValidation,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func scoreStockWithType(symbol, timeframe string, numBars int, criteria ScreenerCriteria, newsStorage *NewsStorage, assetType string) (score float64, signals []string, rsi, atr *float64, longSignal, shortSignal *TradeSignal, srValidation *signalsPkg.SignalValidationWithSR, err error) {

	bars, err := datafeed.GetAlpacaBarsWithType(symbol, timeframe, numBars, "", assetType)
	if err != nil {
		return 0, nil, nil, nil, nil, nil, nil, err
	}

	if len(bars) < 2 {
		return 0, nil, nil, nil, nil, nil, nil, fmt.Errorf("insufficient data for %s (need 2 bars, got %d)", symbol, len(bars))
	}

	startTime := time.Now().AddDate(0, 0, -180)
	endTime := time.Now()

	if len(bars) > 0 {
		oldestTime, err := time.Parse(time.RFC3339, bars[len(bars)-1].Timestamp)
		if err == nil {
			startTime = oldestTime
		}
	}

	rsiMap, rsiErr := datafeed.FetchRSIByTimestampRange(symbol, startTime, endTime)
	if rsiErr != nil {
		log.Printf("RSI fetch failed for %s: %v (continuing with other signals)", symbol, rsiErr)
	} else if len(rsiMap) > 0 {
		rsi = findLatestValue(rsiMap)
	}

	atrMap, atrErr := datafeed.FetchATRByTimestampRange(symbol, startTime, endTime)
	if atrErr != nil {
		log.Printf("ATR fetch failed for %s: %v (continuing with other signals)", symbol, atrErr)
	} else if len(atrMap) > 0 {
		atr = findLatestValue(atrMap)
	}

	latestBar := bars[0]
	volumes := make([]int64, len(bars))
	for i, bar := range bars {
		volumes[i] = bar.Volume
	}
	avgVol20 := utils.CalculateAvgVolume(volumes, 20)

	score = 0
	signals = []string{}

	if rsi != nil {
		if *rsi < criteria.MinOversoldRSI {
			score += 20
			signals = append(signals, fmt.Sprintf("RSI Oversold: %.2f", *rsi))
		} else if *rsi > criteria.MaxRSI {
			score -= 10
			signals = append(signals, fmt.Sprintf("RSI Overbought: %.2f", *rsi))
		} else {
			score += 5
		}
	}

	if atr != nil && *atr > criteria.MinATR {
		score += 10
		signals = append(signals, fmt.Sprintf("High Volatility ATR: %.2f", *atr))
	}

	if avgVol20 > 0 {
		volRatio := float64(latestBar.Volume) / avgVol20
		if volRatio > criteria.MinVolumeRatio {
			score += 15
			signals = append(signals, fmt.Sprintf("High Volume: %.1fx avg", volRatio))
		}
	}

	if newsStorage != nil {
		news, err := newsStorage.GetLatestNews(context.Background(), symbol, 1)
		if err == nil && len(news) > 0 && news[0].Sentiment == Positive {
			score += 10
		}
	}

	whales := detection.DetectWhales(symbol, bars)
	if len(whales) > 0 {
		for _, whale := range whales {
			if whale.Conviction == "HIGH" {
				score += 5
				signals = append(signals, fmt.Sprintf("🐋 Whale %s: Z=%.2f", whale.Direction, whale.ZScore))
			}
		}
	}

	patternDetector := detection.NewPatternDetector()
	patterns := patternDetector.DetectAllPatterns(bars)
	for _, pattern := range patterns {
		if pattern.Detected {
			switch pattern.Direction {
			case "LONG":
				score += pattern.Confidence / 10
				signals = append(signals, fmt.Sprintf(" %s [%.0f%% confidence]", pattern.Pattern, pattern.Confidence))
			case "SHORT":
				score += pattern.Confidence / 15
				signals = append(signals, fmt.Sprintf(" %s [%.0f%% confidence]", pattern.Pattern, pattern.Confidence))
			case "NONE":
				signals = append(signals, fmt.Sprintf("  %s [%.0f%% confidence]", pattern.Pattern, pattern.Confidence))
			}
		}
	}

	support := indicators.FindSupport(bars)
	resistance := indicators.FindResistance(bars)

	currentPrice := latestBar.Close
	if currentPrice < support*1.01 {
		score += 15 //  buy signal
		signals = append(signals, fmt.Sprintf("Near Support: $%.2f", support))
	}
	if currentPrice > resistance*0.99 {
		score -= 10 // Sell signal
		signals = append(signals, fmt.Sprintf("Near Resistance: $%.2f", resistance))
	}

	combinedSignal := signalsPkg.CalculateSignal(rsi, atr, bars, symbol, "")

	// Apply signal quality filtering
	filter := signalsPkg.NewSignalQualityFilter()
	filter.MinConfidenceThreshold = 65.0 // Require 65% confidence for screener
	filter.VerboseLogging = false

	tradeSignal := signalsPkg.ConvertToTradeSignal(combinedSignal)
	filteredResult := filter.FilterSignal(tradeSignal)

	if filteredResult.Passed {
		signals = append(signals, fmt.Sprintf("\n[FINAL] %s [Quality: %.1f%% ✓]",
			signalsPkg.FormatSignal(combinedSignal), filteredResult.QualityScore))
		score += 10 // Bonus for high-quality signal
	} else {
		signals = append(signals, fmt.Sprintf("\n[WARNING] SIGNAL FILTERED: %s (Reason: %s)",
			signalsPkg.FormatSignal(combinedSignal), filteredResult.FailureReason))
		score -= 5 // Penalty for low-quality signal
	}

	longSignal = AnalyzeForLongs(latestBar, rsi, atr, criteria)
	shortSignal = AnalyzeForShorts(latestBar, rsi, atr, criteria)

	// Perform S/R validation on the best signal
	var signalToValidate *TradeSignal
	if longSignal != nil && (shortSignal == nil || longSignal.Confidence >= shortSignal.Confidence) {
		signalToValidate = longSignal
	} else if shortSignal != nil {
		signalToValidate = shortSignal
	}

	if signalToValidate != nil {
		// Convert TradeSignal to types.TradeSignal for validation
		typesSignal := &types.TradeSignal{
			Direction:  signalToValidate.Direction,
			Confidence: signalToValidate.Confidence,
			Reasoning:  signalToValidate.Reasoning,
		}

		// Create S/R validator
		srValidator := signalsPkg.NewSupportResistanceValidator()
		srValidator.MinValidationScore = 50.0

		// Validate signal with S/R levels
		srValidation = srValidator.ValidateSignalWithSR(typesSignal, bars, currentPrice)

		// Adjust score based on S/R validation
		if srValidation.IsValidLocation {
			score += (srValidation.ValidationScore / 10.0) // Add up to 10 points for perfect S/R
			signals = append(signals, fmt.Sprintf("[VALID] S/R: %.0f%% - %s", srValidation.ValidationScore, srValidation.DetailedAnalysis))
		} else {
			score -= 5 // Penalty for poor S/R positioning
			signals = append(signals, fmt.Sprintf("[WARNING] S/R: %.0f%% - %s", srValidation.ValidationScore, srValidation.DetailedAnalysis))
		}
	}

	return score, signals, rsi, atr, longSignal, shortSignal, srValidation, nil
}

func GetTradableAssets() ([]string, error) {
	client := datafeed.GetAlpacaClient()
	if client == nil {
		return nil, fmt.Errorf("alpaca client not initialized - call InitAlpacaClient() first")
	}

	assets, err := client.GetAssets(alpaca.GetAssetsRequest{
		Status: "active",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assets from Alpaca: %v", err)
	}

	symbols := make([]string, 0, len(assets))
	for _, asset := range assets {
		if asset.Class == "us_equity" && asset.Tradable {
			symbols = append(symbols, asset.Symbol)
		}
	}

	log.Printf("Fetched %d tradeable assets from Alpaca", len(symbols))
	return symbols, nil
}

func findLatestValue(m map[string]float64) *float64 {
	if len(m) == 0 {
		return nil
	}
	var latestKey string
	var latestVal float64
	for k, v := range m {
		if latestKey == "" || k > latestKey {
			latestKey = k
			latestVal = v
		}
	}
	return &latestVal
}

func AnalyzeForShorts(bar datafeed.Bar, rsi *float64, atr *float64, criteria ScreenerCriteria) *TradeSignal {
	if rsi == nil || atr == nil {
		return nil
	}
	if *rsi > criteria.MaxRSI && *atr >= criteria.MinATR {
		confidence := ((*rsi - criteria.MaxRSI) / (100 - criteria.MaxRSI)) * 100
		if confidence > 100 {
			confidence = 100
		}
		reasoning := "RSI indicates overbought conditions with sufficient volatility."
		return &TradeSignal{
			Direction:  "SHORT",
			Confidence: confidence,
			Reasoning:  reasoning,
		}
	}
	return nil
}

func AnalyzeForLongs(bar datafeed.Bar, rsi *float64, atr *float64, criteria ScreenerCriteria) *TradeSignal {
	if rsi == nil || atr == nil {
		return nil
	}
	if *rsi < criteria.MinOversoldRSI && *atr >= criteria.MinATR {
		confidence := (1 - (*rsi / criteria.MinOversoldRSI)) * 100
		if confidence > 100 {
			confidence = 100
		}

		reasoning := fmt.Sprintf("RSI oversold (%.1f) with ATR %.2f", *rsi, *atr)
		return &TradeSignal{
			Direction:  "LONG",
			Confidence: confidence,
			Reasoning:  reasoning,
		}
	}
	return nil
}
