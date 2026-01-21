package strategy

import (
	"fmt"
	"log"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
)

type OrderConfig struct {
	MaxPortfolioPercent   float64 //(default 20%)
	MaxOpenPositions      int     //(default 5)
	StopLossPercent       float64 // (default 2%)
	TakeProfitPercent     float64 //(default 5%)
	SafeBailPercent       float64 //(default 3%)
	MaxDailyLossPercent   float64 //(default -2%)
	PartialExitPercentage float64 //(default 0.5 = 50%)
}

type OrderRequest struct {
	Symbol           string
	Quantity         int64
	Direction        string
	SignalConfidence float64
	TradeReason      string
	StopLossPrice    float64
	TakeProfitPrice  float64
	EntryPrice       float64
	UseStopOrder     bool
	UseLimitOrder    bool
	LimitPrice       float64
}

type OrderValidation struct {
	IsValid       bool
	Quantity      int64
	RiskAmount    float64
	PotentialGain float64
	PortfolioRisk float64
	Issues        []string
}

// ValidateOrder checks if order meets safety requirements
func ValidateOrder(req *OrderRequest, cfg *OrderConfig, accountValue float64, openPositions int, dailyLoss float64) *OrderValidation {
	validation := &OrderValidation{
		IsValid: true,
		Issues:  []string{},
	}

	// Check 1: Max open positions
	if openPositions >= cfg.MaxOpenPositions {
		validation.IsValid = false
		validation.Issues = append(validation.Issues,
			fmt.Sprintf("Max open positions reached (%d/%d)", openPositions, cfg.MaxOpenPositions))
	}

	// Check 2: Daily loss limit
	if dailyLoss < (cfg.MaxDailyLossPercent * accountValue / 100) {
		validation.IsValid = false
		validation.Issues = append(validation.Issues,
			fmt.Sprintf("Daily loss limit exceeded: %.2f%% (limit: %.2f%%)",
				(dailyLoss/accountValue)*100, cfg.MaxDailyLossPercent))
	}

	// Check 3: Validate prices
	if req.StopLossPrice <= 0 || req.TakeProfitPrice <= 0 || req.EntryPrice <= 0 {
		validation.IsValid = false
		validation.Issues = append(validation.Issues, "Invalid price levels (must be > 0)")
	}

	// Check 4: Stop loss below entry for long, above entry for short
	if req.Direction == "LONG" && req.StopLossPrice >= req.EntryPrice {
		validation.IsValid = false
		validation.Issues = append(validation.Issues, "Stop loss must be below entry price for LONG")
	}
	if req.Direction == "SHORT" && req.StopLossPrice <= req.EntryPrice {
		validation.IsValid = false
		validation.Issues = append(validation.Issues, "Stop loss must be above entry price for SHORT")
	}

	// Check 5: Take profit on correct side
	if req.Direction == "LONG" && req.TakeProfitPrice <= req.EntryPrice {
		validation.IsValid = false
		validation.Issues = append(validation.Issues, "Take profit must be above entry price for LONG")
	}
	if req.Direction == "SHORT" && req.TakeProfitPrice >= req.EntryPrice {
		validation.IsValid = false
		validation.Issues = append(validation.Issues, "Take profit must be below entry price for SHORT")
	}

	// Check 6: Quantity validation
	if req.Quantity <= 0 {
		validation.IsValid = false
		validation.Issues = append(validation.Issues, "Quantity must be > 0")
	}

	// Check 7: Calculate risk and portfolio impact
	var riskPerShare float64
	if req.Direction == "LONG" {
		riskPerShare = req.EntryPrice - req.StopLossPrice
	} else {
		riskPerShare = req.StopLossPrice - req.EntryPrice
	}

	validation.RiskAmount = float64(req.Quantity) * riskPerShare
	portfolioRiskPercent := (validation.RiskAmount / accountValue) * 100

	// Check 8: Max portfolio % per trade
	if portfolioRiskPercent > cfg.MaxPortfolioPercent {
		validation.IsValid = false
		validation.Issues = append(validation.Issues,
			fmt.Sprintf("Risk %.2f%% exceeds max %.2f%% of portfolio",
				portfolioRiskPercent, cfg.MaxPortfolioPercent))
	}

	validation.PortfolioRisk = portfolioRiskPercent

	// Calculate potential gain
	var gainPerShare float64
	if req.Direction == "LONG" {
		gainPerShare = req.TakeProfitPrice - req.EntryPrice
	} else {
		gainPerShare = req.EntryPrice - req.TakeProfitPrice
	}
	validation.PotentialGain = float64(req.Quantity) * gainPerShare
	validation.Quantity = req.Quantity

	return validation
}

