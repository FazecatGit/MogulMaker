package datafeed

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"

	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
)

func LogTradeExecution(ctx context.Context, symbol string, side string, quantity int64, price decimal.Decimal, alpacaOrderID string, status string) error {
	if Queries == nil {
		return fmt.Errorf("database queries not initialized")
	}

	totalValue := decimal.NewFromInt(quantity).Mul(price)

	err := Queries.LogTrade(ctx, database.LogTradeParams{
		Symbol:        symbol,
		Side:          side,
		Quantity:      decimal.NewFromInt(quantity).String(),
		Price:         price.String(),
		TotalValue:    totalValue.String(),
		AlpacaOrderID: sql.NullString{String: alpacaOrderID, Valid: true},
		Status:        sql.NullString{String: status, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to log trade: %w", err)
	}

	log.Printf("✅ Trade logged to database: %s %s x%d @ %s (Order ID: %s)\n",
		side, symbol, quantity, price.String(), alpacaOrderID)
	return nil
}

func GetTradeHistory(ctx context.Context, symbol string, limit int32) ([]database.GetTradeHistoryRow, error) {
	if Queries == nil {
		return nil, fmt.Errorf("database queries not initialized")
	}

	trades, err := Queries.GetTradeHistory(ctx, database.GetTradeHistoryParams{
		Symbol: symbol,
		Limit:  limit,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch trade history: %w", err)
	}

	return trades, nil
}

func GetOpenTrades(ctx context.Context) ([]database.GetAllOpenTradesRow, error) {
	if Queries == nil {
		return nil, fmt.Errorf("database queries not initialized")
	}

	trades, err := Queries.GetAllOpenTrades(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open trades: %w", err)
	}

	return trades, nil
}

// GetAllTradesDebug returns ALL trades from database (for debugging)
func GetAllTradesDebug(ctx context.Context) ([]database.GetAllTradesRow, error) {
	if Queries == nil {
		return nil, fmt.Errorf("database queries not initialized")
	}

	trades, err := Queries.GetAllTrades(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all trades: %w", err)
	}

	return trades, nil
}

func UpdateTradeStatus(ctx context.Context, alpacaOrderID string, status string) error {
	if Queries == nil {
		return fmt.Errorf("database queries not initialized")
	}

	err := Queries.UpdateTradeStatus(ctx, database.UpdateTradeStatusParams{
		Status:        sql.NullString{String: status, Valid: true},
		AlpacaOrderID: sql.NullString{String: alpacaOrderID, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to update trade status: %w", err)
	}

	log.Printf("✅ Trade status updated: Order ID %s -> %s\n", alpacaOrderID, status)
	return nil
}

type TradeStats struct {
	TotalTrades      int
	WinningTrades    int
	LosingTrades     int
	TotalVolume      decimal.Decimal
	TotalProfit      decimal.Decimal
	WinRate          float64
	AverageTradeSize decimal.Decimal
}

func GetTradeStats(ctx context.Context, symbol string, lookbackDays int) (*TradeStats, error) {
	trades, err := GetTradeHistory(ctx, symbol, 1000)
	if err != nil {
		return nil, err
	}

	stats := &TradeStats{
		TotalVolume:      decimal.NewFromInt(0),
		TotalProfit:      decimal.NewFromInt(0),
		AverageTradeSize: decimal.NewFromInt(0),
	}

	// Filter trades within lookback period
	cutoff := time.Now().AddDate(0, 0, -lookbackDays)
	var recentTrades []database.GetTradeHistoryRow

	for _, trade := range trades {
		if trade.CreatedAt.Valid && trade.CreatedAt.Time.After(cutoff) {
			recentTrades = append(recentTrades, trade)
		}
	}

	if len(recentTrades) == 0 {
		return stats, nil
	}

	stats.TotalTrades = len(recentTrades)

	// Calculate stats - pair longs and shorts
	for i := 0; i < len(recentTrades)-1; i += 2 {
		entry := recentTrades[i]
		var exit database.GetTradeHistoryRow

		// Find matching exit trade
		found := false
		for j := i + 1; j < len(recentTrades); j++ {
			if recentTrades[j].Symbol == entry.Symbol &&
				((entry.Side == "BUY" && recentTrades[j].Side == "SELL") ||
					(entry.Side == "SELL" && recentTrades[j].Side == "BUY")) {
				exit = recentTrades[j]
				found = true
				break
			}
		}

		if !found {
			continue
		}

		// Calculate P&L
		entryPrice, _ := decimal.NewFromString(entry.Price)
		exitPrice, _ := decimal.NewFromString(exit.Price)
		entryQty, _ := decimal.NewFromString(entry.Quantity)
		exitQty, _ := decimal.NewFromString(exit.Quantity)

		entryValue := entryPrice.Mul(entryQty)
		exitValue := exitPrice.Mul(exitQty)
		var profit decimal.Decimal

		if entry.Side == "BUY" {
			profit = exitValue.Sub(entryValue)
		} else {
			profit = entryValue.Sub(exitValue)
		}

		stats.TotalProfit = stats.TotalProfit.Add(profit)
		if profit.IsPositive() {
			stats.WinningTrades++
		} else if profit.IsNegative() {
			stats.LosingTrades++
		}
	}

	// Calculate win rate
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinningTrades) / float64(stats.TotalTrades) * 100
	}

	// Calculate average trade size
	if stats.TotalTrades > 0 {
		totalQty := decimal.NewFromInt(0)
		for _, trade := range recentTrades {
			qty, _ := decimal.NewFromString(trade.Quantity)
			totalQty = totalQty.Add(qty)
		}
		stats.AverageTradeSize = totalQty.Div(decimal.NewFromInt(int64(stats.TotalTrades)))
	}

	return stats, nil
}
