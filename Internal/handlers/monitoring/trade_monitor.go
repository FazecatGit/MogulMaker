package monitoring

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	"github.com/fazecat/mogulmaker/Internal/handlers/risk"
	"github.com/fazecat/mogulmaker/Internal/strategy/position"
	"github.com/fazecat/mogulmaker/Internal/utils/formatting"
)

// P&L tracking, and analytics
type Monitor struct {
	positionManager *position.PositionManager
	riskManager     *risk.Manager
	queries         *database.Queries
}

type TradeRecord struct {
	ID                 string
	Symbol             string
	EntryTime          time.Time
	ExitTime           time.Time
	Direction          string
	EntryPrice         float64
	ExitPrice          float64
	Quantity           int64
	EntryReason        string
	ExitReason         string
	RealizedPnL        float64
	RealizedPnLPercent float64
	Commission         float64
	Duration           time.Duration
	Status             string
	Tags               []string
}

type PortfolioStats struct {
	TotalTrades           int
	WinningTrades         int
	LosingTrades          int
	BreakevenTrades       int
	WinRate               float64
	TotalProfit           float64
	TotalLoss             float64
	NetProfit             float64
	AverageProfitPerTrade float64
	AverageLossPerTrade   float64
	LargestWin            float64
	LargestLoss           float64
	ProfitFactor          float64
	AvgTradeLength        time.Duration
	MaxConsecutiveWins    int
	MaxConsecutiveLosses  int
	Sharpe                float64
	MaxDrawdown           float64
	MaxDrawdownPercent    float64
	RecoveryTime          time.Duration
	LastUpdated           time.Time
}


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
	AlertLevel           string
	AlertMessage         string
}

// STATISTICS & REPORTING

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

func NewMonitor(positionManager *position.PositionManager, riskManager *risk.Manager, queries *database.Queries) *Monitor {
	return &Monitor{
		positionManager: positionManager,
		riskManager:     riskManager,
		queries:         queries,
	}
}

// REPORTING

