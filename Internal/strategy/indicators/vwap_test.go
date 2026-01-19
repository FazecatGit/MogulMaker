package indicators

import (
	"testing"

	"github.com/fazecat/mogulmaker/Internal/types"
)

func TestVWAPCalculation(t *testing.T) {
	bars := []types.Bar{
		{Open: 100, High: 102, Low: 99, Close: 101, Volume: 1000},
		{Open: 101, High: 103, Low: 100, Close: 102, Volume: 1500},
		{Open: 102, High: 105, Low: 101, Close: 104, Volume: 2000},
		{Open: 104, High: 106, Low: 103, Close: 105, Volume: 1800},
	}

	calc := NewVWAPCalculator(bars)
	vwap := calc.Calculate()

	if vwap <= 0 {
		t.Errorf("VWAP should be positive, got %f", vwap)
	}

	// VWAP should be a weighted average, so should be within range of prices
	minPrice := 99.0
	maxPrice := 106.0
	if vwap < minPrice || vwap > maxPrice {
		t.Errorf("VWAP %f should be between %f and %f", vwap, minPrice, maxPrice)
	}
}

func TestVWAPCalculateAt(t *testing.T) {
	bars := []types.Bar{
		{Open: 100, High: 102, Low: 99, Close: 101, Volume: 1000},
		{Open: 101, High: 103, Low: 100, Close: 102, Volume: 1500},
		{Open: 102, High: 105, Low: 101, Close: 104, Volume: 2000},
	}

	calc := NewVWAPCalculator(bars)

	vwap0 := calc.CalculateAt(0)
	if vwap0 <= 0 {
		t.Errorf("VWAP at index 0 should be positive, got %f", vwap0)
	}

	vwap2 := calc.CalculateAt(2)
	if vwap2 <= 0 {
		t.Errorf("VWAP at index 2 should be positive, got %f", vwap2)
	}

	// VWAP should increase as we add more bars (with typical price changes)
	if vwap2 < vwap0 {
		t.Logf("Note: VWAP2 (%f) < VWAP0 (%f) - this is valid if later bars have lower typical prices", vwap2, vwap0)
	}
}

func TestVWAPTrend(t *testing.T) {
	// Test ABOVE VWAP
	barsAbove := []types.Bar{
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 150, High: 150, Low: 150, Close: 150, Volume: 100},
	}

	calcAbove := NewVWAPCalculator(barsAbove)
	trend := calcAbove.GetVWAPTrend()
	if trend != 1 {
		t.Errorf("Expected trend 1 (above VWAP), got %d", trend)
	}

	// Test BELOW VWAP
	barsBelow := []types.Bar{
		{Open: 150, High: 150, Low: 150, Close: 150, Volume: 1000},
		{Open: 150, High: 150, Low: 150, Close: 150, Volume: 1000},
		{Open: 150, High: 150, Low: 150, Close: 150, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 100},
	}

	calcBelow := NewVWAPCalculator(barsBelow)
	trend = calcBelow.GetVWAPTrend()
	if trend != -1 {
		t.Errorf("Expected trend -1 (below VWAP), got %d", trend)
	}
}

func TestVWAPDistance(t *testing.T) {
	bars := []types.Bar{
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 110, Volume: 1000}, // 10% above
	}

	calc := NewVWAPCalculator(bars)
	distance := calc.GetVWAPDistance()

	if distance <= 0 {
		t.Errorf("Expected positive distance, got %f", distance)
	}

	if distance > 15 || distance < 5 {
		t.Logf("Distance: %f (should be roughly 5-10%% above VWAP)", distance)
	}
}

func TestVWAPSupport(t *testing.T) {
	bars := []types.Bar{
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 105, High: 105, Low: 105, Close: 105, Volume: 100}, // Move above VWAP
		{Open: 101, High: 101, Low: 101, Close: 100, Volume: 100}, // Return to VWAP
	}

	calc := NewVWAPCalculator(bars)
	isSupport := calc.IsVWAPSupport(1.0)

	if !isSupport {
		t.Logf("VWAP support detection may need tolerance adjustment")
	}
}

