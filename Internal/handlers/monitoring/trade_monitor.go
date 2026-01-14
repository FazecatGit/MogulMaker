package monitoring

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/fazecat/mongelmaker/Internal/handlers/risk"
	"github.com/fazecat/mongelmaker/Internal/strategy/position"
	"github.com/fazecat/mongelmaker/Internal/utils/formatting"
)

// Real-time trade monitoring, P&L tracking, and analytics
type Monitor struct {
	positionManager *position.PositionManager
	riskManager     *risk.Manager
	tradeHistory    []*TradeRecord
	historyMutex    sync.RWMutex
	portfolioStats  *PortfolioStats
	statsMutex      sync.RWMutex
	isMonitoring    bool
	stopChan        chan bool
	updateInterval  time.Duration
}

// comprehensive record of a completed trade
type TradeRecord struct {
	ID                 string
	Symbol             string
	EntryTime          time.Time
	ExitTime           time.Time
	Direction          string // "LONG" or "SHORT"
	EntryPrice         float64
	ExitPrice          float64
	Quantity           int64
	EntryReason        string
	ExitReason         string
	RealizedPnL        float64
	RealizedPnLPercent float64
	Commission         float64
	Duration           time.Duration
	Status             string // "COMPLETED", "CANCELLED", "PARTIAL_EXIT"
	Tags               []string
}

// running statistics across all trades
type PortfolioStats struct {
	TotalTrades           int
	WinningTrades         int
	LosingTrades          int
	BreakevenTrades       int
	WinRate               float64 // 0-100%
	TotalProfit           float64
	TotalLoss             float64
	NetProfit             float64
	AverageProfitPerTrade float64
	AverageLossPerTrade   float64
	LargestWin            float64
	LargestLoss           float64
	ProfitFactor          float64 // Total profit / Total loss
	AvgTradeLength        time.Duration
	MaxConsecutiveWins    int
	MaxConsecutiveLosses  int
	Sharpe                float64
	MaxDrawdown           float64
	MaxDrawdownPercent    float64
	RecoveryTime          time.Duration
	LastUpdated           time.Time
}

// represents real-time position monitoring data
type PositionMonitor struct {
	Symbol               string
	Direction            string
	EntryPrice           float64
	CurrentPrice         float64
	Quantity             int64
	UnrealizedPnL        float64
	UnrealizedPnLPercent float64
	TimeInTrade          time.Duration
	RiskRewardRatio      float64
	Status               string
	AlertLevel           string // "NONE", "INFO", "WARNING", "CRITICAL"
	AlertMessage         string
}

// creates a new trade monitor
func NewMonitor(pm *position.PositionManager, rm *risk.Manager) *Monitor {
	return &Monitor{
		positionManager: pm,
		riskManager:     rm,
		tradeHistory:    make([]*TradeRecord, 0),
		portfolioStats: &PortfolioStats{
			TotalTrades:          0,
			WinningTrades:        0,
			LosingTrades:         0,
			MaxConsecutiveWins:   0,
			MaxConsecutiveLosses: 0,
		},
		isMonitoring:   false,
		stopChan:       make(chan bool),
		updateInterval: 1 * time.Second,
	}
}

// ============================================================================
// TRADE RECORDING
// ============================================================================

