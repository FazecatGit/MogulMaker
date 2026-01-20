package risk

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/fazecat/mogulmaker/Internal/strategy/position"
	"github.com/fazecat/mogulmaker/Internal/utils/formatting"
)

// Portfolio-level risk controls and limits
type Manager struct {
	// Daily loss limits
	MaxDailyLossPercent    float64 // -2% of account
	MaxDailyLossAmount     float64 // Actual dollar amount
	CurrentDailyLossAmount float64 // Current cumulative loss
	DailyLossResetTime     time.Time

	// Position limits
	MaxOpenPositions        int     // 5 trades max
	MaxPositionSizePercent  float64 // 20% of account per trade
	MaxPortfolioRiskPercent float64 // Overall portfolio risk cap

	// Sector diversification
	MaxSameSectorPositions int            // 3 trades max in same sector
	PositionsBySymbol      map[string]int // Track positions per symbol
	PositionsBySector      map[string]int // Track positions per sector
	positionsMutex         sync.RWMutex

	// Account tracking
	accountBalance        float64
	accountBalanceMutex   sync.RWMutex
	client                *alpaca.Client
	lastAccountUpdateTime time.Time

	// Risk events log
	riskEvents      []*Event
	riskEventsMutex sync.RWMutex

	// Alerts
	alertCallbacks      []AlertCallback
	alertCallbacksMutex sync.RWMutex
}

// represents a significant risk event
type Event struct {
	Timestamp           time.Time
	EventType           string // "MAX_DAILY_LOSS_HIT", "MAX_POSITIONS_HIT", "POSITION_SIZE_EXCEEDED", etc.
	Severity            string
	Symbol              string
	Details             string
	CurrentAccountValue float64
	CurrentDailyLoss    float64
}

// callback function for risk alerts
type AlertCallback func(*Alert)

type Alert struct {
	Level     string // "INFO", "WARNING", "CRITICAL"
	Title     string
	Message   string
	Timestamp time.Time
	Symbol    string
	Data      map[string]interface{}
}

// default limits for trades
func NewManager(client *alpaca.Client, accountBalance float64) *Manager {
	return &Manager{
		MaxDailyLossPercent:     2.0,
		MaxDailyLossAmount:      accountBalance * 0.02, // Calculate dollar amount
		CurrentDailyLossAmount:  0,
		DailyLossResetTime:      time.Now(),
		MaxOpenPositions:        150,
		MaxPositionSizePercent:  20.0,
		MaxPortfolioRiskPercent: 10.0,
		MaxSameSectorPositions:  3,
		PositionsBySymbol:       make(map[string]int),
		PositionsBySector:       make(map[string]int),
		accountBalance:          accountBalance,
		client:                  client,
		lastAccountUpdateTime:   time.Now(),
		riskEvents:              make([]*Event, 0),
		alertCallbacks:          make([]AlertCallback, 0),
	}
}

// ACCOUNT BALANCE MANAGEMENT

func (rm *Manager) UpdateAccountBalance(newBalance float64) {
	rm.accountBalanceMutex.Lock()
	defer rm.accountBalanceMutex.Unlock()

	rm.accountBalance = newBalance

	// Reset daily loss
	now := time.Now()
	if now.Sub(rm.lastAccountUpdateTime) > 24*time.Hour {
		rm.CurrentDailyLossAmount = 0
		rm.DailyLossResetTime = now
		log.Printf("📊 Daily loss reset. New account balance: $%.2f\n", newBalance)
	}

	rm.lastAccountUpdateTime = now
}

// returns current account balance
func (rm *Manager) GetAccountBalance() float64 {
	rm.accountBalanceMutex.RLock()
	defer rm.accountBalanceMutex.RUnlock()
	return rm.accountBalance
}

// DAILY LOSS TRACKING

