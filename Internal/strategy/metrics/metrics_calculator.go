package metrics

import (
	"math"
	"time"

	"github.com/fazecat/mogulmaker/Internal/utils"
)

type SymbolStats struct {
	Symbol       string
	TotalTrades  int
	Wins         int
	Losses       int
	SharpeRatio  float64
	SortinoRatio float64
	CalmarRatio  float64
}

type TradeResult struct {
	Symbol        string
	EntryPrice    float64
	ExitPrice     float64
	Quantity      float64
	PnL           float64
	ReturnPercent float64
	Duration      time.Duration
	EntryTime     time.Time
	ExitTime      time.Time
}

func CalculateSharpeRatio(trades []TradeResult, riskFreeRate float64) float64 {
	if len(trades) == 0 {
		return 0.0
	}
	var returns []float64
	for _, trade := range trades {
		returns = append(returns, trade.ReturnPercent)
	}
	avgReturn := utils.Average(returns)

	stdDev := calculateStandardDeviation(returns)

	if stdDev == 0 {
		return 0.0
	}
	return (avgReturn - riskFreeRate) / stdDev
}

func CalculateSortinoRatio(trades []TradeResult, riskFreeRate float64) float64 {
	if len(trades) == 0 {
		return 0.0
	}
	var returns []float64
	var negativeReturns []float64
	for _, trade := range trades {
		if trade.ReturnPercent < 0 {
			negativeReturns = append(negativeReturns, trade.ReturnPercent)
		}
		returns = append(returns, trade.ReturnPercent)
	}
	avgReturn := utils.Average(returns)
	downsideDev := calculateStandardDeviation(negativeReturns)
	if downsideDev == 0 {
		return 0.0
	}
	return (avgReturn - riskFreeRate) / downsideDev
}

func CalculateCalmarRatio(trades []TradeResult, annualReturn float64, maxDrawdown float64) float64 {
	if maxDrawdown == 0 {
		return 0.0
	}
	var returns []float64
	for _, trade := range trades {
		returns = append(returns, trade.ReturnPercent)
	}

	return annualReturn / maxDrawdown
}

func CalculateSymbolStats(trades []TradeResult) map[string]*SymbolStats {

	Trademap := make(map[string][]TradeResult)
	Results := make(map[string]*SymbolStats)

	for _, trade := range trades {
		Trademap[trade.Symbol] = append(Trademap[trade.Symbol], trade)
	}

	for symbol, tradesForSymbol := range Trademap {
		wins := 0
		losses := 0
		for _, trade := range tradesForSymbol {
			if trade.PnL > 0 {
				wins++
			} else if trade.PnL < 0 {
				losses++
			}
		}
		// 2% risk-free rate assumed
		Sharpe := CalculateSharpeRatio(tradesForSymbol, 0.02)
		Sortino := CalculateSortinoRatio(tradesForSymbol, 0.02)
		Calmar := CalculateCalmarRatio(tradesForSymbol, Sharpe*0.02, 0.1)
		symbolStats := &SymbolStats{
			Symbol:       symbol,
			TotalTrades:  len(tradesForSymbol),
			SharpeRatio:  Sharpe,
			SortinoRatio: Sortino,
			CalmarRatio:  Calmar,
			Wins:         wins,
			Losses:       losses,
		}
		Results[symbol] = symbolStats

	}
	return Results
}

func calculateStandardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	mean := utils.Average(values)
	varianceSum := 0.0
	for _, v := range values {
		varianceSum += (v - mean) * (v - mean)
	}
	variance := varianceSum / float64(len(values))
	return math.Sqrt(variance)
}