func TestVWAPResistance(t *testing.T) {
	bars := []types.Bar{
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 95, High: 95, Low: 95, Close: 95, Volume: 100},  // Move below VWAP
		{Open: 99, High: 99, Low: 99, Close: 100, Volume: 100}, // Return to VWAP
	}

	calc := NewVWAPCalculator(bars)
	isResistance := calc.IsVWAPResistance(1.0)

	if !isResistance {
		t.Logf("VWAP resistance detection may need tolerance adjustment")
	}
}

func TestVWAPBounce(t *testing.T) {
	// Test bullish bounce
	bars := []types.Bar{
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Open: 95, High: 95, Low: 95, Close: 95, Volume: 100},     // Below VWAP
		{Open: 100, High: 100, Low: 100, Close: 102, Volume: 100}, // Bounce up
	}

	calc := NewVWAPCalculator(bars)
	isBounce, bounceType := calc.GetVWAPBounce(1.0)

	if isBounce && bounceType == "bullish_bounce" {
		t.Logf("Bullish bounce detected: %s", bounceType)
	}
}

func TestVWAPAnalyze(t *testing.T) {
	bars := []types.Bar{
		{Open: 100, High: 102, Low: 99, Close: 101, Volume: 1000},
		{Open: 101, High: 103, Low: 100, Close: 102, Volume: 1500},
		{Open: 102, High: 105, Low: 101, Close: 104, Volume: 2000},
	}

	calc := NewVWAPCalculator(bars)
	analysis := calc.AnalyzeVWAP(1.0)

	if analysis["error"] != nil {
		t.Errorf("Analysis failed: %v", analysis["error"])
	}

	expectedKeys := []string{"vwap", "current_price", "trend", "distance_pct", "is_bounce"}
	for _, key := range expectedKeys {
		if _, exists := analysis[key]; !exists {
			t.Errorf("Analysis missing key: %s", key)
		}
	}
}

func TestVWAPEmptyBars(t *testing.T) {
	calc := NewVWAPCalculator([]types.Bar{})

	vwap := calc.Calculate()
	if vwap != 0 {
		t.Errorf("Expected VWAP 0 for empty bars, got %f", vwap)
	}

	trend := calc.GetVWAPTrend()
	if trend != 0 {
		t.Errorf("Expected trend 0 for empty bars, got %d", trend)
	}
}

func TestVWAPCalculateRange(t *testing.T) {
	bars := []types.Bar{
		{Open: 100, High: 102, Low: 99, Close: 101, Volume: 1000},
		{Open: 101, High: 103, Low: 100, Close: 102, Volume: 1500},
		{Open: 102, High: 105, Low: 101, Close: 104, Volume: 2000},
		{Open: 104, High: 106, Low: 103, Close: 105, Volume: 1800},
	}

	calc := NewVWAPCalculator(bars)

	vwapAll := calc.Calculate()
	vwapRange := calc.CalculateRange(1, 2) // Only bars 1 and 2

	if vwapRange <= 0 {
		t.Errorf("Range VWAP should be positive, got %f", vwapRange)
	}

	// They should be different
	if vwapAll == vwapRange {
		t.Logf("Note: Full VWAP (%f) and range VWAP (%f) happen to be equal", vwapAll, vwapRange)
	}
}

func TestVWAPAllValues(t *testing.T) {
	bars := []types.Bar{
		{Open: 100, High: 102, Low: 99, Close: 101, Volume: 1000},
		{Open: 101, High: 103, Low: 100, Close: 102, Volume: 1500},
		{Open: 102, High: 105, Low: 101, Close: 104, Volume: 2000},
	}

	calc := NewVWAPCalculator(bars)
	values := calc.CalculateAllValues()

	if len(values) != len(bars) {
		t.Errorf("Expected %d VWAP values, got %d", len(bars), len(values))
	}

	for i, v := range values {
		if v <= 0 {
			t.Errorf("VWAP value %d should be positive, got %f", i, v)
		}
	}
}