// updates daily loss with a realized loss
func (rm *Manager) LogTradeLoss(symbol string, loss float64) {
	rm.accountBalanceMutex.Lock()
	defer rm.accountBalanceMutex.Unlock()

	if loss > 0 {
		rm.CurrentDailyLossAmount += loss
		lossPercent := (rm.CurrentDailyLossAmount / rm.accountBalance) * 100

		log.Printf("📉 Trade loss logged: $%.2f. Daily loss: $%.2f (%.2f%%)\n",
			loss, rm.CurrentDailyLossAmount, lossPercent)

		// check if daily loss limit hit
		if lossPercent >= rm.MaxDailyLossPercent {
			rm.recordRiskEvent(&Event{
				Timestamp:           time.Now(),
				EventType:           "MAX_DAILY_LOSS_HIT",
				Severity:            "CRITICAL",
				Symbol:              symbol,
				Details:             fmt.Sprintf("Daily loss %.2f%% hit maximum of %.2f%%", lossPercent, rm.MaxDailyLossPercent),
				CurrentAccountValue: rm.accountBalance,
				CurrentDailyLoss:    rm.CurrentDailyLossAmount,
			})

			rm.SendAlert(&Alert{
				Level:   "CRITICAL",
				Title:   "DAILY LOSS LIMIT HIT",
				Message: fmt.Sprintf("Daily loss has reached %.2f%% (%.2f%% limit). Auto-closing %s to prevent further losses.", lossPercent, rm.MaxDailyLossPercent, symbol),
				Symbol:  symbol,
				Data: map[string]interface{}{
					"dailyLoss": rm.CurrentDailyLossAmount,
					"limit":     rm.accountBalance * (rm.MaxDailyLossPercent / 100.0),
				},
			})

			// Auto-close the losing position
			go rm.ClosePositionBySymbol(symbol)
		}
	}
}

func (rm *Manager) GetDailyLossPercent() float64 {
	rm.accountBalanceMutex.RLock()
	defer rm.accountBalanceMutex.RUnlock()

	if rm.accountBalance == 0 {
		return 0
	}
	return (rm.CurrentDailyLossAmount / rm.accountBalance) * 100
}

func (rm *Manager) IsDailyLossLimitHit() bool {
	return rm.GetDailyLossPercent() >= rm.MaxDailyLossPercent
}

// closes position if risk is hit
func (rm *Manager) ClosePositionBySymbol(symbol string) error {
	if rm.client == nil {
		return fmt.Errorf("alpaca client not initialized")
	}

	log.Printf("AUTO-CLOSING %s - Daily loss limit hit\n", symbol)

	_, err := rm.client.ClosePosition(symbol, alpaca.ClosePositionRequest{})
	if err != nil {
		log.Printf("Failed to auto-close %s: %v\n", symbol, err)
		return err
	}

	log.Printf("Position %s closed automatically\n", symbol)
	return nil
}

// PORTFOLIO RISK ASSESSMENT

// calculates total portfolio risk across all open positions
func (rm *Manager) CalculatePortfolioRisk(positions []*position.OpenPosition) PortfolioRisk {
	risk := PortfolioRisk{
		TotalRiskAmount:  0,
		TotalRiskPercent: 0,
		PositionRisks:    []PositionRisk{},
		MaxAllowedRisk:   rm.GetAccountBalance() * (rm.MaxPortfolioRiskPercent / 100.0),
		IsOverRisk:       false,
	}

	for _, pos := range positions {
		riskPerShare := pos.EntryPrice - pos.StopLossPrice
		positionRisk := float64(pos.Quantity) * riskPerShare

		risk.PositionRisks = append(risk.PositionRisks, PositionRisk{
			Symbol:      pos.Symbol,
			RiskAmount:  positionRisk,
			RiskPercent: (positionRisk / rm.GetAccountBalance()) * 100,
		})

		risk.TotalRiskAmount += positionRisk
	}

	risk.TotalRiskPercent = (risk.TotalRiskAmount / rm.GetAccountBalance()) * 100
	risk.IsOverRisk = risk.TotalRiskAmount > risk.MaxAllowedRisk

	if risk.IsOverRisk {
		rm.recordRiskEvent(&Event{
			Timestamp:           time.Now(),
			EventType:           "PORTFOLIO_RISK_EXCEEDED",
			Severity:            "WARNING",
			Details:             fmt.Sprintf("Portfolio risk %.2f%% exceeds max %.2f%%", risk.TotalRiskPercent, rm.MaxPortfolioRiskPercent),
			CurrentAccountValue: rm.GetAccountBalance(),
		})
	}

	return risk
}

