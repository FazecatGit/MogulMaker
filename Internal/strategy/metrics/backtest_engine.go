package metrics

import (
	"time"

	"github.com/fazecat/mogulmaker/Internal/strategy/indicators"
	"github.com/fazecat/mogulmaker/Internal/types"
)

type Position struct {
	Symbol     string
	InTrade    bool
	EntryPrice float64
	Quantity   float64
	EntryTime  time.Time
}

func RunBacktest(symbol string, bars []types.Bar, startingCapital float64) ([]TradeResult, error) {
	if len(bars) == 0 {
		return nil, nil
	}

	var trades []TradeResult
	currentPosition := Position{InTrade: false}
	capital := startingCapital

	for i := 14; i < len(bars); i++ {
		currentBar := bars[i]

		closingPrices := make([]float64, i+1)
		for j := 0; j <= i; j++ {
			closingPrices[j] = bars[j].Close
		}
		rsiValues, err := indicators.CalculateRSI(closingPrices, 14)
		if err != nil {
			continue
		}
		rsi := rsiValues[len(rsiValues)-1]

		if !currentPosition.InTrade && rsi < 30 {
			// Enter long position
			quantity := capital / currentBar.Close
			currentPosition = Position{
				InTrade:    true,
				EntryPrice: currentBar.Close,
				Quantity:   quantity,
				EntryTime:  time.Now(),
			}
		} else if currentPosition.InTrade && rsi > 70 {
			trade := createTradeResult(symbol, currentPosition, currentBar.Close, time.Now())
			trades = append(trades, trade)
			currentPosition = Position{InTrade: false}
		}
	}
	if currentPosition.InTrade {
		trade := createTradeResult(symbol, currentPosition, bars[len(bars)-1].Close, time.Now())
		trades = append(trades, trade)
	}
	return trades, nil
}

func createTradeResult(symbol string, pos Position, exitPrice float64, exitTime time.Time) TradeResult {
	pnl := (exitPrice - pos.EntryPrice) * pos.Quantity
	returnPercent := ((exitPrice - pos.EntryPrice) / pos.EntryPrice) * 100

	return TradeResult{
		Symbol:        symbol,
		EntryPrice:    pos.EntryPrice,
		ExitPrice:     exitPrice,
		Quantity:      pos.Quantity,
		PnL:           pnl,
		ReturnPercent: returnPercent,
		Duration:      exitTime.Sub(pos.EntryTime),
		EntryTime:     pos.EntryTime,
		ExitTime:      exitTime,
	}
}
