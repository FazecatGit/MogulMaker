package risk

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/fazecat/mongelmaker/Internal/strategy/position"
	"github.com/fazecat/mongelmaker/Internal/utils/formatting"
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
	Severity            string // "CRITICAL", "WARNING", "INFO"
	Symbol              string
	Details             string
	CurrentAccountValue float64
	CurrentDailyLoss    float64
}

// callback function for risk alerts
type AlertCallback func(*Alert)

// alert that can be sent to users/handlers
type Alert struct {
	Level     string // "INFO", "WARNING", "CRITICAL"
	Title     string
	Message   string
	Timestamp time.Time
	Symbol    string
	Data      map[string]interface{}
}

// creates a new risk manager with default limits
func NewManager(client *alpaca.Client, accountBalance float64) *Manager {
	return &Manager{
		MaxDailyLossPercent:     2.0,                   // -2% daily loss limit
		MaxDailyLossAmount:      accountBalance * 0.02, // Calculate dollar amount
		CurrentDailyLossAmount:  0,
		DailyLossResetTime:      time.Now(),
		MaxOpenPositions:        5,
		MaxPositionSizePercent:  20.0, // 20% per trade
		MaxPortfolioRiskPercent: 10.0, // 10% total risk
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

// ============================================================================
// ACCOUNT BALANCE MANAGEMENT
// ============================================================================

// updates the account balance and resets daily loss if new trading day
func (rm *Manager) UpdateAccountBalance(newBalance float64) {
	rm.accountBalanceMutex.Lock()
	defer rm.accountBalanceMutex.Unlock()

	rm.accountBalance = newBalance

	// Reset daily loss if it's a new trading day (9:30 AM ET)
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

// ============================================================================
// POSITION SIZING & VALIDATION
// ============================================================================

// validates if a position size is within risk limits
func (rm *Manager) ValidatePositionSize(symbol string, quantity int64, entryPrice float64) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Warnings: []string{},
		Errors:   []string{},
		Details:  map[string]interface{}{},
	}

	positionValue := float64(quantity) * entryPrice
	accountBalance := rm.GetAccountBalance()
	positionSizePercent := (positionValue / accountBalance) * 100

	result.Details["positionValue"] = positionValue
	result.Details["accountBalance"] = accountBalance
	result.Details["positionSizePercent"] = positionSizePercent

	// Check 1: Max position size (20% of account)
	if positionSizePercent > rm.MaxPositionSizePercent {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"Position size %.1f%% exceeds maximum %.1f%% of account",
			positionSizePercent, rm.MaxPositionSizePercent))
	}

	// Check 2: Warning if >15%
	if positionSizePercent > 15 {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"Position size %.1f%% is approaching max limit (20%%)",
			positionSizePercent))
	}

	return result
}

// calculates safe position size based on account and risk parameters
func (rm *Manager) CalculateSafePositionSize(
	symbol string,
	entryPrice float64,
	stopLossPrice float64,
	maxRiskPercent float64) int64 {

	accountBalance := rm.GetAccountBalance()
	riskAmount := accountBalance * (maxRiskPercent / 100.0)
	priceRisk := entryPrice - stopLossPrice

	if priceRisk <= 0 {
		log.Printf("⚠️  Invalid stop loss: %.2f >= entry: %.2f", stopLossPrice, entryPrice)
		return 0
	}

	quantity := int64(riskAmount / priceRisk)

	// Cap at max position size (20% of account)
	maxPositionValue := accountBalance * (rm.MaxPositionSizePercent / 100.0)
	maxQuantity := int64(maxPositionValue / entryPrice)

	if quantity > maxQuantity {
		quantity = maxQuantity
		log.Printf("⚠️  Position size capped at max portfolio allocation")
	}

	return quantity
}

// ============================================================================
// DAILY LOSS TRACKING
// ============================================================================

