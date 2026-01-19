package watchlist

import (
	"context"
	"database/sql"
	"encoding/json"

	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
)

func AddToWatchlist(ctx context.Context, q *database.Queries, symbol string, assetType string, score float64, reason string) (int32, error) {
	params := database.AddToWatchlistParams{
		Symbol:    symbol,
		AssetType: assetType,
		Score:     float32(score),
		Reason:    sql.NullString{String: reason, Valid: reason != ""},
	}

	id, err := q.AddToWatchlist(ctx, params)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func UpdateWatchlistScoreWithHistory(ctx context.Context, q *database.Queries, symbol string, newScore float64, reason string, analysisData map[string]interface{}) error {
	//Get current watchlist item by symbol
	watchlistItem, err := q.GetWatchlistBySymbol(ctx, symbol)
	if err != nil {
		return err
	}

	// Update the score
	updateParams := database.UpdateWatchlistScoreParams{
		Symbol: symbol,
		Score:  float32(newScore),
	}
	err = q.UpdateWatchlistScore(ctx, updateParams)
	if err != nil {
		return err
	}

	// Add history entry with analysis data as JSON
	jsonData, _ := json.Marshal(analysisData)
	historyParams := database.AddWatchlistHistoryParams{
		WatchlistID:  watchlistItem.ID,
		OldScore:     sql.NullFloat64{Float64: float64(watchlistItem.Score), Valid: true},
		NewScore:     float32(newScore),
		AnalysisData: sql.NullString{String: string(jsonData), Valid: len(jsonData) > 0},
	}
	err = q.AddWatchlistHistory(ctx, historyParams)
	if err != nil {
		return err
	}

	return nil
}

func GetWatchlist(ctx context.Context, q *database.Queries) ([]database.GetWatchlistRow, error) {
	getwatchlist, err := q.GetWatchlist(ctx)
	if err != nil {
		return nil, err
	}
	return getwatchlist, nil
}

func SkipSymbol(ctx context.Context, q *database.Queries, symbol, assetType, reason string) error {
	skipSymbol := database.SkipSymbolParams{
		Symbol:    symbol,
		AssetType: assetType,
		Reason:    sql.NullString{String: reason, Valid: reason != ""},
	}
	err := q.SkipSymbol(ctx, skipSymbol)
	if err != nil {
		return err
	}
	return nil
}

func GetRecheckableSymbols(ctx context.Context, q *database.Queries) ([]database.GetRecheckableSymbolsRow, error) {
	recheckableSymbols, err := q.GetRecheckableSymbols(ctx)
	if err != nil {
		return nil, err
	}
	return recheckableSymbols, nil
}
