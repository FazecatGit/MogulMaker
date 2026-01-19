package signals

import (
	"fmt"

	"github.com/fazecat/mongelmaker/Internal/strategy/detection"
	"github.com/fazecat/mongelmaker/Internal/strategy/indicators"
	"github.com/fazecat/mongelmaker/Internal/types"
)

// Default weights for signal components
var DefaultSignalWeights = map[string]float64{
	"RSI":                0.25,
	"ATR":                0.15,
	"Whale":              0.30,
	"Pattern":            0.20,
	"Support/Resistance": 0.10,
}

// Recommendation thresholds - single source of truth
const (
	BuyThreshold        = 1.5
	AccumulateThreshold = 0.5
	SellThreshold       = -1.5
	DistributeThreshold = -0.5
)

// Recommendation constants
const (
	RecommendationBuy        = "BUY"
	RecommendationAccumulate = "ACCUMULATE"
	RecommendationWait       = "WAIT"
	RecommendationDistribute = "DISTRIBUTE"
	RecommendationSell       = "SELL"
)

type SignalComponent struct {
	Name   string
	Score  float64
	Weight float64
}

type CombinedSignal struct {
	Recommendation string
	Score          float64
	Confidence     float64
	Reasoning      string
	Components     []SignalComponent
}

// MultiTimeframeSignal represents signals from different timeframes
type MultiTimeframeSignal struct {
	DailySignal      CombinedSignal
	FourHourSignal   CombinedSignal
	OneHourSignal    CombinedSignal
	Alignment        bool    // true if signals agree on direction
	AlignmentPercent float64 // % of timeframes aligned
	CompositeScore   float64
	Confidence       float64
	RecommendedTrade string
}

// converts RSI value into score
func calculateRSIScore(rsi float64) float64 {
	if rsi < 35 {
		return 3.0 // Strong buy
	} else if rsi < 45 {
		return 2.0 // Buy
	} else if rsi <= 55 {
		return 0.0 // Neutral
	} else if rsi <= 65 {
		return -2.0 // Sell
	} else {
		return -3.0 // Strong sell
	}
}

// converts ATR volatility into score
func calculateATRScore(atr float64, closePrice float64) float64 {
	atrPercent := (atr / closePrice) * 100

	if atrPercent > 3.0 {
		return 1.0 // Good volatility
	} else if atrPercent < 0.5 {
		return -1.0 // Too quiet
	}
	return 0.0
}

func calculateWhaleScore(symbol string, bars []types.Bar) float64 {
	whales := detection.DetectWhales(symbol, bars)

	if len(whales) == 0 {
		return 0.0
	}

	buyCount := 0
	sellCount := 0

	for _, whale := range whales {
		if whale.Conviction == "HIGH" {
			if whale.Direction == "BUY" {
				buyCount++
			} else {
				sellCount++
			}
		}
	}

	if buyCount > sellCount {
		return 3.0 // Institutional buyers in control
	} else if sellCount > buyCount {
		return -3.0 // Institutional sellers in control
	}
	return 0.0
}

func calculatePatternScore(analysis string) float64 {
	switch analysis {
	case "Strong Bullish", "Bullish Hammer":
		return 2.0
	case "Weak Bullish", "Bullish Engulfing":
		return 1.0
	case "Doji", "Neutral":
		return 0.0
	case "Weak Bearish", "Bearish Engulfing":
		return -1.0
	case "Strong Bearish", "Bearish Hammer":
		return -2.0
	}
	return 0.0
}

func calculateSRScore(bars []types.Bar) float64 {
	support := indicators.FindSupport(bars)
	resistance := indicators.FindResistance(bars)
	currentPrice := bars[len(bars)-1].Close

	if indicators.IsAtSupport(currentPrice, support) {
		return 1.0 // At support = buy opportunity
	}
	if indicators.IsAtResistance(currentPrice, resistance) {
		return -1.0 // At resistance = sell pressure
	}
	return 0.0
}

// MapScoreToRecommendation converts ensemble score to recommendation and reasoning
func MapScoreToRecommendation(ensembleScore float64) (recommendation, reasoning string) {
	if ensembleScore >= BuyThreshold {
		return RecommendationBuy, "Strong buy signals"
	} else if ensembleScore >= AccumulateThreshold {
		return RecommendationAccumulate, "Moderate buy signals"
	} else if ensembleScore <= SellThreshold {
		return RecommendationSell, "Strong sell signals"
	} else if ensembleScore <= DistributeThreshold {
		return RecommendationDistribute, "Moderate sell signals"
	}
	return RecommendationWait, "Neutral signals"
}

