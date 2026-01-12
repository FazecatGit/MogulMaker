package strategy

import (
	"fmt"

	"github.com/fazecat/mongelmaker/Internal/types"
)

// comment for future (adjust the values instead of hardcoding)

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

// calculateWhaleScore determines whale signal from detected whale events
func calculateWhaleScore(symbol string, bars []types.Bar) float64 {
	whales := DetectWhales(symbol, bars)

	if len(whales) == 0 {
		return 0.0 // No whales detected
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
	support := FindSupport(bars)
	resistance := FindResistance(bars)
	currentPrice := bars[len(bars)-1].Close

	if IsAtSupport(currentPrice, support) {
		return 1.0 // At support = buy opportunity
	}
	if IsAtResistance(currentPrice, resistance) {
		return -1.0 // At resistance = sell pressure
	}
	return 0.0
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
			Weight: 0.25,
		})
	}

	atrScore := 0.0
	if atrValue != nil && len(bars) > 0 {
		atrScore = calculateATRScore(*atrValue, bars[0].Close)
		components = append(components, SignalComponent{
			Name:   "ATR",
			Score:  atrScore,
			Weight: 0.15,
		})
	}

	whaleScore := calculateWhaleScore(symbol, bars)
	components = append(components, SignalComponent{
		Name:   "Whale",
		Score:  whaleScore,
		Weight: 0.30,
	})

	patternScore := calculatePatternScore(analysis)
	components = append(components, SignalComponent{
		Name:   "Pattern",
		Score:  patternScore,
		Weight: 0.20,
	})

	srScore := calculateSRScore(bars)
	components = append(components, SignalComponent{
		Name:   "Support/Resistance",
		Score:  srScore,
		Weight: 0.10,
	})

	// Calculate weighted ensemble score
	ensembleScore := (rsiScore * 0.25) + (atrScore * 0.15) + (whaleScore * 0.30) + (patternScore * 0.20) + (srScore * 0.10)

	// Map to recommendation
	recommendation := "WAIT"
	reasoning := "Neutral signals"

	if ensembleScore >= 1.5 {
		recommendation = "BUY"
		reasoning = "Strong buy signals"
	} else if ensembleScore >= 0.5 {
		recommendation = "ACCUMULATE"
		reasoning = "Moderate buy signals"
	} else if ensembleScore <= -1.5 {
		recommendation = "SELL"
		reasoning = "Strong sell signals"
	} else if ensembleScore <= -0.5 {
		recommendation = "DISTRIBUTE"
		reasoning = "Moderate sell signals"
	}

	confidence := (ensembleScore / 3.0) * 100
	if confidence < 0 {
		confidence = -confidence
	}

	return CombinedSignal{
		Recommendation: recommendation,
		Score:          ensembleScore,
		Confidence:     confidence,
		Reasoning:      reasoning,
		Components:     components,
	}
}

// CombineMultiTimeframeSignals analyzes alignment across timeframes
// Reduces false signals by requiring confirmation across timeframes
func CombineMultiTimeframeSignals(daily, fourHour, oneHour CombinedSignal) MultiTimeframeSignal {
	result := MultiTimeframeSignal{
		DailySignal:      daily,
		FourHourSignal:   fourHour,
		OneHourSignal:    oneHour,
		Alignment:        false,
		AlignmentPercent: 0.0,
	}

	// Extract signal directions (BUY/ACCUMULATE = bullish, SELL/DISTRIBUTE = bearish)
	dailyBullish := daily.Recommendation == "BUY" || daily.Recommendation == "ACCUMULATE"
	dailyBearish := daily.Recommendation == "SELL" || daily.Recommendation == "DISTRIBUTE"

	fourHourBullish := fourHour.Recommendation == "BUY" || fourHour.Recommendation == "ACCUMULATE"
	fourHourBearish := fourHour.Recommendation == "SELL" || fourHour.Recommendation == "DISTRIBUTE"

	oneHourBullish := oneHour.Recommendation == "BUY" || oneHour.Recommendation == "ACCUMULATE"
	oneHourBearish := oneHour.Recommendation == "SELL" || oneHour.Recommendation == "DISTRIBUTE"

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

// IsMultiTimeframeConfirmed checks if signal is strong enough to execute
// Goal: Reduce false signals by ~60% through multi-timeframe confirmation
func (m *MultiTimeframeSignal) IsMultiTimeframeConfirmed(requireStrongAlignment bool) bool {
	if requireStrongAlignment {
		// Strict: Daily + 4H must agree, 1H should not contradict
		if !m.Alignment {
			return false
		}
		// Additional check: daily trend is primary
		return m.DailySignal.Confidence > 50.0 && m.FourHourSignal.Confidence > 50.0
	}

	// Loose: Any 2 timeframes aligned
	return m.AlignmentPercent >= 66.0 // 2 out of 3 aligned
}

func FormatSignal(signal CombinedSignal) string {
	emoji := "⏸️"
	if signal.Recommendation == "BUY" || signal.Recommendation == "ACCUMULATE" {
		emoji = "🟢"
	} else if signal.Recommendation == "SELL" || signal.Recommendation == "DISTRIBUTE" {
		emoji = "🔴"
	}

	return fmt.Sprintf("%s %s (%.0f%% confidence) - %s",
		emoji,
		signal.Recommendation,
		signal.Confidence,
		signal.Reasoning,
	)
}

// FormatMultiTimeframeSignal provides detailed multi-timeframe analysis output
func FormatMultiTimeframeSignal(signal MultiTimeframeSignal) string {
	alignment := "❌"
	if signal.Alignment {
		alignment = "✅"
	}

	return fmt.Sprintf(`
📊 Multi-Timeframe Analysis:
  Daily:    %s (%.0f%%)
  4H:       %s (%.0f%%)
  1H:       %s (%.0f%%)
  
  %s Alignment: %.0f%% | Composite Score: %.2f | Confidence: %.0f%%
  📈 Recommended: %s
`,
		signal.DailySignal.Recommendation, signal.DailySignal.Confidence,
		signal.FourHourSignal.Recommendation, signal.FourHourSignal.Confidence,
		signal.OneHourSignal.Recommendation, signal.OneHourSignal.Confidence,
		alignment, signal.AlignmentPercent, signal.CompositeScore, signal.Confidence,
		signal.RecommendedTrade,
	)
}
