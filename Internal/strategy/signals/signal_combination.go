package signals

import (
	"fmt"

	"github.com/fazecat/mogulmaker/Internal/strategy/detection"
	"github.com/fazecat/mogulmaker/Internal/strategy/indicators"
	"github.com/fazecat/mogulmaker/Internal/types"
)

// Default weights for signal components
var DefaultSignalWeights = map[string]float64{
	"RSI":                0.20,
	"ATR":                0.12,
	"Whale":              0.25,
	"Pattern":            0.18,
	"Support/Resistance": 0.10,
	"Divergence":         0.15,
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
	Recommendation    string
	Score             float64
	Confidence        float64
	Reasoning         string
	Components        []SignalComponent
	DivergenceDetails string // Details about detected divergence
}

type MultiTimeframeSignal struct {
	DailySignal      CombinedSignal
	FourHourSignal   CombinedSignal
	OneHourSignal    CombinedSignal
	Alignment        bool
	AlignmentPercent float64
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

func calculateDivergenceScore(bars []types.Bar, rsiValues []float64) (float64, string) {
	if len(bars) < 20 || len(rsiValues) < 20 {
		return 0.0, "" // Not enough data
	}

	detector := detection.NewDivergenceDetector()

	// Check for regular divergence
	regularDiv := detector.DetectRSIDivergence(bars, rsiValues)
	if regularDiv.Detected {
		if regularDiv.Type == detection.DivergenceBullish {
			details := fmt.Sprintf("BULLISH Divergence (%.0f%% confidence): %s", regularDiv.Confidence, regularDiv.Reasoning)
			return 2.5 * (regularDiv.Confidence / 100.0), details
		} else if regularDiv.Type == detection.DivergenceBearish {
			details := fmt.Sprintf("BEARISH Divergence (%.0f%% confidence): %s", regularDiv.Confidence, regularDiv.Reasoning)
			return -2.5 * (regularDiv.Confidence / 100.0), details
		}
	}

	// Check for hidden divergence (trend continuation)
	hiddenDiv := detector.DetectHiddenDivergence(bars, rsiValues)
	if hiddenDiv.Detected {
		if hiddenDiv.Type == detection.DivergenceBullish {
			details := fmt.Sprintf("Hidden BULLISH Divergence (%.0f%% confidence): %s", hiddenDiv.Confidence, hiddenDiv.Reasoning)
			return 1.5 * (hiddenDiv.Confidence / 100.0), details
		} else if hiddenDiv.Type == detection.DivergenceBearish {
			details := fmt.Sprintf("Hidden BEARISH Divergence (%.0f%% confidence): %s", hiddenDiv.Confidence, hiddenDiv.Reasoning)
			return -1.5 * (hiddenDiv.Confidence / 100.0), details
		}
	}

	// Check for exaggerated conditions
	exaggeratedDiv := detector.DetectExaggeratedDivergence(rsiValues)
	if exaggeratedDiv.Detected {
		if exaggeratedDiv.Direction == "SHORT" {
			details := fmt.Sprintf("RSI Overbought (%.0f%% confidence): %s", exaggeratedDiv.Confidence, exaggeratedDiv.Reasoning)
			return -1.0 * (exaggeratedDiv.Confidence / 100.0), details
		} else if exaggeratedDiv.Direction == "LONG" {
			details := fmt.Sprintf("RSI Oversold (%.0f%% confidence): %s", exaggeratedDiv.Confidence, exaggeratedDiv.Reasoning)
			return 1.0 * (exaggeratedDiv.Confidence / 100.0), details
		}
	}

	return 0.0, "" // No divergence detected
}

// it converts ensemble score to recommendation and reasoning
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
	rsiValues []float64,
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

	// Calculate divergence score if enough RSI data is available
	divergenceScore := 0.0
	divergenceDetails := ""
	if len(rsiValues) >= 20 {
		divergenceScore, divergenceDetails = calculateDivergenceScore(bars, rsiValues)
		components = append(components, SignalComponent{
			Name:   "Divergence",
			Score:  divergenceScore,
			Weight: DefaultSignalWeights["Divergence"],
		})
	}

	ensembleScore := (rsiScore * DefaultSignalWeights["RSI"]) +
		(atrScore * DefaultSignalWeights["ATR"]) +
		(whaleScore * DefaultSignalWeights["Whale"]) +
		(patternScore * DefaultSignalWeights["Pattern"]) +
		(srScore * DefaultSignalWeights["Support/Resistance"]) +
		(divergenceScore * DefaultSignalWeights["Divergence"])

	recommendation, reasoning := MapScoreToRecommendation(ensembleScore)

	confidence := 50.0 // Default for WAIT
	absScore := ensembleScore
	if absScore < 0 {
		absScore = -absScore
	}

	if recommendation == RecommendationBuy {
		confidence = 85.0 + ((ensembleScore - BuyThreshold) / BuyThreshold * 15.0)
		if confidence > 100 {
			confidence = 100
		}
	} else if recommendation == RecommendationAccumulate {
		confidence = 70.0 + ((ensembleScore - AccumulateThreshold) / 1.0 * 15.0)
	} else if recommendation == RecommendationSell {
		confidence = 85.0 + ((absScore + SellThreshold) / (-SellThreshold) * 15.0)
		if confidence > 100 {
			confidence = 100
		}
	} else if recommendation == RecommendationDistribute {
		confidence = 70.0 + ((absScore + DistributeThreshold) / 1.0 * 15.0)
	} else {
		confidence = 50.0 + (absScore / (-DistributeThreshold) * 20.0)
	}

	return CombinedSignal{
		Recommendation:    recommendation,
		Score:             ensembleScore,
		Confidence:        confidence,
		Reasoning:         reasoning,
		Components:        components,
		DivergenceDetails: divergenceDetails,
	}
}

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

