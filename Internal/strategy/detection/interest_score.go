package detection

import (
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils"
	"github.com/fazecat/mogulmaker/Internal/utils/config"
)

func CalculateInterestScore(input types.ScoringInput, weights config.SignalWeights) float64 {
	var rsiScore float64
	if input.RSIValue < 30 {
		rsiScore = 9.0 + (1.0 - (input.RSIValue / 30)) // 9-10 (oversold = good for longs)
	} else if input.RSIValue > 70 {
		// HIGH RSI (overbought): bad for longs, good for shorts
		// Default (shorts disabled): low score
		rsiScore = 1.0 - ((input.RSIValue - 70) / 30) // 0-1
		if rsiScore < 0 {
			rsiScore = 0
		}
	} else {
		rsiScore = 5.0 + ((input.RSIValue - 50) / 10) // Neutral
	}

	var atrScore float64
	if input.ATRCategory == "LOW" {
		atrScore = 4.0
	} else {
		atrScore = 8.0
	}

	var whaleScore float64
	if input.WhaleCount > 0 {
		whaleScore = 9.0
	} else {
		whaleScore = 5.0
	}

	// VWAP distance score
	vWapDistanceScore := (input.VWAPPrice - input.CurrentPrice) / input.VWAPPrice * 100
	var vwapScore float64
	if vWapDistanceScore >= 0 {
		vwapScore = 5.0
		distanceAbs := utils.Abs(vWapDistanceScore)
		if distanceAbs <= 5 {
			vwapScore = 6.0 + (distanceAbs / 5)
		} else if distanceAbs <= 15 {
			vwapScore = 7.0 + ((distanceAbs - 5) / 10)
		} else if distanceAbs <= 30 {
			vwapScore = 8.0 + ((distanceAbs - 15) / 15)
			vwapScore = 9.0
		}
	}

	// Volume score: Current volume relative to average
	// Ratio < 0.5: low volume = 3.0
	// Ratio 0.5-1.0: average volume = 5.0
	// Ratio 1.0-1.5: above average = 7.0
	// Ratio > 1.5: high volume = 9.0
	volumeScore := calculateVolumeScore(input.VolumeRatio)

	// News sentiment score (0-10, default 5.0 = neutral)
	newsScore := input.NewsSentimentScore

	// Calculate total weight to normalize
	totalWeight := weights.RSIWeight + weights.ATRWeight + weights.WhaleActivityWeight +
		0.15 + weights.VolumeWeight + weights.NewsSentimentWeight

	// Calculate weighted sum and normalize to 0-10 scale
	weightedSum := (rsiScore * weights.RSIWeight) +
		(atrScore * weights.ATRWeight) +
		(whaleScore * weights.WhaleActivityWeight) +
		(vwapScore * 0.15) +
		(volumeScore * weights.VolumeWeight) +
		(newsScore * weights.NewsSentimentWeight)

	// Normalize: divide by total weight to get average, then scale to 0-10
	finalScore := (weightedSum / totalWeight)

	if finalScore > 10.0 {
		finalScore = 10.0
	}
	if finalScore < 0.0 {
		finalScore = 0.0
	}

	return finalScore
}

// calculateVolumeScore converts volume ratio to 0-10 score
func calculateVolumeScore(volumeRatio float64) float64 {
	if volumeRatio < 0.5 {
		return 3.0
	} else if volumeRatio < 1.0 {
		return 5.0
	} else if volumeRatio < 1.5 {
		return 7.0
	}
	return 9.0
}
