package analyzer

import (
	"context"
	"fmt"
	"time"

	datafeed "github.com/fazecat/mogulmaker/Internal/database"
	"github.com/fazecat/mogulmaker/Internal/strategy/detection"
	"github.com/fazecat/mogulmaker/Internal/strategy/indicators"
	signalsPkg "github.com/fazecat/mogulmaker/Internal/strategy/signals"
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils/config"
)

func extractClosingPrices(bars []types.Bar) []float64 {
	closes := make([]float64, len(bars))
	for i, bar := range bars {
		closes[i] = bar.Close
	}
	return closes
}

func CalculateCandidateMetrics(ctx context.Context, symbol string, bars []types.Bar, cfg *config.Config, weights config.SignalWeights) (*types.Candidate, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("no bars provided for %s", symbol)
	}
	closingPrices := extractClosingPrices(bars)
	rsiValues, err := indicators.CalculateRSI(closingPrices, 14)
	if err != nil {
		return nil, err
	}

	atrMap, err := datafeed.FetchATRForDisplay(symbol, 1)
	if err != nil {
		return nil, err
	}

	var atrValue float64
	if len(atrMap) > 0 {
		for _, v := range atrMap {
			atrValue = v
			break
		}
	}

	interestScoreInput := types.ScoringInput{
		CurrentPrice: bars[len(bars)-1].Close,
		RSIValue:     rsiValues[len(rsiValues)-1],
		ATRValue:     atrValue,
		PriceDrop:    (bars[len(bars)-2].Close - bars[len(bars)-1].Close) / bars[len(bars)-2].Close * 100,
	}
	interestScore := detection.CalculateInterestScore(interestScoreInput, weights)

	latestPattern := GetLatestCandlePattern(bars, 5)

	candidate := &types.Candidate{
		Symbol:   symbol,
		Score:    interestScore,
		RSI:      rsiValues[len(rsiValues)-1],
		ATR:      atrValue,
		Analysis: latestPattern,
		Bars:     bars,
	}

	return candidate, nil
}

func analyzeRecentCandles(bars []types.Bar, numCandles int) (int, int, string) {
	if len(bars) == 0 {
		return 0, 0, "N/A"
	}

	if numCandles > len(bars) {
		numCandles = len(bars)
	}

	startIdx := len(bars) - numCandles
	recentBars := bars[startIdx:]

	latestBar := recentBars[len(recentBars)-1]
	candle := Candlestick{
		Open:  latestBar.Open,
		Close: latestBar.Close,
		High:  latestBar.High,
		Low:   latestBar.Low,
	}

	_, analysisMap := AnalyzeCandlestick(candle)
	latestPattern := analysisMap["Analysis"]

	bullishCount := 0
	bearishCount := 0
	for _, bar := range recentBars {
		if bar.Close > bar.Open {
			bullishCount++
		} else if bar.Close < bar.Open {
			bearishCount++
		}
	}

	return bullishCount, bearishCount, latestPattern
}

func GetLatestCandlePattern(bars []types.Bar, numCandles int) string {
	if len(bars) == 0 {
		return "N/A"
	}

	bullishCount, bearishCount, latestPattern := analyzeRecentCandles(bars, numCandles)

	if numCandles == 1 {
		return latestPattern
	}

	if bullishCount > bearishCount {
		return fmt.Sprintf("Bullish Trend (%d/%d, Latest: %s)", bullishCount, numCandles, latestPattern)
	} else if bearishCount > bullishCount {
		return fmt.Sprintf("Bearish Trend (%d/%d, Latest: %s)", bearishCount, numCandles, latestPattern)
	}

	return fmt.Sprintf("Mixed Trend (Latest: %s)", latestPattern)
}

func GetPatternConfidence(ctx context.Context, symbol string, bars []types.Bar) (float64, error) {
	if len(bars) == 0 {
		return 0, fmt.Errorf("no bars provided for %s", symbol)
	}

	latestBar := bars[len(bars)-1]

	atrMap, err := datafeed.FetchATRForDisplay(symbol, 1)
	if err != nil {
		return 0, err
	}

	var atrValue *float64
	if len(atrMap) > 0 {
		for _, v := range atrMap {
			atrValue = &v
			break
		}
	}

	volumes := make([]int64, len(bars))
	for i, bar := range bars {
		volumes[i] = bar.Volume
	}

	_, confidence := PatternAnalyzeCandle(latestBar, atrValue, 0, latestBar.Volume)

	return confidence, nil
}

