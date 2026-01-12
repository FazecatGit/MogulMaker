package strategy

import (
	"fmt"
	"math"

	"github.com/fazecat/mongelmaker/Internal/types"
)

// DivergenceType defines the type of divergence detected
type DivergenceType string

const (
	DivergenceBullish     DivergenceType = "BULLISH"
	DivergenceBearish     DivergenceType = "BEARISH"
	DivergenceNone        DivergenceType = "NONE"
	DivergenceHidden      DivergenceType = "HIDDEN"
	DivergenceExaggerated DivergenceType = "EXAGGERATED"
)

// DivergenceSignal represents a detected divergence between price and indicator
type DivergenceSignal struct {
	Type                DivergenceType
	Detected            bool
	Confidence          float64 // 0-100
	Direction           string  // "LONG", "SHORT", "NONE"
	IndicatorName       string  // "RSI", "MACD", "Stochastic", etc.
	PriceAction         string  // "HIGHER_HIGH", "LOWER_LOW", "LOWER_HIGH", "HIGHER_LOW"
	IndicatorAction     string  // Same as price
	Reasoning           string
	ReversalProbability float64
	FormationBars       int
}

// analyzes price and indicator divergence
type DivergenceDetector struct {
	MinFormationBars   int
	VerboseLogging     bool
	DivergenceLookback int // Number of bars to analyze
}

// creates a new divergence detector
func NewDivergenceDetector() *DivergenceDetector {
	return &DivergenceDetector{
		MinFormationBars:   3,
		VerboseLogging:     false,
		DivergenceLookback: 20,
	}
}

// a price level with its bar index
type PricePoint struct {
	Index int
	Price float64
}

// an indicator level with its bar index
type IndicatorPoint struct {
	Index int
	Value float64
}

// identifies divergence between price and RSI
// Bullish: Lower lows in price but higher lows in RSI (reversal up expected)
// Bearish: Higher highs in price but lower highs in RSI (reversal down expected)
func (dd *DivergenceDetector) DetectRSIDivergence(bars []types.Bar, rsiValues []float64) DivergenceSignal {
	signal := DivergenceSignal{
		Type:                DivergenceNone,
		Detected:            false,
		IndicatorName:       "RSI",
		Direction:           "NONE",
		ReversalProbability: 0.0,
	}

	if len(bars) < dd.MinFormationBars || len(rsiValues) != len(bars) {
		return signal
	}

	// Look back N bars
	lookback := dd.DivergenceLookback
	if lookback > len(bars) {
		lookback = len(bars)
	}

	startIdx := len(bars) - lookback
	if startIdx < 0 {
		startIdx = 0
	}

	recentBars := bars[startIdx:]
	recentRSI := rsiValues[startIdx:]

	// Find local lows and highs in price
	priceLows := findLocalExtrema(recentBars, true)   // true = lows
	priceHighs := findLocalExtrema(recentBars, false) // false = highs

	// Find local lows and highs in RSI
	rsiLows := findIndicatorExtrema(recentRSI, true)
	rsiHighs := findIndicatorExtrema(recentRSI, false)

	// BULLISH DIVERGENCE: Lower price lows but higher RSI lows
	if len(priceLows) >= 2 && len(rsiLows) >= 2 {
		priceLow1 := priceLows[len(priceLows)-2]
		priceLow2 := priceLows[len(priceLows)-1]
		rsiLow1 := rsiLows[len(rsiLows)-2]
		rsiLow2 := rsiLows[len(rsiLows)-1]

		if priceLow2.Price < priceLow1.Price && rsiLow2.Value > rsiLow1.Value {
			signal.Type = DivergenceBullish
			signal.Detected = true
			signal.Direction = "LONG"
			signal.PriceAction = "LOWER_LOW"
			signal.IndicatorAction = "HIGHER_LOW"
			signal.Confidence = calculateDivergenceConfidence(priceLow1.Price, priceLow2.Price, rsiLow1.Value, rsiLow2.Value, true)
			signal.FormationBars = priceLow2.Index - priceLow1.Index + 1
			signal.Reasoning = fmt.Sprintf("Bullish divergence: Price makes lower low (%.2f → %.2f) but RSI makes higher low (%.1f → %.1f)",
				priceLow1.Price, priceLow2.Price, rsiLow1.Value, rsiLow2.Value)
			signal.ReversalProbability = 0.65 + (signal.Confidence / 100 * 0.20) // 65-85%

			if dd.VerboseLogging {
				fmt.Printf("🟢 Bullish RSI Divergence detected: %.1f%% confidence\n", signal.Confidence)
			}

			return signal
		}
	}

	// BEARISH DIVERGENCE: Higher price highs but lower RSI highs
	if len(priceHighs) >= 2 && len(rsiHighs) >= 2 {
		priceHigh1 := priceHighs[len(priceHighs)-2]
		priceHigh2 := priceHighs[len(priceHighs)-1]
		rsiHigh1 := rsiHighs[len(rsiHighs)-2]
		rsiHigh2 := rsiHighs[len(rsiHighs)-1]

		if priceHigh2.Price > priceHigh1.Price && rsiHigh2.Value < rsiHigh1.Value {
			signal.Type = DivergenceBearish
			signal.Detected = true
			signal.Direction = "SHORT"
			signal.PriceAction = "HIGHER_HIGH"
			signal.IndicatorAction = "LOWER_HIGH"
			signal.Confidence = calculateDivergenceConfidence(priceHigh1.Price, priceHigh2.Price, rsiHigh1.Value, rsiHigh2.Value, false)
			signal.FormationBars = priceHigh2.Index - priceHigh1.Index + 1
			signal.Reasoning = fmt.Sprintf("Bearish divergence: Price makes higher high (%.2f → %.2f) but RSI makes lower high (%.1f → %.1f)",
				priceHigh1.Price, priceHigh2.Price, rsiHigh1.Value, rsiHigh2.Value)
			signal.ReversalProbability = 0.65 + (signal.Confidence / 100 * 0.20) // 65-85%

			if dd.VerboseLogging {
				fmt.Printf("🔴 Bearish RSI Divergence detected: %.1f%% confidence\n", signal.Confidence)
			}

			return signal
		}
	}

	return signal
}

