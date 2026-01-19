package signals

import (
	"fmt"
	"math"

	"github.com/fazecat/mogulmaker/Internal/strategy/indicators"
	"github.com/fazecat/mogulmaker/Internal/types"
)

type SignalValidationWithSR struct {
	Signal               *types.TradeSignal
	SupportLevel         float64
	ResistanceLevel      float64
	CurrentPrice         float64
	DistanceToSupport    float64 // Positive = above support
	DistanceToResistance float64 // Positive = below resistance
	IsValidLocation      bool
	ValidationScore      float64
	RecommendedAction    string
	DetailedAnalysis     string
}

type SupportResistanceValidator struct {
	MinValidationScore     float64 // Minimum score to consider signal valid (0-100)
	PreferNearSupport      bool    // Prefer longs near support, shorts near resistance
	RequireSignalAlignment bool    // Require signal direction to match S/R location
	TolerancePercent       float64 // Tolerance for "near" support/resistance (%)
}

func NewSupportResistanceValidator() *SupportResistanceValidator {
	return &SupportResistanceValidator{
		MinValidationScore:     50.0, // 50% minimum validation score
		PreferNearSupport:      true,
		RequireSignalAlignment: false,
		TolerancePercent:       2.0, // 2% tolerance
	}
}

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

	score := 50.0 // neutral starting point

	// higher the score the better for LONG signals
	if signal.Direction == "LONG" {
		if atSupport {
			score = 90.0
			validation.DetailedAnalysis = fmt.Sprintf("Price at support (%.2f) - excellent entry", support)
		} else if validation.DistanceToSupport > 0 && validation.DistanceToSupport < srv.TolerancePercent {
			score = 75.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% above support - good entry", validation.DistanceToSupport)
		} else if validation.DistanceToSupport > srv.TolerancePercent && validation.DistanceToSupport < 5.0 {
			score = 60.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% above support - acceptable entry", validation.DistanceToSupport)
		} else if validation.DistanceToSupport >= 5.0 {
			score = 30.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% above support - far from support level", validation.DistanceToSupport)
		}

		if atResistance {
			score -= 20.0
			validation.DetailedAnalysis += " (but at resistance - reduce position)"
		}

		//if strong breakout above include more points
		if signal.Confidence >= 80 && resistance > 0 && currentPrice > resistance*1.005 {
			score += 15.0
			validation.DetailedAnalysis = "Breakout above resistance with strong signal"
		}
	}

	// higher the score the better for SHORT signals
	if signal.Direction == "SHORT" {
		if atResistance {
			score = 90.0
			validation.DetailedAnalysis = fmt.Sprintf("Price at resistance (%.2f) - excellent entry", resistance)
		} else if validation.DistanceToResistance > 0 && validation.DistanceToResistance < srv.TolerancePercent {
			score = 75.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% below resistance - good entry", validation.DistanceToResistance)
		} else if validation.DistanceToResistance > srv.TolerancePercent && validation.DistanceToResistance < 5.0 {
			score = 60.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% below resistance - acceptable entry", validation.DistanceToResistance)
		} else if validation.DistanceToResistance >= 5.0 {
			score = 30.0
			validation.DetailedAnalysis = fmt.Sprintf("Price %.1f%% below resistance - far from resistance level", validation.DistanceToResistance)
		}

		if atSupport {
			score -= 20.0
			validation.DetailedAnalysis += " (but at support - reduce position)"
		}

		if signal.Confidence >= 80 && support > 0 && currentPrice < support*0.995 {
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

	return validation
}

func (srv *SignalValidationWithSR) IsBreakoutAboveResistance(currentPrice, resistance float64) bool {
	if resistance == 0 {
		return false
	}
	return currentPrice > resistance*1.005
}

func (srv *SignalValidationWithSR) IsBreakoutBelowSupport(currentPrice, support float64) bool {
	if support == 0 {
		return false
	}
	return currentPrice < support*0.995
}

// views multiple signals and returns those passing S/R validation
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

func FilterValidSignals(validations []*SignalValidationWithSR) []*SignalValidationWithSR {
	var filtered []*SignalValidationWithSR
	for _, v := range validations {
		if v.IsValidLocation {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

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
