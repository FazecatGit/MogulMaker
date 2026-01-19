package analyzer

import (
	"math"

	"github.com/fazecat/mogulmaker/Internal/types"
)

type Candlestick struct {
	Open  float64
	Close float64
	High  float64
	Low   float64
}
type CommonMetrics struct {
	Body        float64
	RangeVal    float64
	BodyPct     float64
	UpperWick   float64
	LowerWick   float64
	BodyToUpper float64
	BodyToLower float64
}

func calculateCommonMetrics(candle Candlestick) CommonMetrics {
	body := math.Abs(candle.Close - candle.Open)
	rangeVal := candle.High - candle.Low
	bodyPct := 0.0
	if rangeVal != 0 {
		bodyPct = (body / rangeVal) * 100
	}
	upperWick := candle.High - math.Max(candle.Open, candle.Close)
	lowerWick := math.Min(candle.Open, candle.Close) - candle.Low
	bodyToUpper := 0.0
	bodyToLower := 0.0
	if upperWick != 0 {
		bodyToUpper = body / upperWick
	}
	if lowerWick != 0 {
		bodyToLower = body / lowerWick
	}
	return CommonMetrics{
		Body:        body,
		RangeVal:    rangeVal,
		BodyPct:     bodyPct,
		UpperWick:   upperWick,
		LowerWick:   lowerWick,
		BodyToUpper: bodyToUpper,
		BodyToLower: bodyToLower,
	}
}

func AnalyzeCandlestick(candle Candlestick) (map[string]float64, map[string]string) {
	common := calculateCommonMetrics(candle)
	upperPct := 0.0
	lowerPct := 0.0
	if common.RangeVal != 0 {
		upperPct = (common.UpperWick / common.RangeVal) * 100
		lowerPct = (common.LowerWick / common.RangeVal) * 100
	}
	metrics := map[string]float64{
		"Body":        common.Body,
		"Range":       common.RangeVal,
		"BodyPct":     common.BodyPct,
		"UpperWick":   common.UpperWick,
		"LowerWick":   common.LowerWick,
		"UpperPct":    upperPct,
		"LowerPct":    lowerPct,
		"BodyToUpper": common.BodyToUpper,
		"BodyToLower": common.BodyToLower,
	}
	analysis := "Neutral"
	if common.BodyPct < 10 {
		if upperPct > 70 {
			analysis = "Bearish Rejection"
		} else if lowerPct > 70 {
			analysis = "Bullish Rejection"
		} else {
			analysis = "Doji (indecision)"
		}
	} else if candle.Close > candle.Open {
		if common.BodyPct > 60 {
			analysis = "Strong Bullish"
		} else {
			analysis = "Bullish"
		}
	} else {
		if common.BodyPct > 60 {
			analysis = "Strong Bearish"
		} else {
			analysis = "Bearish"
		}
	}

	stringResults := map[string]string{
		"Analysis": analysis,
	}

	return metrics, stringResults
}

func PatternAnalyzeCandle(bar types.Bar, atr *float64, avgVol20 float64, vol int64) (signal string, confidence float64) {
	candle := Candlestick{
		Open:  bar.Open,
		Close: bar.Close,
		High:  bar.High,
		Low:   bar.Low,
	}

	common := calculateCommonMetrics(candle)
	bodyOverATR := 0.0
	if atr != nil && *atr != 0 {
		bodyOverATR = common.Body / *atr
	}
	volRatio := 0.0
	if avgVol20 != 0 {
		volRatio = float64(vol) / avgVol20
	}

	// Determine signal based on patterns
	confidence = 0.0
	signal = "Neutral"

	// Bull patterns
	if common.LowerWick > common.Body*1.5 && common.UpperWick < common.Body*0.5 && candle.Close > candle.Open {
		signal = "Bullish Hammer"
		confidence = 0.8
	} else if common.BodyPct > 70 && candle.Close > candle.Open && volRatio > 1.2 {
		signal = "Strong Bullish"
		confidence = 0.9
	} else if common.BodyToLower > 2.0 && candle.Close > candle.Open {
		signal = "Bullish Engulfing Potential"
		confidence = 0.7
	}

	// Bear patterns
	if common.UpperWick > common.Body*1.5 && common.LowerWick < common.Body*0.5 && candle.Close < candle.Open {
		signal = "Bearish Shooting Star"
		confidence = 0.8
	} else if common.BodyPct > 70 && candle.Close < candle.Open && volRatio > 1.2 {
		signal = "Strong Bearish"
		confidence = 0.9
	}

	// Adjust confidence values based on ATR and volume
	if atr != nil && bodyOverATR > 1.5 {
		confidence += 0.1
	}
	if volRatio > 1.5 {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return signal, confidence
}
