package strategy

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

// tracks an active trade
type OpenPosition struct {
	Symbol               string
	OrderID              string
	Direction            string // "LONG" or "SHORT"
	EntryPrice           float64
	Quantity             int64
	StopLossPrice        float64
	TakeProfitPrice      float64
	SafeBailPrice        float64 // Partial exit price
	EntryTime            time.Time
	CurrentPrice         float64
	UnrealizedPnL        float64
	UnrealizedPnLPercent float64
	Status               string // "OPEN", "PARTIAL_EXIT", "CLOSED"
}

// tracks all open positions and enforces limits
type PositionManager struct {
	positions      map[string]*OpenPosition // Key: OrderID
	positionsMutex sync.RWMutex
	config         *OrderConfig
	client         *alpaca.Client
	dailyLoss      float64
	dailyLossMutex sync.RWMutex
}

// creates a new position manager
func NewPositionManager(client *alpaca.Client, cfg *OrderConfig) *PositionManager {
	return &PositionManager{
		positions: make(map[string]*OpenPosition),
		config:    cfg,
		client:    client,
		dailyLoss: 0,
	}
}

// adds a new open position
func (pm *PositionManager) AddPosition(order *alpaca.Order, signal *TradeSignal, entryPrice float64,
	stopLoss float64, takeProfit float64, safeBail float64) *OpenPosition {

	pm.positionsMutex.Lock()
	defer pm.positionsMutex.Unlock()

	// Convert FilledQty to int64
	qtyFloat, _ := order.FilledQty.Float64()
	qty := int64(qtyFloat)

	position := &OpenPosition{
		Symbol:          order.Symbol,
		OrderID:         order.ID,
		Direction:       signal.Direction,
		EntryPrice:      entryPrice,
		Quantity:        qty,
		StopLossPrice:   stopLoss,
		TakeProfitPrice: takeProfit,
		SafeBailPrice:   safeBail,
		EntryTime:       order.CreatedAt,
		CurrentPrice:    entryPrice,
		Status:          "OPEN",
	}

	pm.positions[order.ID] = position
	log.Printf("✅ Position added: %s x%d @ $%.2f (ID: %s)\n",
		position.Symbol, position.Quantity, position.EntryPrice, position.OrderID)

	return position
}

// returns all open positions
func (pm *PositionManager) GetOpenPositions() []*OpenPosition {
	pm.positionsMutex.RLock()
	defer pm.positionsMutex.RUnlock()

	positions := make([]*OpenPosition, 0, len(pm.positions))
	for _, pos := range pm.positions {
		if pos.Status == "OPEN" {
			positions = append(positions, pos)
		}
	}
	return positions
}

// returns number of open trades
func (pm *PositionManager) CountOpenPositions() int {
	return len(pm.GetOpenPositions())
}

// updates current price and calculates P&L
func (pm *PositionManager) UpdatePosition(orderID string, currentPrice float64) error {
	pm.positionsMutex.Lock()
	defer pm.positionsMutex.Unlock()

	position, exists := pm.positions[orderID]
	if !exists {
		return fmt.Errorf("position not found: %s", orderID)
	}

	position.CurrentPrice = currentPrice

	// Calculate unrealized P&L
	if position.Direction == "LONG" {
		position.UnrealizedPnL = (currentPrice - position.EntryPrice) * float64(position.Quantity)
		position.UnrealizedPnLPercent = ((currentPrice - position.EntryPrice) / position.EntryPrice) * 100
	} else {
		position.UnrealizedPnL = (position.EntryPrice - currentPrice) * float64(position.Quantity)
		position.UnrealizedPnLPercent = ((position.EntryPrice - currentPrice) / position.EntryPrice) * 100
	}

	return nil
}

// checks if any positions hit stop loss
func (pm *PositionManager) CheckStopLosses() []*OpenPosition {
	pm.positionsMutex.RLock()
	defer pm.positionsMutex.RUnlock()

	hitStopLoss := make([]*OpenPosition, 0)

	for _, pos := range pm.positions {
		if pos.Status != "OPEN" {
			continue
		}

		shouldExit := false
		if pos.Direction == "LONG" && pos.CurrentPrice <= pos.StopLossPrice {
			shouldExit = true
		} else if pos.Direction == "SHORT" && pos.CurrentPrice >= pos.StopLossPrice {
			shouldExit = true
		}

		if shouldExit {
			hitStopLoss = append(hitStopLoss, pos)
			log.Printf("🛑 STOP LOSS HIT: %s @ $%.2f\n", pos.Symbol, pos.CurrentPrice)
		}
	}

	return hitStopLoss
}

