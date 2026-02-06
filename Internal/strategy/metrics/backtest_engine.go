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
	EntryDate  string // Store the bar date as string (YYYY-MM-DD)
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

		// Parse the bar date for trade record
		barDate := "1970-01-01"
		if t, err := time.Parse(time.RFC3339, currentBar.Timestamp); err == nil {
			barDate = t.Format("2006-01-02")
		}

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
			entryTime, _ := time.Parse("2006-01-02", barDate)
			if entryTime.IsZero() {
				entryTime = time.Now()
			}
			currentPosition = Position{
				InTrade:    true,
				EntryPrice: currentBar.Close,
				Quantity:   quantity,
				EntryTime:  entryTime,
				EntryDate:  barDate,
			}
		} else if currentPosition.InTrade && rsi > 70 {
			trade := createTradeResult(symbol, currentPosition, currentBar.Close, barDate)
			trades = append(trades, trade)
			currentPosition = Position{InTrade: false}
		}
	}
	if currentPosition.InTrade {
		// Use last bar's date for exit
		barDate := "1970-01-01"
		if t, err := time.Parse(time.RFC3339, bars[len(bars)-1].Timestamp); err == nil {
			barDate = t.Format("2006-01-02")
		}
		trade := createTradeResult(symbol, currentPosition, bars[len(bars)-1].Close, barDate)
		trades = append(trades, trade)
	}

	return trades, nil
}

func createTradeResult(symbol string, pos Position, exitPrice float64, exitDate string) TradeResult {
	pnl := (exitPrice - pos.EntryPrice) * pos.Quantity
	returnPercent := ((exitPrice - pos.EntryPrice) / pos.EntryPrice) * 100

	// Parse exit date to create proper exit time for duration calculation
	exitTime, _ := time.Parse("2006-01-02", exitDate)
	if exitTime.IsZero() {
		exitTime = time.Now()
	}

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

func CalculateWinRate(trades []TradeResult) float64 {
	if len(trades) == 0 {
		return 0.0
	}
	wins := 0
	for _, trade := range trades {
		if trade.PnL > 0 {
			wins++
		}
	}
	return (float64(wins) / float64(len(trades))) * 100
}