// updates daily loss with a realized loss
func (rm *Manager) LogTradeLoss(symbol string, loss float64) {
	rm.accountBalanceMutex.Lock()
	defer rm.accountBalanceMutex.Unlock()

	if loss > 0 {
		rm.CurrentDailyLossAmount += loss
		lossPercent := (rm.CurrentDailyLossAmount / rm.accountBalance) * 100

		log.Printf("📉 Trade loss logged: $%.2f. Daily loss: $%.2f (%.2f%%)\n",
			loss, rm.CurrentDailyLossAmount, lossPercent)

		// Check if daily loss limit hit
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
				Title:   "⛔ DAILY LOSS LIMIT HIT",
				Message: fmt.Sprintf("Daily loss has reached %.2f%% (%.2f%% limit). Trading halted.", lossPercent, rm.MaxDailyLossPercent),
				Symbol:  symbol,
				Data: map[string]interface{}{
					"dailyLoss": rm.CurrentDailyLossAmount,
					"limit":     rm.accountBalance * (rm.MaxDailyLossPercent / 100.0),
				},
			})
		}
	}
}

// returns current daily loss
func (rm *Manager) GetDailyLoss() float64 {
	rm.accountBalanceMutex.RLock()
	defer rm.accountBalanceMutex.RUnlock()
	return rm.CurrentDailyLossAmount
}

// returns daily loss as percentage
func (rm *Manager) GetDailyLossPercent() float64 {
	rm.accountBalanceMutex.RLock()
	defer rm.accountBalanceMutex.RUnlock()

	if rm.accountBalance == 0 {
		return 0
	}
	return (rm.CurrentDailyLossAmount / rm.accountBalance) * 100
}

// checks if daily loss limit has been hit
func (rm *Manager) IsDailyLossLimitHit() bool {
	return rm.GetDailyLossPercent() >= rm.MaxDailyLossPercent
}

// ============================================================================
// OPEN POSITIONS TRACKING
// ============================================================================

// records a new position entry
func (rm *Manager) AddPosition(symbol string, sector string) {
	rm.positionsMutex.Lock()
	defer rm.positionsMutex.Unlock()

	rm.PositionsBySymbol[symbol]++
	rm.PositionsBySector[sector]++
}

// records a position closure
func (rm *Manager) RemovePosition(symbol string, sector string) {
	rm.positionsMutex.Lock()
	defer rm.positionsMutex.Unlock()

	if count, exists := rm.PositionsBySymbol[symbol]; exists && count > 0 {
		rm.PositionsBySymbol[symbol]--
	}
	if count, exists := rm.PositionsBySector[sector]; exists && count > 0 {
		rm.PositionsBySector[sector]--
	}
}

// returns count of open positions
func (rm *Manager) CountOpenPositions() int {
	rm.positionsMutex.RLock()
	defer rm.positionsMutex.RUnlock()

	count := 0
	for _, c := range rm.PositionsBySymbol {
		count += c
	}
	return count
}

// validates if a new position can be opened
func (rm *Manager) CanOpenPosition(symbol string, sector string) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Warnings: []string{},
		Errors:   []string{},
		Details:  map[string]interface{}{},
	}

	rm.positionsMutex.RLock()
	openPositions := len(rm.PositionsBySymbol)
	sectorCount := rm.PositionsBySector[sector]
	rm.positionsMutex.RUnlock()

	result.Details["openPositions"] = openPositions
	result.Details["sectorCount"] = sectorCount
	result.Details["maxOpenPositions"] = rm.MaxOpenPositions

	// Check 1: Max open positions
	if openPositions >= rm.MaxOpenPositions {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"Cannot open new position: %d/%d max open positions reached",
			openPositions, rm.MaxOpenPositions))
		return result
	}

	// Check 2: Sector diversification
	if sectorCount >= rm.MaxSameSectorPositions {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"Cannot open new position: %d/%d positions in %s sector",
			sectorCount, rm.MaxSameSectorPositions, sector))
	}

	// Check 3: Daily loss limit
	if rm.IsDailyLossLimitHit() {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"Cannot open new position: Daily loss limit (%.2f%%) hit",
			rm.MaxDailyLossPercent))
	}

	return result
}

// ============================================================================
// PORTFOLIO RISK ASSESSMENT
// ============================================================================

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
		// Calculate position risk: quantity * (entry - stop loss)
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

// ============================================================================
// RISK EVENTS & ALERTS
// ============================================================================