// checks if any positions hit take profit
func (pm *PositionManager) CheckTakeProfits() []*OpenPosition {
	pm.positionsMutex.RLock()
	defer pm.positionsMutex.RUnlock()

	hitTakeProfit := make([]*OpenPosition, 0)

	for _, pos := range pm.positions {
		if pos.Status != "OPEN" {
			continue
		}

		shouldExit := false
		if pos.Direction == "LONG" && pos.CurrentPrice >= pos.TakeProfitPrice {
			shouldExit = true
		} else if pos.Direction == "SHORT" && pos.CurrentPrice <= pos.TakeProfitPrice {
			shouldExit = true
		}

		if shouldExit {
			hitTakeProfit = append(hitTakeProfit, pos)
			log.Printf("🎯 TAKE PROFIT HIT: %s @ $%.2f\n", pos.Symbol, pos.CurrentPrice)
		}
	}

	return hitTakeProfit
}

// checks if positions should partially exit at safe bail price
func (pm *PositionManager) CheckSafeBails() []*OpenPosition {
	pm.positionsMutex.RLock()
	defer pm.positionsMutex.RUnlock()

	readyForBail := make([]*OpenPosition, 0)

	for _, pos := range pm.positions {
		if pos.Status != "OPEN" || pos.SafeBailPrice <= 0 {
			continue
		}

		shouldBail := false
		if pos.Direction == "LONG" && pos.CurrentPrice >= pos.SafeBailPrice {
			shouldBail = true
		} else if pos.Direction == "SHORT" && pos.CurrentPrice <= pos.SafeBailPrice {
			shouldBail = true
		}

		if shouldBail {
			readyForBail = append(readyForBail, pos)
			log.Printf("💰 SAFE BAIL READY: %s @ $%.2f (profit lock)\n", pos.Symbol, pos.CurrentPrice)
		}
	}

	return readyForBail
}

// marks a position as closed and tracks P&L
func (pm *PositionManager) ClosePosition(orderID string, exitPrice float64, reason string) error {
	pm.positionsMutex.Lock()
	defer pm.positionsMutex.Unlock()

	position, exists := pm.positions[orderID]
	if !exists {
		return fmt.Errorf("position not found: %s", orderID)
	}

	position.CurrentPrice = exitPrice
	position.Status = "CLOSED"

	// Calculate realized P&L
	realizedPnL := 0.0
	if position.Direction == "LONG" {
		realizedPnL = (exitPrice - position.EntryPrice) * float64(position.Quantity)
	} else {
		realizedPnL = (position.EntryPrice - exitPrice) * float64(position.Quantity)
	}

	// Update daily loss tracking
	pm.dailyLossMutex.Lock()
	if realizedPnL < 0 {
		pm.dailyLoss += realizedPnL // Add negative value
	}
	pm.dailyLossMutex.Unlock()

	log.Printf("✅ Position closed: %s | Exit: $%.2f | P&L: $%.2f | Reason: %s\n",
		position.Symbol, exitPrice, realizedPnL, reason)

	return nil
}

// PartialExit reduces position size
func (pm *PositionManager) PartialExit(orderID string, exitQty int64, exitPrice float64) error {
	pm.positionsMutex.Lock()
	defer pm.positionsMutex.Unlock()

	position, exists := pm.positions[orderID]
	if !exists {
		return fmt.Errorf("position not found: %s", orderID)
	}

	if exitQty > position.Quantity {
		return fmt.Errorf("exit quantity (%d) exceeds position size (%d)", exitQty, position.Quantity)
	}

	position.Quantity -= exitQty
	position.Status = "PARTIAL_EXIT"

	log.Printf("📤 Partial exit: %s | Exited: %d @ $%.2f | Remaining: %d\n",
		position.Symbol, exitQty, exitPrice, position.Quantity)

	return nil
}

// returns total loss for the day
func (pm *PositionManager) GetDailyLoss() float64 {
	pm.dailyLossMutex.RLock()
	defer pm.dailyLossMutex.RUnlock()
	return pm.dailyLoss
}