func (tm *Monitor) PrintStatsReport() {
	if tm.queries == nil {
		fmt.Println("Database queries not available")
		return
	}

	ctx := context.Background()
	trades, err := tm.queries.GetAllTrades(ctx)
	if err != nil {
		fmt.Printf("Error fetching trades: %v\n", err)
		return
	}

	if len(trades) == 0 {
		fmt.Println("\nNo trades found in database")
		return
	}

	stats := tm.calculateStatsFromTrades(trades)

	width := 70
	fmt.Println("\n" + formatting.Separator(width))
	fmt.Println("TRADE STATISTICS REPORT")
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

// generates portfolio statistics from trade records
func (tm *Monitor) calculateStatsFromTrades(trades []database.GetAllTradesRow) *PortfolioStats {
	stats := &PortfolioStats{}

	// Group trades by symbol to match buy/sell pairs
	type Position struct {
		buys  []database.GetAllTradesRow
		sells []database.GetAllTradesRow
	}
	positions := make(map[string]*Position)

	for _, trade := range trades {
		if _, exists := positions[trade.Symbol]; !exists {
			positions[trade.Symbol] = &Position{}
		}

		side := strings.ToUpper(trade.Side)
		if side == "BUY" || side == "LONG" {
			positions[trade.Symbol].buys = append(positions[trade.Symbol].buys, trade)
		} else if side == "SELL" || side == "SHORT" {
			positions[trade.Symbol].sells = append(positions[trade.Symbol].sells, trade)
		}
	}

	// Calculate P&L for completed trades
	var completedTrades []float64
	var tradeDurations []time.Duration
	consecutiveWins := 0
	consecutiveLosses := 0

	for _, pos := range positions {
		// FIFO matching
		for i := 0; i < len(pos.buys) && i < len(pos.sells); i++ {
			buy := pos.buys[i]
			sell := pos.sells[i]

			buyPrice, _ := strconv.ParseFloat(buy.Price, 64)
			sellPrice, _ := strconv.ParseFloat(sell.Price, 64)
			qty, _ := strconv.ParseFloat(buy.Quantity, 64)

			pnl := (sellPrice - buyPrice) * qty
			completedTrades = append(completedTrades, pnl)

			stats.TotalTrades++
			if pnl > 0 {
				stats.WinningTrades++
				stats.TotalProfit += pnl
				if pnl > stats.LargestWin {
					stats.LargestWin = pnl
				}
				consecutiveWins++
				if consecutiveWins > stats.MaxConsecutiveWins {
					stats.MaxConsecutiveWins = consecutiveWins
				}
				consecutiveLosses = 0
			} else if pnl < 0 {
				stats.LosingTrades++
				stats.TotalLoss += pnl
				if pnl < stats.LargestLoss {
					stats.LargestLoss = pnl
				}
				consecutiveLosses++
				if consecutiveLosses > stats.MaxConsecutiveLosses {
					stats.MaxConsecutiveLosses = consecutiveLosses
				}
				consecutiveWins = 0
			} else {
				stats.BreakevenTrades++
				consecutiveWins = 0
				consecutiveLosses = 0
			}

			// Calculate trade duration
			if buy.CreatedAt.Valid && sell.CreatedAt.Valid {
				duration := sell.CreatedAt.Time.Sub(buy.CreatedAt.Time)
				tradeDurations = append(tradeDurations, duration)
			}
		}
	}

	// Show helpful message if no completed trades
	if stats.TotalTrades == 0 {
		fmt.Println("\n NOTE: No completed trade pairs found (need both entry and exit trades)")
		fmt.Println("   Your database has entry trades (LONG/BUY) but no exit trades (SHORT/SELL)")
		fmt.Println("\n💡 TIP: To see current P&L, check 'Open Positions' which shows unrealized gains/losses")
	}

	// Calculate derived statistics
	if stats.TotalTrades > 0 {
		stats.WinRate = (float64(stats.WinningTrades) / float64(stats.TotalTrades)) * 100
		stats.NetProfit = stats.TotalProfit + stats.TotalLoss // TotalLoss is negative
	}

	if stats.WinningTrades > 0 {
		stats.AverageProfitPerTrade = stats.TotalProfit / float64(stats.WinningTrades)
	}

	if stats.LosingTrades > 0 {
		stats.AverageLossPerTrade = stats.TotalLoss / float64(stats.LosingTrades)
	}

	if stats.TotalLoss != 0 {
		stats.ProfitFactor = stats.TotalProfit / -stats.TotalLoss
	}

	// Calculate average trade duration
	if len(tradeDurations) > 0 {
		var totalDuration time.Duration
		for _, d := range tradeDurations {
			totalDuration += d
		}
		stats.AvgTradeLength = totalDuration / time.Duration(len(tradeDurations))
	}

	// Calculate max drawdown
	if len(completedTrades) > 0 {
		var peak, trough float64
		runningTotal := 0.0
		for _, pnl := range completedTrades {
			runningTotal += pnl
			if runningTotal > peak {
				peak = runningTotal
			}
			drawdown := peak - runningTotal
			if drawdown > stats.MaxDrawdown {
				stats.MaxDrawdown = drawdown
				trough = runningTotal
			}
		}
		if peak > 0 {
			stats.MaxDrawdownPercent = (stats.MaxDrawdown / peak) * 100
		}
		_ = trough
	}

	return stats
}

// UpdatePositionAlerts records CRITICAL positions as risk events (called from API endpoints)
func (tm *Monitor) UpdatePositionAlerts() {
	// Sync with Alpaca first
	if tm.positionManager != nil {
		ctx := context.Background()
		if err := tm.positionManager.SyncFromAlpaca(ctx); err != nil {
			log.Printf("Warning: Could not sync positions from Alpaca: %v\n", err)
		}
	}

	monitors := tm.GetPositionMonitors()

	for _, m := range monitors {
		if m.AlertLevel == "CRITICAL" {
			// Record CRITICAL position as risk event
			if tm.riskManager != nil {
				tm.riskManager.RecordCriticalPosition(&risk.Event{
					Timestamp:           time.Now(),
					EventType:           "POSITION_CRITICAL",
					Severity:            "CRITICAL",
					Symbol:              m.Symbol,
					Details:             fmt.Sprintf("Position in CRITICAL status: %s %.2f%% loss. Entry: $%.2f, Current: $%.2f", m.Direction, m.UnrealizedPnLPercent, m.EntryPrice, m.CurrentPrice),
					CurrentAccountValue: 0,
					CurrentDailyLoss:    m.UnrealizedPnL,
				})
			}
		}
	}
}

func (tm *Monitor) PrintOpenPositions() {
	// Sync with Alpaca first
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
	fmt.Println(" OPEN POSITIONS")
	fmt.Println(formatting.Separator(width))
	fmt.Printf("%-8s %-6s %-8s %-8s %-10s %-8s %-8s %-12s %-12s %-10s\n",
		"Symbol", "Dir", "Entry", "Current", "Qty", "U/R P&L", "U/R %", "Time", "R/R Ratio", "Alert")

	criticalPositions := []string{}

	for _, m := range monitors {
		indicator := "[OK]"
		if m.AlertLevel == "WARNING" {
			indicator = "[!]"
		} else if m.AlertLevel == "CRITICAL" {
			indicator = "[X]"
		} else if m.AlertLevel == "INFO" {
			indicator = "[i]"
		}

		fmt.Printf("%-8s %-6s $%-7.2f $%-7.2f %-10d $%-7.2f %-7.2f%% %-12v %.2f %-4s %s\n",
			m.Symbol, m.Direction, m.EntryPrice, m.CurrentPrice, m.Quantity,
			m.UnrealizedPnL, m.UnrealizedPnLPercent, m.TimeInTrade, m.RiskRewardRatio,
			indicator, m.AlertLevel)

		if m.AlertLevel == "CRITICAL" {
			criticalPositions = append(criticalPositions, m.Symbol)

			// Record CRITICAL position as risk event
			if tm.riskManager != nil {
				tm.riskManager.RecordCriticalPosition(&risk.Event{
					Timestamp:           time.Now(),
					EventType:           "POSITION_CRITICAL",
					Severity:            "CRITICAL",
					Symbol:              m.Symbol,
					Details:             fmt.Sprintf("Position in CRITICAL status: %s %.2f%% loss. Entry: $%.2f, Current: $%.2f", m.Direction, m.UnrealizedPnLPercent, m.EntryPrice, m.CurrentPrice),
					CurrentAccountValue: 0, // Will be set by risk manager if needed
					CurrentDailyLoss:    m.UnrealizedPnL,
				})
			}
		}
	}

	fmt.Println(formatting.Separator(width) + "\n")

	// Prompt for critical positions
	if len(criticalPositions) > 0 {
		fmt.Printf("\n Found %d position(s) in CRITICAL status: %v\n", len(criticalPositions), criticalPositions)
		fmt.Print("Do you want to close any of these positions? (y/n): ")
		var response string
		fmt.Scanln(&response)

		if response == "y" || response == "Y" || response == "yes" {
			for _, symbol := range criticalPositions {
				fmt.Printf("\nClose %s? (y/n): ", symbol)
				var confirmClose string
				fmt.Scanln(&confirmClose)

				if confirmClose == "y" || confirmClose == "Y" || confirmClose == "yes" {
					if tm.riskManager != nil {
						fmt.Printf(" Closing %s...\n", symbol)
						err := tm.riskManager.ClosePositionBySymbol(symbol)
						if err != nil {
							// Check if the position may already be closed
							errMsg := err.Error()
							if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found") {
								fmt.Printf("  %s appears to already be closed or doesn't exist\n", symbol)
							} else {
								fmt.Printf("  Failed to close %s: %v\n", symbol, err)
							}
						} else {
							fmt.Printf("  %s closed successfully\n", symbol)
						}
					} else {
						fmt.Println(" Risk Manager not available")
					}
				}
			}
		}
	}
}

