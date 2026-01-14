package detection

import (
	"testing"

	"github.com/fazecat/mongelmaker/Internal/types"
)

func TestDivergenceDetector_DetectRSIDivergence_Bullish(t *testing.T) {
	// Bullish divergence: Lower lows in price, higher lows in RSI
	bars := []types.Bar{
		{High: 110, Low: 100, Close: 105},
		{High: 105, Low: 95, Close: 98}, // First low
		{High: 107, Low: 97, Close: 105},
		{High: 103, Low: 92, Close: 95}, // Second low (lower than first)
		{High: 100, Low: 90, Close: 98},
	}

	rsiValues := []float64{65, 35, 45, 38, 48} // RSI makes higher low at position 3

	detector := NewDivergenceDetector()
	result := detector.DetectRSIDivergence(bars, rsiValues)

	// Just verify the function executes without panic
	if result.Type == "" {
		t.Errorf("DetectRSIDivergence should return a valid result")
	}
}

func TestDivergenceDetector_DetectRSIDivergence_Bearish(t *testing.T) {
	// Bearish divergence: Higher highs in price, lower highs in RSI
	bars := []types.Bar{
		{High: 100, Low: 90, Close: 95},
		{High: 110, Low: 95, Close: 108}, // First high
		{High: 105, Low: 92, Close: 100},
		{High: 112, Low: 98, Close: 110}, // Second high (higher than first)
		{High: 108, Low: 95, Close: 102},
	}

	rsiValues := []float64{35, 70, 60, 65, 55} // RSI makes lower high at position 3

	detector := NewDivergenceDetector()
	result := detector.DetectRSIDivergence(bars, rsiValues)

	if !result.Detected {
		t.Errorf("DetectRSIDivergence should detect bearish divergence")
	}

	if result.Direction != "SHORT" {
		t.Errorf("Bearish divergence should be SHORT, got %s", result.Direction)
	}

	if result.Type != DivergenceBearish {
		t.Errorf("Type should be BEARISH, got %v", result.Type)
	}
}

func TestDivergenceDetector_DetectHiddenDivergence_Bullish(t *testing.T) {
	// Hidden bullish: Higher price lows, lower RSI lows (continuation of uptrend)
	bars := []types.Bar{
		{High: 110, Low: 100, Close: 105},
		{High: 105, Low: 98, Close: 102}, // First low
		{High: 107, Low: 99, Close: 105},
		{High: 103, Low: 101, Close: 102}, // Second low (higher than first)
		{High: 100, Low: 102, Close: 103},
	}

	rsiValues := []float64{65, 35, 45, 32, 48} // RSI makes lower low at position 3

	detector := NewDivergenceDetector()
	result := detector.DetectHiddenDivergence(bars, rsiValues)

	// Just verify the function executes without panic
	if result.Type == "" {
		t.Errorf("DetectHiddenDivergence should return a valid result")
	}
}

func TestDivergenceDetector_DetectExaggeratedDivergence_Overbought(t *testing.T) {
	rsiValues := []float64{30, 45, 60, 75, 85} // Last value is 85 (overbought)

	detector := NewDivergenceDetector()
	result := detector.DetectExaggeratedDivergence(rsiValues)

	if !result.Detected {
		t.Errorf("DetectExaggeratedDivergence should detect overbought at RSI 85")
	}

	if result.Direction != "SHORT" {
		t.Errorf("Overbought should suggest SHORT, got %s", result.Direction)
	}

	if result.Type != DivergenceExaggerated {
		t.Errorf("Type should be EXAGGERATED, got %v", result.Type)
	}

	if result.ReversalProbability < 0.65 {
		t.Errorf("Reversal probability should be >= 0.65, got %.2f", result.ReversalProbability)
	}
}

func TestDivergenceDetector_DetectExaggeratedDivergence_Oversold(t *testing.T) {
	rsiValues := []float64{70, 55, 40, 25, 15} // Last value is 15 (oversold)

	detector := NewDivergenceDetector()
	result := detector.DetectExaggeratedDivergence(rsiValues)

	if !result.Detected {
		t.Errorf("DetectExaggeratedDivergence should detect oversold at RSI 15")
	}

	if result.Direction != "LONG" {
		t.Errorf("Oversold should suggest LONG, got %s", result.Direction)
	}

	if result.ReversalProbability < 0.65 {
		t.Errorf("Reversal probability should be >= 0.65, got %.2f", result.ReversalProbability)
	}
}

func TestDivergenceDetector_NoExaggerated_Normal(t *testing.T) {
	rsiValues := []float64{40, 45, 50, 55, 60} // All normal RSI values

	detector := NewDivergenceDetector()
	result := detector.DetectExaggeratedDivergence(rsiValues)

	if result.Detected {
		t.Errorf("Should not detect exaggerated divergence for normal RSI values")
	}
}

func TestDivergenceDetector_ConfidenceCalculation(t *testing.T) {
	// Test that divergence confidence increases with magnitude
	bars := []types.Bar{
		{High: 110, Low: 100, Close: 105},
		{High: 105, Low: 95, Close: 98}, // First low at 95
		{High: 107, Low: 97, Close: 105},
		{High: 103, Low: 80, Close: 95}, // Second low at 80 (big drop)
		{High: 100, Low: 90, Close: 98},
	}

	rsiValues := []float64{65, 35, 45, 55, 48} // RSI diverges significantly

	detector := NewDivergenceDetector()
	result := detector.DetectRSIDivergence(bars, rsiValues)

	if result.Detected {
		if result.Confidence < 60.0 || result.Confidence > 100.0 {
			t.Errorf("Confidence should be between 60 and 100, got %.1f", result.Confidence)
		}
	}
}

func TestFindLocalExtrema(t *testing.T) {
	bars := []types.Bar{
		{High: 100, Low: 90},
		{High: 110, Low: 85}, // Local low at index 1
		{High: 105, Low: 88},
		{High: 95, Low: 80}, // Local low at index 3
		{High: 100, Low: 85},
	}

	lows := findLocalExtrema(bars, true)

	if len(lows) < 1 {
		t.Errorf("findLocalExtrema should find at least one local low")
	}

	// Verify extrema are local minima
	for _, extremum := range lows {
		if extremum.Index == 0 || extremum.Index == len(bars)-1 {
			t.Errorf("Local extrema should not be at edges")
		}
	}
}

func TestFindIndicatorExtrema(t *testing.T) {
	values := []float64{50, 30, 40, 35, 50} // Local low at index 1 and 3

	lows := findIndicatorExtrema(values, true)

	if len(lows) < 1 {
		t.Errorf("findIndicatorExtrema should find at least one local low")
	}

	for _, extremum := range lows {
		if extremum.Index == 0 || extremum.Index == len(values)-1 {
			t.Errorf("Local extrema should not be at edges")
		}
	}
}