// records a completed trade
func (tm *Monitor) RecordTrade(
	symbol string,
	direction string,
	entryPrice float64,
	exitPrice float64,
	quantity int64,
	entryReason string,
	exitReason string,
	commission float64,
	entryTime time.Time,
	exitTime time.Time) *TradeRecord {

	// Calculate P&L
	var realizedPnL float64
	if direction == "LONG" {
		realizedPnL = (exitPrice - entryPrice) * float64(quantity)
	} else { // SHORT
		realizedPnL = (entryPrice - exitPrice) * float64(quantity)
	}
	realizedPnL -= commission

	realizedPnLPercent := (realizedPnL / (entryPrice * float64(quantity))) * 100

	trade := &TradeRecord{
		ID:                 fmt.Sprintf("%s_%d", symbol, time.Now().UnixNano()),
		Symbol:             symbol,
		Direction:          direction,
		EntryPrice:         entryPrice,
		ExitPrice:          exitPrice,
		Quantity:           quantity,
		EntryReason:        entryReason,
		ExitReason:         exitReason,
		RealizedPnL:        realizedPnL,
		RealizedPnLPercent: realizedPnLPercent,
		Commission:         commission,
		EntryTime:          entryTime,
		ExitTime:           exitTime,
		Duration:           exitTime.Sub(entryTime),
		Status:             "COMPLETED",
	}

	// Add to history
	tm.historyMutex.Lock()
	tm.tradeHistory = append(tm.tradeHistory, trade)
	tm.historyMutex.Unlock()

	// Update stats
	tm.updateStats()

	// Log trade
	emoji := "🟢"
	if realizedPnL < 0 {
		emoji = "🔴"
	}
	log.Printf("%s Trade recorded: %s %s $%.2f->$%.2f (%+.2f, %+.2f%%)\n",
		emoji, symbol, direction, entryPrice, exitPrice, realizedPnL, realizedPnLPercent)

	return trade
}

// updates portfolio statistics
func (tm *Monitor) updateStats() {
	tm.historyMutex.RLock()
	trades := tm.tradeHistory
	tm.historyMutex.RUnlock()

	if len(trades) == 0 {
		return
	}

	stats := &PortfolioStats{
		TotalTrades:          len(trades),
		MaxConsecutiveWins:   0,
		MaxConsecutiveLosses: 0,
	}

	currentWinStreak := 0
	currentLossStreak := 0

	for _, trade := range trades {
		if trade.RealizedPnL > 0.01 { // Winning trade
			stats.WinningTrades++
			stats.TotalProfit += trade.RealizedPnL
			stats.AverageProfitPerTrade += trade.RealizedPnL

			if trade.RealizedPnL > stats.LargestWin {
				stats.LargestWin = trade.RealizedPnL
			}

			currentWinStreak++
			if currentWinStreak > stats.MaxConsecutiveWins {
				stats.MaxConsecutiveWins = currentWinStreak
			}
			currentLossStreak = 0

		} else if trade.RealizedPnL < -0.01 { // Losing trade
			stats.LosingTrades++
			stats.TotalLoss += math.Abs(trade.RealizedPnL)
			stats.AverageLossPerTrade += math.Abs(trade.RealizedPnL)

			if math.Abs(trade.RealizedPnL) > stats.LargestLoss {
				stats.LargestLoss = math.Abs(trade.RealizedPnL)
			}

			currentLossStreak++
			if currentLossStreak > stats.MaxConsecutiveLosses {
				stats.MaxConsecutiveLosses = currentLossStreak
			}
			currentWinStreak = 0

		} else { // Breakeven
			stats.BreakevenTrades++
		}
	}

	// Calculate averages
	if stats.WinningTrades > 0 {
		stats.AverageProfitPerTrade /= float64(stats.WinningTrades)
	}
	if stats.LosingTrades > 0 {
		stats.AverageLossPerTrade /= float64(stats.LosingTrades)
	}

	// Win rate
	stats.WinRate = (float64(stats.WinningTrades) / float64(stats.TotalTrades)) * 100

	// Net profit
	stats.NetProfit = stats.TotalProfit - stats.TotalLoss

	// Profit factor
	if stats.TotalLoss > 0 {
		stats.ProfitFactor = stats.TotalProfit / stats.TotalLoss
	} else {
		stats.ProfitFactor = 0
	}

	// Average trade duration
	totalDuration := time.Duration(0)
	for _, trade := range trades {
		totalDuration += trade.Duration
	}
	stats.AvgTradeLength = totalDuration / time.Duration(len(trades))

	// Calculate max drawdown
	stats.MaxDrawdown, stats.MaxDrawdownPercent = tm.calculateMaxDrawdown(trades)

	stats.LastUpdated = time.Now()

	tm.statsMutex.Lock()
	tm.portfolioStats = stats
	tm.statsMutex.Unlock()
}

