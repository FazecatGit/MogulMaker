package internal

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	"github.com/fazecat/mogulmaker/Internal/handlers/monitoring"
	"github.com/fazecat/mogulmaker/Internal/handlers/risk"
	"github.com/fazecat/mogulmaker/Internal/strategy/metrics"
	"github.com/fazecat/mogulmaker/Internal/strategy/position"
)

type API struct {
	PositionManager *position.PositionManager
	RiskManager     *risk.Manager
	Queries         *database.Queries
	TradeMonitor    *monitoring.Monitor
}

func (api *API) HandleGetPositions(w http.ResponseWriter, r *http.Request) {
	positions := api.PositionManager.GetOpenPositions()

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

	response := map[string]interface{}{
		"positions":   positions,
		"risk_status": riskStatus,
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
