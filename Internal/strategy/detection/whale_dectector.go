package detection

import (
	"github.com/fazecat/mongelmaker/Internal/types"
	"github.com/fazecat/mongelmaker/Internal/utils"
)

type VolumeStats struct {
	Timestamp   string
	TotalVolume int64
	MeanVolume  float64
	StdDev      float64
	Zscore      float64
	IsAnomalous bool
}

type WhaleEvent struct {
	Timestamp   string
	Symbol      string
	Direction   string // buy or sell
	Volume      int64
	ZScore      float64
	ClosePrice  float64
	PriceChange float64
	Conviction  string
}

func CalculateVolumeStats(volumes []int64) (mean float64, stdDev float64) {
	floatVolumes := make([]float64, len(volumes))
	for i, v := range volumes {
		floatVolumes[i] = float64(v)
	}

	mean = utils.Average(floatVolumes)
	stdDev = utils.StandardDeviation(floatVolumes)

	return mean, stdDev
}

func CalculateZScore(currentVolume int64, meanVolume float64, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}

	zScore := (float64(currentVolume) - meanVolume) / stdDev
	return zScore
}

func DetectWhales(symbol string, bars []types.Bar) []WhaleEvent {
	whales := make([]WhaleEvent, 0)

	if len(bars) < 20 {
		return whales
	}
	for i := 20; i < len(bars); i++ {
		currentBar := bars[i]

		historicalBars := bars[i-20 : i]
		volumes := extractVolumes(historicalBars)

		meanVolume, stdDev := CalculateVolumeStats(volumes)

		zScore := CalculateZScore(currentBar.Volume, meanVolume, stdDev)

		if zScore > 2.0 {
			whale := createWhaleEvent(symbol, currentBar, zScore, meanVolume)
			whales = append(whales, whale)
		}
	}

	return whales
}

func extractVolumes(bars []types.Bar) []int64 {
	volumes := make([]int64, len(bars))

	for i, bar := range bars {
		volumes[i] = bar.Volume
	}
	return volumes
}

func createWhaleEvent(symbol string, bar types.Bar, zScore float64, meanVolume float64) WhaleEvent {
	direction := DetectDirection(bar)
	conviction := DetermineConviction(zScore)

	whaleEvent := WhaleEvent{
		Timestamp:   bar.Timestamp,
		Symbol:      symbol,
		Direction:   direction, // "BUY" or "SELL"
		Volume:      bar.Volume,
		ZScore:      zScore,
		ClosePrice:  bar.Close,
		PriceChange: 0,
		Conviction:  conviction,
	}

	return whaleEvent

}
func DetectDirection(bar types.Bar) string {
	if bar.Close > bar.Open {
		return "BUY"
	} else if bar.Close < bar.Open {
		return "SELL"
	}
	return "NEUTRAL"
}

func DetermineConviction(zScore float64) string {
	if zScore > 3.0 {
		return "HIGH"
	} else if zScore > 2.0 {
		return "MEDIUM"
	}
	return "LOW"
}