// calculates maximum drawdown from historical trades
func (tm *Monitor) calculateMaxDrawdown(trades []*TradeRecord) (float64, float64) {
	if len(trades) == 0 {
		return 0, 0
	}

	balance := 0.0
	peak := 0.0
	maxDrawdown := 0.0
	maxDrawdownPercent := 0.0

	for _, trade := range trades {
		balance += trade.RealizedPnL

		if balance > peak {
			peak = balance
		}

		drawdown := peak - balance
		drawdownPercent := (drawdown / peak) * 100

		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
			maxDrawdownPercent = drawdownPercent
		}
	}

	return maxDrawdown, maxDrawdownPercent
}

// ============================================================================
// REAL-TIME MONITORING
// ============================================================================

// starts continuous position monitoring
func (tm *Monitor) StartMonitoring() {
	if tm.isMonitoring {
		log.Println("⚠️  Monitoring already active")
		return
	}

	tm.isMonitoring = true
	log.Println("🟢 Trade monitoring started")

	go tm.monitoringLoop()
}

// stops monitoring
func (tm *Monitor) StopMonitoring() {
	if !tm.isMonitoring {
		return
	}

	tm.isMonitoring = false
	tm.stopChan <- true
	log.Println("🔴 Trade monitoring stopped")
}

// main monitoring loop - runs continuously
func (tm *Monitor) monitoringLoop() {
	ticker := time.NewTicker(tm.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tm.stopChan:
			return
		case <-ticker.C:
			tm.updatePositionMetrics()
		}
	}
}

// updates real-time metrics for all open positions
func (tm *Monitor) updatePositionMetrics() {
	positions := tm.positionManager.GetOpenPositions()

	for _, pos := range positions {
		// Update unrealized P&L
		var unrealizedPnL float64
		if pos.Direction == "LONG" {
			unrealizedPnL = (pos.CurrentPrice - pos.EntryPrice) * float64(pos.Quantity)
		} else { // SHORT
			unrealizedPnL = (pos.EntryPrice - pos.CurrentPrice) * float64(pos.Quantity)
		}

		pos.UnrealizedPnL = unrealizedPnL
		pos.UnrealizedPnLPercent = (unrealizedPnL / (pos.EntryPrice * float64(pos.Quantity))) * 100

		// Check for alerts
		tm.checkPositionAlerts(pos)
	}
}

