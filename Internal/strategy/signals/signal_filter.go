package signals

import (
	"fmt"

	"github.com/fazecat/mongelmaker/Internal/types"
)

// Reduce false signals by filtering low-confidence signals
type SignalQualityFilter struct {
	MinConfidenceThreshold float64 //default: 70%
	MaxConfidenceThreshold float64
	RequireIndicatorMatch  bool // Require multiple indicators to align
	VerboseLogging         bool
}

type FilteredSignal struct {
	Original           *types.TradeSignal
	Passed             bool
	FailureReason      string
	QualityScore       float64
	IndicatorAlignment int
	RecommendedAction  string
}

func NewSignalQualityFilter() *SignalQualityFilter {
	return &SignalQualityFilter{
		MinConfidenceThreshold: 70.0,
		MaxConfidenceThreshold: 100.0,
		RequireIndicatorMatch:  true,
		VerboseLogging:         false,
	}
}

func (f *SignalQualityFilter) FilterSignal(signal *types.TradeSignal) *FilteredSignal {
	result := &FilteredSignal{
		Original:           signal,
		Passed:             false,
		FailureReason:      "",
		QualityScore:       signal.Confidence,
		IndicatorAlignment: 1,
		RecommendedAction:  "PASS",
	}

	// Confidence threshold
	if signal.Confidence < f.MinConfidenceThreshold {
		result.Passed = false
		result.FailureReason = fmt.Sprintf("Confidence %.1f%% below minimum threshold %.1f%%",
			signal.Confidence, f.MinConfidenceThreshold)
		result.RecommendedAction = "REJECT - Low Confidence"
		return result
	}

	//Confidence sanity check
	if signal.Confidence > f.MaxConfidenceThreshold {
		signal.Confidence = f.MaxConfidenceThreshold
		result.QualityScore = f.MaxConfidenceThreshold
	}

	if signal.Reasoning == "" {
		result.Passed = false
		result.FailureReason = "Signal has no reasoning attached"
		result.RecommendedAction = "REJECT - No Reasoning"
		return result
	}

	// Direction validation
	if signal.Direction != "LONG" && signal.Direction != "SHORT" {
		result.Passed = false
		result.FailureReason = fmt.Sprintf("Invalid direction: %s (must be LONG or SHORT)", signal.Direction)
		result.RecommendedAction = "REJECT - Invalid Direction"
		return result
	}

	// All checks passed
	result.Passed = true
	result.RecommendedAction = fmt.Sprintf("EXECUTE %s", signal.Direction)

	if f.VerboseLogging {
		fmt.Printf("Signal quality check PASSED: %s @ %.1f%% confidence\n",
			signal.Direction, signal.Confidence)
	}

	return result
}

func (f *SignalQualityFilter) FilterSignalBatch(signals []*types.TradeSignal) []*FilteredSignal {
	filtered := make([]*FilteredSignal, len(signals))
	for i, signal := range signals {
		filtered[i] = f.FilterSignal(signal)
	}
	return filtered
}

func (f *SignalQualityFilter) GetHighestConfidenceSignal(signals []*types.TradeSignal) *FilteredSignal {
	if len(signals) == 0 {
		return &FilteredSignal{
			Passed:            false,
			FailureReason:     "No signals provided",
			RecommendedAction: "WAIT - No Signals",
		}
	}

	var bestSignal *FilteredSignal
	var bestConfidence float64 = -1

	for _, signal := range signals {
		filtered := f.FilterSignal(signal)
		if filtered.Passed && filtered.QualityScore > bestConfidence {
			bestSignal = filtered
			bestConfidence = filtered.QualityScore
		}
	}

	if bestSignal == nil {
		return &FilteredSignal{
			Passed:            false,
			FailureReason:     "No signals passed quality filter",
			RecommendedAction: "WAIT - All Signals Filtered",
		}
	}

	return bestSignal
}
