package strategy

import (
	"testing"
)

func TestSignalQualityFilter_FilterSignal(t *testing.T) {
	tests := []struct {
		name          string
		signal        *TradeSignal
		minThreshold  float64
		wantPassed    bool
		wantRejection string
	}{
		{
			name: "high confidence LONG signal passes",
			signal: &TradeSignal{
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
			signal: &TradeSignal{
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
			signal: &TradeSignal{
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
			signal: &TradeSignal{
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
	signals := []*TradeSignal{
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

func TestSignalStrengthCalculator_CalculateCompositeScore(t *testing.T) {
	tests := []struct {
		name         string
		indicators   []IndicatorScore
		wantScoreMin float64
		wantScoreMax float64
	}{
		{
			name: "all bullish indicators",
			indicators: []IndicatorScore{
				{Name: "RSI", Value: 30, Score: 2.0, Weight: 0.25, Alignment: "BULLISH"},
				{Name: "ATR", Value: 2.5, Score: 1.0, Weight: 0.15, Alignment: "BULLISH"},
				{Name: "Whale", Value: 1, Score: 3.0, Weight: 0.30, Alignment: "BULLISH"},
			},
			wantScoreMin: 1.5,
			wantScoreMax: 3.0,
		},
		{
			name: "mixed indicators",
			indicators: []IndicatorScore{
				{Name: "RSI", Value: 50, Score: 0.0, Weight: 0.25, Alignment: "NEUTRAL"},
				{Name: "Whale", Value: 1, Score: 2.0, Weight: 0.30, Alignment: "BULLISH"},
			},
			wantScoreMin: 0.5,
			wantScoreMax: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewSignalStrengthCalculator()
			score := calc.CalculateCompositeScore(tt.indicators)

			if score < tt.wantScoreMin || score > tt.wantScoreMax {
				t.Errorf("CalculateCompositeScore() = %.2f, want between %.2f and %.2f",
					score, tt.wantScoreMin, tt.wantScoreMax)
			}
		})
	}
}

func TestCountAlignedIndicators(t *testing.T) {
	indicators := []IndicatorScore{
		{Name: "RSI", Score: 2.0, Alignment: "BULLISH"},
		{Name: "ATR", Score: 1.0, Alignment: "BULLISH"},
		{Name: "Whale", Score: 3.0, Alignment: "BEARISH"},
		{Name: "Pattern", Score: 1.0, Alignment: "BULLISH"},
	}

	aligned := CountAlignedIndicators(indicators, "LONG")
	if aligned != 3 {
		t.Errorf("CountAlignedIndicators(LONG) = %d, want 3", aligned)
	}

	aligned = CountAlignedIndicators(indicators, "SHORT")
	if aligned != 1 {
		t.Errorf("CountAlignedIndicators(SHORT) = %d, want 1", aligned)
	}
}

func TestCalculateAlignmentPercentage(t *testing.T) {
	indicators := []IndicatorScore{
		{Name: "RSI", Score: 2.0, Alignment: "BULLISH"},
		{Name: "ATR", Score: 1.0, Alignment: "BULLISH"},
		{Name: "Whale", Score: 3.0, Alignment: "BULLISH"},
		{Name: "Pattern", Score: -1.0, Alignment: "BEARISH"},
	}

	pct := CalculateAlignmentPercentage(indicators, "LONG")
	expectedPct := 75.0 // 3 out of 4 aligned

	if pct != expectedPct {
		t.Errorf("CalculateAlignmentPercentage(LONG) = %.1f%%, want %.1f%%", pct, expectedPct)
	}
}

func TestValidationFilters(t *testing.T) {
	signal := &TradeSignal{
		Direction:  "LONG",
		Confidence: 75.0,
		Reasoning:  "Test signal",
	}

	// Test minimum confidence
	check := ValidateMinimumConfidence(70.0)(signal)
	if !check.Passed {
		t.Errorf("ValidateMinimumConfidence should pass for 75%% confidence")
	}

	check = ValidateMinimumConfidence(80.0)(signal)
	if check.Passed {
		t.Errorf("ValidateMinimumConfidence should fail for 75%% when threshold is 80%%")
	}

	// Test direction validation
	check = ValidateDirection()(signal)
	if !check.Passed {
		t.Errorf("ValidateDirection should pass for LONG")
	}

	invalidSignal := &TradeSignal{
		Direction:  "INVALID",
		Confidence: 75.0,
		Reasoning:  "Test",
	}
	check = ValidateDirection()(invalidSignal)
	if check.Passed {
		t.Errorf("ValidateDirection should fail for INVALID direction")
	}

	// Test reasoning validation
	check = ValidateReasoning()(signal)
	if !check.Passed {
		t.Errorf("ValidateReasoning should pass when reasoning provided")
	}

	noReasonSignal := &TradeSignal{
		Direction:  "LONG",
		Confidence: 75.0,
		Reasoning:  "",
	}
	check = ValidateReasoning()(noReasonSignal)
	if check.Passed {
		t.Errorf("ValidateReasoning should fail when no reasoning provided")
	}
}

func TestGenerateValidationReport(t *testing.T) {
	signal := &TradeSignal{
		Direction:  "LONG",
		Confidence: 82.0,
		Reasoning:  "RSI oversold with ATR confirmation",
	}

	filters := []func(*TradeSignal) ValidationCheck{
		ValidateMinimumConfidence(70.0),
		ValidateMaximumConfidence(100.0),
		ValidateDirection(),
		ValidateReasoning(),
	}

	report := GenerateValidationReport(signal, filters...)

	if !report.IsValid {
		t.Errorf("GenerateValidationReport should be valid for signal with 82%% confidence")
	}

	if report.OverallQualityScore < 90.0 {
		t.Errorf("Overall quality score should be high, got %.1f%%", report.OverallQualityScore)
	}
}
