package strategy

import (
	"fmt"
)

// signals based on quality metrics
// Goal: Reduce false signals by filtering low-confidence signals
type SignalQualityFilter struct {
	MinConfidenceThreshold float64 // Minimum confidence to execute (default: 70%)
	MaxConfidenceThreshold float64 // Maximum confidence (safety check)
	RequireIndicatorMatch  bool    // Require multiple indicators to align
	VerboseLogging         bool    // Enable detailed logging
}

// a signal after quality filtering
type FilteredSignal struct {
	Original           *TradeSignal
	Passed             bool
	FailureReason      string
	QualityScore       float64
	IndicatorAlignment int // How many indicators aligned with this signal
	RecommendedAction  string
}

// filter with default settings
func NewSignalQualityFilter() *SignalQualityFilter {
	return &SignalQualityFilter{
		MinConfidenceThreshold: 70.0,
		MaxConfidenceThreshold: 100.0,
		RequireIndicatorMatch:  true,
		VerboseLogging:         false,
	}
}

// quality checks to a single signal
func (f *SignalQualityFilter) FilterSignal(signal *TradeSignal) *FilteredSignal {
	result := &FilteredSignal{
		Original:           signal,
		Passed:             false,
		FailureReason:      "",
		QualityScore:       signal.Confidence,
		IndicatorAlignment: 1, // At least one indicator (the primary one) is present
		RecommendedAction:  "PASS",
	}

	// Check 1: Confidence threshold
	if signal.Confidence < f.MinConfidenceThreshold {
		result.Passed = false
		result.FailureReason = fmt.Sprintf("Confidence %.1f%% below minimum threshold %.1f%%",
			signal.Confidence, f.MinConfidenceThreshold)
		result.RecommendedAction = "REJECT - Low Confidence"
		return result
	}

	// Check 2: Confidence sanity check
	if signal.Confidence > f.MaxConfidenceThreshold {
		signal.Confidence = f.MaxConfidenceThreshold
		result.QualityScore = f.MaxConfidenceThreshold
	}

	// Check 3: Signal reasoning validation (must have reasoning)
	if signal.Reasoning == "" {
		result.Passed = false
		result.FailureReason = "Signal has no reasoning attached"
		result.RecommendedAction = "REJECT - No Reasoning"
		return result
	}

	// Check 4: Direction validation
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
		fmt.Printf("✅ Signal quality check PASSED: %s @ %.1f%% confidence\n",
			signal.Direction, signal.Confidence)
	}

	return result
}

// quality filtering to multiple signals
func (f *SignalQualityFilter) FilterSignalBatch(signals []*TradeSignal) []*FilteredSignal {
	filtered := make([]*FilteredSignal, len(signals))
	for i, signal := range signals {
		filtered[i] = f.FilterSignal(signal)
	}
	return filtered
}

// returns the best signal from a batch that passes filtering
func (f *SignalQualityFilter) GetHighestConfidenceSignal(signals []*TradeSignal) *FilteredSignal {
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
		// No signals passed filtering
		return &FilteredSignal{
			Passed:            false,
			FailureReason:     "No signals passed quality filter",
			RecommendedAction: "WAIT - All Signals Filtered",
		}
	}

	return bestSignal
}

//	combines multiple signal components
//
// Used for multi-indicator confirmation
type SignalStrengthCalculator struct {
	Weights map[string]float64 // Weight for each indicator type
}

// creates a calculator with default weights
func NewSignalStrengthCalculator() *SignalStrengthCalculator {
	return &SignalStrengthCalculator{
		Weights: map[string]float64{
			"RSI":                0.25,
			"ATR":                0.15,
			"Whale":              0.30,
			"Pattern":            0.20,
			"Support/Resistance": 0.10,
		},
	}
}

// represents a single indicator's contribution
type IndicatorScore struct {
	Name      string
	Value     float64
	Score     float64
	Weight    float64
	Alignment string // "BULLISH", "BEARISH", "NEUTRAL"
}

