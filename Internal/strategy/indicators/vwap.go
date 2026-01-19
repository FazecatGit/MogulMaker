package indicators

import (
	"fmt"

	"github.com/fazecat/mogulmaker/Internal/types"
)

// computes Volume Weighted Average Price
type VWAPCalculator struct {
	bars []types.Bar
}

// creates a new VWAP calculator
func NewVWAPCalculator(bars []types.Bar) *VWAPCalculator {
	return &VWAPCalculator{
		bars: bars,
	}
}

// returns the VWAP value for the full dataset
func (v *VWAPCalculator) Calculate() float64 {
	if len(v.bars) == 0 {
		return 0
	}
	return v.CalculateAt(len(v.bars) - 1)
}

// returns the VWAP value at a specific bar index
func (v *VWAPCalculator) CalculateAt(index int) float64 {
	if index < 0 || index >= len(v.bars) {
		return 0
	}

	typicalPrice := 0.0
	volume := 0.0

	for i := 0; i <= index; i++ {
		tp := v.typicalPrice(v.bars[i])
		typicalPrice += tp * float64(v.bars[i].Volume)
		volume += float64(v.bars[i].Volume)
	}

	if volume == 0 {
		return 0
	}

	return typicalPrice / volume
}

// returns VWAP over a specific range of bars
func (v *VWAPCalculator) CalculateRange(startIndex, endIndex int) float64 {
	if startIndex < 0 || endIndex >= len(v.bars) || startIndex > endIndex {
		return 0
	}

	typicalPrice := 0.0
	volume := 0.0

	for i := startIndex; i <= endIndex; i++ {
		tp := v.typicalPrice(v.bars[i])
		typicalPrice += tp * float64(v.bars[i].Volume)
		volume += float64(v.bars[i].Volume)
	}

	if volume == 0 {
		return 0
	}

	return typicalPrice / volume
}

// returns VWAP values for each bar in the dataset
func (v *VWAPCalculator) CalculateAllValues() []float64 {
	vwapValues := make([]float64, len(v.bars))

	for i := 0; i < len(v.bars); i++ {
		vwapValues[i] = v.CalculateAt(i)
	}

	return vwapValues
}

// returns whether price is above or below VWAP
func (v *VWAPCalculator) GetVWAPTrend() int {
	if len(v.bars) == 0 {
		return 0
	}

	currentPrice := v.bars[len(v.bars)-1].Close
	vwap := v.Calculate()

	if currentPrice > vwap {
		return 1
	} else if currentPrice < vwap {
		return -1
	}
	return 0
}

// returns the percentage distance from current price to VWAP
func (v *VWAPCalculator) GetVWAPDistance() float64 {
	if len(v.bars) == 0 {
		return 0
	}

	currentPrice := v.bars[len(v.bars)-1].Close
	vwap := v.Calculate()

	if vwap == 0 {
		return 0
	}

	return ((currentPrice - vwap) / vwap) * 100
}

// checks if price is touching VWAP from above with tolerance
func (v *VWAPCalculator) IsVWAPSupport(tolerance float64) bool {
	if len(v.bars) < 2 {
		return false
	}

	currentPrice := v.bars[len(v.bars)-1].Close
	previousPrice := v.bars[len(v.bars)-2].Close
	vwap := v.Calculate()

	// Price crossed below VWAP or touching it within tolerance
	return previousPrice >= vwap && currentPrice <= vwap*(1+tolerance/100)
}

// checks if price is touching VWAP from below with tolerance
func (v *VWAPCalculator) IsVWAPResistance(tolerance float64) bool {
	if len(v.bars) < 2 {
		return false
	}

	currentPrice := v.bars[len(v.bars)-1].Close
	previousPrice := v.bars[len(v.bars)-2].Close
	vwap := v.Calculate()

	// Price crossed above VWAP or touching it within tolerance
	return previousPrice <= vwap && currentPrice >= vwap*(1-tolerance/100)
}

// detects if price bounced off VWAP
// Returns true if: price touched VWAP and then moved away
func (v *VWAPCalculator) GetVWAPBounce(tolerance float64) (bool, string) {
	if len(v.bars) < 3 {
		return false, ""
	}

	current := v.bars[len(v.bars)-1].Close
	previous := v.bars[len(v.bars)-2].Close
	twoBack := v.bars[len(v.bars)-3].Close
	vwap := v.Calculate()

	// Bounce from below: price was below, touched VWAP, now above
	if twoBack < vwap && previous <= vwap*(1+tolerance/100) && current > vwap {
		return true, "bullish_bounce"
	}

	// Bounce from above: price was above, touched VWAP, now below
	if twoBack > vwap && previous >= vwap*(1-tolerance/100) && current < vwap {
		return true, "bearish_bounce"
	}

	return false, ""
}

// provides a comprehensive VWAP analysis
func (v *VWAPCalculator) AnalyzeVWAP(tolerance float64) map[string]interface{} {
	if len(v.bars) == 0 {
		return map[string]interface{}{
			"error": "no bars available",
		}
	}

	vwap := v.Calculate()
	currentPrice := v.bars[len(v.bars)-1].Close
	distance := v.GetVWAPDistance()
	trend := v.GetVWAPTrend()
	isBounce, bounceType := v.GetVWAPBounce(tolerance)

	var trendStr string
	switch trend {
	case 1:
		trendStr = "ABOVE (Bullish)"
	case -1:
		trendStr = "BELOW (Bearish)"
	default:
		trendStr = "AT VWAP (Neutral)"
	}

	return map[string]interface{}{
		"vwap":           fmt.Sprintf("%.2f", vwap),
		"current_price":  fmt.Sprintf("%.2f", currentPrice),
		"trend":          trendStr,
		"distance_pct":   fmt.Sprintf("%.2f%%", distance),
		"is_bounce":      isBounce,
		"bounce_type":    bounceType,
		"is_support":     v.IsVWAPSupport(tolerance),
		"is_resistance":  v.IsVWAPResistance(tolerance),
		"bars_processed": len(v.bars),
	}
}

// calculates (High + Low + Close) / 3
func (v *VWAPCalculator) typicalPrice(bar types.Bar) float64 {
	return (bar.High + bar.Low + bar.Close) / 3
}
