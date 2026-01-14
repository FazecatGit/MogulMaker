package indicators

import (
	"testing"

	datafeed "github.com/fazecat/mongelmaker/Internal/database"
)

func TestRSICalculation(t *testing.T) {
	tests := []struct {
		name     string
		closes   []float64
		period   int
		wantErr  bool
		checkRSI func([]float64) bool
	}{
		{
			name:    "rising prices should give high RSI",
			closes:  []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114},
			period:  14,
			wantErr: false,
			checkRSI: func(rsi []float64) bool {
				// Last RSI should be high (approaching 100)
				return rsi[len(rsi)-1] > 70
			},
		},
		{
			name:    "falling prices should give low RSI",
			closes:  []float64{114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100},
			period:  14,
			wantErr: false,
			checkRSI: func(rsi []float64) bool {
				// Last RSI should be low (approaching 0)
				return rsi[len(rsi)-1] < 30
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsi, err := CalculateRSI(tt.closes, tt.period)

			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateRSI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkRSI(rsi) {
				t.Errorf("RSI values don't meet expectations: %v", rsi)
			}
		})
	}
}
func TestDetermineRSISignal(t *testing.T) {
	tests := []struct {
		rsi      float64
		expected string
	}{
		{25, "oversold"},     // Below 30
		{75, "overbought"},   // Above 70
		{50, "neutral"},      // Between 30-70
		{29.9, "oversold"},   // Edge case
		{70.1, "overbought"}, // Edge case
	}

	for _, test := range tests {
		result := DetermineRSISignal(test.rsi)
		if result != test.expected {
			t.Errorf("For RSI %f, expected %q but got %q", test.rsi, test.expected, result)
		}
	}
}
func TestCalculateAndStoreRSI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	// Setup database connection first
	err := datafeed.InitDatabase()
	if err != nil {
		t.Skip("Database not available:", err)
	}
	defer datafeed.CloseDatabase()

	symbol := "AAPL"
	timeframe := "1Day"
	limit := 100

	bars, err := datafeed.GetAlpacaBars(symbol, timeframe, limit, "")
	if err != nil {
		t.Errorf("GetAlpacaBars() error = %v", err)
		return
	}

	err = datafeed.CalculateAndStoreRSI(symbol, bars)
	if err != nil {
		t.Errorf("CalculateAndStoreRSI() error = %v", err)
	}
}

func TestEdgeCases(t *testing.T) {
	// Test with insufficient data
	closes := []float64{100, 102}
	period := 5
	_, err := CalculateRSI(closes, period)
	if err == nil {
		t.Error("Expected error for insufficient data, but got none")
	}
	// Test with constant prices
	closes = []float64{100, 100, 100, 100, 100, 100}
	rsi, err := CalculateRSI(closes, period)
	if err != nil {
		t.Error("Error calculating RSI for constant prices:", err)
		return
	}
	t.Log("RSI values for constant prices:", rsi)
}

func TestRSIMultipleValues(t *testing.T) {
	// Simple test data
	closes := []float64{100, 102, 101, 103, 102, 104, 103, 105, 104, 106, 105, 107, 106, 108, 107}
	period := 5

	rsi, err := CalculateRSI(closes, period)
	if err != nil {
		t.Fatal(err)
	}

	// Check that RSI values CHANGE over time
	firstRSI := rsi[period]
	lastRSI := rsi[len(rsi)-1]

	if firstRSI == lastRSI {
		t.Errorf("RSI should vary! Got same value %.2f for all positions", firstRSI)
	}

	// Log all RSI values to see the progression
	t.Logf("RSI values: %v", rsi[period:])
}