// checks for alert conditions on a position
func (tm *Monitor) checkPositionAlerts(pos *position.OpenPosition) {
	// Check stop loss
	if pos.Direction == "LONG" && pos.CurrentPrice <= pos.StopLossPrice {
		tm.riskManager.SendAlert(&risk.Alert{
			Level:   "CRITICAL",
			Title:   "🛑 STOP LOSS HIT",
			Message: fmt.Sprintf("%s stop loss triggered at $%.2f", pos.Symbol, pos.CurrentPrice),
			Symbol:  pos.Symbol,
			Data: map[string]interface{}{
				"price":    pos.CurrentPrice,
				"stopLoss": pos.StopLossPrice,
				"loss":     pos.UnrealizedPnL,
			},
		})
	} else if pos.Direction == "SHORT" && pos.CurrentPrice >= pos.StopLossPrice {
		tm.riskManager.SendAlert(&risk.Alert{
			Level:   "CRITICAL",
			Title:   "🛑 STOP LOSS HIT",
			Message: fmt.Sprintf("%s stop loss triggered at $%.2f", pos.Symbol, pos.CurrentPrice),
			Symbol:  pos.Symbol,
			Data: map[string]interface{}{
				"price":    pos.CurrentPrice,
				"stopLoss": pos.StopLossPrice,
				"loss":     pos.UnrealizedPnL,
			},
		})
	}

	// Check take profit
	if pos.Direction == "LONG" && pos.CurrentPrice >= pos.TakeProfitPrice {
		tm.riskManager.SendAlert(&risk.Alert{
			Level:   "INFO",
			Title:   "🎯 TAKE PROFIT TARGET HIT",
			Message: fmt.Sprintf("%s take profit target reached at $%.2f", pos.Symbol, pos.CurrentPrice),
			Symbol:  pos.Symbol,
			Data: map[string]interface{}{
				"price":     pos.CurrentPrice,
				"target":    pos.TakeProfitPrice,
				"profit":    pos.UnrealizedPnL,
				"profitPct": pos.UnrealizedPnLPercent,
			},
		})
	} else if pos.Direction == "SHORT" && pos.CurrentPrice <= pos.TakeProfitPrice {
		tm.riskManager.SendAlert(&risk.Alert{
			Level:   "INFO",
			Title:   "🎯 TAKE PROFIT TARGET HIT",
			Message: fmt.Sprintf("%s take profit target reached at $%.2f", pos.Symbol, pos.CurrentPrice),
			Symbol:  pos.Symbol,
			Data: map[string]interface{}{
				"price":     pos.CurrentPrice,
				"target":    pos.TakeProfitPrice,
				"profit":    pos.UnrealizedPnL,
				"profitPct": pos.UnrealizedPnLPercent,
			},
		})
	}

	// Check safe bail (partial exit)
	if pos.SafeBailPrice > 0 {
		if pos.Direction == "LONG" && pos.CurrentPrice >= pos.SafeBailPrice {
			tm.riskManager.SendAlert(&risk.Alert{
				Level:   "INFO",
				Title:   "🟡 SAFE BAIL LEVEL REACHED",
				Message: fmt.Sprintf("%s safe bail level reached at $%.2f", pos.Symbol, pos.CurrentPrice),
				Symbol:  pos.Symbol,
				Data: map[string]interface{}{
					"price":     pos.CurrentPrice,
					"bailLevel": pos.SafeBailPrice,
				},
			})
		} else if pos.Direction == "SHORT" && pos.CurrentPrice <= pos.SafeBailPrice {
			tm.riskManager.SendAlert(&risk.Alert{
				Level:   "INFO",
				Title:   "🟡 SAFE BAIL LEVEL REACHED",
				Message: fmt.Sprintf("%s safe bail level reached at $%.2f", pos.Symbol, pos.CurrentPrice),
				Symbol:  pos.Symbol,
				Data: map[string]interface{}{
					"price":     pos.CurrentPrice,
					"bailLevel": pos.SafeBailPrice,
				},
			})
		}
	}
}

// ============================================================================
// STATISTICS & REPORTING
// ============================================================================

// returns current portfolio statistics
func (tm *Monitor) GetStats() *PortfolioStats {
	tm.statsMutex.RLock()
	defer tm.statsMutex.RUnlock()
	return tm.portfolioStats
}

// returns trade history with optional filtering
func (tm *Monitor) GetTradeHistory(limit int) []*TradeRecord {
	tm.historyMutex.RLock()
	defer tm.historyMutex.RUnlock()

	trades := tm.tradeHistory
	if limit > 0 && len(trades) > limit {
		trades = trades[len(trades)-limit:]
	}
	return trades
}

// returns current position monitors for all open trades
func (tm *Monitor) GetPositionMonitors() []*PositionMonitor {
	positions := tm.positionManager.GetOpenPositions()
	monitors := make([]*PositionMonitor, len(positions))

	for i, pos := range positions {
		timeInTrade := time.Since(pos.EntryTime)
		riskReward := 0.0
		if pos.EntryPrice > pos.StopLossPrice {
			riskReward = (pos.TakeProfitPrice - pos.EntryPrice) / (pos.EntryPrice - pos.StopLossPrice)
		}

		alertLevel, alertMsg := tm.determineAlertLevel(pos.UnrealizedPnLPercent)

		monitors[i] = &PositionMonitor{
			Symbol:               pos.Symbol,
			Direction:            pos.Direction,
			EntryPrice:           pos.EntryPrice,
			CurrentPrice:         pos.CurrentPrice,
			Quantity:             pos.Quantity,
			UnrealizedPnL:        pos.UnrealizedPnL,
			UnrealizedPnLPercent: pos.UnrealizedPnLPercent,
			TimeInTrade:          timeInTrade,
			RiskRewardRatio:      riskReward,
			Status:               pos.Status,
			AlertLevel:           alertLevel,
			AlertMessage:         alertMsg,
		}
	}

	return monitors
}

// ============================================================================
// REPORTING
// ============================================================================

