package indicators

import (
	"testing"

	"github.com/fazecat/mogulmaker/Internal/utils"
)

func TestCalculateATR(t *testing.T) {
	atrBars := []ATRBar{
		{High: 15.0, Low: 10.0, Close: 12.0},
		{High: 16.0, Low: 11.0, Close: 15.0},
		{High: 18.0, Low: 14.0, Close: 17.0},
		{High: 50.0, Low: 20.0, Close: 35.0},
		{High: 25.0, Low: 12.0, Close: 20.0},
	}
	period := 3
	expectedATR := []float64{0, 0, 0, 14.0, 20.0}

	atrValues, err := CalculateATR(atrBars, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, expected := range expectedATR {
		if utils.Abs(atrValues[i]-expected) > 1e-5 {
			t.Errorf("unexpected ATR value at index %d: got %v, want %v", i, atrValues[i], expected)
		}
	}
}

func TestCalculateTrueRange(t *testing.T) {
	tests := []struct {
		high      float64
		low       float64
		prevClose float64
		expected  float64
	}{
		{high: 15.0, low: 10.0, prevClose: 12.0, expected: 5.0},
		{high: 16.0, low: 11.0, prevClose: 15.0, expected: 5.0},
		{high: 18.0, low: 14.0, prevClose: 17.0, expected: 4.0},
		{high: 20.0, low: 15.0, prevClose: 19.0, expected: 5.0},
		{high: 22.0, low: 18.0, prevClose: 21.0, expected: 4.0},
	}
	for _, tt := range tests {
		tr := CalculateTrueRange(tt.high, tt.low, tt.prevClose)
		if utils.Abs(tr-tt.expected) > 1e-5 {
			t.Errorf("CalculateTrueRange(%v, %v, %v) = %v; want %v", tt.high, tt.low, tt.prevClose, tr, tt.expected)
		}
	}
}

func TestCalculateTrueRangeWithGaps(t *testing.T) {
	tests := []struct {
		high      float64
		low       float64
		prevClose float64
		expected  float64
	}{
		// Gap up
		{high: 18.0, low: 14.0, prevClose: 10.0, expected: 8.0},
		// Gap down
		{high: 12.0, low: 8.0, prevClose: 15.0, expected: 7.0},
		// No gap
		{high: 16.0, low: 12.0, prevClose: 14.0, expected: 4.0},
	}
	for _, tt := range tests {
		tr := CalculateTrueRange(tt.high, tt.low, tt.prevClose)
		if utils.Abs(tr-tt.expected) > 1e-5 {
			t.Errorf("CalculateTrueRange(%v, %v, %v) = %v; want %v", tt.high, tt.low, tt.prevClose, tr, tt.expected)
		}
	}
}

func TestCalculateATR_HighVolatility(t *testing.T) {
	atrBars := []ATRBar{
		// Large ranges, big movements
		{High: 110.0, Low: 100.0, Close: 105.0},
		{High: 120.0, Low: 105.0, Close: 118.0},
		{High: 125.0, Low: 112.0, Close: 115.0},
		{High: 130.0, Low: 115.0, Close: 125.0},
	}
	period := 3
	expectedATR := []float64{0, 0, 0, 14.333333333}
	atrValues, err := CalculateATR(atrBars, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, expected := range expectedATR {
		if utils.Abs(atrValues[i]-expected) > 1e-5 {
			t.Errorf("unexpected ATR value at index %d: got %v, want %v", i, atrValues[i], expected)
		}
	}
}

func TestCalculateATR_LowVolatility(t *testing.T) {
	atrBars := []ATRBar{
		// Small ranges, consistent movement
		{High: 100.5, Low: 100.0, Close: 100.2},
		{High: 100.8, Low: 100.3, Close: 100.6},
		{High: 101.0, Low: 100.5, Close: 100.8},
		{High: 101.2, Low: 100.7, Close: 101.0},
	}
	period := 3
	expectedATR := []float64{0, 0, 0, 0.5333333333}
	atrValues, err := CalculateATR(atrBars, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, expected := range expectedATR {
		if utils.Abs(atrValues[i]-expected) > 1e-5 {
			t.Errorf("unexpected ATR value at index %d: got %v, want %v", i, atrValues[i], expected)
		}
	}
}

func TestCalculateATR_InsufficientData(t *testing.T) {
	atrBars := []ATRBar{
		{High: 100.0, Low: 95.0, Close: 98.0},
		{High: 102.0, Low: 97.0, Close: 100.0},
		{High: 101.0, Low: 96.0, Close: 99.0},
		{High: 103.0, Low: 98.0, Close: 101.0},
	}
	period := 14 // Not enough bars!

	_, err := CalculateATR(atrBars, period)
	if err == nil {
		t.Error("expected error for insufficient data, got nil")
	}
}