// records a risk event
func (rm *Manager) recordRiskEvent(event *Event) {
	rm.riskEventsMutex.Lock()
	defer rm.riskEventsMutex.Unlock()
	rm.riskEvents = append(rm.riskEvents, event)
	log.Printf("🚨 Risk Event: [%s] %s - %s\n", event.Severity, event.EventType, event.Details)
}

// returns recent risk events
func (rm *Manager) GetRiskEvents(limit int) []*Event {
	rm.riskEventsMutex.RLock()
	defer rm.riskEventsMutex.RUnlock()

	events := rm.riskEvents
	if len(events) > limit {
		events = events[len(events)-limit:]
	}
	return events
}

// registers a callback for alerts
func (rm *Manager) RegisterAlertCallback(callback AlertCallback) {
	rm.alertCallbacksMutex.Lock()
	defer rm.alertCallbacksMutex.Unlock()
	rm.alertCallbacks = append(rm.alertCallbacks, callback)
}

// SendAlert sends an alert to all registered callbacks
func (rm *Manager) SendAlert(alert *Alert) {
	alert.Timestamp = time.Now()

	rm.alertCallbacksMutex.RLock()
	callbacks := rm.alertCallbacks
	rm.alertCallbacksMutex.RUnlock()

	for _, callback := range callbacks {
		go callback(alert) // Non-blocking
	}
}

// ============================================================================
// RISK REPORT & MONITORING
// ============================================================================

// generates a comprehensive risk report
func (rm *Manager) GenerateRiskReport(positions []*position.OpenPosition) Report {
	accountBalance := rm.GetAccountBalance()
	dailyLossPercent := rm.GetDailyLossPercent()
	portfolioRisk := rm.CalculatePortfolioRisk(positions)
	openPosCount := rm.CountOpenPositions()

	report := Report{
		Timestamp:           time.Now(),
		AccountBalance:      accountBalance,
		OpenPositions:       openPosCount,
		DailyLoss:           rm.GetDailyLoss(),
		DailyLossPercent:    dailyLossPercent,
		MaxDailyLossPercent: rm.MaxDailyLossPercent,
		DailyLossRemaining:  (rm.MaxDailyLossPercent - dailyLossPercent),
		PortfolioRisk:       portfolioRisk,
		HealthStatus:        "HEALTHY",
		Alerts:              []string{},
	}

	// Determine health status
	if dailyLossPercent >= rm.MaxDailyLossPercent {
		report.HealthStatus = "CRITICAL - DAILY LOSS LIMIT HIT"
		report.Alerts = append(report.Alerts, "🛑 Daily loss limit reached. No new trades.")
	} else if dailyLossPercent >= rm.MaxDailyLossPercent*0.75 {
		report.HealthStatus = "WARNING"
		report.Alerts = append(report.Alerts, fmt.Sprintf("⚠️  Daily loss at %.1f%% of limit", dailyLossPercent/rm.MaxDailyLossPercent*100))
	}

	if portfolioRisk.IsOverRisk {
		report.Alerts = append(report.Alerts, fmt.Sprintf("⚠️  Portfolio risk at %.2f%% (max %.2f%%)", portfolioRisk.TotalRiskPercent, rm.MaxPortfolioRiskPercent))
	}

	if openPosCount >= rm.MaxOpenPositions {
		report.Alerts = append(report.Alerts, fmt.Sprintf("⚠️  Max open positions (%d/%d) reached", openPosCount, rm.MaxOpenPositions))
	}

	return report
}

// ============================================================================
// TYPES & STRUCTS
// ============================================================================

// validation result for position/order checks
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
	Details  map[string]interface{}
}

// portfolio-wide risk assessment
type PortfolioRisk struct {
	TotalRiskAmount  float64
	TotalRiskPercent float64
	PositionRisks    []PositionRisk
	MaxAllowedRisk   float64
	IsOverRisk       bool
}

// individual position risk
type PositionRisk struct {
	Symbol      string
	RiskAmount  float64
	RiskPercent float64
}

// comprehensive risk report
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
	fmt.Println(formatting.Separator(width) + "\n")
}