// resets daily P&L (call at market open)
func (pm *PositionManager) ResetDailyLoss() {
	pm.dailyLossMutex.Lock()
	defer pm.dailyLossMutex.Unlock()
	pm.dailyLoss = 0
	log.Println("📊 Daily loss reset to $0.00")
}

// calculates portfolio statistics
func (pm *PositionManager) GetPortfolioStats(accountValue float64) map[string]interface{} {
	pm.positionsMutex.RLock()
	defer pm.positionsMutex.RUnlock()

	totalUnrealizedPnL := 0.0
	maxDrawdown := 0.0
	winningTrades := 0
	losingTrades := 0

	for _, pos := range pm.positions {
		totalUnrealizedPnL += pos.UnrealizedPnL

		if pos.UnrealizedPnLPercent < maxDrawdown {
			maxDrawdown = pos.UnrealizedPnLPercent
		}

		if pos.Status == "CLOSED" {
			if pos.UnrealizedPnL > 0 {
				winningTrades++
			} else if pos.UnrealizedPnL < 0 {
				losingTrades++
			}
		}
	}

	var winRate float64
	totalTrades := winningTrades + losingTrades
	if totalTrades > 0 {
		winRate = float64(winningTrades) / float64(totalTrades) * 100
	}

	return map[string]interface{}{
		"open_positions":        len(pm.GetOpenPositions()),
		"total_unrealized_pnl":  totalUnrealizedPnL,
		"portfolio_pnl_percent": (totalUnrealizedPnL / accountValue) * 100,
		"daily_loss":            pm.GetDailyLoss(),
		"max_drawdown":          maxDrawdown,
		"winning_trades":        winningTrades,
		"losing_trades":         losingTrades,
		"win_rate":              winRate,
	}
}

// continuously checks for stop loss/take profit hits
func (pm *PositionManager) MonitorPositions(ctx context.Context, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Position monitor stopped")
			return
		case <-ticker.C:
			// Check stop losses
			stopLossHits := pm.CheckStopLosses()
			for _, pos := range stopLossHits {
				log.Printf("🛑 STOP LOSS HIT: %s @ $%.2f - Go to menu option 8 to close\n", pos.Symbol, pos.CurrentPrice)
			}

			// Check take profits
			takeProfitHits := pm.CheckTakeProfits()
			for _, pos := range takeProfitHits {
				log.Printf("🎯 TAKE PROFIT HIT: %s @ $%.2f - Go to menu option 8 to close\n", pos.Symbol, pos.CurrentPrice)
			}

			// Check safe bails
			safeBails := pm.CheckSafeBails()
			for _, pos := range safeBails {
				log.Printf("💰 SAFE BAIL READY: %s @ $%.2f - Go to menu option 8 to partial exit\n", pos.Symbol, pos.CurrentPrice)
			}
		}
	}
}

// checks and displays alerts when returning to main menu
func (pm *PositionManager) CheckMenuAlerts() {
	separator := "============================================================"
	fmt.Println("\n" + separator)
	fmt.Println("📍 POSITION ALERTS")
	fmt.Println(separator)

	hasAlerts := false

	// Check stop losses
	stopLossHits := pm.CheckStopLosses()
	for _, pos := range stopLossHits {
		fmt.Printf("🛑 STOP LOSS HIT: %s @ $%.2f\n", pos.Symbol, pos.CurrentPrice)
		hasAlerts = true
	}

	// Check take profits
	takeProfitHits := pm.CheckTakeProfits()
	for _, pos := range takeProfitHits {
		fmt.Printf("🎯 TAKE PROFIT HIT: %s @ $%.2f\n", pos.Symbol, pos.CurrentPrice)
		hasAlerts = true
	}

	// Check safe bails
	safeBails := pm.CheckSafeBails()
	for _, pos := range safeBails {
		fmt.Printf("💰 SAFE BAIL READY: %s @ $%.2f\n", pos.Symbol, pos.CurrentPrice)
		hasAlerts = true
	}

	if hasAlerts {
		fmt.Println("\n👉 Select menu option 8 to close/sell positions")
		fmt.Println(separator)
	} else {
		fmt.Println("✅ No alerts - all positions are normal")
		fmt.Println(separator)
	}
}
