package detection

import (
	"fmt"
	"math"

	"github.com/fazecat/mogulmaker/Internal/types"
)

// defines the type of detected pattern
type PatternType string

const (
	PatternDoubleBottom       PatternType = "DOUBLE_BOTTOM"
	PatternDoubleTip          PatternType = "DOUBLE_TOP"
	PatternHeadAndShoulders   PatternType = "HEAD_AND_SHOULDERS"
	PatternInverseHeadShould  PatternType = "INVERSE_HEAD_AND_SHOULDERS"
	PatternConsolidation      PatternType = "CONSOLIDATION"
	PatternConsolidationBreak PatternType = "CONSOLIDATION_BREAKOUT"
	PatternTriangle           PatternType = "TRIANGLE"
	PatternFlagPole           PatternType = "FLAG_POLE"
	PatternNone               PatternType = "NONE"
)

// represents a detected chart pattern and its trading signal
type PatternSignal struct {
	Pattern         PatternType
	Detected        bool
	Confidence      float64 // 0-100
	Direction       string  // "LONG", "SHORT", "NONE"
	SupportLevel    float64
	ResistanceLevel float64
	Reasoning       string
	FormationBars   int // Number of bars in the pattern
	PriceTargetUp   float64
	PriceTargetDown float64
	StopLossLevel   float64
	RiskRewardRatio float64
}

// analyzes price bars for chart patterns
type PatternDetector struct {
	MinFormationBars int     // Minimum bars to form a pattern
	TolerancePercent float64 // Tolerance for support/resistance levels (%)
	VerboseLogging   bool
}

// creates a new pattern detector with default settings
func NewPatternDetector() *PatternDetector {
	return &PatternDetector{
		MinFormationBars: 3,
		TolerancePercent: 1.5, // 1.5% tolerance
		VerboseLogging:   false,
	}
}

// runs all pattern detection algorithms
func (pd *PatternDetector) DetectAllPatterns(bars []types.Bar) []PatternSignal {
	signals := []PatternSignal{}

	if len(bars) < pd.MinFormationBars {
		return signals
	}

	// Check each pattern type
	if db := pd.DetectDoubleBottom(bars); db.Detected {
		signals = append(signals, db)
	}

	if dt := pd.DetectDoubleTop(bars); dt.Detected {
		signals = append(signals, dt)
	}

	if hs := pd.DetectHeadAndShoulders(bars); hs.Detected {
		signals = append(signals, hs)
	}

	if ihs := pd.DetectInverseHeadAndShoulders(bars); ihs.Detected {
		signals = append(signals, ihs)
	}

	if cb := pd.DetectConsolidationBreakout(bars); cb.Detected {
		signals = append(signals, cb)
	}

	if tri := pd.DetectTriangle(bars); tri.Detected {
		signals = append(signals, tri)
	}

	return signals
}

//	identifies a double bottom pattern (bullish reversal)
//
// Pattern: Low -> Recovery -> Low (similar height) -> Recovery up
func (pd *PatternDetector) DetectDoubleBottom(bars []types.Bar) PatternSignal {
	signal := PatternSignal{
		Pattern:   PatternDoubleBottom,
		Detected:  false,
		Direction: "NONE",
	}

	if len(bars) < 5 {
		return signal
	}

	// Find the lowest points (potential bottoms)
	lows := make([]float64, len(bars))
	for i, bar := range bars {
		lows[i] = bar.Low
	}

	for i := 1; i < len(bars)-2; i++ {
		for j := i + 2; j < len(bars)-1; j++ {
			bottom1 := lows[i]
			bottom2 := lows[j]

			// Check if both lows are similar
			pctDiff := math.Abs((bottom1-bottom2)/bottom1) * 100
			if pctDiff <= pd.TolerancePercent && pctDiff > 0.1 {
				// Check if there's recovery between the lows
				recovery := bars[i+1].High > bottom1*1.02 || bars[j-1].High > bottom1*1.02

				if recovery && j-i >= 3 { // At least 3 bars between
					signal.Detected = true
					signal.Pattern = PatternDoubleBottom
					signal.Direction = "LONG"
					signal.SupportLevel = math.Min(bottom1, bottom2)
					signal.ResistanceLevel = bars[i+1].High
					signal.FormationBars = j - i + 1
					signal.Confidence = 75.0 - (pctDiff * 5) // Lower confidence if bottoms differ more
					signal.Reasoning = fmt.Sprintf("Double bottom at %.2f with recovery to %.2f", signal.SupportLevel, signal.ResistanceLevel)

					// Calculate price targets
					neckline := signal.ResistanceLevel
					height := neckline - signal.SupportLevel
					signal.PriceTargetUp = neckline + height
					signal.StopLossLevel = signal.SupportLevel * 0.98
					signal.RiskRewardRatio = (signal.PriceTargetUp - neckline) / (neckline - signal.StopLossLevel)

					if pd.VerboseLogging {
						fmt.Printf("🟢 Double Bottom detected: %.2f support, target %.2f\n",
							signal.SupportLevel, signal.PriceTargetUp)
					}

					return signal
				}
			}
		}
	}

	return signal
}