// This function converts OrderRequest to Alpaca PlaceOrderRequest
func BuildPlaceOrderRequest(req *OrderRequest) (*alpaca.PlaceOrderRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("order request is nil")
	}

	side := alpaca.Buy
	if req.Direction == "SHORT" {
		side = alpaca.Sell
	} else if req.Direction != "LONG" {
		return nil, fmt.Errorf("invalid direction: %s (must be LONG or SHORT)", req.Direction)
	}

	orderType := alpaca.Market
	if req.UseLimitOrder {
		orderType = alpaca.Limit
	}

	placeOrderReq := &alpaca.PlaceOrderRequest{
		Symbol:      req.Symbol,
		Qty:         &decimal.Decimal{},
		Side:        side,
		Type:        orderType,
		TimeInForce: alpaca.Day,
	}

	*placeOrderReq.Qty = decimal.NewFromInt(req.Quantity)

	if req.UseLimitOrder {
		limitPrice := decimal.NewFromFloat(req.LimitPrice)
		placeOrderReq.LimitPrice = &limitPrice
	}

	return placeOrderReq, nil
}

// checks safe quantity based on account size and risk
func CalculatePositionSize(accountValue float64, entryPrice float64, stopLossPrice float64,
	maxRiskPercent float64, cfg *OrderConfig) int64 {

	riskPerShare := entryPrice - stopLossPrice
	if riskPerShare < 0 {
		riskPerShare = -riskPerShare
	}

	maxRiskDollars := (maxRiskPercent / 100) * accountValue

	positionSize := int64(maxRiskDollars / riskPerShare)

	if positionSize < 1 {
		positionSize = 1
	}

	// Verify it doesn't exceed portfolio percent limit
	totalRisk := float64(positionSize) * riskPerShare
	portfolioRiskPercent := (totalRisk / accountValue) * 100

	if portfolioRiskPercent > cfg.MaxPortfolioPercent {
		// Recalculate with max portfolio percent
		maxRiskDollars = (cfg.MaxPortfolioPercent / 100) * accountValue
		positionSize = int64(maxRiskDollars / riskPerShare)
	}

	return positionSize
}

// computes stop loss and take profit levels
func CalculatePriceTargets(entryPrice float64, direction string, cfg *OrderConfig) (stopLoss float64, takeProfit float64) {
	if direction == "LONG" {
		stopLoss = entryPrice * (1 - (cfg.StopLossPercent / 100))
		takeProfit = entryPrice * (1 + (cfg.TakeProfitPercent / 100))
	} else if direction == "SHORT" {
		stopLoss = entryPrice * (1 + (cfg.StopLossPercent / 100))
		takeProfit = entryPrice * (1 - (cfg.TakeProfitPercent / 100))
	}
	return
}

func LogOrderExecution(req *OrderRequest, validation *OrderValidation, orderId string) {
	log.Printf("========== ORDER EXECUTED ==========\n")
	log.Printf("Symbol: %s | Direction: %s | Qty: %d\n", req.Symbol, req.Direction, req.Quantity)
	log.Printf("Entry: $%.2f | SL: $%.2f | TP: $%.2f\n", req.EntryPrice, req.StopLossPrice, req.TakeProfitPrice)
	log.Printf("Portfolio Risk: %.2f%% | Potential Gain: $%.2f\n", validation.PortfolioRisk, validation.PotentialGain)
	log.Printf("Confidence: %.0f%% | Reason: %s\n", req.SignalConfidence, req.TradeReason)
	log.Printf("Order ID: %s\n", orderId)
	log.Printf("====================================\n")
}
