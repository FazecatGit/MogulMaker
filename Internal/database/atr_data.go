package datafeed

import (
	"context"
	"strconv"
	"time"

	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils/scoring"
)

type ATRPoint struct {
	ATR       float64
	Timestamp time.Time
}

type ATRBar struct {
	High      float64
	Low       float64
	Close     float64
	Timestamp time.Time
}

func FetchATRPrices(symbol string, limit int, timeframe string) ([]ATRBar, error) {
	params := database.GetATRPricesParams{
		Symbol:    symbol,
		Timeframe: timeframe,
		Limit:     int32(limit),
	}
	ctx := context.Background()
	rows, err := Queries.GetATRPrices(ctx, params)

	var atrBars []ATRBar
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		atrHigh, err := strconv.ParseFloat(row.HighPrice, 64)
		if err != nil {
			return nil, err
		}
		atrLow, err := strconv.ParseFloat(row.LowPrice, 64)
		if err != nil {
			return nil, err
		}
		atrClose, err := strconv.ParseFloat(row.ClosePrice, 64)
		if err != nil {
			return nil, err
		}

		atrBars = append(atrBars, ATRBar{
			High:      atrHigh,
			Low:       atrLow,
			Close:     atrClose,
			Timestamp: row.Timestamp,
		})
	}
	return atrBars, nil
}

func SaveATR(symbol string, timestamp time.Time, atrValue float64) error {
	params := database.SaveATRParams{
		Symbol:               symbol,
		CalculationTimestamp: timestamp,
		AtrValue:             float32(atrValue),
	}
	ctx := context.Background()
	err := Queries.SaveATR(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func FetchATRForDisplay(symbol string, limit int) (map[string]float64, error) {
	params := database.GetATRForDateRangeParams{
		Symbol: symbol,
		Limit:  int32(limit),
	}
	ctx := context.Background()
	rows, err := Queries.GetATRForDateRange(ctx, params)
	if err != nil {
		return nil, err
	}

	atrMap := make(map[string]float64)
	for _, row := range rows {
		dateStr := row.CalculationTimestamp.Format("2006-01-02 15:04:05")
		atrMap[dateStr] = float64(row.AtrValue)
	}
	return atrMap, nil
}

func FetchATRByTimestampRange(symbol string, startTime, endTime time.Time) (map[string]float64, error) {
	params := database.GetATRByTimestampRangeParams{
		Symbol:                 symbol,
		CalculationTimestamp:   startTime,
		CalculationTimestamp_2: endTime,
	}
	ctx := context.Background()
	rows, err := Queries.GetATRByTimestampRange(ctx, params)
	if err != nil {
		return nil, err
	}

	atrMap := make(map[string]float64)
	for _, row := range rows {
		dateStr := row.CalculationTimestamp.Format("2006-01-02 15:04:05")
		atrMap[dateStr] = float64(row.AtrValue)
	}
	return atrMap, nil
}
func CalculateAndStoreATR(symbol string, bars []types.Bar) error {
	if len(bars) == 0 {
		return nil
	}

	atrValue := scoring.CalculateATRFromBars(bars)

	latestBar := bars[len(bars)-1]
	latestTime, err := time.Parse(time.RFC3339, latestBar.Timestamp)
	if err != nil {
		return err
	}

	err = SaveATR(symbol, latestTime, atrValue)
	if err != nil {
		return err
	}

	return nil
}