func CombineMultiTimeframeSignals(daily, fourHour, oneHour CombinedSignal) MultiTimeframeSignal {
	result := MultiTimeframeSignal{
		DailySignal:      daily,
		FourHourSignal:   fourHour,
		OneHourSignal:    oneHour,
		Alignment:        false,
		AlignmentPercent: 0.0,
	}

	dailyBullish := daily.Recommendation == RecommendationBuy || daily.Recommendation == RecommendationAccumulate
	dailyBearish := daily.Recommendation == RecommendationSell || daily.Recommendation == RecommendationDistribute

	fourHourBullish := fourHour.Recommendation == RecommendationBuy || fourHour.Recommendation == RecommendationAccumulate
	fourHourBearish := fourHour.Recommendation == RecommendationSell || fourHour.Recommendation == RecommendationDistribute

	oneHourBullish := oneHour.Recommendation == RecommendationBuy || oneHour.Recommendation == RecommendationAccumulate
	oneHourBearish := oneHour.Recommendation == RecommendationSell || oneHour.Recommendation == RecommendationDistribute

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

	result.Alignment = alignedCount >= 2

	result.CompositeScore = (daily.Score * 0.5) + (fourHour.Score * 0.35) + (oneHour.Score * 0.15)

	result.Confidence = (daily.Confidence + fourHour.Confidence + oneHour.Confidence) / 3.0

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

func CalculateTradingRecommendation(price, rsi, support, resistance float64, trend string, pattern *detection.PatternSignal) map[string]interface{} {
	recommendation := "HOLD"
	confidence := 50.0
	reasoning := ""

	if rsi < 30 {
		recommendation = "BUY"
		confidence = 65.0
		reasoning = "RSI is oversold"

		if price < support*1.01 {
			confidence = 80.0
			reasoning += " and price is at support level"
		}
	} else if rsi > 70 {
		recommendation = "SELL"
		confidence = 65.0
		reasoning = "RSI is overbought"

		if price > resistance*0.99 {
			confidence = 80.0
			reasoning += " and price is at resistance level"
		}
	} else {
		// Check trend
		if trend == "bullish" {
			recommendation = "BUY"
			confidence = 60.0
			reasoning = "Bullish trend with RSI in neutral zone"
		} else if trend == "bearish" {
			recommendation = "SELL"
			confidence = 60.0
			reasoning = "Bearish trend with RSI in neutral zone"
		}
	}

	// Adjust based on pattern if available
	if pattern != nil && pattern.Detected {
		if pattern.Direction == "LONG" && (recommendation == "BUY" || recommendation == "HOLD") {
			confidence += (pattern.Confidence / 100.0) * 20
			reasoning += fmt.Sprintf(" - %s pattern supports upside", pattern.Pattern)
			recommendation = "BUY"
		} else if pattern.Direction == "SHORT" && (recommendation == "SELL" || recommendation == "HOLD") {
			confidence += (pattern.Confidence / 100.0) * 20
			reasoning += fmt.Sprintf(" - %s pattern suggests downside", pattern.Pattern)
			recommendation = "SELL"
		}
	}

	// Cap confidence at 100
	if confidence > 100 {
		confidence = 100
	}

	return map[string]interface{}{
		"action":     recommendation,
		"confidence": confidence,
		"reasoning":  reasoning,
	}
}