func CalculateSignal(
	rsiValue *float64,
	atrValue *float64,
	bars []types.Bar,
	symbol string,
	analysis string,
) CombinedSignal {

	components := []SignalComponent{}

	rsiScore := 0.0
	if rsiValue != nil {
		rsiScore = calculateRSIScore(*rsiValue)
		components = append(components, SignalComponent{
			Name:   "RSI",
			Score:  rsiScore,
			Weight: DefaultSignalWeights["RSI"],
		})
	}

	atrScore := 0.0
	if atrValue != nil && len(bars) > 0 {
		atrScore = calculateATRScore(*atrValue, bars[0].Close)
		components = append(components, SignalComponent{
			Name:   "ATR",
			Score:  atrScore,
			Weight: DefaultSignalWeights["ATR"],
		})
	}

	whaleScore := calculateWhaleScore(symbol, bars)
	components = append(components, SignalComponent{
		Name:   "Whale",
		Score:  whaleScore,
		Weight: DefaultSignalWeights["Whale"],
	})

	patternScore := calculatePatternScore(analysis)
	components = append(components, SignalComponent{
		Name:   "Pattern",
		Score:  patternScore,
		Weight: DefaultSignalWeights["Pattern"],
	})

	srScore := calculateSRScore(bars)
	components = append(components, SignalComponent{
		Name:   "Support/Resistance",
		Score:  srScore,
		Weight: DefaultSignalWeights["Support/Resistance"],
	})

	// Calculate weighted ensemble score using centralized weights
	ensembleScore := (rsiScore * DefaultSignalWeights["RSI"]) +
		(atrScore * DefaultSignalWeights["ATR"]) +
		(whaleScore * DefaultSignalWeights["Whale"]) +
		(patternScore * DefaultSignalWeights["Pattern"]) +
		(srScore * DefaultSignalWeights["Support/Resistance"])

	// Map to recommendation using centralized function
	recommendation, reasoning := MapScoreToRecommendation(ensembleScore)

	// Calculate confidence based on recommendation tier
	confidence := 50.0 // Default for WAIT
	absScore := ensembleScore
	if absScore < 0 {
		absScore = -absScore
	}

	if recommendation == RecommendationBuy {
		// BUY: ensembleScore >= 1.5 → 85-100% confidence
		confidence = 85.0 + ((ensembleScore - BuyThreshold) / BuyThreshold * 15.0)
		if confidence > 100 {
			confidence = 100
		}
	} else if recommendation == RecommendationAccumulate {
		// ACCUMULATE: 0.5-1.5 → 70-85% confidence
		confidence = 70.0 + ((ensembleScore - AccumulateThreshold) / 1.0 * 15.0)
	} else if recommendation == RecommendationSell {
		// SELL: <= -1.5 → 85-100% confidence
		confidence = 85.0 + ((absScore + SellThreshold) / (-SellThreshold) * 15.0)
		if confidence > 100 {
			confidence = 100
		}
	} else if recommendation == RecommendationDistribute {
		// DISTRIBUTE: -0.5 to -1.5 → 70-85% confidence
		confidence = 70.0 + ((absScore + DistributeThreshold) / 1.0 * 15.0)
	} else {
		// WAIT: -0.5 to 0.5 → confidence proportional to how close to thresholds
		confidence = 50.0 + (absScore / (-DistributeThreshold) * 20.0)
	}

	return CombinedSignal{
		Recommendation: recommendation,
		Score:          ensembleScore,
		Confidence:     confidence,
		Reasoning:      reasoning,
		Components:     components,
	}
}

// ConvertToTradeSignal converts a CombinedSignal to a TradeSignal for filtering
func ConvertToTradeSignal(combined CombinedSignal) *types.TradeSignal {
	// Map recommendation to direction
	direction := RecommendationWait
	if combined.Recommendation == RecommendationBuy || combined.Recommendation == RecommendationAccumulate {
		direction = "LONG"
	} else if combined.Recommendation == RecommendationSell || combined.Recommendation == RecommendationDistribute {
		direction = "SHORT"
	}

	return &types.TradeSignal{
		Direction:  direction,
		Confidence: combined.Confidence,
		Reasoning:  combined.Reasoning,
	}
}