// prints formatted statistics report
func (tm *Monitor) PrintStatsReport() {
	stats := tm.GetStats()

	width := 70
	fmt.Println("\n" + formatting.Separator(width))
	fmt.Println("📈 TRADE STATISTICS REPORT")
	fmt.Println(formatting.Separator(width))
	fmt.Printf("Total Trades:          %d\n", stats.TotalTrades)
	fmt.Printf("Winning Trades:        %d (%.1f%% win rate)\n", stats.WinningTrades, stats.WinRate)
	fmt.Printf("Losing Trades:         %d\n", stats.LosingTrades)
	fmt.Printf("Breakeven Trades:      %d\n", stats.BreakevenTrades)
	fmt.Printf("\n")
	fmt.Printf("Total Profit:          $%.2f\n", stats.TotalProfit)
	fmt.Printf("Total Loss:            $%.2f\n", stats.TotalLoss)
	fmt.Printf("Net Profit:            $%.2f\n", stats.NetProfit)
	fmt.Printf("Profit Factor:         %.2f (revenue/losses ratio)\n", stats.ProfitFactor)
	fmt.Printf("\n")
	fmt.Printf("Avg Profit/Trade:      $%.2f\n", stats.AverageProfitPerTrade)
	fmt.Printf("Avg Loss/Trade:        $%.2f\n", stats.AverageLossPerTrade)
	fmt.Printf("Largest Win:           $%.2f\n", stats.LargestWin)
	fmt.Printf("Largest Loss:          $%.2f\n", stats.LargestLoss)
	fmt.Printf("\n")
	fmt.Printf("Max Consecutive Wins:  %d\n", stats.MaxConsecutiveWins)
	fmt.Printf("Max Consecutive Losses: %d\n", stats.MaxConsecutiveLosses)
	fmt.Printf("Avg Trade Duration:    %v\n", stats.AvgTradeLength)
	fmt.Printf("Max Drawdown:          $%.2f (%.2f%%)\n", stats.MaxDrawdown, stats.MaxDrawdownPercent)
	fmt.Println(formatting.Separator(width) + "\n")
}

// prints current open positions with real-time P&L
func (tm *Monitor) PrintOpenPositions() {
	// Sync with Alpaca to get latest positions
	if tm.positionManager != nil {
		ctx := context.Background()
		if err := tm.positionManager.SyncFromAlpaca(ctx); err != nil {
			log.Printf("Warning: Could not sync positions from Alpaca: %v\n", err)
		}
	}

	monitors := tm.GetPositionMonitors()

	if len(monitors) == 0 {
		fmt.Println("No open positions")
		return
	}

	width := 100
	fmt.Println("\n" + formatting.Separator(width))
	fmt.Println("📊 OPEN POSITIONS")
	fmt.Println(formatting.Separator(width))
	fmt.Printf("%-8s %-6s %-8s %-8s %-10s %-8s %-8s %-12s %-12s %-10s\n",
		"Symbol", "Dir", "Entry", "Current", "Qty", "U/R P&L", "U/R %", "Time", "R/R Ratio", "Alert")

	for _, m := range monitors {
		emoji := "🟢"
		if m.UnrealizedPnLPercent < 0 {
			emoji = "🔴"
		}

		fmt.Printf("%-8s %-6s $%-7.2f $%-7.2f %-10d $%-7.2f %-7.2f%% %-12v %.2f %s\n",
			m.Symbol, m.Direction, m.EntryPrice, m.CurrentPrice, m.Quantity,
			m.UnrealizedPnL, m.UnrealizedPnLPercent, m.TimeInTrade, m.RiskRewardRatio,
			emoji+" "+m.AlertLevel)
	}
	fmt.Println(formatting.Separator(width) + "\n")
}

// determineAlertLevel evaluates position P&L and returns appropriate alert level and message
func (tm *Monitor) determineAlertLevel(unrealizedPnLPercent float64) (string, string) {
	switch {
	case unrealizedPnLPercent <= -2:
		return "CRITICAL", "Critical loss threshold"
	case unrealizedPnLPercent <= -1:
		return "WARNING", "Approaching stop loss"
	case unrealizedPnLPercent >= 3:
		return "INFO", "Good profit, consider partial exit"
	default:
		return "NONE", ""
	}
}
