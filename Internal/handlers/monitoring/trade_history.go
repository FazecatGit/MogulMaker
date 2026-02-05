package monitoring

import (
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type TradeHistoryRecord struct {
	Order       alpaca.Order
	PnL         float64
	ReturnPct   float64
	PairedWith  *alpaca.Order
	IsClosed    bool
	TradePairID string
}

// PairTradesAndCalculatePnL pairs buy and sell orders and calculates P&L for each pair
func PairTradesAndCalculatePnL(allOrders []alpaca.Order) []TradeHistoryRecord {
	type TradeRecord struct {
		order       alpaca.Order
		pnl         float64
		returnPct   float64
		pairedWith  *alpaca.Order
		isClosed    bool
		tradePairID string
	}

	filledBySymbol := make(map[string][]alpaca.Order)
	var allFilledOrders []alpaca.Order

	for _, order := range allOrders {
		if order.Status == "filled" || order.Status == "closed" {
			filledBySymbol[order.Symbol] = append(filledBySymbol[order.Symbol], order)
			allFilledOrders = append(allFilledOrders, order)
		}
	}

	var tradeRecords []TradeRecord
	usedOrderIDs := make(map[string]bool)

	// Pair buy and sell orders and calculate P&L
	for _, symbolOrders := range filledBySymbol {
		var buyOrders []alpaca.Order
		var sellOrders []alpaca.Order

		for _, order := range symbolOrders {
			if order.Side == "buy" {
				buyOrders = append(buyOrders, order)
			} else {
				sellOrders = append(sellOrders, order)
			}
		}

		// Pair them up
		for i := 0; i < len(buyOrders) && i < len(sellOrders); i++ {
			buyOrder := buyOrders[i]
			sellOrder := sellOrders[i]

			buyQty, _ := buyOrder.FilledQty.Float64()
			buyPrice, _ := buyOrder.FilledAvgPrice.Float64()
			sellQty, _ := sellOrder.FilledQty.Float64()
			sellPrice, _ := sellOrder.FilledAvgPrice.Float64()

			qty := buyQty
			if sellQty < qty {
				qty = sellQty
			}

			pnl := (sellPrice - buyPrice) * qty
			returnPct := ((sellPrice - buyPrice) / buyPrice) * 100

			pairID := buyOrder.ID + "-" + sellOrder.ID

			// Add buy order record
			tradeRecords = append(tradeRecords, TradeRecord{
				order:       buyOrder,
				pnl:         pnl,
				returnPct:   returnPct,
				pairedWith:  &sellOrder,
				isClosed:    true,
				tradePairID: pairID,
			})

			// Add sell order record
			tradeRecords = append(tradeRecords, TradeRecord{
				order:       sellOrder,
				pnl:         pnl,
				returnPct:   returnPct,
				pairedWith:  &buyOrder,
				isClosed:    true,
				tradePairID: pairID,
			})

			usedOrderIDs[buyOrder.ID] = true
			usedOrderIDs[sellOrder.ID] = true
		}
	}

	// Add unpaired orders as open trades
	for _, order := range allFilledOrders {
		if !usedOrderIDs[order.ID] {
			tradeRecords = append(tradeRecords, TradeRecord{
				order:     order,
				pnl:       0,
				returnPct: 0,
				isClosed:  false,
			})
		}
	}

	// Convert to output format
	var result []TradeHistoryRecord
	for _, rec := range tradeRecords {
		result = append(result, TradeHistoryRecord{
			Order:       rec.order,
			PnL:         rec.pnl,
			ReturnPct:   rec.returnPct,
			PairedWith:  rec.pairedWith,
			IsClosed:    rec.isClosed,
			TradePairID: rec.tradePairID,
		})
	}

	return result
}

// FormatTradeRecordsAsJSON converts trade history records to JSON-friendly format
func FormatTradeRecordsAsJSON(records []TradeHistoryRecord) []map[string]interface{} {
	var trades []map[string]interface{}

	for _, rec := range records {
		order := rec.Order
		filledQty, _ := order.FilledQty.Float64()
		filledAvgPrice, _ := order.FilledAvgPrice.Float64()

		status := "open"
		if rec.IsClosed {
			status = "closed"
		}

		trade := map[string]interface{}{
			"id":            order.ID,
			"symbol":        order.Symbol,
			"exchange":      "NASDAQ",
			"entry_time":    order.CreatedAt.Format(time.RFC3339),
			"exit_time":     nil,
			"entry_price":   filledAvgPrice,
			"exit_price":    nil,
			"qty":           filledQty,
			"side":          order.Side,
			"status":        status,
			"realized_pl":   rec.PnL,
			"realized_plpc": rec.ReturnPct / 100,
			"duration_ms":   nil,
			"submitted_at":  order.SubmittedAt.Format(time.RFC3339),
			"filled_at":     nil,
		}

		if order.FilledAt != nil {
			trade["filled_at"] = order.FilledAt.Format(time.RFC3339)
		}

		trades = append(trades, trade)
	}

	return trades
}
