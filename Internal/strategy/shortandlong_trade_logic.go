package strategy

import (
	"fmt"

	datafeed "github.com/fazecat/mongelmaker/Internal/database"
)

type TradeSignal struct {
	Direction  string
	Confidence float64
	Reasoning  string
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