// DetectHiddenDivergence identifies hidden divergence (continuation patterns)
// Bullish hidden: Higher price lows but lower RSI lows (continuation up)
// Bearish hidden: Lower price highs but higher RSI highs (continuation down)
func (dd *DivergenceDetector) DetectHiddenDivergence(bars []types.Bar, rsiValues []float64) DivergenceSignal {
	signal := DivergenceSignal{
		Type:                DivergenceNone,
		Detected:            false,
		IndicatorName:       "RSI",
		Direction:           "NONE",
		ReversalProbability: 0.0,
	}

	if len(bars) < dd.MinFormationBars || len(rsiValues) != len(bars) {
		return signal
	}

	lookback := dd.DivergenceLookback
	if lookback > len(bars) {
		lookback = len(bars)
	}

	startIdx := len(bars) - lookback
	if startIdx < 0 {
		startIdx = 0
	}

	recentBars := bars[startIdx:]
	recentRSI := rsiValues[startIdx:]

	priceLows := findLocalExtrema(recentBars, true)
	priceHighs := findLocalExtrema(recentBars, false)

	rsiLows := findIndicatorExtrema(recentRSI, true)
	rsiHighs := findIndicatorExtrema(recentRSI, false)

	// BULLISH HIDDEN: Higher price lows + lower RSI lows = continuation of uptrend
	if len(priceLows) >= 2 && len(rsiLows) >= 2 {
		priceLow1 := priceLows[len(priceLows)-2]
		priceLow2 := priceLows[len(priceLows)-1]
		rsiLow1 := rsiLows[len(rsiLows)-2]
		rsiLow2 := rsiLows[len(rsiLows)-1]

		if priceLow2.Price > priceLow1.Price && rsiLow2.Value < rsiLow1.Value {
			signal.Type = DivergenceHidden
			signal.Detected = true
			signal.Direction = "LONG"
			signal.PriceAction = "HIGHER_LOW"
			signal.IndicatorAction = "LOWER_LOW"
			signal.Confidence = 60.0
			signal.FormationBars = priceLow2.Index - priceLow1.Index + 1
			signal.Reasoning = "Hidden bullish divergence: Uptrend continuation expected"
			signal.ReversalProbability = 0.55

			if dd.VerboseLogging {
				fmt.Printf("🟢 Hidden Bullish Divergence detected\n")
			}

			return signal
		}
	}

	// BEARISH HIDDEN: Lower price highs + higher RSI highs = continuation of downtrend
	if len(priceHighs) >= 2 && len(rsiHighs) >= 2 {
		priceHigh1 := priceHighs[len(priceHighs)-2]
		priceHigh2 := priceHighs[len(priceHighs)-1]
		rsiHigh1 := rsiHighs[len(rsiHighs)-2]
		rsiHigh2 := rsiHighs[len(rsiHighs)-1]

		if priceHigh2.Price < priceHigh1.Price && rsiHigh2.Value > rsiHigh1.Value {
			signal.Type = DivergenceHidden
			signal.Detected = true
			signal.Direction = "SHORT"
			signal.PriceAction = "LOWER_HIGH"
			signal.IndicatorAction = "HIGHER_HIGH"
			signal.Confidence = 60.0
			signal.FormationBars = priceHigh2.Index - priceHigh1.Index + 1
			signal.Reasoning = "Hidden bearish divergence: Downtrend continuation expected"
			signal.ReversalProbability = 0.55

			if dd.VerboseLogging {
				fmt.Printf("🔴 Hidden Bearish Divergence detected\n")
			}

			return signal
		}
	}

	return signal
}