// RISK EVENTS & ALERTS

func (rm *Manager) recordRiskEvent(event *Event) {
	rm.riskEventsMutex.Lock()
	defer rm.riskEventsMutex.Unlock()
	rm.riskEvents = append(rm.riskEvents, event)
	log.Printf("Risk Event: [%s] %s - %s\n", event.Severity, event.EventType, event.Details)
}

func (rm *Manager) GetRiskEvents(limit int) []*Event {
	rm.riskEventsMutex.RLock()
	defer rm.riskEventsMutex.RUnlock()

	events := rm.riskEvents
	if len(events) > limit {
		events = events[len(events)-limit:]
	}
	return events
}

// PrintRiskEvents displays recent risk events in a formatted table
func (rm *Manager) PrintRiskEvents(limit int) {
	events := rm.GetRiskEvents(limit)

	if len(events) == 0 {
		fmt.Println("✅ No risk events recorded")
		return
	}

	width := 120
	fmt.Println("\n" + formatting.Separator(width))
	fmt.Println("🚨 RECENT RISK EVENTS")
	fmt.Println(formatting.Separator(width))
	fmt.Printf("%-10s %-25s %-10s %-40s %-20s\n",
		"Severity", "Event Type", "Symbol", "Details", "Timestamp")
	fmt.Println(formatting.Separator(width))

	for _, event := range events {
		// Truncate details if too long
		details := event.Details
		if len(details) > 40 {
			details = details[:37] + "..."
		}

		fmt.Printf("%-10s %-25s %-10s %-40s %s\n",
			event.Severity, event.EventType, event.Symbol, details,
			event.Timestamp.Format("2006-01-02 15:04:05"))
	}
	fmt.Println(formatting.Separator(width) + "\n")
}

func (rm *Manager) SendAlert(alert *Alert) {
	alert.Timestamp = time.Now()

	rm.alertCallbacksMutex.RLock()
	callbacks := rm.alertCallbacks
	rm.alertCallbacksMutex.RUnlock()

	for _, callback := range callbacks {
		go callback(alert)
	}
}

// RISK REPORT & MONITORING

// generates a comprehensive risk report
func (rm *Manager) GenerateRiskReport(positions []*position.OpenPosition) Report {
	accountBalance := rm.GetAccountBalance()
	dailyLossPercent := rm.GetDailyLossPercent()
	portfolioRisk := rm.CalculatePortfolioRisk(positions)

	rm.accountBalanceMutex.RLock()
	dailyLoss := rm.CurrentDailyLossAmount
	rm.accountBalanceMutex.RUnlock()

	report := Report{
		Timestamp:           time.Now(),
		AccountBalance:      accountBalance,
		OpenPositions:       len(positions),
		DailyLoss:           dailyLoss,
		DailyLossPercent:    dailyLossPercent,
		MaxDailyLossPercent: rm.MaxDailyLossPercent,
		DailyLossRemaining:  (rm.MaxDailyLossPercent - dailyLossPercent),
		PortfolioRisk:       portfolioRisk,
		HealthStatus:        "HEALTHY",
		Alerts:              []string{},
		RecentEvents:        rm.GetRiskEvents(5),
	}

	if dailyLossPercent >= rm.MaxDailyLossPercent {
		report.HealthStatus = "CRITICAL - DAILY LOSS LIMIT HIT"
		report.Alerts = append(report.Alerts, " Daily loss limit reached. No new trades.")
	} else if dailyLossPercent >= rm.MaxDailyLossPercent*0.75 {
		report.HealthStatus = "WARNING"
		report.Alerts = append(report.Alerts, fmt.Sprintf("  Daily loss at %.1f%% of limit", dailyLossPercent/rm.MaxDailyLossPercent*100))
	}

	if portfolioRisk.IsOverRisk {
		report.Alerts = append(report.Alerts, fmt.Sprintf("  Portfolio risk at %.2f%% (max %.2f%%)", portfolioRisk.TotalRiskPercent, rm.MaxPortfolioRiskPercent))
	}

	if len(positions) >= rm.MaxOpenPositions {
		report.Alerts = append(report.Alerts, fmt.Sprintf("  Max open positions (%d/%d) reached", len(positions), rm.MaxOpenPositions))
	}

	return report
}