// computes weighted average of multiple indicators
func (c *SignalStrengthCalculator) CalculateCompositeScore(indicators []IndicatorScore) float64 {
	if len(indicators) == 0 {
		return 0.0
	}

	var totalScore float64
	var totalWeight float64

	for _, indicator := range indicators {
		weight := c.Weights[indicator.Name]
		if weight == 0 {
			// Use equal weight if not specified
			weight = 1.0 / float64(len(indicators))
		}
		totalScore += indicator.Score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// counts how many indicators align with the signal direction
func CountAlignedIndicators(indicators []IndicatorScore, direction string) int {
	count := 0
	for _, indicator := range indicators {
		if direction == "LONG" && indicator.Alignment == "BULLISH" {
			count++
		} else if direction == "SHORT" && indicator.Alignment == "BEARISH" {
			count++
		}
	}
	return count
}

// returns % of indicators aligned with the signal
func CalculateAlignmentPercentage(indicators []IndicatorScore, direction string) float64 {
	if len(indicators) == 0 {
		return 0.0
	}
	aligned := CountAlignedIndicators(indicators, direction)
	return (float64(aligned) / float64(len(indicators))) * 100.0
}

// detailed feedback on signal quality
type SignalValidationReport struct {
	SignalID                string
	Direction               string
	Confidence              float64
	IsValid                 bool
	ValidationChecks        []ValidationCheck
	OverallQualityScore     float64
	ExecutionRecommendation string
}

// a single validation test
type ValidationCheck struct {
	Name     string
	Passed   bool
	Details  string
	Severity string // "CRITICAL", "WARNING", "INFO"
}

// a comprehensive report on signal quality
func GenerateValidationReport(signal *TradeSignal, filters ...func(*TradeSignal) ValidationCheck) *SignalValidationReport {
	report := &SignalValidationReport{
		SignalID:         fmt.Sprintf("%s_%d", signal.Direction, int(signal.Confidence)),
		Direction:        signal.Direction,
		Confidence:       signal.Confidence,
		ValidationChecks: []ValidationCheck{},
	}

	// Run all validation filters
	for _, filter := range filters {
		check := filter(signal)
		report.ValidationChecks = append(report.ValidationChecks, check)
	}

	// Calculate overall quality
	passed := 0
	for _, check := range report.ValidationChecks {
		if check.Passed {
			passed++
		}
	}

	report.OverallQualityScore = (float64(passed) / float64(len(report.ValidationChecks))) * 100.0
	report.IsValid = report.OverallQualityScore >= 70.0

	if report.IsValid {
		report.ExecutionRecommendation = fmt.Sprintf("EXECUTE %s", signal.Direction)
	} else {
		report.ExecutionRecommendation = "REJECT - Quality below threshold"
	}

	return report
}

// Common validation filters
func ValidateMinimumConfidence(minThreshold float64) func(*TradeSignal) ValidationCheck {
	return func(signal *TradeSignal) ValidationCheck {
		return ValidationCheck{
			Name:     "Minimum Confidence",
			Passed:   signal.Confidence >= minThreshold,
			Details:  fmt.Sprintf("Confidence: %.1f%% (required: >= %.1f%%)", signal.Confidence, minThreshold),
			Severity: "CRITICAL",
		}
	}
}

func ValidateMaximumConfidence(maxThreshold float64) func(*TradeSignal) ValidationCheck {
	return func(signal *TradeSignal) ValidationCheck {
		return ValidationCheck{
			Name:     "Maximum Confidence Sanity Check",
			Passed:   signal.Confidence <= maxThreshold,
			Details:  fmt.Sprintf("Confidence: %.1f%% (max allowed: %.1f%%)", signal.Confidence, maxThreshold),
			Severity: "WARNING",
		}
	}
}

func ValidateDirection() func(*TradeSignal) ValidationCheck {
	return func(signal *TradeSignal) ValidationCheck {
		valid := signal.Direction == "LONG" || signal.Direction == "SHORT"
		return ValidationCheck{
			Name:     "Valid Direction",
			Passed:   valid,
			Details:  fmt.Sprintf("Direction: %s", signal.Direction),
			Severity: "CRITICAL",
		}
	}
}

func ValidateReasoning() func(*TradeSignal) ValidationCheck {
	return func(signal *TradeSignal) ValidationCheck {
		return ValidationCheck{
			Name:     "Has Reasoning",
			Passed:   signal.Reasoning != "",
			Details:  fmt.Sprintf("Reasoning provided: %v", signal.Reasoning != ""),
			Severity: "WARNING",
		}
	}
}
