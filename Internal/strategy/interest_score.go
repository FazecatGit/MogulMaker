package strategy

import (
	"github.com/fazecat/mongelmaker/Internal/types"
	"github.com/fazecat/mongelmaker/Internal/utils"
	"github.com/fazecat/mongelmaker/Internal/utils/config"
)

func CalculateInterestScore(input types.ScoringInput, weights config.SignalWeights) float64 {
	var rsiScore float64
	if input.RSIValue < 30 {
		rsiScore = 9.0 + (1.0 - (input.RSIValue / 30)) // 9-10
	} else if input.RSIValue > 70 {
		rsiScore = 1.0 - ((input.RSIValue - 70) / 30) // 0-1
		if rsiScore < 0 {
			rsiScore = 0
		}
	} else {
		rsiScore = 5.0 + ((input.RSIValue - 50) / 10)
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

	// Volume score placeholder (0-10)
	volumeScore := 5.0
	newsScore := 5.0

	finalScore := (rsiScore * weights.RSIWeight) +
		(atrScore * weights.ATRWeight) +
		(whaleScore * weights.WhaleActivityWeight) +
		(vwapScore * 0.15) +
		(volumeScore * weights.VolumeWeight) +
		(newsScore * weights.NewsSentimentWeight)

	if finalScore > 10.0 {
		finalScore = 10.0
	}
	if finalScore < 0.0 {
		finalScore = 0.0
	}

	return finalScore
}