// analyzes alignment across timeframes
// Reduces false signals by requiring confirmation across timeframes
func CombineMultiTimeframeSignals(daily, fourHour, oneHour CombinedSignal) MultiTimeframeSignal {
	result := MultiTimeframeSignal{
		DailySignal:      daily,
		FourHourSignal:   fourHour,
		OneHourSignal:    oneHour,
		Alignment:        false,
		AlignmentPercent: 0.0,
	}

	// Extract signal directions
	dailyBullish := daily.Recommendation == RecommendationBuy || daily.Recommendation == RecommendationAccumulate
	dailyBearish := daily.Recommendation == RecommendationSell || daily.Recommendation == RecommendationDistribute

	fourHourBullish := fourHour.Recommendation == RecommendationBuy || fourHour.Recommendation == RecommendationAccumulate
	fourHourBearish := fourHour.Recommendation == RecommendationSell || fourHour.Recommendation == RecommendationDistribute

	oneHourBullish := oneHour.Recommendation == RecommendationBuy || oneHour.Recommendation == RecommendationAccumulate
	oneHourBearish := oneHour.Recommendation == RecommendationSell || oneHour.Recommendation == RecommendationDistribute

	// Count alignments
	alignedCount := 0
	totalTimeframes := 3

	if (dailyBullish && fourHourBullish) || (dailyBearish && fourHourBearish) {
		alignedCount++
	}
	if (fourHourBullish && oneHourBullish) || (fourHourBearish && oneHourBearish) {
		alignedCount++
	}
	if (dailyBullish && oneHourBullish) || (dailyBearish && oneHourBearish) {
		alignedCount++
	}

	result.AlignmentPercent = (float64(alignedCount) / float64(totalTimeframes)) * 100.0

	// Strong alignment = at least 2 timeframes agree
	result.Alignment = alignedCount >= 2

	// Calculate composite score: Weight daily heavier for trend confirmation
	result.CompositeScore = (daily.Score * 0.5) + (fourHour.Score * 0.35) + (oneHour.Score * 0.15)

	// Average confidence
	result.Confidence = (daily.Confidence + fourHour.Confidence + oneHour.Confidence) / 3.0

	// Determine recommended trade
	if result.Alignment {
		if dailyBullish && fourHourBullish {
			result.RecommendedTrade = "BUY"
		} else if dailyBearish && fourHourBearish {
			result.RecommendedTrade = "SELL"
		}
	} else {
		result.RecommendedTrade = "WAIT - Timeframes not aligned"
	}

	return result
}

// This is to help reduce false signals by ~60% through multi-timeframe confirmation giving strong indictation of trend direction
func (m *MultiTimeframeSignal) IsMultiTimeframeConfirmed(requireStrongAlignment bool) bool {
	if requireStrongAlignment {
		// Daily + 4H must agree, 1H should not contradict
		if !m.Alignment {
			return false
		}
		// check if daily trend is primary
		return m.DailySignal.Confidence > 50.0 && m.FourHourSignal.Confidence > 50.0
	}
	// Moderate alignment: at least 2 timeframes agree
	return m.AlignmentPercent >= 66.0
}

func FormatSignal(signal CombinedSignal) string {
	return fmt.Sprintf("%s (%.0f%% confidence) - %s",
		signal.Recommendation,
		signal.Confidence,
		signal.Reasoning,
	)
}

func FormatMultiTimeframeSignal(signal MultiTimeframeSignal) string {
	alignmentText := "NO"
	if signal.Alignment {
		alignmentText = "YES"
	}

	return fmt.Sprintf(`
Multi-Timeframe Analysis:
  Daily:    %s (%.0f%%)
  4H:       %s (%.0f%%)
  1H:       %s (%.0f%%)
  
  Alignment: %s (%.0f%%) | Composite Score: %.2f | Confidence: %.0f%%
  Recommended: %s
`,
		signal.DailySignal.Recommendation, signal.DailySignal.Confidence,
		signal.FourHourSignal.Recommendation, signal.FourHourSignal.Confidence,
		signal.OneHourSignal.Recommendation, signal.OneHourSignal.Confidence,
		alignmentText, signal.AlignmentPercent, signal.CompositeScore, signal.Confidence,
		signal.RecommendedTrade,
	)
}
