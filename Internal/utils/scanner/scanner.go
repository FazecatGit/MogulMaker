package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	db "github.com/fazecat/mogulmaker/Internal/database"
	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils/config"
)

func ShouldScan(ctx context.Context, profileName string, cfg *config.Config, q *database.Queries) (bool, error) {
	scan, err := q.GetScanLog(ctx, profileName)
	if err != nil {
		return false, err
	}
	nextDue := GetNextScanDue(scan.LastScanTimestamp, profileName, cfg)
	if time.Now().After(nextDue) || time.Now().Equal(nextDue) {
		return true, nil
	}
	return false, nil
}

func PerformScan(ctx context.Context, profileName string, cfg *config.Config, q *database.Queries) (int, error) {
	watchlist, err := q.GetWatchlist(ctx)
	if err != nil {
		return 0, err
	}

	scannedCount := 0
	criteria := DefaultScreenerCriteria()

	for _, item := range watchlist {
		symbol := item.Symbol

		// Use the advanced screener logic
		stockScores, err := ScreenStocksWithType([]string{symbol}, "1Day", 100, criteria, nil, "stock")
		if err != nil || len(stockScores) == 0 {
			continue
		}

		result := stockScores[0]

		// Skip if no meaningful data
		if result.Score == 0 && len(result.Signals) == 0 {
			continue
		}

		// Update watchlist score
		err = q.UpdateWatchlistScore(ctx, database.UpdateWatchlistScoreParams{
			Score:  float32(result.Score),
			Symbol: symbol,
		})
		if err != nil {
			continue
		}

		scannedCount++
	}

	err = q.UpsertScanLog(ctx, database.UpsertScanLogParams{
		ProfileName:       profileName,
		LastScanTimestamp: time.Now(),
		NextScanDue:       GetNextScanDue(time.Now(), profileName, cfg),
		SymbolsScanned:    sql.NullInt32{Int32: int32(scannedCount), Valid: true},
	})
	if err != nil {
		return 0, err
	}

	return scannedCount, nil
}

func CalculateScanInterval(profileName string, cfg *config.Config) time.Duration {
	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return 24 * time.Hour
	}
	return time.Duration(profile.ScanIntervalDays) * 24 * time.Hour
}

func GetNextScanDue(lastScan time.Time, profileName string, cfg *config.Config) time.Time {
	interval := CalculateScanInterval(profileName, cfg)
	return lastScan.Add(interval)
}

func PerformProfileScan(ctx context.Context, profileName string, minScore float64, offset int, batchSize int, cfg *config.Config) ([]types.Candidate, int, error) {
	symbols, err := GetTradableAssets()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch tradeable assets: %v", err)
	}

	totalSymbols := len(symbols)

	end := offset + batchSize
	if end > totalSymbols {
		end = totalSymbols
	}

	if offset >= totalSymbols {
		return []types.Candidate{}, totalSymbols, nil
	}

	candidates := []types.Candidate{}
	criteria := DefaultScreenerCriteria()

	for i := offset; i < end; i++ {
		symbol := symbols[i]

		// Use the advanced screener logic instead of simple scoring
		stockScores, err := ScreenStocksWithType([]string{symbol}, "1Day", 100, criteria, nil, "stock")
		if err != nil || len(stockScores) == 0 {
			continue
		}

		result := stockScores[0]

		if result.Score == 0 && len(result.Signals) == 0 {
			continue
		}

		analysis := "No signals"
		if len(result.Signals) > 0 {
			analysis = result.Signals[0] // Use first signal as primary analysis
		}

		bars, err := db.GetAlpacaBars(symbol, "1Day", 100, "")
		if err != nil {
			continue
		}

		candidate := types.Candidate{
			Symbol:   symbol,
			Score:    result.Score,
			Analysis: analysis,
			Bars:     bars,
		}

		if result.RSI != nil {
			candidate.RSI = *result.RSI
		}
		if result.ATR != nil {
			candidate.ATR = *result.ATR
		}

		if candidate.Score >= minScore {
			candidates = append(candidates, candidate)
		}
	}

	return candidates, totalSymbols, nil
}
