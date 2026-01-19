package scoring

import (
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils"
)

func BuildScoringInput(bars []types.Bar, vwapPrice float64, rsiValue float64, whaleCount int, atrValue float64, atrCategory string) (types.ScoringInput, error) {
	if len(bars) < 2 {
		return types.ScoringInput{}, nil
	}

	currentBar := bars[len(bars)-1]
	currentPrice := currentBar.Close
	openPrice := currentBar.Open

	priceDrop := 0.0
	if openPrice != 0 {
		priceDrop = ((openPrice - currentPrice) / openPrice) * 100
	}

	// Calculate volume ratio (current volume / average volume)
	volumeRatio := calculateVolumeRatio(bars)

	return types.ScoringInput{
		CurrentPrice:       currentPrice,
		VWAPPrice:          vwapPrice,
		ATRValue:           atrValue,
		RSIValue:           rsiValue,
		WhaleCount:         float64(whaleCount),
		PriceDrop:          priceDrop,
		ATRCategory:        atrCategory,
		VolumeRatio:        volumeRatio,
		NewsSentimentScore: 5.0,   // Default neutral, can be updated from news data
		ShortSignalActive:  false, // Will be set by caller
	}, nil
}

func calculateVolumeRatio(bars []types.Bar) float64 {
	if len(bars) < 2 {
		return 1.0
	}

	currentVolume := float64(bars[len(bars)-1].Volume)

	// Calculate average volume from last 20 bars
	period := 20
	if len(bars) < period {
		period = len(bars) - 1
	}

	totalVolume := 0.0
	for i := len(bars) - period; i < len(bars)-1; i++ {
		totalVolume += float64(bars[i].Volume)
	}

	avgVolume := totalVolume / float64(period)
	if avgVolume == 0 {
		return 1.0
	}

	return currentVolume / avgVolume
}

func CalculateATRFromBars(bars []types.Bar) float64 {
	if len(bars) < 2 {
		return 0
	}

	trueRanges := make([]float64, len(bars))
	for i := 1; i < len(bars); i++ {
		high := bars[i].High
		low := bars[i].Low
		prevClose := bars[i-1].Close
		tr := utils.Max(high-low, utils.Abs(high-prevClose), utils.Abs(low-prevClose))
		trueRanges[i] = tr
	}

	period := 14
	if len(trueRanges) < period {
		period = len(trueRanges) - 1
	}

	return utils.Average(trueRanges[len(trueRanges)-period:])
}

func CategorizeATRValue(currentATR float64, bars []types.Bar) string {
	if len(bars) < 15 {
		return "NORMAL"
	}

	atrValues := make([]float64, 0)
	for i := len(bars) - 14; i < len(bars); i++ {
		subset := bars[:i+1]
		atr := CalculateATRFromBars(subset)
		if atr > 0 {
			atrValues = append(atrValues, atr)
		}
	}

	if len(atrValues) == 0 {
		return "NORMAL"
	}

	avgATR := utils.Average(atrValues)

	if currentATR < avgATR*0.5 {
		return "LOW"
	} else if currentATR > avgATR*1.5 {
		return "HIGH"
	}
	return "NORMAL"
}

func ScoreCategory(score float64) string {
	if score >= 8.0 {
		return "🟢 Excellent"
	}
	if score >= 6.0 {
		return "🟢 Good"
	}
	if score >= 4.0 {
		return "🟡 Fair"
	}
	if score >= 2.0 {
		return "🟠 Moderate"
	}
	return "🔴 Poor"
}
