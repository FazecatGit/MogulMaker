package indicators

import "github.com/fazecat/mogulmaker/Internal/types"

type PriceLevel struct {
	Price      float64
	BouncCount int
	Strength   float64
}

func FindSupport(bars []types.Bar) float64 {
	if len(bars) < 3 {
		return 0
	}
	lowestLow := bars[0].Low
	for _, bar := range bars {
		if bar.Low < lowestLow {
			lowestLow = bar.Low
		}
	}
	return lowestLow
}

func FindResistance(bars []types.Bar) float64 {
	if len(bars) < 3 {
		return 0
	}
	highestHigh := bars[0].High
	for _, bar := range bars {
		if bar.High > highestHigh {
			highestHigh = bar.High
		}
	}
	return highestHigh
}

func GetSupportLevels(bars []types.Bar) []PriceLevel {
	levels := []PriceLevel{}

	for i := 1; i < len(bars)-1; i++ {
		if bars[i].Low < bars[i-1].Low && bars[i].Low < bars[i+1].Low {
			levels = append(levels, PriceLevel{
				Price:      bars[i].Low,
				BouncCount: 1,
			})
		}
	}
	return levels
}

func GetResistanceLevels(bars []types.Bar) []PriceLevel {
	levels := []PriceLevel{}

	for i := 1; i < len(bars)-1; i++ {
		if bars[i].High > bars[i-1].High && bars[i].High > bars[i+1].High {
			levels = append(levels, PriceLevel{
				Price:      bars[i].High,
				BouncCount: 1,
			})
		}
	}
	return levels
}

func IsAtSupport(currentPrice float64, support float64) bool {
	tolerance := support * 0.01
	return currentPrice >= support-tolerance && currentPrice <= support+tolerance
}

func IsAtResistance(currentPrice float64, resistance float64) bool {
	tolerance := resistance * 0.01
	return currentPrice >= resistance-tolerance && currentPrice <= resistance+tolerance
}

func DistanceToSupport(currentPrice float64, support float64) float64 {
	if support == 0 {
		return 0
	}
	return ((currentPrice - support) / support) * 100
}

func DistanceToResistance(currentPrice float64, resistance float64) float64 {
	if resistance == 0 {
		return 0
	}
	return ((resistance - currentPrice) / resistance) * 100
}

func FindPivotPoint(bars []types.Bar) float64 {
	if len(bars) == 0 {
		return 0
	}

	latestBar := bars[0]
	return (latestBar.High + latestBar.Low + latestBar.Close) / 3
}

func IsBreakoutAboveResistance(currentPrice float64, resistance float64) bool {
	if resistance == 0 {
		return false
	}
	return currentPrice > resistance*1.005 // 0.5% above resistance = breakout
}

func IsBreakoutBelowSupport(currentPrice float64, support float64) bool {
	if support == 0 {
		return false
	}
	return currentPrice < support*0.995 // 0.5% below support = breakdown
}
