package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	datafeed "github.com/fazecat/mogulmaker/Internal/database"
	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	"github.com/fazecat/mogulmaker/Internal/handlers/monitoring"
	"github.com/fazecat/mogulmaker/Internal/handlers/risk"
	"github.com/fazecat/mogulmaker/Internal/strategy/metrics"
	"github.com/fazecat/mogulmaker/Internal/strategy/position"
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

	response := map[string]interface{}{
		"count":     len(alpacaPositions),
		"positions": alpacaPositions,
		"timestamp": time.Now().Unix(),
		"risk_status": map[string]interface{}{
			"enabled": api.RiskManager != nil,
		},
	}
	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleGetRiskStatus(w http.ResponseWriter, r *http.Request) {
	var riskStatus interface{}
	if api.RiskManager != nil {
		riskStatus = map[string]interface{}{
			"enabled": true,
		}
	} else {
		riskStatus = map[string]interface{}{
			"enabled": false,
		}
	}

	WriteJSON(w, http.StatusOK, riskStatus)
}

func (api *API) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	dbTrades, err := api.Queries.GetAllTrades(context.Background())
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
	symbol := r.URL.Query().Get("symbol")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	allTrades, err := api.Queries.GetAllTrades(context.Background())
	if err != nil {
		log.Printf("Error fetching trades: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch trades")
		return
	}

	var filteredTrades []database.GetAllTradesRow
	if symbol != "" {
		for _, trade := range allTrades {
			if trade.Symbol == symbol {
				filteredTrades = append(filteredTrades, trade)
			}
		}
	} else {
		filteredTrades = allTrades
	}

	if len(filteredTrades) > limit {
		filteredTrades = filteredTrades[:limit]
	}

	WriteJSON(w, http.StatusOK, filteredTrades)
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

	qty, _ := position.Qty.Float64()

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

	// Fetch historical bars for the symbol
	historicalBars, err := datafeed.GetAlpacaBars(symbol, "1Day", 100, "")
	if err != nil || len(historicalBars) == 0 {
		log.Printf("Error fetching historical bars: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch historical data for backtest")
		return
	}

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
	watchlist, err := api.Queries.GetWatchlist(context.Background())
	if err != nil {
		log.Printf("Error fetching watchlist: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch watchlist")
		return
	}

	// Extract just the symbols and scores
	var symbols []map[string]interface{}
	for _, item := range watchlist {
		symbols = append(symbols, map[string]interface{}{
			"symbol":  item.Symbol,
			"score":   item.Score,
			"type":    item.AssetType,
			"reason":  item.Reason,
			"added":   item.AddedDate,
			"updated": item.LastUpdated,
		})
	}

	response := map[string]interface{}{
		"watchlist": symbols,
		"count":     len(watchlist),
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleAddToWatchlist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string  `json:"symbol"`
		Score  float32 `json:"score"`
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

	// Add to watchlist in database
	params := database.AddToWatchlistParams{
		Symbol:    req.Symbol,
		AssetType: "stock",
		Score:     req.Score,
		Reason: sql.NullString{
			String: req.Reason,
			Valid:  req.Reason != "",
		},
	}

	watchlistID, err := api.Queries.AddToWatchlist(context.Background(), params)
	if err != nil {
		log.Printf("Error adding to watchlist: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to add to watchlist")
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"watchlist_id": watchlistID,
		"symbol":       req.Symbol,
		"score":        req.Score,
		"message":      "Symbol added to watchlist",
	}

	WriteJSON(w, http.StatusCreated, response)
}

func (api *API) HandleRemoveFromWatchlist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string `json:"symbol"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Symbol == "" {
		WriteError(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	// Remove from watchlist in database
	err := api.Queries.RemoveFromWatchlist(context.Background(), req.Symbol)
	if err != nil {
		log.Printf("Error removing from watchlist: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to remove from watchlist")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"symbol":  req.Symbol,
		"message": "Symbol removed from watchlist",
	}

	WriteJSON(w, http.StatusOK, response)
}

func (api *API) HandleScoutStocks(w http.ResponseWriter, r *http.Request) {
	// Scout/scanner endpoint - find opportunities
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Fetch scan logs from database
	scanLogs, err := api.Queries.GetAllScanLogs(context.Background())
	if err != nil {
		log.Printf("Error fetching scan logs: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch scan logs")
		return
	}

	// Limit results
	if len(scanLogs) > limit {
		scanLogs = scanLogs[:limit]
	}

	response := map[string]interface{}{
		"scanned_count": len(scanLogs),
		"limit":         limit,
		"opportunities": scanLogs,
		"message":       "Scanner results from database",
	}

	WriteJSON(w, http.StatusOK, response)
}
