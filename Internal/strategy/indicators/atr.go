package indicators

import (
	"fmt"

	"github.com/fazecat/mongelmaker/Internal/utils"
)

type ATRBar struct {
	High  float64
	Low   float64
	Close float64
}

func CalculateATR(atrBars []ATRBar, period int) ([]float64, error) {
	if len(atrBars) < period+1 {
		return nil, fmt.Errorf("not enough data")
	}
	atrValues := make([]float64, len(atrBars))
	trueRanges := make([]float64, len(atrBars))

	// Calculate True Ranges
	for i := 1; i < len(atrBars); i++ {
		high := atrBars[i].High
		low := atrBars[i].Low
		prevClose := atrBars[i-1].Close

		trueRange := utils.Max(high-low, utils.Abs(high-prevClose), utils.Abs(low-prevClose))
		trueRanges[i] = trueRange
	}

	for i := period; i < len(atrBars); i++ {
		atrValues[i] = utils.Average(trueRanges[i-period+1 : i+1])
	}

	return atrValues, nil
}

func CalculateTrueRange(high, low, prevClose float64) float64 {
	return utils.Max(high-low, utils.Abs(high-prevClose), utils.Abs(low-prevClose))
}

// Will do later
func DetermineATRSignal(atrValue float64, threshold float64) string {
	if atrValue > threshold {
		return "high volatility"
	}
	return "low volatility"
}