//	identifies a double top pattern (bearish reversal)
//
// Pattern: High -> Pullback -> High (similar height) -> Pullback down
func (pd *PatternDetector) DetectDoubleTop(bars []types.Bar) PatternSignal {
	signal := PatternSignal{
		Pattern:   PatternDoubleTip,
		Detected:  false,
		Direction: "NONE",
	}

	if len(bars) < 5 {
		return signal
	}

	// Find the highest points (potential tops)
	highs := make([]float64, len(bars))
	for i, bar := range bars {
		highs[i] = bar.High
	}

	// Look for two highs that are approximately equal
	for i := 1; i < len(bars)-2; i++ {
		for j := i + 2; j < len(bars)-1; j++ {
			top1 := highs[i]
			top2 := highs[j]

			// Check if both highs are similar
			pctDiff := math.Abs((top1-top2)/top1) * 100
			if pctDiff <= pd.TolerancePercent && pctDiff > 0.1 {
				// Check if there's pullback between the tops
				pullback := bars[i+1].Low < top1*0.98 || bars[j-1].Low < top1*0.98

				if pullback && j-i >= 3 {
					signal.Detected = true
					signal.Pattern = PatternDoubleTip
					signal.Direction = "SHORT"
					signal.ResistanceLevel = math.Max(top1, top2)
					signal.SupportLevel = bars[i+1].Low
					signal.FormationBars = j - i + 1
					signal.Confidence = 75.0 - (pctDiff * 5)
					signal.Reasoning = fmt.Sprintf("Double top at %.2f with pullback to %.2f", signal.ResistanceLevel, signal.SupportLevel)

					// Calculate price targets
					neckline := signal.SupportLevel
					height := signal.ResistanceLevel - neckline
					signal.PriceTargetDown = neckline - height
					signal.StopLossLevel = signal.ResistanceLevel * 1.02
					signal.RiskRewardRatio = (signal.StopLossLevel - neckline) / (neckline - signal.PriceTargetDown)

					if pd.VerboseLogging {
						fmt.Printf("🔴 Double Top detected: %.2f resistance, target %.2f\n",
							signal.ResistanceLevel, signal.PriceTargetDown)
					}

					return signal
				}
			}
		}
	}

	return signal
}

// identifies a head and shoulders pattern (bearish reversal)
// Pattern: Shoulder (high) -> Head (higher high) -> Shoulder (similar to first)
func (pd *PatternDetector) DetectHeadAndShoulders(bars []types.Bar) PatternSignal {
	signal := PatternSignal{
		Pattern:   PatternHeadAndShoulders,
		Detected:  false,
		Direction: "NONE",
	}

	if len(bars) < 7 {
		return signal
	}

	// Look for: Low, High (shoulder), Low, High (head), Low, High (shoulder), Low
	for i := 1; i < len(bars)-5; i++ {
		shoulder1High := bars[i].High
		headHigh := bars[i+2].High
		shoulder2High := bars[i+4].High

		// Head should be highest
		if headHigh > shoulder1High && headHigh > shoulder2High {
			// Shoulders should be similar
			shoulderDiff := math.Abs((shoulder1High-shoulder2High)/shoulder1High) * 100
			if shoulderDiff <= pd.TolerancePercent*2 {
				signal.Detected = true
				signal.Pattern = PatternHeadAndShoulders
				signal.Direction = "SHORT"
				signal.ResistanceLevel = headHigh
				signal.SupportLevel = math.Min(bars[i+1].Low, math.Min(bars[i+3].Low, bars[i+5].Low))
				signal.FormationBars = 6
				signal.Confidence = 70.0
				signal.Reasoning = "Head and Shoulders pattern detected - bearish reversal"

				// Neckline is the support level between shoulders
				neckline := signal.SupportLevel
				height := signal.ResistanceLevel - neckline
				signal.PriceTargetDown = neckline - height
				signal.StopLossLevel = signal.ResistanceLevel * 1.02
				signal.RiskRewardRatio = (signal.StopLossLevel - neckline) / (neckline - signal.PriceTargetDown)

				if pd.VerboseLogging {
					fmt.Printf("🔴 Head and Shoulders detected\n")
				}

				return signal
			}
		}
	}

	return signal
}