// checks if portfolio risk is exceeded and offers to close high-risk positions
func (tm *Monitor) CheckPortfolioRiskAndClose() {
	if tm.positionManager == nil || tm.riskManager == nil {
		fmt.Println("❌ Position or Risk Manager not available")
		return
	}

	ctx := context.Background()
	if err := tm.positionManager.SyncFromAlpaca(ctx); err != nil {
		log.Printf("Warning: Could not sync positions: %v\n", err)
	}

	openPositions := tm.positionManager.GetOpenPositions()
	if len(openPositions) == 0 {
		fmt.Println("No open positions to analyze")
		return
	}

	portfolioRisk := tm.riskManager.CalculatePortfolioRisk(openPositions)

	if !portfolioRisk.IsOverRisk {
		fmt.Printf(" Portfolio risk is within limits: %.2f%% (max %.2f%%)\n",
			portfolioRisk.TotalRiskPercent, tm.riskManager.MaxPortfolioRiskPercent)
		return
	}

	fmt.Printf("\n  PORTFOLIO RISK EXCEEDED: %.2f%% (max %.2f%%)\n",
		portfolioRisk.TotalRiskPercent, tm.riskManager.MaxPortfolioRiskPercent)
	fmt.Printf("Max allowed risk: $%.2f\n", portfolioRisk.MaxAllowedRisk)
	fmt.Printf("Current total risk: $%.2f\n\n", portfolioRisk.TotalRiskAmount)

	fmt.Println("Position Risk Breakdown:")
	fmt.Println("────────────────────────────────────────────────────")
	fmt.Printf("%-8s %-12s %-12s\n", "Symbol", "Risk $", "Risk %")
	fmt.Println("────────────────────────────────────────────────────")

	for _, pr := range portfolioRisk.PositionRisks {
		fmt.Printf("%-8s $%-11.2f %-11.2f%%\n", pr.Symbol, pr.RiskAmount, pr.RiskPercent)
	}

	fmt.Print("Close positions to reduce risk? (y/n): ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" || response == "yes" {
		// Sort positions by risk with highest being first
		riskMap := make(map[string]float64)
		for _, pr := range portfolioRisk.PositionRisks {
			riskMap[pr.Symbol] = pr.RiskAmount
		}

		for _, pr := range portfolioRisk.PositionRisks {
			fmt.Printf("\nClose %s (Risk: $%.2f)? (y/n): ", pr.Symbol, pr.RiskAmount)
			var closeConfirm string
			fmt.Scanln(&closeConfirm)

			if closeConfirm == "y" || closeConfirm == "Y" || closeConfirm == "yes" {
				if tm.riskManager != nil {
					fmt.Printf(" Closing %s...\n", pr.Symbol)
					err := tm.riskManager.ClosePositionBySymbol(pr.Symbol)
					if err != nil {
						errMsg := err.Error()
						if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found") {
							fmt.Printf("%s appears to already be closed or doesn't exist\n", pr.Symbol)
						} else {
							fmt.Printf("Failed to close %s: %v\n", pr.Symbol, err)
						}
					} else {
						fmt.Printf("%s closed successfully\n", pr.Symbol)
					}
				}
			}
		}
	}
}


