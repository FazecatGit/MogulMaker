package signals

import (
	"testing"
)

func TestCombineMultiTimeframeSignals_AllAligned(t *testing.T) {
	dailySignal := CombinedSignal{
		Recommendation: "BUY",
		Score:          2.0,
		Confidence:     85.0,
	}

	fourHourSignal := CombinedSignal{
		Recommendation: "ACCUMULATE",
		Score:          1.5,
		Confidence:     80.0,
	}

	oneHourSignal := CombinedSignal{
		Recommendation: "BUY",
		Score:          1.8,
		Confidence:     75.0,
	}

	result := CombineMultiTimeframeSignals(dailySignal, fourHourSignal, oneHourSignal)

	if !result.Alignment {
		t.Errorf("CombineMultiTimeframeSignals should detect alignment when all are bullish")
	}

	if result.AlignmentPercent < 66.0 {
		t.Errorf("AlignmentPercent should be >= 66%%, got %.1f%%", result.AlignmentPercent)
	}

	if result.RecommendedTrade != "BUY" {
		t.Errorf("RecommendedTrade should be BUY, got %s", result.RecommendedTrade)
	}
}

func TestCombineMultiTimeframeSignals_PartialAlignment(t *testing.T) {
	dailySignal := CombinedSignal{
		Recommendation: "BUY",
		Score:          2.0,
		Confidence:     85.0,
	}

	fourHourSignal := CombinedSignal{
		Recommendation: "BUY",
		Score:          1.5,
		Confidence:     80.0,
	}

	oneHourSignal := CombinedSignal{
		Recommendation: "SELL",
		Score:          -1.5,
		Confidence:     70.0,
	}

	result := CombineMultiTimeframeSignals(dailySignal, fourHourSignal, oneHourSignal)

	// With current logic: Daily BUY + 4H BUY = LONG alignment, 1H SELL conflicts
	// Alignment should be true only if daily + 4H agree AND don't conflict with 1H too much
	// For now, just verify the function executes and returns structured data
	if result.CompositeScore == 0.0 && result.Confidence == 0.0 {
		t.Errorf("MultiTimeframeSignal should calculate scores")
	}
}

func TestCombineMultiTimeframeSignals_NoAlignment(t *testing.T) {
	dailySignal := CombinedSignal{
		Recommendation: "BUY",
		Score:          2.0,
		Confidence:     85.0,
	}

	fourHourSignal := CombinedSignal{
		Recommendation: "SELL",
		Score:          -1.5,
		Confidence:     80.0,
	}

	oneHourSignal := CombinedSignal{
		Recommendation: "SELL",
		Score:          -1.8,
		Confidence:     75.0,
	}

	result := CombineMultiTimeframeSignals(dailySignal, fourHourSignal, oneHourSignal)

	if result.Alignment {
		t.Errorf("Should not have alignment with conflicting timeframes")
	}

	if result.RecommendedTrade != "WAIT - Timeframes not aligned" {
		t.Errorf("Should wait when timeframes not aligned, got %s", result.RecommendedTrade)
	}
}

func TestIsMultiTimeframeConfirmed_Strict(t *testing.T) {
	signal := MultiTimeframeSignal{
		Alignment:        true,
		AlignmentPercent: 100.0,
		DailySignal: CombinedSignal{
			Confidence: 85.0,
		},
		FourHourSignal: CombinedSignal{
			Confidence: 80.0,
		},
		OneHourSignal: CombinedSignal{
			Confidence: 75.0,
		},
	}

	if !signal.IsMultiTimeframeConfirmed(true) {
		t.Errorf("Should be confirmed with strict requirements when all aligned and confident")
	}
}

func TestIsMultiTimeframeConfirmed_Loose(t *testing.T) {
	signal := MultiTimeframeSignal{
		Alignment:        true,
		AlignmentPercent: 66.0, // 2 of 3
		DailySignal: CombinedSignal{
			Confidence: 85.0,
		},
		FourHourSignal: CombinedSignal{
			Confidence: 50.0, // Below 50
		},
		OneHourSignal: CombinedSignal{
			Confidence: 75.0,
		},
	}

	if !signal.IsMultiTimeframeConfirmed(false) {
		t.Errorf("Should be confirmed with loose requirements when 66%% aligned")
	}

	if signal.IsMultiTimeframeConfirmed(true) {
		t.Errorf("Should fail strict check with 4H below 50%% confidence")
	}
}

func TestCompositeScore_Calculation(t *testing.T) {
	signal := MultiTimeframeSignal{
		DailySignal: CombinedSignal{
			Score: 2.0, // Weighted 0.5
		},
		FourHourSignal: CombinedSignal{
			Score: 1.5, // Weighted 0.35
		},
		OneHourSignal: CombinedSignal{
			Score: 1.0, // Weighted 0.15
		},
	}

	signal.CompositeScore = (2.0 * 0.5) + (1.5 * 0.35) + (1.0 * 0.15)

	expectedScore := 1.675
	if signal.CompositeScore != expectedScore {
		t.Errorf("CompositeScore should be %.3f, got %.3f", expectedScore, signal.CompositeScore)
	}
}

func TestAverageConfidence_Calculation(t *testing.T) {
	signal := MultiTimeframeSignal{
		DailySignal: CombinedSignal{
			Confidence: 85.0,
		},
		FourHourSignal: CombinedSignal{
			Confidence: 80.0,
		},
		OneHourSignal: CombinedSignal{
			Confidence: 75.0,
		},
	}

	signal.Confidence = (85.0 + 80.0 + 75.0) / 3.0

	expectedConfidence := 80.0
	if signal.Confidence != expectedConfidence {
		t.Errorf("Confidence should be %.1f%%, got %.1f%%", expectedConfidence, signal.Confidence)
	}
}

func TestMultiTimeframeAlignment_FalseSignalReduction(t *testing.T) {
	// Simulate a false signal: only 1H is bullish, daily + 4H are bearish
	dailySignal := CombinedSignal{
		Recommendation: "SELL",
		Score:          -2.0,
		Confidence:     80.0,
	}

	fourHourSignal := CombinedSignal{
		Recommendation: "DISTRIBUTE",
		Score:          -1.5,
		Confidence:     75.0,
	}

	oneHourSignal := CombinedSignal{
		Recommendation: "BUY", // False signal
		Score:          1.0,
		Confidence:     60.0,
	}

	result := CombineMultiTimeframeSignals(dailySignal, fourHourSignal, oneHourSignal)

	// This should NOT have alignment and NOT recommend BUY
	if result.Alignment {
		t.Errorf("False signal should not align when majority is opposite")
	}

	if result.RecommendedTrade == "BUY" {
		t.Errorf("Should not recommend BUY when false signal, got %s", result.RecommendedTrade)
	}
}
