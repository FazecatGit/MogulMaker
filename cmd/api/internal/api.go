package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	datafeed "github.com/fazecat/mogulmaker/Internal/database"
	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	"github.com/fazecat/mogulmaker/Internal/handlers/monitoring"
	"github.com/fazecat/mogulmaker/Internal/handlers/risk"
	"github.com/fazecat/mogulmaker/Internal/strategy/detection"
	"github.com/fazecat/mogulmaker/Internal/strategy/indicators"
	"github.com/fazecat/mogulmaker/Internal/strategy/metrics"
	"github.com/fazecat/mogulmaker/Internal/strategy/position"
	"github.com/fazecat/mogulmaker/Internal/utils/analyzer"
	"github.com/fazecat/mogulmaker/Internal/utils/config"
	"github.com/fazecat/mogulmaker/Internal/utils/scanner"
	"github.com/fazecat/mogulmaker/Internal/utils/scoring"
	"github.com/shopspring/decimal"
)

type API struct {
	PositionManager *position.PositionManager
	RiskManager     *risk.Manager
	Queries         *database.Queries
	TradeMonitor    *monitoring.Monitor
	AlpacaClient    *alpaca.Client
	JWTManager      *JWTManager
	backtestCache   map[string]map[string]interface{} // backtestID -> results
	backtestMutex   sync.RWMutex
}

func (api *API) HandleGetPositions(w http.ResponseWriter, r *http.Request) {
	alpacaPositions, err := api.AlpacaClient.GetPositions()
	if err != nil {
		log.Printf("Error fetching positions from Alpaca: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch positions")
		return
	}

	pendingOrders, err := api.AlpacaClient.GetOrders(alpaca.GetOrdersRequest{
		Status: "open",
		Limit:  100,
		Nested: true,
	})
	if err != nil {
		log.Printf("Warning: Could not fetch pending orders: %v", err)
		pendingOrders = []alpaca.Order{}
	} else {
		log.Printf("Found %d pending orders", len(pendingOrders))
	}

	response := map[string]interface{}{
		"count":          len(alpacaPositions),
		"positions":      alpacaPositions,
		"pending_orders": pendingOrders,
		"timestamp":      time.Now().Unix(),
		"risk_status": map[string]interface{}{
			"enabled": api.RiskManager != nil,
		},
	}
	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleGetRiskStatus(w http.ResponseWriter, r *http.Request) {
	if api.RiskManager == nil {
		WriteError(w, http.StatusInternalServerError, "Risk manager not initialized")
		return
	}

	// Get Alpaca account info
	account, err := api.AlpacaClient.GetAccount()
	if err != nil {
		log.Printf("Error fetching account: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch account data")
		return
	}

	// Get open positions from Alpaca
	alpacaPositions, err := api.AlpacaClient.GetPositions()
	if err != nil {
		log.Printf("Error fetching positions: %v", err)
		alpacaPositions = []alpaca.Position{}
	}

	// Calculate portfolio risk metrics
	accountBalance := decimal.NewFromFloat(account.Cash.InexactFloat64() + account.PortfolioValue.InexactFloat64()).InexactFloat64()
	portfolioValue := account.PortfolioValue.InexactFloat64()
	buyingPower := account.BuyingPower.InexactFloat64()
	dayTradingBuyingPower := account.DaytradingBuyingPower.InexactFloat64()

	// Calculate total unrealized P&L
	totalUnrealizedPnL := 0.0
	for _, pos := range alpacaPositions {
		totalUnrealizedPnL += pos.UnrealizedPL.InexactFloat64()
	}

	// Get daily loss from risk manager
	dailyLoss := api.RiskManager.GetDailyLossPercent()
	isDailyLimitHit := api.RiskManager.IsDailyLossLimitHit()

	// Get account balance from risk manager for portfolio risk calc
	accountBal := api.RiskManager.GetAccountBalance()
	if accountBal == 0 {
		accountBal = accountBalance
	}

	portfolioRisk := (totalUnrealizedPnL / accountBalance) * 100
	if portfolioRisk < 0 {
		portfolioRisk = -portfolioRisk
	}

	// Determine status based on risk levels
	status := "HEALTHY"
	if isDailyLimitHit || portfolioRisk > 10.0 {
		status = "CRITICAL"
	} else if portfolioRisk > 7.0 {
		status = "WARNING"
	}

	// Build position details
	positionCount := len(alpacaPositions)
	positionLimit := 5
	if positionCount > positionLimit {
		positionLimit = positionCount // Allow flexibility but flag alert
	}

	// Format positions for response
	var positions []map[string]interface{}
	for _, pos := range alpacaPositions {
		qty, _ := pos.Qty.Float64()
		costBasis, _ := pos.CostBasis.Float64()
		avgFillPrice := 0.0
		if qty > 0 {
			avgFillPrice = costBasis / qty
		}

		posDetail := map[string]interface{}{
			"symbol":          pos.Symbol,
			"side":            pos.Side,
			"qty":             qty,
			"avg_fill_price":  avgFillPrice,
			"current_price":   pos.CurrentPrice.InexactFloat64(),
			"unrealized_pl":   pos.UnrealizedPL.InexactFloat64(),
			"unrealized_plpc": pos.UnrealizedPLPC.InexactFloat64(),
			"change_today":    pos.ChangeToday.InexactFloat64(),
		}
		positions = append(positions, posDetail)
	}

	riskStatus := map[string]interface{}{
		"enabled":              true,
		"account_balance":      accountBalance,
		"portfolio_value":      portfolioValue,
		"buying_power":         buyingPower,
		"day_trading_bp":       dayTradingBuyingPower,
		"daily_loss_percent":   dailyLoss,
		"daily_loss_limit_hit": isDailyLimitHit,
		"total_unrealized_pnl": totalUnrealizedPnL,
		"portfolio_risk_pct":   portfolioRisk,
		"status":               status,
		"open_positions":       positionCount,
		"position_limit":       positionLimit,
		"positions":            positions,
		"timestamp":            time.Now().Unix(),
	}

	WriteJSON(w, http.StatusOK, riskStatus)
}

func (api *API) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	dbTrades, err := api.Queries.GetAllTrades(r.Context())
	if err != nil {
		log.Printf("Error fetching trades: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch trades")
		return
	}

	totalTrades := len(dbTrades)

	trades := convertToTradeResults(dbTrades)
	completedTrades := len(trades)

	sharpe := 0.0
	sortino := 0.0
	winRate := 0.0
	totalPnL := 0.0

	if len(trades) > 0 {
		sharpe = metrics.CalculateSharpeRatio(trades, 0.02)
		sortino = metrics.CalculateSortinoRatio(trades, 0.02)
		winRate = metrics.CalculateWinRate(trades)

		for _, trade := range trades {
			totalPnL += trade.PnL
		}
	}

	response := map[string]interface{}{
		"total_trades":     totalTrades,
		"completed_trades": completedTrades,
		"total_pnl":        totalPnL,
		"sharpe_ratio":     sharpe,
		"sortino_ratio":    sortino,
		"win_rate":         winRate,
	}

	WriteJSON(w, http.StatusOK, response)
}

