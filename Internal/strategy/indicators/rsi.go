package indicators

import (
	"fmt"

	"github.com/fazecat/mongelmaker/Internal/utils"
)

func CalculateRSI(closes []float64, period int) ([]float64, error) {

	if len(closes) < period+1 {
		return nil, fmt.Errorf("not enough data")
	}
	rsi := make([]float64, len(closes))

	gains := make([]float64, len(closes))
	losses := make([]float64, len(closes))

	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	for i := period; i < len(closes); i++ {
		windowGains := gains[i-period+1 : i+1]
		windowLosses := losses[i-period+1 : i+1]

		avgGain := utils.Average(windowGains)
		avgLoss := utils.Average(windowLosses)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi, nil
}

func DetermineRSISignal(rsiValue float64) string {
	if rsiValue < 30 {
		return "oversold"
	} else if rsiValue > 70 {
		return "overbought"
	}
	return "neutral"
}
