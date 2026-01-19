package signals

import (
	"testing"

	"github.com/fazecat/mongelmaker/Internal/types"
)

func TestSignalQualityFilter_FilterSignal(t *testing.T) {
	tests := []struct {
		name          string
		signal        *types.TradeSignal
		minThreshold  float64
		wantPassed    bool
		wantRejection string
	}{
		{
			name: "high confidence LONG signal passes",
			signal: &types.TradeSignal{
				Direction:  "LONG",
				Confidence: 85.0,
				Reasoning:  "Strong RSI oversold",
			},
			minThreshold:  70.0,
			wantPassed:    true,
			wantRejection: "",
		},
		{
			name: "low confidence signal rejected",
			signal: &types.TradeSignal{
				Direction:  "LONG",
				Confidence: 50.0,
				Reasoning:  "Weak signal",
			},
			minThreshold:  70.0,
			wantPassed:    false,
			wantRejection: "below minimum threshold",
		},
		{
			name: "signal with no reasoning rejected",
			signal: &types.TradeSignal{
				Direction:  "SHORT",
				Confidence: 80.0,
				Reasoning:  "",
			},
			minThreshold:  70.0,
			wantPassed:    false,
			wantRejection: "no reasoning",
		},
		{
			name: "signal with invalid direction rejected",
			signal: &types.TradeSignal{
				Direction:  "INVALID",
				Confidence: 85.0,
				Reasoning:  "Test",
			},
			minThreshold:  70.0,
			wantPassed:    false,
			wantRejection: "Invalid direction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &SignalQualityFilter{
				MinConfidenceThreshold: tt.minThreshold,
				MaxConfidenceThreshold: 100.0,
			}

			result := filter.FilterSignal(tt.signal)

			if result.Passed != tt.wantPassed {
				t.Errorf("FilterSignal() Passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if !tt.wantPassed && tt.wantRejection != "" {
				if result.FailureReason == "" {
					t.Errorf("Expected failure reason containing '%s', got empty", tt.wantRejection)
				}
			}
		})
	}
}

func TestSignalQualityFilter_GetHighestConfidenceSignal(t *testing.T) {
	signals := []*types.TradeSignal{
		{
			Direction:  "LONG",
			Confidence: 65.0, // Below threshold
			Reasoning:  "Weak signal",
		},
		{
			Direction:  "LONG",
			Confidence: 85.0, // Above threshold
			Reasoning:  "Strong signal",
		},
		{
			Direction:  "SHORT",
			Confidence: 92.0, // Highest
			Reasoning:  "Very strong signal",
		},
	}

	filter := NewSignalQualityFilter()
	best := filter.GetHighestConfidenceSignal(signals)

	if !best.Passed {
		t.Errorf("Expected signal to pass filter")
	}

	if best.Original.Confidence != 92.0 {
		t.Errorf("GetHighestConfidenceSignal() returned confidence %.1f, want 92.0", best.Original.Confidence)
	}

	if best.Original.Direction != "SHORT" {
		t.Errorf("GetHighestConfidenceSignal() returned direction %s, want SHORT", best.Original.Direction)
	}
}
