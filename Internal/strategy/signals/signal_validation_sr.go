package signals

import (
	"fmt"
	"math"

	"github.com/fazecat/mongelmaker/Internal/strategy/indicators"
	"github.com/fazecat/mongelmaker/Internal/types"
)

// SignalValidationWithSR combines signal analysis with support/resistance levels
type SignalValidationWithSR struct {
	Signal               *types.TradeSignal
	SupportLevel         float64
	ResistanceLevel      float64
	CurrentPrice         float64
	DistanceToSupport    float64 // Positive = above support
	DistanceToResistance float64 // Positive = below resistance
	IsValidLocation      bool
	ValidationScore      float64 // 0-100
	RecommendedAction    string
	DetailedAnalysis     string
}

// SupportResistanceValidator validates signals against key price levels
type SupportResistanceValidator struct {
	MinValidationScore     float64 // Minimum score to consider signal valid (0-100)
	PreferNearSupport      bool    // Prefer longs near support, shorts near resistance
	RequireSignalAlignment bool    // Require signal direction to match S/R location
	TolerancePercent       float64 // Tolerance for "near" support/resistance (%)
	VerboseLogging         bool
}

// NewSupportResistanceValidator creates a new validator with defaults
func NewSupportResistanceValidator() *SupportResistanceValidator {
	return &SupportResistanceValidator{
		MinValidationScore:     50.0, // 50% minimum validation score
		PreferNearSupport:      true,
		RequireSignalAlignment: false,
		TolerancePercent:       2.0, // 2% tolerance
		VerboseLogging:         false,
	}
}

// ValidateSignalWithSR checks if a signal is valid given current support/resistance levels
func (srv *SupportResistanceValidator) ValidateSignalWithSR(
	signal *types.TradeSignal,
	bars []types.Bar,
	currentPrice float64,
) *SignalValidationWithSR {

	if len(bars) == 0 {
		return &SignalValidationWithSR{
			Signal:            signal,
			IsValidLocation:   false,
			ValidationScore:   0.0,
			RecommendedAction: "REJECT - No price data",
		}
	}

	// Calculate support and resistance levels
	support := indicators.FindSupport(bars)
	resistance := indicators.FindResistance(bars)

	validation := &SignalValidationWithSR{
		Signal:          signal,
		SupportLevel:    support,
		ResistanceLevel: resistance,
		CurrentPrice:    currentPrice,
	}

	// Calculate distances
	validation.DistanceToSupport = indicators.DistanceToSupport(currentPrice, support)
	validation.DistanceToResistance = indicators.DistanceToResistance(currentPrice, resistance)

	// Determine if price is at support or resistance
	atSupport := indicators.IsAtSupport(currentPrice, support)
	atResistance := indicators.IsAtResistance(currentPrice, resistance)

	// Scoring logic
	score := 50.0 // Start with neutral score

	// LONG signal validation
	if signal.Direction == "LONG" {
		// Ideal: Price near support
		if atSupport {
			score = 90.0 // Excellent location for long
			validation.DetailedAnalysis = fmt.Sprintf("Price at support (%.2f) - excellent entry", support)
		} else if validation.DistanceToSupport > 0 && validation.DistanceToSupport < srv.TolerancePercent {
			// Close to support but not exactly at it
			score = 75.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% above support - good entry", validation.DistanceToSupport)
		} else if validation.DistanceToSupport > srv.TolerancePercent && validation.DistanceToSupport < 5.0 {
			// Moderately away from support
			score = 60.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% above support - acceptable entry", validation.DistanceToSupport)
		} else if validation.DistanceToSupport >= 5.0 {
			// Far from support - risky
			score = 30.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% above support - far from support level", validation.DistanceToSupport)
		}

		// Penalty if at resistance
		if atResistance {
			score -= 20.0
			validation.DetailedAnalysis += " (but at resistance - reduce position)"
		}

		// Bonus if price is near resistance but signal is strong
		if signal.Confidence >= 80 && validation.IsBreakoutAboveResistance(currentPrice, resistance) {
			score += 15.0
			validation.DetailedAnalysis = "Breakout above resistance with strong signal"
		}
	}

	// SHORT signal validation
	if signal.Direction == "SHORT" {
		// Ideal: Price near resistance
		if atResistance {
			score = 90.0 // Excellent location for short
			validation.DetailedAnalysis = fmt.Sprintf("Price at resistance (%.2f) - excellent entry", resistance)
		} else if validation.DistanceToResistance > 0 && validation.DistanceToResistance < srv.TolerancePercent {
			// Close to resistance but not exactly at it
			score = 75.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% below resistance - good entry", validation.DistanceToResistance)
		} else if validation.DistanceToResistance > srv.TolerancePercent && validation.DistanceToResistance < 5.0 {
			// Moderately away from resistance
			score = 60.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% below resistance - acceptable entry", validation.DistanceToResistance)
		} else if validation.DistanceToResistance >= 5.0 {
			// Far from resistance - risky
			score = 30.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% below resistance - far from resistance level", validation.DistanceToResistance)
		}

		// Penalty if at support
		if atSupport {
			score -= 20.0
			validation.DetailedAnalysis += " (but at support - reduce position)"
		}

		// Bonus if price is near support but signal is strong
		if signal.Confidence >= 80 && validation.IsBreakoutBelowSupport(currentPrice, support) {
			score += 15.0
			validation.DetailedAnalysis = "Breakdown below support with strong signal"
		}
	}

	// Apply confidence boost
	score += (signal.Confidence / 100.0) * 20.0

	// Cap score
	score = math.Min(100.0, math.Max(0.0, score))

	validation.ValidationScore = score
	validation.IsValidLocation = score >= srv.MinValidationScore

	// Determine action
	if validation.IsValidLocation {
		validation.RecommendedAction = fmt.Sprintf("EXECUTE %s", signal.Direction)
	} else {
		validation.RecommendedAction = fmt.Sprintf("CAUTION - Score %.0f%% (threshold %.0f%%)", validation.ValidationScore, srv.MinValidationScore)
	}

	if srv.VerboseLogging {
		fmt.Printf("%s\n", FormatSRValidation(*validation))
	}

	return validation
}