// TYPES & STRUCTS

type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
	Details  map[string]interface{}
}

type PortfolioRisk struct {
	TotalRiskAmount  float64
	TotalRiskPercent float64
	PositionRisks    []PositionRisk
	MaxAllowedRisk   float64
	IsOverRisk       bool
}

type PositionRisk struct {
	Symbol      string
	RiskAmount  float64
	RiskPercent float64
}

type Report struct {
	Timestamp           time.Time
	AccountBalance      float64
	OpenPositions       int
	DailyLoss           float64
	DailyLossPercent    float64
	MaxDailyLossPercent float64
	DailyLossRemaining  float64
	PortfolioRisk       PortfolioRisk
	HealthStatus        string
	Alerts              []string
	RecentEvents        []*Event
}

// prints a formatted risk report
func (r *Report) Print() {
	width := 70
	fmt.Println("\n" + formatting.Separator(width))
	fmt.Println("📊 PORTFOLIO RISK REPORT")
	fmt.Println(formatting.Separator(width))
	fmt.Printf("Account Balance:       $%.2f\n", r.AccountBalance)
	fmt.Printf("Open Positions:        %d/%d\n", r.OpenPositions, 5)
	fmt.Printf("Daily Loss:            $%.2f (%.2f%% of %.2f%% limit)\n",
		r.DailyLoss, r.DailyLossPercent, r.MaxDailyLossPercent)
	fmt.Printf("Portfolio Risk:        $%.2f (%.2f%% of max 10%%)\n",
		r.PortfolioRisk.TotalRiskAmount, r.PortfolioRisk.TotalRiskPercent)
	fmt.Printf("Status:                %s\n", r.HealthStatus)

	if len(r.Alerts) > 0 {
		fmt.Println("\nAlerts:")
		for _, alert := range r.Alerts {
			fmt.Printf("  %s\n", alert)
		}
	}

	if len(r.RecentEvents) > 0 {
		fmt.Println("\nRecent Risk Events (Last 5):")
		for _, event := range r.RecentEvents {
			fmt.Printf("  [%s] %s - %s (%s)\n",
				event.Severity, event.EventType, event.Details, event.Timestamp.Format("15:04:05"))
		}
	}
	fmt.Println(formatting.Separator(width) + "\n")
}

// GetRecentEvents returns recent risk events for monitoring
func (rm *Manager) GetRecentEvents() []string {
	rm.riskEventsMutex.RLock()
	defer rm.riskEventsMutex.RUnlock()

	// Get events from last 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	var recentEvents []string

	for _, event := range rm.riskEvents {
		if event.Timestamp.After(cutoff) {
			recentEvents = append(recentEvents, fmt.Sprintf("[%s] %s: %s - %s",
				event.Timestamp.Format("15:04:05"),
				event.Severity,
				event.EventType,
				event.Details))
		}
	}

	return recentEvents
}