func convertToTradeResults(dbTrades []database.GetAllTradesRow) []metrics.TradeResult {
	var results []metrics.TradeResult

	tradesBySymbol := make(map[string][]database.GetAllTradesRow)
	for _, trade := range dbTrades {
		symbol := trade.Symbol
		tradesBySymbol[symbol] = append(tradesBySymbol[symbol], trade)
	}

	for symbol, trades := range tradesBySymbol {
		var buyTrades []database.GetAllTradesRow

		for _, trade := range trades {
			if trade.Side == "" || trade.Price == "" || trade.Quantity == "" {
				continue
			}

			side := trade.Side

			if side == "buy" {
				buyTrades = append(buyTrades, trade)
			} else if side == "sell" && len(buyTrades) > 0 {

				buyTrade := buyTrades[0]
				buyTrades = buyTrades[1:]

				buyPrice, _ := strconv.ParseFloat(buyTrade.Price, 64)
				sellPrice, _ := strconv.ParseFloat(trade.Price, 64)
				qty, _ := strconv.ParseFloat(trade.Quantity, 64)

				pnl := (sellPrice - buyPrice) * qty
				returnPercent := ((sellPrice - buyPrice) / buyPrice) * 100

				var duration time.Duration
				if buyTrade.CreatedAt.Valid && trade.CreatedAt.Valid {
					duration = trade.CreatedAt.Time.Sub(buyTrade.CreatedAt.Time)
				}

				result := metrics.TradeResult{
					Symbol:        symbol,
					EntryPrice:    buyPrice,
					ExitPrice:     sellPrice,
					Quantity:      qty,
					PnL:           pnl,
					ReturnPercent: returnPercent,
					Duration:      duration,
					EntryTime:     buyTrade.CreatedAt.Time,
					ExitTime:      trade.CreatedAt.Time,
				}

				results = append(results, result)
			}
		}
	}

	return results
}