// IsBreakoutAboveResistance checks if price has broken above resistance
func (srv *SignalValidationWithSR) IsBreakoutAboveResistance(currentPrice, resistance float64) bool {
	if resistance == 0 {
		return false
	}
	return currentPrice > resistance*1.005 // 0.5% above = breakout
}

// IsBreakoutBelowSupport checks if price has broken below support
func (srv *SignalValidationWithSR) IsBreakoutBelowSupport(currentPrice, support float64) bool {
	if support == 0 {
		return false
	}
	return currentPrice < support*0.995 // 0.5% below = breakdown
}

// ValidateBatchWithSR validates multiple signals and returns those passing S/R validation
func (srv *SupportResistanceValidator) ValidateBatchWithSR(
	signals []*types.TradeSignal,
	bars []types.Bar,
	currentPrice float64,
) []*SignalValidationWithSR {

	validations := make([]*SignalValidationWithSR, len(signals))
	for i, signal := range signals {
		validations[i] = srv.ValidateSignalWithSR(signal, bars, currentPrice)
	}
	return validations
}

// FilterValidSignals returns only signals that pass validation
func FilterValidSignals(validations []*SignalValidationWithSR) []*SignalValidationWithSR {
	var filtered []*SignalValidationWithSR
	for _, v := range validations {
		if v.IsValidLocation {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// GetBestSignal returns the highest-scoring validated signal
func GetBestSignal(validations []*SignalValidationWithSR) *SignalValidationWithSR {
	if len(validations) == 0 {
		return nil
	}

	var best *SignalValidationWithSR
	for _, v := range validations {
		if best == nil || v.ValidationScore > best.ValidationScore {
			best = v
		}
	}
	return best
}

// CalculateSRStrength determines how strong the support/resistance level is
func CalculateSRStrength(bars []types.Bar, level float64, isSupportLevel bool) float64 {
	if level == 0 {
		return 0.0
	}

	touchCount := 0
	tolerance := level * 0.01 // 1% tolerance

	for _, bar := range bars {
		if isSupportLevel {
			if bar.Low >= level-tolerance && bar.Low <= level+tolerance {
				touchCount++
			}
		} else {
			if bar.High >= level-tolerance && bar.High <= level+tolerance {
				touchCount++
			}
		}
	}

	// More touches = stronger level
	// Cap at 100%
	strength := (float64(touchCount) / float64(len(bars))) * 100.0
	return math.Min(100.0, strength)
}

// SRQualityReport provides detailed S/R analysis
type SRQualityReport struct {
	SupportLevel       float64
	SupportStrength    float64
	ResistanceLevel    float64
	ResistanceStrength float64
	CurrentPrice       float64
	PricePosition      string // "AT_SUPPORT", "AT_RESISTANCE", "IN_RANGE", "ABOVE_ALL", "BELOW_ALL"
	RangeWidth         float64
	RangeWidthPercent  float64
}

// GenerateSRQualityReport analyzes S/R quality
func GenerateSRQualityReport(bars []types.Bar) *SRQualityReport {
	report := &SRQualityReport{}

	if len(bars) == 0 {
		return report
	}

	report.SupportLevel = indicators.FindSupport(bars)
	report.ResistanceLevel = indicators.FindResistance(bars)
	report.CurrentPrice = bars[len(bars)-1].Close

	report.SupportStrength = CalculateSRStrength(bars, report.SupportLevel, true)
	report.ResistanceStrength = CalculateSRStrength(bars, report.ResistanceLevel, false)

	report.RangeWidth = report.ResistanceLevel - report.SupportLevel
	report.RangeWidthPercent = (report.RangeWidth / report.SupportLevel) * 100

	// Determine price position
	if indicators.IsAtSupport(report.CurrentPrice, report.SupportLevel) {
		report.PricePosition = "AT_SUPPORT"
	} else if indicators.IsAtResistance(report.CurrentPrice, report.ResistanceLevel) {
		report.PricePosition = "AT_RESISTANCE"
	} else if report.CurrentPrice > report.SupportLevel && report.CurrentPrice < report.ResistanceLevel {
		report.PricePosition = "IN_RANGE"
	} else if report.CurrentPrice > report.ResistanceLevel {
		report.PricePosition = "ABOVE_ALL"
	} else {
		report.PricePosition = "BELOW_ALL"
	}

	return report
}

// FormatSRValidation provides a formatted report
func FormatSRValidation(validation SignalValidationWithSR) string {
	return fmt.Sprintf(`
📍 Signal S/R Validation:
   Signal: %s @ %.0f%% confidence
   Support: %.2f | Resistance: %.2f | Current: %.2f
   To Support: %.1f%% | To Resistance: %.1f%%
   
   Score: %.0f/100 | Valid: %v
   %s
   
   Recommendation: %s
`,
		validation.Signal.Direction, validation.Signal.Confidence,
		validation.SupportLevel, validation.ResistanceLevel, validation.CurrentPrice,
		validation.DistanceToSupport, validation.DistanceToResistance,
		validation.ValidationScore, validation.IsValidLocation,
		validation.DetailedAnalysis,
		validation.RecommendedAction,
	)
}

// FormatSRQualityReport provides formatted S/R analysis
func FormatSRQualityReport(report *SRQualityReport) string {
	return fmt.Sprintf(`
📊 Support/Resistance Quality:
   Support: %.2f (%.0f%% strength)
   Resistance: %.2f (%.0f%% strength)
   Range: %.2f (%.2f%%)
   
   Current Price: %.2f (%s)
`,
		report.SupportLevel, report.SupportStrength,
		report.ResistanceLevel, report.ResistanceStrength,
		report.RangeWidth, report.RangeWidthPercent,
		report.CurrentPrice, report.PricePosition,
	)
}
