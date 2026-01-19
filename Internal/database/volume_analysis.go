package datafeed

import (
	"context"
	"database/sql"
	"time"

	sqlc "github.com/fazecat/mogulmaker/Internal/database/sqlc"
)

type WhaleEventData struct {
	Symbol      string
	Timestamp   string
	Direction   string
	Volume      int64
	ZScore      string
	ClosePrice  string
	PriceChange string
	Conviction  string
}

func SaveWhaleEvent(ctx context.Context, q *sqlc.Queries, whale WhaleEventData) error {
	timestamp, err := time.Parse("2006-01-02 15:04:05", whale.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	priceChange := sql.NullString{String: whale.PriceChange, Valid: whale.PriceChange != ""}

	err = q.CreateWhaleEvent(ctx, sqlc.CreateWhaleEventParams{
		Symbol:      whale.Symbol,
		Timestamp:   timestamp,
		Direction:   whale.Direction,
		Volume:      whale.Volume,
		ZScore:      whale.ZScore,
		ClosePrice:  whale.ClosePrice,
		PriceChange: priceChange,
		Conviction:  whale.Conviction,
	})
	return err
}

func GetRecentWhales(ctx context.Context, q *sqlc.Queries, symbol string, limit int32) ([]sqlc.WhaleEvent, error) {
	return q.GetWhaleEventsBySymbol(ctx, sqlc.GetWhaleEventsBySymbolParams{
		Symbol: symbol,
		Limit:  limit,
	})
}

func GetHighConvictionWhales(ctx context.Context, q *sqlc.Queries, symbol string) ([]sqlc.WhaleEvent, error) {
	return q.GetHighConvictionWhales(ctx, symbol)
}