func (tm *Monitor) PrintTradeHistory() {
	if tm.queries == nil {
		fmt.Println(" Database queries not available")
		return
	}

	ctx := context.Background()
	trades, err := tm.queries.GetAllTrades(ctx)
	if err != nil {
		fmt.Printf("Error fetching trades: %v\n", err)
		return
	}

	if len(trades) == 0 {
		fmt.Println("\n No trades found in database")
		return
	}

	width := 100
	fmt.Println("\n" + formatting.Separator(width))
	fmt.Println("TRADE HISTORY")
	fmt.Println(formatting.Separator(width))
	fmt.Printf("%-6s %-8s %-6s %-10s %-10s %-12s %-15s %-20s\n",
		"ID", "Symbol", "Side", "Quantity", "Price", "Total", "Status", "Created")
	fmt.Println(formatting.Separator(width))

	for _, trade := range trades {
		status := "UNKNOWN"
		if trade.Status.Valid {
			status = trade.Status.String
		}

		createdAt := "N/A"
		if trade.CreatedAt.Valid {
			createdAt = trade.CreatedAt.Time.Format("2006-01-02 15:04")
		}

		fmt.Printf("%-6d %-8s %-6s %-10s $%-9s $%-11s %-15s %-20s\n",
			trade.ID, trade.Symbol, trade.Side, trade.Quantity,
			trade.Price, trade.TotalValue, status, createdAt)
	}

	fmt.Println(formatting.Separator(width) + "\n")
}

func (tm *Monitor) PrintRiskEvents() {
	if tm.riskManager == nil {
		fmt.Println("Risk Manager not available")
		return
	}

	width := 80
	fmt.Println("\n" + formatting.Separator(width))
	fmt.Println("RISK EVENTS & ALERTS")
	fmt.Println(formatting.Separator(width))


	events := tm.riskManager.GetRecentEvents()

	if len(events) == 0 {
		fmt.Println("No recent risk events")
	} else {
		for _, event := range events {
			fmt.Printf("%s\n", event)
		}
	}

	fmt.Println(formatting.Separator(width) + "\n")
}

func (tm *Monitor) determineAlertLevel(unrealizedPnLPercent float64) (string, string) {
	switch {
	case unrealizedPnLPercent <= -2:
		return "CRITICAL", "Critical loss threshold"
	case unrealizedPnLPercent <= -1:
		return "WARNING", "Approaching stop loss"
	case unrealizedPnLPercent >= 3:
		return "POSITIVE", "Good profit, consider partial exit"
	default:
		return "NONE", ""
	}
}