// extreme divergence (overbought/oversold)
func (dd *DivergenceDetector) DetectExaggeratedDivergence(rsiValues []float64) DivergenceSignal {
	signal := DivergenceSignal{
		Type:                DivergenceNone,
		Detected:            false,
		IndicatorName:       "RSI",
		Direction:           "NONE",
		ReversalProbability: 0.0,
	}

	if len(rsiValues) == 0 {
		return signal
	}

	currentRSI := rsiValues[len(rsiValues)-1]

	// Extremely overbought
	if currentRSI > 80 {
		signal.Type = DivergenceExaggerated
		signal.Detected = true
		signal.Direction = "SHORT"
		signal.Confidence = math.Min(95, (currentRSI/100)*100) // Cap at 95%
		signal.Reasoning = fmt.Sprintf("Extreme RSI overbought: %.1f (pullback likely)", currentRSI)
		signal.ReversalProbability = 0.70

		if dd.VerboseLogging {
			fmt.Printf("⚠️  Exaggerated overbought condition: RSI %.1f\n", currentRSI)
		}

		return signal
	}

	// Extremely oversold
	if currentRSI < 20 {
		signal.Type = DivergenceExaggerated
		signal.Detected = true
		signal.Direction = "LONG"
		signal.Confidence = math.Min(95, ((100-currentRSI)/100)*100) // Cap at 95%
		signal.Reasoning = fmt.Sprintf("Extreme RSI oversold: %.1f (bounce expected)", currentRSI)
		signal.ReversalProbability = 0.70

		if dd.VerboseLogging {
			fmt.Printf("⚠️  Exaggerated oversold condition: RSI %.1f\n", currentRSI)
		}

		return signal
	}

	return signal
}

// Helper functions

func findLocalExtrema(bars []types.Bar, findLows bool) []PricePoint {
	extrema := []PricePoint{}

	if len(bars) < 3 {
		return extrema
	}

	for i := 1; i < len(bars)-1; i++ {
		if findLows {
			// Find local lows
			if bars[i].Low < bars[i-1].Low && bars[i].Low < bars[i+1].Low {
				extrema = append(extrema, PricePoint{Index: i, Price: bars[i].Low})
			}
		} else {
			// Find local highs
			if bars[i].High > bars[i-1].High && bars[i].High > bars[i+1].High {
				extrema = append(extrema, PricePoint{Index: i, Price: bars[i].High})
			}
		}
	}

	return extrema
}

func findIndicatorExtrema(values []float64, findLows bool) []IndicatorPoint {
	extrema := []IndicatorPoint{}

	if len(values) < 3 {
		return extrema
	}

	for i := 1; i < len(values)-1; i++ {
		if findLows {
			// Find local lows
			if values[i] < values[i-1] && values[i] < values[i+1] {
				extrema = append(extrema, IndicatorPoint{Index: i, Value: values[i]})
			}
		} else {
			// Find local highs
			if values[i] > values[i-1] && values[i] > values[i+1] {
				extrema = append(extrema, IndicatorPoint{Index: i, Value: values[i]})
			}
		}
	}

	return extrema
}

func calculateDivergenceConfidence(price1, price2, indicator1, indicator2 float64, isLow bool) float64 {
	// Calculate the magnitude of divergence
	priceMagnitude := math.Abs((price2-price1)/price1) * 100
	indicatorMagnitude := math.Abs((indicator2-indicator1)/100) * 100

	// Larger divergence = higher confidence
	divergenceMagnitude := (priceMagnitude + indicatorMagnitude) / 2

	// Base confidence: 60% + magnitude bonus (up to 35%)
	confidence := 60.0 + math.Min(35.0, divergenceMagnitude*2)

	return math.Min(95.0, confidence)
}

// returns a formatted string representation
func FormatDivergenceSignal(signal DivergenceSignal) string {
	if !signal.Detected {
		return "❌ No divergence detected"
	}

	emoji := "⏸️"
	if signal.Direction == "LONG" {
		emoji = "🟢"
	} else if signal.Direction == "SHORT" {
		emoji = "🔴"
	}

	return fmt.Sprintf(`%s %s %s (%.0f%% confidence)
   Type: %s | Reversal Probability: %.0f%%
   Formation: %d bars
   %s`,
		emoji, signal.IndicatorName, signal.Type, signal.Confidence,
		signal.Type, signal.ReversalProbability*100,
		signal.FormationBars,
		signal.Reasoning,
	)
}
