package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	db "github.com/fazecat/mongelmaker/Internal/database"
	database "github.com/fazecat/mongelmaker/Internal/database/sqlc"
	"github.com/fazecat/mongelmaker/Internal/strategy"
	"github.com/fazecat/mongelmaker/Internal/types"
	"github.com/fazecat/mongelmaker/Internal/utils/analyzer"
	"github.com/fazecat/mongelmaker/Internal/utils/config"
	"github.com/fazecat/mongelmaker/Internal/utils/scoring"
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

// PerformScan scans all watchlist symbols and updates scores
func PerformScan(ctx context.Context, profileName string, cfg *config.Config, q *database.Queries) (int, error) {
	watchlist, err := q.GetWatchlist(ctx)
	if err != nil {
		return 0, err
	}

	scannedCount := 0

	for _, item := range watchlist {
		symbol := item.Symbol

		bars, err := db.GetAlpacaBars(symbol, "1Day", 100, "")
		if err != nil {
			continue
		}

		// Calculate indicators
		vwapCalc := strategy.NewVWAPCalculator(bars)
		vwapPrice := vwapCalc.Calculate()

		closes := make([]float64, len(bars))
		for i, bar := range bars {
			closes[i] = bar.Close
		}
		rsiValues, err := strategy.CalculateRSI(closes, 14)
		if err != nil {
			rsiValues = []float64{50}
		}
		rsiValue := rsiValues[len(rsiValues)-1]

		if len(rsiValues) > 0 && len(bars) >= 14 {
			startIdx := len(bars) - len(rsiValues)
			for i, rsi := range rsiValues {
				barIdx := startIdx + i
				if barIdx >= 0 && barIdx < len(bars) {
					timestamp, _ := time.Parse(time.RFC3339, bars[barIdx].Timestamp)
					db.SaveRSI(symbol, timestamp, rsi)
				}
			}
		}

		atrValue := scoring.CalculateATRFromBars(bars)
		atrCategory := scoring.CategorizeATRValue(atrValue, bars)

		if len(bars) > 0 {
			latestTimestamp, _ := time.Parse(time.RFC3339, bars[len(bars)-1].Timestamp)
			db.SaveATR(symbol, latestTimestamp, atrValue)
		}

		whaleEvents := strategy.DetectWhales("", bars)
		whaleCount := len(whaleEvents)

		scoringInput, err := scoring.BuildScoringInput(bars, vwapPrice, rsiValue, whaleCount, atrValue, atrCategory)
		if err != nil {
			continue
		}

		weights := cfg.Profiles[profileName].SignalWeights
		score := strategy.CalculateInterestScore(scoringInput, weights)

		err = q.UpdateWatchlistScore(ctx, database.UpdateWatchlistScoreParams{
			Score:  float32(score),
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
	symbols, err := strategy.GetTradableAssets()
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

	for i := offset; i < end; i++ {
		symbol := symbols[i]

		bars, err := db.GetAlpacaBars(symbol, "1Day", 100, "")
		if err != nil {
			continue
		}

		if len(bars) == 0 {
			continue
		}

		weights := cfg.Profiles[profileName].SignalWeights
		candidate, err := analyzer.CalculateCandidateMetrics(ctx, symbol, bars, cfg, weights)
		if err != nil {
			continue
		}

		if candidate.Score >= minScore {
			candidates = append(candidates, *candidate)
		}
	}

	return candidates, totalSymbols, nil
}