// inverse head and shoulders (bullish reversal)
func (pd *PatternDetector) DetectInverseHeadAndShoulders(bars []types.Bar) PatternSignal {
	signal := PatternSignal{
		Pattern:   PatternInverseHeadShould,
		Detected:  false,
		Direction: "NONE",
	}

	if len(bars) < 7 {
		return signal
	}

	// Look for: High, Low (shoulder), High, Low (head), High, Low (shoulder), High
	for i := 1; i < len(bars)-5; i++ {
		shoulder1Low := bars[i].Low
		headLow := bars[i+2].Low
		shoulder2Low := bars[i+4].Low

		// Head should be lowest
		if headLow < shoulder1Low && headLow < shoulder2Low {
			// Shoulders should be similar
			shoulderDiff := math.Abs((shoulder1Low-shoulder2Low)/shoulder1Low) * 100
			if shoulderDiff <= pd.TolerancePercent*2 {
				signal.Detected = true
				signal.Pattern = PatternInverseHeadShould
				signal.Direction = "LONG"
				signal.SupportLevel = headLow
				signal.ResistanceLevel = math.Max(bars[i+1].High, math.Max(bars[i+3].High, bars[i+5].High))
				signal.FormationBars = 6
				signal.Confidence = 70.0
				signal.Reasoning = "Inverse Head and Shoulders pattern detected - bullish reversal"

				// Neckline is the resistance level between shoulders
				neckline := signal.ResistanceLevel
				height := neckline - signal.SupportLevel
				signal.PriceTargetUp = neckline + height
				signal.StopLossLevel = signal.SupportLevel * 0.98
				signal.RiskRewardRatio = (signal.PriceTargetUp - neckline) / (neckline - signal.StopLossLevel)

				if pd.VerboseLogging {
					fmt.Printf("🟢 Inverse Head and Shoulders detected\n")
				}

				return signal
			}
		}
	}

	return signal
}

// a consolidation phase (narrow range)
func (pd *PatternDetector) DetectConsolidation(bars []types.Bar) PatternSignal {
	signal := PatternSignal{
		Pattern:   PatternConsolidation,
		Detected:  false,
		Direction: "NONE",
	}

	if len(bars) < 5 {
		return signal
	}

	// Calculate average range over last N bars
	recentBars := bars[len(bars)-5:]
	avgHigh := 0.0
	avgLow := 0.0

	for _, bar := range recentBars {
		avgHigh += bar.High
		avgLow += bar.Low
	}

	avgHigh /= float64(len(recentBars))
	avgLow /= float64(len(recentBars))

	// Range should be less than 1% of price
	currentPrice := bars[len(bars)-1].Close
	rangePercent := ((avgHigh - avgLow) / currentPrice) * 100

	if rangePercent < 1.0 {
		signal.Detected = true
		signal.Pattern = PatternConsolidation
		signal.Direction = "NONE"
		signal.ResistanceLevel = avgHigh
		signal.SupportLevel = avgLow
		signal.Confidence = 60.0
		signal.Reasoning = fmt.Sprintf("Consolidation detected: range %.2f%%", rangePercent)

		if pd.VerboseLogging {
			fmt.Printf("⏸️  Consolidation pattern detected: %.2f%% range\n", rangePercent)
		}

		return signal
	}

	return signal
}

