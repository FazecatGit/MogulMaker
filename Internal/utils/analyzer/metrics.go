package analyzer

import (
	"context"
	"fmt"

	datafeed "github.com/fazecat/mogulmaker/Internal/database"
	"github.com/fazecat/mogulmaker/Internal/strategy/detection"
	"github.com/fazecat/mogulmaker/Internal/strategy/indicators"
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
