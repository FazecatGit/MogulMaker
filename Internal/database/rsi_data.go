package datafeed

import (
	"context"
	"fmt"
	"strconv"
	"time"

	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils"
)

type PricePoint struct {
	Price     float64
	Timestamp time.Time
}

func FetchClosingPrices(symbol string, days int, timeframe string) ([]float64, error) {
	params := database.GetClosingPricesParams{
		Symbol:    symbol,
		Timeframe: timeframe,
		Limit:     int32(days),
	}

	ctx := context.Background()
	rows, err := Queries.GetClosingPrices(ctx, params)
	if err != nil {
		return nil, err
	}

	var closingPrices []float64
	for _, row := range rows {
		price, err := strconv.ParseFloat(row.ClosePrice, 64)
		if err != nil {
			return nil, err
		}
		closingPrices = append(closingPrices, price)
	}

	return closingPrices, nil
}
func SaveRSI(symbol string, timestamp time.Time, rsiValue float64) error {
	params := database.SaveRSIParams{
		Symbol:               symbol,
		CalculationTimestamp: timestamp,
		RsiValue:             float32(rsiValue),
	}
	ctx := context.Background()
	err := Queries.SaveRSI(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func FetchPricePoints(symbol string, days int, timeframe string) ([]PricePoint, error) {
	params := database.GetClosingPricesParams{
		Symbol:    symbol,
		Timeframe: timeframe,
		Limit:     int32(days),
	}
	ctx := context.Background()
	rows, err := Queries.GetClosingPrices(ctx, params)
	if err != nil {
		return nil, err
	}

	var pricePoints []PricePoint
	for _, row := range rows {
		price, err := strconv.ParseFloat(row.ClosePrice, 64)
		if err != nil {
			return nil, err
		}
		pricePoints = append(pricePoints, PricePoint{
			Price:     price,
			Timestamp: row.Timestamp,
		})
	}

	return pricePoints, nil
}

func FetchRSIForDisplay(symbol string, limit int) (map[string]float64, error) {
	params := database.GetRSIForDateRangeParams{
		Symbol: symbol,
		Limit:  int32(limit),
	}
	ctx := context.Background()
	rows, err := Queries.GetRSIForDateRange(ctx, params)
	if err != nil {
		return nil, err
	}

	rsiMap := make(map[string]float64)
	for _, row := range rows {
		dateStr := row.CalculationTimestamp.Format("2006-01-02 15:04:05")
		rsiMap[dateStr] = float64(row.RsiValue)
	}
	return rsiMap, nil
}

func FetchRSIByTimestampRange(symbol string, startTime, endTime time.Time) (map[string]float64, error) {
	params := database.GetRSIByTimestampRangeParams{
		Symbol:                 symbol,
		CalculationTimestamp:   startTime,
		CalculationTimestamp_2: endTime,
	}
	ctx := context.Background()
	rows, err := Queries.GetRSIByTimestampRange(ctx, params)
	if err != nil {
		return nil, err
	}

	rsiMap := make(map[string]float64)
	for _, row := range rows {
		dateStr := row.CalculationTimestamp.Format("2006-01-02 15:04:05")
		rsiMap[dateStr] = float64(row.RsiValue)
	}
	return rsiMap, nil
}

// calculateRSI calculates RSI values locally to avoid import cycle
func calculateRSI(closes []float64, period int) ([]float64, error) {
	if len(closes) < period+1 {
		return nil, fmt.Errorf("not enough data")
	}
	rsi := make([]float64, len(closes))

	// Calculate gains and losses
	gains := make([]float64, len(closes))
	losses := make([]float64, len(closes))

	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	for i := period; i < len(closes); i++ {
		windowGains := gains[i-period+1 : i+1]
		windowLosses := losses[i-period+1 : i+1]

		// Calculate average window
		avgGain := utils.Average(windowGains)
		avgLoss := utils.Average(windowLosses)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi, nil
}

func CalculateAndStoreRSI(symbol string, bars []types.Bar) error {
	if len(bars) == 0 {
		return nil
	}
	closingPrices := make([]float64, len(bars))
	for i, bar := range bars {
		closingPrices[i] = bar.Close
	}

	rsiValues, err := calculateRSI(closingPrices, 14)
	if err != nil {
		return err
	}
	for i, bar := range bars {
		timestamp, err := time.Parse(time.RFC3339, bar.Timestamp)
		if err != nil {
			return err
		}
		err = SaveRSI(symbol, timestamp, rsiValues[i])
		if err != nil {
			return err
		}
	}
	return nil
}