func (api *API) HandleGetTrades(w http.ResponseWriter, r *http.Request) {
	_ = r.URL.Query().Get("symbol") // Symbol filter available if needed
	limitStr := r.URL.Query().Get("limit")
	statusFilter := r.URL.Query().Get("status") // all, open, closed

	limit := 100
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get all orders from Alpaca (includes full trading history)
	orders, err := api.AlpacaClient.GetOrders(alpaca.GetOrdersRequest{
		Status: "all",          // Get all orders: open, closed, etc.
		Limit:  int(limit * 2), // Get more to account for filtering
		Nested: true,
	})
	if err != nil {
		log.Printf("Error fetching Alpaca orders: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	// Group filled orders by symbol to pair buy/sell
	filledBySymbol := make(map[string][]alpaca.Order)
	allOrders := []alpaca.Order{}

	for _, order := range orders {
		if order.Status == "filled" || order.Status == "closed" {
			filledBySymbol[order.Symbol] = append(filledBySymbol[order.Symbol], order)
		}
		allOrders = append(allOrders, order)
	}

	// Create trade records with P&L calculations
	// Pair trades and calculate P&L using the monitoring package
	tradeRecords := monitoring.PairTradesAndCalculatePnL(allOrders)
	trades := monitoring.FormatTradeRecordsAsJSON(tradeRecords)

	// Filter by status if provided
	if statusFilter != "" && statusFilter != "all" {
		var filtered []map[string]interface{}
		for _, trade := range trades {
			if tradeStatus, ok := trade["status"].(string); ok {
				if tradeStatus == statusFilter {
					filtered = append(filtered, trade)
				}
			}
		}
		trades = filtered
	}

	// Sort by submitted_at descending
	sort.Slice(trades, func(i, j int) bool {
		iTime, _ := trades[i]["submitted_at"].(string)
		jTime, _ := trades[j]["submitted_at"].(string)
		return iTime > jTime
	})

	// Limit results
	if len(trades) > limit {
		trades = trades[:limit]
	}

	response := map[string]interface{}{
		"count":       len(trades),
		"trades":      trades,
		"timestamp":   time.Now().Unix(),
		"risk_status": map[string]interface{}{"enabled": true},
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleTradeStatistics(w http.ResponseWriter, r *http.Request) {
	// Get all orders from Alpaca
	orders, err := api.AlpacaClient.GetOrders(alpaca.GetOrdersRequest{
		Status: "all",
		Limit:  1000, // Get more orders for better statistics
		Nested: true,
	})
	if err != nil {
		log.Printf("Error fetching Alpaca orders: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	// Group orders by symbol and pair buy/sell to calculate P&L
	tradesBySymbol := make(map[string][]alpaca.Order)
	for _, order := range orders {
		// Only use filled orders
		if order.Status == "filled" || order.Status == "closed" {
			tradesBySymbol[order.Symbol] = append(tradesBySymbol[order.Symbol], order)
		}
	}

	// Calculate P&L by pairing buy/sell trades
	var pnlResults []float64
	var completedTrades []map[string]interface{}
	totalPnL := 0.0
	largestWin := 0.0
	largestLoss := 0.0

	for symbol, trades := range tradesBySymbol {
		// Separate buys and sells
		var buyTrades []alpaca.Order
		var sellTrades []alpaca.Order

		for _, trade := range trades {
			if trade.Side == "buy" {
				buyTrades = append(buyTrades, trade)
			} else {
				sellTrades = append(sellTrades, trade)
			}
		}

		// Pair buys with sells
		minPairs := len(buyTrades)
		if len(sellTrades) < minPairs {
			minPairs = len(sellTrades)
		}

		for i := 0; i < minPairs; i++ {
			buyOrder := buyTrades[i]
			sellOrder := sellTrades[i]

			buyQty, _ := buyOrder.FilledQty.Float64()
			buyPrice, _ := buyOrder.FilledAvgPrice.Float64()
			sellQty, _ := sellOrder.FilledQty.Float64()
			sellPrice, _ := sellOrder.FilledAvgPrice.Float64()

			// Use minimum quantity to pair
			qty := buyQty
			if sellQty < qty {
				qty = sellQty
			}

			pnl := (sellPrice - buyPrice) * qty
			pnlResults = append(pnlResults, pnl)
			totalPnL += pnl

			if pnl > largestWin {
				largestWin = pnl
			}
			if pnl < largestLoss {
				largestLoss = pnl
			}

			completedTrades = append(completedTrades, map[string]interface{}{
				"symbol": symbol,
				"pnl":    pnl,
			})
		}
	}

	// Calculate metrics
	totalFilled := len(orders)
	winningTrades := 0
	for _, pnl := range pnlResults {
		if pnl > 0 {
			winningTrades++
		}
	}
	losingTrades := len(pnlResults) - winningTrades

	winRate := 0.0
	avgPnL := 0.0
	if len(pnlResults) > 0 {
		winRate = (float64(winningTrades) / float64(len(pnlResults))) * 100
		avgPnL = totalPnL / float64(len(pnlResults))
	}

	// Calculate Sharpe ratio from PnL returns using metrics package
	sharpeRatio := metrics.CalculateSharpeFromReturns(pnlResults)
	sortinoRatio := metrics.CalculateSortinoFromReturns(pnlResults)

	// Get open positions for additional context
	openPositions, err := api.AlpacaClient.GetPositions()
	openCount := 0
	openPnL := 0.0
	if err == nil {
		openCount = len(openPositions)
		for _, pos := range openPositions {
			openPnL += pos.UnrealizedPL.InexactFloat64()
		}
	}

	response := map[string]interface{}{
		"total_trades":       totalFilled,
		"winning_trades":     winningTrades,
		"losing_trades":      losingTrades,
		"win_rate":           winRate,
		"total_pnl":          totalPnL,
		"avg_pnl":            avgPnL,
		"largest_win":        largestWin,
		"largest_loss":       largestLoss,
		"avg_trade_duration": "N/A",
		"sharpe_ratio":       sharpeRatio,
		"sortino_ratio":      sortinoRatio,
		"open_positions":     openCount,
		"open_pnl":           openPnL,
		"timestamp":          time.Now().Unix(),
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleSellAllTrades(w http.ResponseWriter, r *http.Request) {
	positions := api.PositionManager.GetOpenPositions()

	if len(positions) == 0 {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"message":      "No open positions to sell",
			"sold_count":   0,
			"failed_count": 0,
		})
		return
	}

	var soldSymbols []string
	var failedSymbols []map[string]interface{}

	for _, pos := range positions {
		_, err := api.AlpacaClient.ClosePosition(pos.Symbol, alpaca.ClosePositionRequest{})
		if err != nil {
			failedSymbols = append(failedSymbols, map[string]interface{}{
				"symbol": pos.Symbol,
				"error":  err.Error(),
			})
		} else {
			soldSymbols = append(soldSymbols, pos.Symbol)
		}
	}

	response := map[string]interface{}{
		"message":      "Sell all trades completed",
		"sold":         soldSymbols,
		"sold_count":   len(soldSymbols),
		"failed":       failedSymbols,
		"failed_count": len(failedSymbols),
		"total_count":  len(positions),
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleGetPositionBySymbol(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	if symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	position, err := api.AlpacaClient.GetPosition(symbol)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Position not found")
		return
	}

	WriteJSON(w, http.StatusOK, position)
}

func (api *API) HandleExecuteTrade(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol   string  `json:"symbol"`
		Side     string  `json:"side"`
		Quantity float64 `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required")
		return
	}
	if req.Side != "buy" && req.Side != "sell" {
		WriteError(w, http.StatusBadRequest, "Side must be 'buy' or 'sell'")
		return
	}
	if req.Quantity <= 0 {
		WriteError(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}

	side := alpaca.Buy
	if req.Side == "sell" {
		side = alpaca.Sell
	}

	qty := decimal.NewFromFloat(req.Quantity)
	order := alpaca.PlaceOrderRequest{
		Symbol:      req.Symbol,
		Qty:         &qty,
		Side:        side,
		Type:        alpaca.Market,
		TimeInForce: alpaca.Day,
	}

	placedOrder, err := api.AlpacaClient.PlaceOrder(order)
	if err != nil {
		log.Printf("Error placing order: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to execute trade")
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"order_id": placedOrder.ID,
		"symbol":   placedOrder.Symbol,
		"side":     placedOrder.Side,
		"quantity": placedOrder.Qty.String(),
		"status":   placedOrder.Status,
	}

	WriteJSON(w, http.StatusCreated, response)
}

func (api *API) HandleClosePosition(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	if symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	position, err := api.AlpacaClient.GetPosition(symbol)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Position not found")
		return
	}

	qty, ok := position.Qty.Float64()
	if !ok {
		log.Printf("Error converting quantity to float64")
		WriteError(w, http.StatusInternalServerError, "Failed to process position quantity")
		return
	}

	qtyDecimal := decimal.NewFromFloat(qty)
	order := alpaca.PlaceOrderRequest{
		Symbol:      symbol,
		Qty:         &qtyDecimal,
		Side:        alpaca.Sell,
		Type:        alpaca.Market,
		TimeInForce: alpaca.Day,
	}

	placedOrder, err := api.AlpacaClient.PlaceOrder(order)
	if err != nil {
		log.Printf("Error closing position: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to close position")
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  "Position closed",
		"order_id": placedOrder.ID,
		"symbol":   placedOrder.Symbol,
		"quantity": placedOrder.Qty.String(),
		"status":   placedOrder.Status,
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleGenerateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.UserID == "" {
		WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	token, err := api.JWTManager.GenerateToken(req.UserID, req.Email, 24)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := map[string]interface{}{
		"token":      token,
		"user_id":    req.UserID,
		"expires_at": time.Now().Add(24 * time.Hour).Unix(),
	}

	WriteJSON(w, http.StatusOK, response)
}

// 5.3
func (api *API) HandlePortfolioSummary(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")

	alpacaPositions, err := api.AlpacaClient.GetPositions()
	if err != nil {
		log.Printf("Error fetching positions from Alpaca: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch portfolio summary")
		return
	}

	var filteredPositions []alpaca.Position
	if symbol != "" {
		for _, pos := range alpacaPositions {
			if pos.Symbol == symbol {
				filteredPositions = append(filteredPositions, pos)
			}
		}
	} else {
		filteredPositions = alpacaPositions
	}

	// Calculate portfolio metrics
	var totalValue decimal.Decimal
	var totalCost decimal.Decimal
	var totalGain decimal.Decimal

	for _, pos := range filteredPositions {
		if pos.MarketValue != nil {
			totalValue = totalValue.Add(*pos.MarketValue)
		}
		totalCost = totalCost.Add(pos.CostBasis)

		if pos.UnrealizedPL != nil {
			totalGain = totalGain.Add(*pos.UnrealizedPL)
		}
	}

	response := map[string]interface{}{
		"total_positions": len(filteredPositions),
		"total_value":     totalValue.String(),
		"total_cost":      totalCost.String(),
		"total_gain":      totalGain.String(),
		"positions":       filteredPositions,
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleRiskAdjustments(w http.ResponseWriter, r *http.Request) {
	if api.RiskManager == nil {
		WriteError(w, http.StatusInternalServerError, "Risk manager not initialized")
		return
	}

	riskEvents := api.RiskManager.GetRiskEvents(50)

	response := map[string]interface{}{
		"account_balance":      api.RiskManager.GetAccountBalance(),
		"daily_loss_percent":   api.RiskManager.GetDailyLossPercent(),
		"daily_loss_limit_hit": api.RiskManager.IsDailyLossLimitHit(),
		"recent_events":        riskEvents,
	}

	WriteJSON(w, http.StatusOK, response)
}
func (api *API) HandlePerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	if api.TradeMonitor == nil {
		WriteError(w, http.StatusInternalServerError, "Trade monitor not initialized")
		return
	}

	monitors := api.TradeMonitor.GetPositionMonitors()

	response := map[string]interface{}{
		"monitors":  monitors,
		"timestamp": time.Now().Unix(),
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleRiskAlerts(w http.ResponseWriter, r *http.Request) {
	if api.RiskManager == nil {
		WriteError(w, http.StatusInternalServerError, "Risk manager not initialized")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	events := api.RiskManager.GetRiskEvents(limit)

	response := map[string]interface{}{
		"count":  len(events),
		"alerts": events,
	}

	WriteJSON(w, http.StatusOK, response)
}

//5.4

func (api *API) HandleBacktest(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required for backtest")
		return
	}

	openPositions := api.PositionManager.GetOpenPositions()
	for _, pos := range openPositions {
		if pos.Symbol == symbol {
			WriteError(w, http.StatusBadRequest, "Cannot run backtest on an open position")
			return
		}
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	capitalStr := r.URL.Query().Get("capital")

	if startDate == "" || endDate == "" {
		WriteError(w, http.StatusBadRequest, "start_date and end_date are required (YYYY-MM-DD)")
		return
	}

	// Parse capital amount
	capital := 100000.0
	if capitalStr != "" {
		if parsedCap, err := strconv.ParseFloat(capitalStr, 64); err == nil && parsedCap > 0 {
			capital = parsedCap
		}
	} else if api.RiskManager != nil {
		capital = api.RiskManager.GetAccountBalance()
	}

	// Fetch historical bars for the symbol using the date range
	historicalBars, err := datafeed.GetAlpacaBars(symbol, "1Day", 10000, startDate)
	if err != nil || len(historicalBars) == 0 {
		log.Printf("Error fetching historical bars: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch historical data for backtest")
		return
	}

	log.Printf("Fetched %d bars for backtest from %s to %s", len(historicalBars), startDate, endDate)

	// Run backtest with TradeResult from metrics.RunBacktest
	trades, err := metrics.RunBacktest(symbol, historicalBars, capital)
	if err != nil {
		log.Printf("Error running backtest: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to execute backtest")
		return
	}

	// Calculate metrics from trades
	winRate := metrics.CalculateWinRate(trades)
	sharpeRatio := metrics.CalculateSharpeRatio(trades, 0.02)
	sortinoRatio := metrics.CalculateSortinoRatio(trades, 0.02)

	winningTrades := 0
	totalPnL := 0.0
	largestWin := 0.0
	largestLoss := 0.0

	for _, trade := range trades {
		totalPnL += trade.PnL
		if trade.PnL > 0 {
			winningTrades++
			if trade.PnL > largestWin {
				largestWin = trade.PnL
			}
		} else if trade.PnL < 0 {
			if trade.PnL < largestLoss {
				largestLoss = trade.PnL
			}
		}
	}

	finalBalance := capital + totalPnL
	totalReturnPct := (totalPnL / capital) * 100
	losingTrades := len(trades) - winningTrades

	backtestID := symbol + "_" + time.Now().Format("20060102150405")

	response := map[string]interface{}{
		"backtest_id":      backtestID,
		"symbol":           symbol,
		"status":           "completed",
		"start_date":       startDate,
		"end_date":         endDate,
		"initial_capital":  capital,
		"final_balance":    finalBalance,
		"total_return_pct": totalReturnPct,
		"sharpe_ratio":     sharpeRatio,
		"sortino_ratio":    sortinoRatio,
		"win_rate":         winRate,
		"total_trades":     len(trades),
		"winning_trades":   winningTrades,
		"losing_trades":    losingTrades,
		"largest_win":      largestWin,
		"largest_loss":     largestLoss,
		"created_at":       time.Now().Unix(),
	}

	// Cache the backtest results
	api.backtestMutex.Lock()
	if api.backtestCache == nil {
		api.backtestCache = make(map[string]map[string]interface{})
	}
	api.backtestCache[backtestID] = response
	api.backtestMutex.Unlock()

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleBacktestResults(w http.ResponseWriter, r *http.Request) {
	backtestID := r.URL.Query().Get("id")
	if backtestID == "" {
		WriteError(w, http.StatusBadRequest, "Backtest ID is required")
		return
	}

	// Retrieve backtest results from cache using backtestID
	api.backtestMutex.RLock()
	results, exists := api.backtestCache[backtestID]
	api.backtestMutex.RUnlock()

	if !exists {
		WriteError(w, http.StatusNotFound, "Backtest results not found")
		return
	}

	WriteJSON(w, http.StatusOK, results)
}

func (api *API) HandleBacktestStatus(w http.ResponseWriter, r *http.Request) {
	backtestID := r.URL.Query().Get("id")
	if backtestID == "" {
		WriteError(w, http.StatusBadRequest, "Backtest ID is required")
		return
	}

	// Check backtest status from cache
	api.backtestMutex.RLock()
	results, exists := api.backtestCache[backtestID]
	api.backtestMutex.RUnlock()

	if !exists {
		WriteError(w, http.StatusNotFound, "Backtest not found")
		return
	}

	status := "completed"
	if resultsStatus, ok := results["status"].(string); ok {
		status = resultsStatus
	}

	response := map[string]interface{}{
		"backtest_id": backtestID,
		"status":      status,
		"progress":    100,
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleSymbolAnalysis(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	// Return basic analysis structure, needs to be implemented
	response := map[string]interface{}{
		"symbol":            symbol,
		"status":            "not_analyzed",
		"rsi_signals":       nil,
		"support_levels":    nil,
		"resistance_levels": nil,
		"trend":             "neutral",
		"message":           "Symbol analysis not implemented yet",
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleAnalysisReport(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required for analysis report")
		return
	}

	response := map[string]interface{}{
		"symbol":          symbol,
		"generated_at":    time.Now().Unix(),
		"analysis":        nil,
		"recommendations": nil,
		"message":         "Analysis report generation not implemented yet",
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleGetWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlist, err := api.Queries.GetWatchlist(r.Context())
	if err != nil {
		log.Printf("Error fetching watchlist: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch watchlist")
		return
	}

	log.Printf("GetWatchlist returned %d items", len(watchlist))

	if watchlist == nil {
		watchlist = []database.GetWatchlistRow{}
	}

	// Extract just the symbols and scores
	symbols := make([]map[string]interface{}, len(watchlist))
	for i, item := range watchlist {
		log.Printf("Watchlist item %d: Symbol=%s, Score=%v", i, item.Symbol, item.Score)
		symbols[i] = map[string]interface{}{
			"symbol":  item.Symbol,
			"score":   item.Score,
			"type":    item.AssetType,
			"reason":  item.Reason,
			"added":   item.AddedDate,
			"updated": item.LastUpdated,
		}
	}

	response := map[string]interface{}{
		"watchlist": symbols,
		"count":     len(symbols),
	}

	log.Printf("Sending response: %d symbols", len(symbols))

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleAddToWatchlist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string  `json:"symbol"`
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	// Validate that the stock exists by fetching asset info from Alpaca
	asset, err := api.AlpacaClient.GetAsset(req.Symbol)
	if err != nil {
		log.Printf("Warning: Stock validation failed for %s: %v", req.Symbol, err)
		// Continue anyway - validation is optional, log the error for debugging
	}
	if asset == nil && err != nil {
		log.Printf("Stock symbol '%s' not found or invalid", req.Symbol)
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Stock symbol '%s' not found. Please verify the symbol is valid.", req.Symbol))
		return
	}

	// Calculate actual score using metrics
	calculatedScore := req.Score // Default to provided score

	// Fetch bars and calculate real metrics
	bars, err := datafeed.GetAlpacaBars(req.Symbol, "1Day", 100, "")
	if err == nil && len(bars) > 0 {
		// Load config for weights
		cfg, cfgErr := config.LoadConfig()
		if cfgErr == nil {
			// Get the balanced profile weights (or default profile)
			weights := cfg.Profiles["balanced"].SignalWeights

			// Calculate metrics
			candidate, metricsErr := analyzer.CalculateCandidateMetrics(r.Context(), req.Symbol, bars, cfg, weights)
			if metricsErr == nil && candidate != nil {
				calculatedScore = candidate.Score
				log.Printf("Calculated score for %s: %.2f", req.Symbol, calculatedScore)
			} else {
				log.Printf("Warning: Could not calculate metrics for %s: %v", req.Symbol, metricsErr)
			}
		} else {
			log.Printf("Warning: Could not load config: %v", cfgErr)
		}
	} else {
		log.Printf("Warning: Could not fetch bars for %s: %v", req.Symbol, err)
	}

	params := database.AddToWatchlistParams{
		Symbol:    req.Symbol,
		AssetType: "stock",
		Score:     float32(calculatedScore),
		Reason: sql.NullString{
			String: req.Reason,
			Valid:  req.Reason != "",
		},
	}

	watchlistID, err := api.Queries.AddToWatchlist(r.Context(), params)
	if err != nil {
		log.Printf("Error adding to watchlist: %v", err)

		// Check for duplicate key constraint error
		if err.Error() == "pq: duplicate key value violates unique constraint \"watchlist_symbol_key\"" {
			WriteError(w, http.StatusConflict, fmt.Sprintf("Stock symbol '%s' is already in your watchlist.", req.Symbol))
			return
		}

		// Check for other PostgreSQL errors
		if err.Error() == "pq: duplicate key value violates unique constraint \"watchlist_pkey\"" {
			WriteError(w, http.StatusConflict, "This watchlist item already exists.")
			return
		}

		// Generic error
		WriteError(w, http.StatusInternalServerError, "Failed to add to watchlist")
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"watchlist_id": watchlistID,
		"symbol":       req.Symbol,
		"score":        calculatedScore,
		"message":      "Symbol added to watchlist",
	}

	WriteJSON(w, http.StatusCreated, response)
}

func (api *API) HandleRemoveFromWatchlist(w http.ResponseWriter, r *http.Request) {
	// Get symbol from query parameter (primary source)
	symbol := r.URL.Query().Get("symbol")

	// Only try to parse body if symbol is not in query parameter
	if symbol == "" {
		var req struct {
			Symbol string `json:"symbol"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
		symbol = req.Symbol
	}

	if symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	log.Printf("DEBUG: Attempting to remove symbol '%s' from watchlist", symbol)
	err := api.Queries.RemoveFromWatchlist(r.Context(), symbol)
	if err != nil {
		log.Printf("Error removing from watchlist: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to remove from watchlist")
		return
	}
	log.Printf("DEBUG: Successfully removed symbol '%s' from watchlist", symbol)

	response := map[string]interface{}{
		"success": true,
		"symbol":  symbol,
		"message": "Symbol removed from watchlist",
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleRefreshWatchlistScores(w http.ResponseWriter, r *http.Request) {
	// Get all watchlist items
	watchlist, err := api.Queries.GetWatchlist(r.Context())
	if err != nil {
		log.Printf("Error fetching watchlist: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch watchlist")
		return
	}

	// Load config for weights
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to load config")
		return
	}

	weights := cfg.Profiles["balanced"].SignalWeights

	updated := 0
	failed := 0
	results := make([]map[string]interface{}, 0)

	// Recalculate score for each symbol using the full scoring logic
	for _, item := range watchlist {
		symbol := item.Symbol

		// Fetch bars
		bars, err := datafeed.GetAlpacaBars(symbol, "1Day", 100, "")
		if err != nil || len(bars) == 0 {
			log.Printf("Failed to fetch bars for %s: %v", symbol, err)
			failed++
			results = append(results, map[string]interface{}{
				"symbol": symbol,
				"status": "failed",
				"error":  "Failed to fetch market data",
			})
			continue
		}

		// Calculate RSI
		closes := make([]float64, len(bars))
		for i, bar := range bars {
			closes[i] = bar.Close
		}
		rsiValues, err := indicators.CalculateRSI(closes, 14)
		if err != nil || len(rsiValues) == 0 {
			log.Printf("Failed to calculate RSI for %s: %v", symbol, err)
			failed++
			results = append(results, map[string]interface{}{
				"symbol": symbol,
				"status": "failed",
				"error":  "Failed to calculate RSI",
			})
			continue
		}

		// Calculate ATR
		atrBars := make([]indicators.ATRBar, len(bars))
		for i, bar := range bars {
			atrBars[i] = indicators.ATRBar{
				High:  bar.High,
				Low:   bar.Low,
				Close: bar.Close,
			}
		}
		atrValues, err := indicators.CalculateATR(atrBars, 14)
		if err != nil || len(atrValues) == 0 {
			log.Printf("Failed to calculate ATR for %s: %v", symbol, err)
			failed++
			results = append(results, map[string]interface{}{
				"symbol": symbol,
				"status": "failed",
				"error":  "Failed to calculate ATR",
			})
			continue
		}

		// Calculate VWAP
		vwapCalc := indicators.NewVWAPCalculator(bars)
		vwapValue := vwapCalc.Calculate()

		// Get whale activity count
		whaleCount := 0 // Default to 0, would fetch from database if needed

		// Build scoring input and calculate score
		atrValue := atrValues[len(atrValues)-1]
		rsiValue := rsiValues[len(rsiValues)-1]
		atrCategory := scoring.CategorizeATRValue(atrValue, bars)

		scoringInput, _ := scoring.BuildScoringInput(bars, vwapValue, rsiValue, whaleCount, atrValue, atrCategory)
		score := detection.CalculateInterestScore(scoringInput, weights)

		// Update the score in database
		updateParams := database.UpdateWatchlistScoreParams{
			Symbol: symbol,
			Score:  float32(score),
		}

		err = api.Queries.UpdateWatchlistScore(r.Context(), updateParams)
		if err != nil {
			log.Printf("Failed to update score for %s: %v", symbol, err)
			failed++
			results = append(results, map[string]interface{}{
				"symbol": symbol,
				"status": "failed",
				"error":  "Failed to update database",
			})
			continue
		}

		updated++
		results = append(results, map[string]interface{}{
			"symbol":    symbol,
			"status":    "updated",
			"old_score": item.Score,
			"new_score": score,
		})

		log.Printf("Updated score for %s: %.2f -> %.2f", symbol, item.Score, score)
	}

	response := map[string]interface{}{
		"success": true,
		"total":   len(watchlist),
		"updated": updated,
		"failed":  failed,
		"results": results,
		"message": fmt.Sprintf("Refreshed scores: %d updated, %d failed", updated, failed),
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleAnalyzeSymbol(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol parameter is required")
		return
	}

	bars, err := datafeed.GetAlpacaBars(symbol, "1Day", 100, "")
	if err != nil {
		log.Printf("Error fetching bars for %s: %v", symbol, err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch market data")
		return
	}

	if len(bars) < 14 {
		WriteError(w, http.StatusBadRequest, "Not enough data to analyze")
		return
	}

	closes := make([]float64, len(bars))
	atrBars := make([]indicators.ATRBar, len(bars))
	for i, bar := range bars {
		closes[i] = bar.Close
		atrBars[i] = indicators.ATRBar{
			High:  bar.High,
			Low:   bar.Low,
			Close: bar.Close,
		}
	}

	rsiValues, err := indicators.CalculateRSI(closes, 14)
	if err != nil || len(rsiValues) == 0 {
		WriteError(w, http.StatusInternalServerError, "Failed to calculate RSI")
		return
	}

	atrValues, err := indicators.CalculateATR(atrBars, 14)
	if err != nil || len(atrValues) == 0 {
		WriteError(w, http.StatusInternalServerError, "Failed to calculate ATR")
		return
	}

	currentPrice := bars[0].Close
	currentRSI := rsiValues[len(rsiValues)-1]
	currentATR := atrValues[len(atrValues)-1]

	// SMA 20 - calculate from most recent 20 bars
	sma20 := 0.0
	barsForSMA := 20
	if len(bars) < barsForSMA {
		barsForSMA = len(bars)
	}
	for i := 0; i < barsForSMA; i++ {
		sma20 += bars[i].Close
	}
	sma20 /= float64(barsForSMA)

	trend := "neutral"
	if currentPrice > sma20*1.02 {
		trend = "bullish"
	} else if currentPrice < sma20*0.98 {
		trend = "bearish"
	}

	support := indicators.FindSupport(bars)
	resistance := indicators.FindResistance(bars)

	distanceToSupport := ((currentPrice - support) / support) * 100
	distanceToResistance := ((resistance - currentPrice) / currentPrice) * 100

	patternDetector := detection.NewPatternDetector()
	patterns := patternDetector.DetectAllPatterns(bars)

	var bestPattern map[string]interface{}
	var bestP *detection.PatternSignal
	if len(patterns) > 0 {
		// Find the best detected pattern
		for i := range patterns {
			if patterns[i].Detected {
				if bestP == nil || patterns[i].Confidence > bestP.Confidence {
					bestP = &patterns[i]
				}
			}
		}

		if bestP != nil {
			bestPattern = map[string]interface{}{
				"pattern":       bestP.Pattern,
				"direction":     bestP.Direction,
				"confidence":    bestP.Confidence,
				"support_level": bestP.SupportLevel,
				"resistance":    bestP.ResistanceLevel,
				"target_up":     bestP.PriceTargetUp,
				"target_down":   bestP.PriceTargetDown,
				"stop_loss":     bestP.StopLossLevel,
				"risk_reward":   bestP.RiskRewardRatio,
				"reasoning":     bestP.Reasoning,
			}
		}
	}

	// Multi-timeframe analysis (placeholder - would require async calls in typescript api)
	var multiTimeframe map[string]interface{}
	// Note: Full multi-timeframe analysis would require fetching 4H and 1H data separately
	// For now, we return empty placeholder
	multiTimeframe = map[string]interface{}{
		"note": "Multi-timeframe analysis requires additional data fetching",
	}

	// Determine RSI signal
	rsiSignal := "neutral"
	if currentRSI > 70 {
		rsiSignal = "overbought"
	} else if currentRSI < 30 {
		rsiSignal = "oversold"
	}

	// Calculate trading recommendation
	tradingRec := calculateTradingRecommendation(currentPrice, currentRSI, support, resistance, trend, bestP)

	response := map[string]interface{}{
		"symbol":                 symbol,
		"current_price":          currentPrice,
		"rsi":                    currentRSI,
		"rsi_signal":             rsiSignal,
		"atr":                    currentATR,
		"sma_20":                 sma20,
		"trend":                  trend,
		"bars_analyzed":          len(bars),
		"timestamp":              time.Now().Unix(),
		"support_level":          support,
		"resistance_level":       resistance,
		"distance_to_support":    distanceToSupport,
		"distance_to_resistance": distanceToResistance,
		"chart_pattern":          bestPattern,
		"multi_timeframe":        multiTimeframe,
		"trading_recommendation": tradingRec,
	}

	WriteJSON(w, http.StatusOK, response)
}

func calculateTradingRecommendation(price, rsi, support, resistance float64, trend string, pattern *detection.PatternSignal) map[string]interface{} {
	recommendation := "HOLD"
	confidence := 50.0
	reasoning := ""

	// Base recommendation on RSI
	if rsi < 30 {
		recommendation = "BUY"
		confidence = 65.0
		reasoning = "RSI is oversold"

		// Strengthen if near support
		if price < support*1.01 {
			confidence = 80.0
			reasoning += " and price is at support level"
		}
	} else if rsi > 70 {
		recommendation = "SELL"
		confidence = 65.0
		reasoning = "RSI is overbought"

		// Strengthen if near resistance
		if price > resistance*0.99 {
			confidence = 80.0
			reasoning += " and price is at resistance level"
		}
	} else {
		// Check trend
		if trend == "bullish" {
			recommendation = "BUY"
			confidence = 60.0
			reasoning = "Bullish trend with RSI in neutral zone"
		} else if trend == "bearish" {
			recommendation = "SELL"
			confidence = 60.0
			reasoning = "Bearish trend with RSI in neutral zone"
		}
	}

	// Adjust based on pattern if available
	if pattern != nil && pattern.Detected {
		if pattern.Direction == "LONG" && (recommendation == "BUY" || recommendation == "HOLD") {
			confidence += (pattern.Confidence / 100.0) * 20
			reasoning += fmt.Sprintf(" - %s pattern supports upside", pattern.Pattern)
			recommendation = "BUY"
		} else if pattern.Direction == "SHORT" && (recommendation == "SELL" || recommendation == "HOLD") {
			confidence += (pattern.Confidence / 100.0) * 20
			reasoning += fmt.Sprintf(" - %s pattern suggests downside", pattern.Pattern)
			recommendation = "SELL"
		}
	}

	// Cap confidence at 100
	if confidence > 100 {
		confidence = 100
	}

	return map[string]interface{}{
		"action":     recommendation,
		"confidence": confidence,
		"reasoning":  reasoning,
	}
}

func (api *API) HandleScoutStocks(w http.ResponseWriter, r *http.Request) {

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	minScoreStr := r.URL.Query().Get("min_score")
	minScore := 50.0
	if minScoreStr != "" {
		if parsedScore, err := strconv.ParseFloat(minScoreStr, 64); err == nil {
			minScore = parsedScore
		}
	}

	log.Printf("Scanning %d stocks with min score %.1f (limit=%s, minScore=%s)", limit, minScore, limitStr, minScoreStr)
	ctx := context.Background()

	candidates, totalScanned, err := scanner.PerformProfileScan(ctx, "api_scout", minScore, 0, limit, nil)
	if err != nil {
		errMsg := err.Error()
		log.Printf("SCANNER ERROR: %v", errMsg)
		WriteError(w, http.StatusInternalServerError, errMsg)
		return
	}

	log.Printf("SCAN COMPLETE: Got %d results from %d total symbols, limit was %d", len(candidates), totalScanned, limit)

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	var opportunities []map[string]interface{}
	for i, candidate := range candidates {
		if i >= limit {
			break
		}

		opp := map[string]interface{}{
			"symbol":    candidate.Symbol,
			"score":     candidate.Score,
			"analysis":  candidate.Analysis,
			"rsi":       candidate.RSI,
			"atr":       candidate.ATR,
			"timestamp": time.Now().Unix(),
			"rank":      i + 1,
		}
		opportunities = append(opportunities, opp)
	}

	response := map[string]interface{}{
		"scanned_count":  len(opportunities),
		"total_symbols":  totalScanned,
		"min_score":      minScore,
		"limit":          limit,
		"opportunities":  opportunities,
		"scan_timestamp": time.Now().Unix(),
		"message":        "Real-time stock screening results",
	}

	WriteJSON(w, http.StatusOK, response)
}