// AnalyzeSymbolDetailed performs comprehensive analysis on a symbol and returns formatted analysis data
func AnalyzeSymbolDetailed(symbol string, bars []types.Bar) (map[string]interface{}, error) {
	if len(bars) < 14 {
		return nil, fmt.Errorf("not enough data to analyze - need at least 14 bars, got %d", len(bars))
	}

	// Calculate RSI
	closes := extractClosingPrices(bars)
	rsiValues, err := indicators.CalculateRSI(closes, 14)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate RSI: %w", err)
	}

	// Calculate ATR
	atrBars := make([]indicators.ATRBar, len(bars))
	for i, bar := range bars {
		atrBars[i] = indicators.ATRBar{
			High:  bar.High,
			Low:   bar.Low,
			Close: bar.Close,
		}
	}
	atrValues, err := indicators.CalculateATR(atrBars, 14)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ATR: %w", err)
	}

	// Get current values
	currentPrice := bars[0].Close
	currentRSI := rsiValues[len(rsiValues)-1]
	currentATR := atrValues[len(atrValues)-1]

	// Calculate SMA 20
	sma20 := 0.0
	barsForSMA := 20
	if len(bars) < barsForSMA {
		barsForSMA = len(bars)
	}
	for i := 0; i < barsForSMA; i++ {
		sma20 += bars[i].Close
	}
	sma20 /= float64(barsForSMA)

	// Determine trend
	trend := "neutral"
	if currentPrice > sma20*1.02 {
		trend = "bullish"
	} else if currentPrice < sma20*0.98 {
		trend = "bearish"
	}

	// Find support and resistance
	support := indicators.FindSupport(bars)
	resistance := indicators.FindResistance(bars)

	distanceToSupport := ((currentPrice - support) / support) * 100
	distanceToResistance := ((resistance - currentPrice) / currentPrice) * 100

	// Detect patterns
	patternDetector := detection.NewPatternDetector()
	patterns := patternDetector.DetectAllPatterns(bars)

	var bestPattern map[string]interface{}
	var bestP *detection.PatternSignal
	if len(patterns) > 0 {
		for i := range patterns {
			if patterns[i].Detected {
				if bestP == nil || patterns[i].Confidence > bestP.Confidence {
					bestP = &patterns[i]
				}
			}
		}

		if bestP != nil {
			bestPattern = map[string]interface{}{
				"pattern":       bestP.Pattern,
				"direction":     bestP.Direction,
				"confidence":    bestP.Confidence,
				"support_level": bestP.SupportLevel,
				"resistance":    bestP.ResistanceLevel,
				"target_up":     bestP.PriceTargetUp,
				"target_down":   bestP.PriceTargetDown,
				"stop_loss":     bestP.StopLossLevel,
				"risk_reward":   bestP.RiskRewardRatio,
				"reasoning":     bestP.Reasoning,
			}
		}
	}

	// Determine RSI signal
	rsiSignal := "neutral"
	if currentRSI > 70 {
		rsiSignal = "overbought"
	} else if currentRSI < 30 {
		rsiSignal = "oversold"
	}

	// Calculate trading recommendation
	tradingRec := signalsPkg.CalculateTradingRecommendation(currentPrice, currentRSI, support, resistance, trend, bestP)

	// Format historical bars
	historicalBars := make([]map[string]interface{}, len(bars))
	for i, bar := range bars {
		rsiVal := 0.0
		if i < len(rsiValues) {
			rsiVal = rsiValues[i]
		}
		atrVal := 0.0
		if i < len(atrValues) {
			atrVal = atrValues[i]
		}

		timestamp := int64(0)
		if t, err := time.Parse(time.RFC3339, bar.Timestamp); err == nil {
			timestamp = t.Unix()
		}

		historicalBars[i] = map[string]interface{}{
			"open":      bar.Open,
			"high":      bar.High,
			"low":       bar.Low,
			"close":     bar.Close,
			"volume":    bar.Volume,
			"timestamp": timestamp,
			"rsi":       rsiVal,
			"atr":       atrVal,
		}
	}

	// Reverse bars so oldest dates appear on left, newest on right
	for i, j := 0, len(historicalBars)-1; i < j; i, j = i+1, j-1 {
		historicalBars[i], historicalBars[j] = historicalBars[j], historicalBars[i]
	}

	// Build response
	response := map[string]interface{}{
		"symbol":                 symbol,
		"current_price":          currentPrice,
		"rsi":                    currentRSI,
		"rsi_signal":             rsiSignal,
		"atr":                    currentATR,
		"sma_20":                 sma20,
		"trend":                  trend,
		"bars_analyzed":          len(bars),
		"timestamp":              time.Now().Unix(),
		"support_level":          support,
		"resistance_level":       resistance,
		"distance_to_support":    distanceToSupport,
		"distance_to_resistance": distanceToResistance,
		"chart_pattern":          bestPattern,
		"multi_timeframe": map[string]interface{}{
			"note": "Multi-timeframe analysis requires additional data fetching",
		},
		"trading_recommendation": tradingRec,
		"historical_bars":        historicalBars,
	}

	return response, nil
}