// identifies a breakout from consolidation
func (pd *PatternDetector) DetectConsolidationBreakout(bars []types.Bar) PatternSignal {
	signal := PatternSignal{
		Pattern:   PatternConsolidationBreak,
		Detected:  false,
		Direction: "NONE",
	}

	if len(bars) < 10 {
		return signal
	}

	// Find consolidation zone (last 5-8 bars with low range)
	consolidationBars := 6
	consolidationRange := bars[len(bars)-consolidationBars:]

	maxPrice := consolidationRange[0].High
	minPrice := consolidationRange[0].Low

	for _, bar := range consolidationRange {
		if bar.High > maxPrice {
			maxPrice = bar.High
		}
		if bar.Low < minPrice {
			minPrice = bar.Low
		}
	}

	rangePercent := ((maxPrice - minPrice) / minPrice) * 100

	// Consolidation should be tight
	if rangePercent > 1.5 {
		return signal
	}

	// Check if current bar breaks out
	currentBar := bars[len(bars)-1]
	prevBar := bars[len(bars)-2]

	// Breakout up
	if currentBar.Close > maxPrice && prevBar.Close < maxPrice && currentBar.Volume > int64(float64(prevBar.Volume)*1.3) {
		signal.Detected = true
		signal.Pattern = PatternConsolidationBreak
		signal.Direction = "LONG"
		signal.ResistanceLevel = maxPrice
		signal.SupportLevel = minPrice
		signal.Confidence = 80.0
		signal.Reasoning = "Upside breakout from consolidation"
		signal.PriceTargetUp = maxPrice + (maxPrice - minPrice)
		signal.StopLossLevel = minPrice * 0.98

		if pd.VerboseLogging {
			fmt.Printf("🟢 Consolidation breakout (UP) detected\n")
		}

		return signal
	}

	// Breakout down
	if currentBar.Close < minPrice && prevBar.Close > minPrice && currentBar.Volume > int64(float64(prevBar.Volume)*1.3) {
		signal.Detected = true
		signal.Pattern = PatternConsolidationBreak
		signal.Direction = "SHORT"
		signal.ResistanceLevel = maxPrice
		signal.SupportLevel = minPrice
		signal.Confidence = 80.0
		signal.Reasoning = "Downside breakout from consolidation"
		signal.PriceTargetDown = minPrice - (maxPrice - minPrice)
		signal.StopLossLevel = maxPrice * 1.02

		if pd.VerboseLogging {
			fmt.Printf("🔴 Consolidation breakout (DOWN) detected\n")
		}

		return signal
	}

	return signal
}

// identifies ascending, descending, or symmetrical triangle patterns
func (pd *PatternDetector) DetectTriangle(bars []types.Bar) PatternSignal {
	signal := PatternSignal{
		Pattern:   PatternTriangle,
		Detected:  false,
		Direction: "NONE",
	}

	if len(bars) < 6 {
		return signal
	}

	// Simplified: Check if highs are descending and lows are ascending (symmetrical/ascending triangle)
	recentBars := bars[len(bars)-6:]

	highs := make([]float64, len(recentBars))
	lows := make([]float64, len(recentBars))

	for i, bar := range recentBars {
		highs[i] = bar.High
		lows[i] = bar.Low
	}

	// Check for narrowing range (triangle characteristic)
	firstRange := highs[0] - lows[0]
	lastRange := highs[len(highs)-1] - lows[len(lows)-1]

	if lastRange < firstRange*0.7 { // Range should contract to 70% or less
		// Check trend direction
		highTrend := highs[0] > highs[len(highs)-1]
		lowTrend := lows[0] < lows[len(lows)-1]

		if highTrend && lowTrend {
			// Symmetrical triangle - potential breakout
			signal.Detected = true
			signal.Pattern = PatternTriangle
			signal.Direction = "NONE" // Awaiting breakout
			signal.ResistanceLevel = highs[len(highs)-1]
			signal.SupportLevel = lows[len(lows)-1]
			signal.Confidence = 60.0
			signal.Reasoning = "Triangle pattern forming - awaiting breakout"

			if pd.VerboseLogging {
				fmt.Printf("📐 Triangle pattern detected\n")
			}

			return signal
		}
	}

	return signal
}

// formatted string representation of a pattern
func FormatPatternSignal(signal PatternSignal) string {
	if !signal.Detected {
		return fmt.Sprintf("❌ No pattern detected")
	}

	emoji := "⏸️"
	if signal.Direction == "LONG" {
		emoji = "🟢"
	} else if signal.Direction == "SHORT" {
		emoji = "🔴"
	}

	return fmt.Sprintf(`%s %s (%.0f%% confidence)
   Formation: %d bars | Risk/Reward: %.2f
   Resistance: %.2f | Support: %.2f
   Reason: %s`,
		emoji, signal.Pattern, signal.Confidence,
		signal.FormationBars, signal.RiskRewardRatio,
		signal.ResistanceLevel, signal.SupportLevel,
		signal.Reasoning,
	)
}
