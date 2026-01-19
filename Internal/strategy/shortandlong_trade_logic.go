package strategy

import (
	"context"
	"fmt"
	"log"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"

	datafeed "github.com/fazecat/mogulmaker/Internal/database"
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

func ExecuteTrade(ctx context.Context, client *alpaca.Client, symbol string, quantity int64, signal *TradeSignal) error {
	if signal == nil {
		return fmt.Errorf("trade signal is nil")
	}

	if client == nil {
		return fmt.Errorf("alpaca client is nil")
	}

	var side alpaca.Side
	if signal.Direction == "LONG" {
		side = alpaca.Buy
		log.Printf("📈 Placing LONG order: %s x %d @ confidence %.2f%%\n", symbol, quantity, signal.Confidence)
	} else if signal.Direction == "SHORT" {
		side = alpaca.Sell
		log.Printf("📉 Placing SHORT order: %s x %d @ confidence %.2f%%\n", symbol, quantity, signal.Confidence)
	} else {
		return fmt.Errorf("unknown trade direction: %s", signal.Direction)
	}

	req := alpaca.PlaceOrderRequest{
		Symbol:      symbol,
		Qty:         &decimal.Decimal{},
		Side:        side,
		Type:        alpaca.Market,
		TimeInForce: alpaca.Day,
	}
	*req.Qty = decimal.NewFromInt(quantity)

	order, err := client.PlaceOrder(req)
	if err != nil {
		return fmt.Errorf("failed to create %s order for %s: %v", signal.Direction, symbol, err)
	}

	log.Printf("✅ Order created: %s | ID: %s | Status: %s\n", symbol, order.ID, order.Status)

	// Log trade to database
	var price decimal.Decimal
	if order.FilledAvgPrice != nil {
		price = *order.FilledAvgPrice
	} else {
		price = decimal.NewFromInt(0)
	}
	err = datafeed.LogTradeExecution(ctx, symbol, string(side), quantity, price, order.ID, order.Status)
	if err != nil {
		log.Printf("⚠️  Failed to log trade to database: %v", err)
		// Don't fail the whole operation if logging fails
	}

	return nil
}

func PlaceLongOrder(ctx context.Context, client *alpaca.Client, symbol string, quantity int64, confidence float64) error {
	signal := &TradeSignal{
		Direction:  "LONG",
		Confidence: confidence,
		Reasoning:  fmt.Sprintf("Manual long trade with %.2f%% confidence", confidence),
	}
	return ExecuteTrade(ctx, client, symbol, quantity, signal)
}

func PlaceShortOrder(ctx context.Context, client *alpaca.Client, symbol string, quantity int64, confidence float64) error {
	signal := &TradeSignal{
		Direction:  "SHORT",
		Confidence: confidence,
		Reasoning:  fmt.Sprintf("Manual short trade with %.2f%% confidence", confidence),
	}
	return ExecuteTrade(ctx, client, symbol, quantity, signal)
}
