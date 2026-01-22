package interactive

import (
	"context"
	"fmt"
	"time"

	datafeed "github.com/fazecat/mogulmaker/Internal/database"
	sqlc "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	"github.com/fazecat/mogulmaker/Internal/export"
	newsscraping "github.com/fazecat/mogulmaker/Internal/news_scraping"
	"github.com/fazecat/mogulmaker/Internal/strategy/detection"
	"github.com/fazecat/mogulmaker/Internal/strategy/indicators"
	"github.com/fazecat/mogulmaker/Internal/strategy/signals"
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils"
	"github.com/fazecat/mogulmaker/Internal/utils/analyzer"
	"github.com/fazecat/mogulmaker/Internal/utils/scanner"
	"github.com/fazecat/mogulmaker/Internal/utils/scoring"
)

func FetchMarketData(symbol string, timeframe string, limit int, startDate string) ([]datafeed.Bar, error) {
	if timeframe == "" {
		return nil, fmt.Errorf("timeframe cannot be empty")
	}

	if limit < 14 {
		limit = 14
	}

	bars, err := datafeed.GetAlpacaBars(symbol, timeframe, limit, startDate)
	if err != nil {
		return nil, err
	}

	if len(bars) < 14 {
		return nil, fmt.Errorf("not enough data to calculate RSI/ATR: fetched %d bars, need at least 14", len(bars))
	}

	return bars, nil
}

func FetchMarketDataWithType(symbol string, timeframe string, limit int, startDate string, assetType string) ([]datafeed.Bar, error) {
	if timeframe == "" {
		return nil, fmt.Errorf("timeframe cannot be empty")
	}

	if limit < 14 {
		limit = 14
	}

	bars, err := datafeed.GetAlpacaBarsWithType(symbol, timeframe, limit, startDate, assetType)
	if err != nil {
		return nil, err
	}

	if len(bars) < 14 {
		return nil, fmt.Errorf("not enough data to calculate RSI/ATR: fetched %d bars, need at least 14", len(bars))
	}

	return bars, nil
}

func FetchMultiTimeframeSignals(symbol string, assetType string) (*signals.MultiTimeframeSignal, error) {
	dailyBars, err := datafeed.GetAlpacaBarsWithType(symbol, "1Day", 100, "", assetType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daily data: %w", err)
	}

	fourHourBars, err := datafeed.GetAlpacaBarsWithType(symbol, "4Hour", 100, "", assetType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch 4H data: %w", err)
	}

	oneHourBars, err := datafeed.GetAlpacaBarsWithType(symbol, "1Hour", 100, "", assetType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch 1H data: %w", err)
	}

	// Extract closes for RSI calculation
	dailyCloses := make([]float64, len(dailyBars))
	for i, bar := range dailyBars {
		dailyCloses[i] = bar.Close
	}
	fourHourCloses := make([]float64, len(fourHourBars))
	for i, bar := range fourHourBars {
		fourHourCloses[i] = bar.Close
	}
	oneHourCloses := make([]float64, len(oneHourBars))
	for i, bar := range oneHourBars {
		oneHourCloses[i] = bar.Close
	}

	// Calculate RSI for each timeframe
	dailyRSIValues, err := indicators.CalculateRSI(dailyCloses, 14)
	if err != nil || len(dailyRSIValues) == 0 {
		return nil, fmt.Errorf("failed to calculate daily RSI: %w", err)
	}
	dailyRSI := dailyRSIValues[len(dailyRSIValues)-1]

	fourHourRSIValues, err := indicators.CalculateRSI(fourHourCloses, 14)
	if err != nil || len(fourHourRSIValues) == 0 {
		return nil, fmt.Errorf("failed to calculate 4H RSI: %w", err)
	}
	fourHourRSI := fourHourRSIValues[len(fourHourRSIValues)-1]

	oneHourRSIValues, err := indicators.CalculateRSI(oneHourCloses, 14)
	if err != nil || len(oneHourRSIValues) == 0 {
		return nil, fmt.Errorf("failed to calculate 1H RSI: %w", err)
	}
	oneHourRSI := oneHourRSIValues[len(oneHourRSIValues)-1]

	// Calculate ATR using scoring helper
	dailyATR := scoring.CalculateATRFromBars(dailyBars)
	fourHourATR := scoring.CalculateATRFromBars(fourHourBars)
	oneHourATR := scoring.CalculateATRFromBars(oneHourBars)

	// Analyze each timeframe
	dailyCandle := analyzer.Candlestick{Open: dailyBars[len(dailyBars)-1].Open, Close: dailyBars[len(dailyBars)-1].Close, High: dailyBars[len(dailyBars)-1].High, Low: dailyBars[len(dailyBars)-1].Low}
	_, dailyResults := analyzer.AnalyzeCandlestick(dailyCandle)
	dailyAnalysis := dailyResults["Analysis"]

	fourHourCandle := analyzer.Candlestick{Open: fourHourBars[len(fourHourBars)-1].Open, Close: fourHourBars[len(fourHourBars)-1].Close, High: fourHourBars[len(fourHourBars)-1].High, Low: fourHourBars[len(fourHourBars)-1].Low}
	_, fourHourResults := analyzer.AnalyzeCandlestick(fourHourCandle)
	fourHourAnalysis := fourHourResults["Analysis"]

	oneHourCandle := analyzer.Candlestick{Open: oneHourBars[len(oneHourBars)-1].Open, Close: oneHourBars[len(oneHourBars)-1].Close, High: oneHourBars[len(oneHourBars)-1].High, Low: oneHourBars[len(oneHourBars)-1].Low}
	_, oneHourResults := analyzer.AnalyzeCandlestick(oneHourCandle)
	oneHourAnalysis := oneHourResults["Analysis"]

	// Generate signals for each timeframe
	dailySignal := signals.CalculateSignal(&dailyRSI, &dailyATR, dailyBars, symbol, dailyAnalysis, dailyRSIValues)
	fourHourSignal := signals.CalculateSignal(&fourHourRSI, &fourHourATR, fourHourBars, symbol, fourHourAnalysis, fourHourRSIValues)
	oneHourSignal := signals.CalculateSignal(&oneHourRSI, &oneHourATR, oneHourBars, symbol, oneHourAnalysis, oneHourRSIValues)

	// Combine multi-timeframe signals
	multiSignal := signals.CombineMultiTimeframeSignals(dailySignal, fourHourSignal, oneHourSignal)
	return &multiSignal, nil
}

func PickStockFromResults(results []scanner.StockScore) (string, error) {
	fmt.Println("\nSelect a stock to analyze in detail:")
	for i, result := range results {
		rsiStr := "-"
		if result.RSI != nil {
			rsiStr = fmt.Sprintf("%.1f", *result.RSI)
		}
		atrStr := "-"
		if result.ATR != nil {
			atrStr = fmt.Sprintf("%.2f", *result.ATR)
		}
		fmt.Printf("%d. %s (Score: %.1f, RSI: %s, ATR: %s)\n",
			i+1, result.Symbol, result.Score, rsiStr, atrStr)
	}
	fmt.Print("Enter choice: ")
	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil || choice < 1 || choice > len(results) {
		fmt.Println("Invalid input.")
		return "", fmt.Errorf("invalid choice")
	}
	return results[choice-1].Symbol, nil
}

func DisplayBasicData(bars []datafeed.Bar, symbol string, timeframe string) {
	fmt.Printf("\n[DATA] Basic Data for %s (%s)\n", symbol, timeframe)
	fmt.Println("Timestamp           | Close Price | Volume")
	fmt.Println("--------------------|-------------|----------")

	for _, bar := range bars {
		fmt.Printf("%-20s | %11.2f | %8d\n", bar.Timestamp, bar.Close, bar.Volume)

	}
}

func DisplayAdvancedData(bars []datafeed.Bar, symbol string, timeframe string) {
	fmt.Printf("\n[DATA] Advanced Data for %s (%s)\n", symbol, timeframe)
	fmt.Println("Timestamp           | Open Price | High Price | Low Price | Close Price | Volume")
	fmt.Println("--------------------|------------|------------|-----------|-------------|----------")

	for _, bar := range bars {
		fmt.Printf("%-20s | %11.2f | %11.2f | %9.2f | %11.2f | %8d\n",
			bar.Timestamp, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
	}
}

func DisplayAnalyticsData(bars []datafeed.Bar, symbol string, timeframe string, tz *time.Location, queries *sqlc.Queries, newsStorage *newsscraping.NewsStorage) {
	fmt.Printf("\n[ANALYTICS] Analytics Data for %s (%s) - Timezone: %s\n", symbol, timeframe, tz.String())

	// Display news first if available
	if newsStorage != nil {
		displayNewsForSymbol(symbol, newsStorage)
	}

	var startTime, endTime time.Time
	if len(bars) > 0 {
		firstBar, err := time.Parse(time.RFC3339, bars[0].Timestamp)
		if err == nil {
			startTime = firstBar
		}
		lastBar, err := time.Parse(time.RFC3339, bars[len(bars)-1].Timestamp)
		if err == nil {
			endTime = lastBar
		}
	}

	var rsiMap map[string]float64
	var atrMap map[string]float64
	var err error

	// Try to fetch from database first
	if !startTime.IsZero() && !endTime.IsZero() {
		rsiMap, err = datafeed.FetchRSIByTimestampRange(symbol, startTime, endTime)
		if err != nil {
			rsiMap = make(map[string]float64)
		}

		atrMap, err = datafeed.FetchATRByTimestampRange(symbol, startTime, endTime)
		if err != nil {
			atrMap = make(map[string]float64)
		}
	} else {
		rsiMap = make(map[string]float64)
		atrMap = make(map[string]float64)
	}

	// If database is empty or has insufficient data, calculate from bars
	if len(rsiMap) == 0 && len(bars) >= 14 {
		closes := make([]float64, len(bars))
		for i, bar := range bars {
			closes[i] = bar.Close
		}

		rsiValues, err := indicators.CalculateRSI(closes, 14)
		if err == nil && len(rsiValues) > 0 {
			// Map RSI values to timestamps
			startIdx := len(bars) - len(rsiValues)
			for i, rsi := range rsiValues {
				barIdx := startIdx + i
				if barIdx >= 0 && barIdx < len(bars) {
					t, _ := time.Parse(time.RFC3339, bars[barIdx].Timestamp)
					timestampStr := t.Format("2006-01-02 15:04:05")
					rsiMap[timestampStr] = rsi
				}
			}
		}
	}

	// Calculate ATR from bars if not in database
	if len(atrMap) == 0 && len(bars) >= 14 {
		atrValue := scoring.CalculateATRFromBars(bars)
		// Store same ATR for all recent bars
		for _, bar := range bars {
			t, _ := time.Parse(time.RFC3339, bar.Timestamp)
			timestampStr := t.Format("2006-01-02 15:04:05")
			atrMap[timestampStr] = atrValue
		}
	}

	fmt.Println("Timestamp           | Close Price | Price Chg | Chg %  | Volume   | RSI    | ATR    | B/U Ratio | B/L Ratio | Analysis                  | Signals             ")
	fmt.Println("--------------------|-------------|-----------|--------|----------|--------|--------|-----------|-----------|--------------------------|---------------------")

	var latestAnalysis string
	var latestRSI *float64
	var latestATR *float64

	for i, bar := range bars {
		priceChange := bar.Close - bar.Open
		priceChangePercent := (bar.Close - bar.Open) / bar.Open * 100

		t, err := time.Parse(time.RFC3339, bar.Timestamp)
		if err != nil {
			fmt.Printf("⚠️  Could not parse timestamp: %v\n", err)
		}

		var timestampStr string
		var displayTimestamp string

		if err == nil {
			localTime := t.In(tz)
			displayTimestamp = localTime.Format("2006-01-02 15:04:05")
			timestampStr = t.Format("2006-01-02 15:04:05")
		} else {
			timestampStr = bar.Timestamp
			displayTimestamp = bar.Timestamp
		}

		rsiVal, hasRSI := rsiMap[timestampStr]
		atrVal, hasATR := atrMap[timestampStr]

		rsiStr := "  -   "
		if hasRSI {
			rsiStr = fmt.Sprintf("%6.2f", rsiVal)
		}

		atrStr := "  -   "
		if hasATR {
			atrStr = fmt.Sprintf("%6.2f", atrVal)
		}

		candle := analyzer.Candlestick{
			Open:  bar.Open,
			Close: bar.Close,
			High:  bar.High,
			Low:   bar.Low,
		}
		metrics, results := analyzer.AnalyzeCandlestick(candle)
		bodyToUpperStr := fmt.Sprintf("%9.2f", metrics["BodyToUpper"])
		bodyToLowerStr := fmt.Sprintf("%9.2f", metrics["BodyToLower"])
		analysisStr := results["Analysis"]

		if i == 0 {
			latestAnalysis = analysisStr
			if hasRSI {
				val := rsiVal
				latestRSI = &val
			}
			if hasATR {
				val := atrVal
				latestATR = &val
			}
		}

		signalStr := ""

		if hasRSI {
			rsiSignal := indicators.DetermineRSISignal(rsiVal)
			switch rsiSignal {
			case "overbought":
				signalStr += "[OVERBOUGHT] RSI High"
			case "oversold":
				signalStr += "[OVERSOLD] RSI Low"
			case "neutral":
				signalStr += "[NEUTRAL] RSI Mid"
			}
		} else {

			analysis := results["Analysis"]
			if analysis == "Strong Bullish" || analysis == "Bullish" {
				signalStr += "[BULLISH] Signal"
			} else if analysis == "Strong Bearish" || analysis == "Bearish" {
				signalStr += "[BEARISH] Signal"
			} else if analysis == "Doji (indecision)" {
				signalStr += "[WAIT] Indecision"
			} else if analysis == "Bullish Rejection" {
				signalStr += "[REVERSAL+] Bullish Setup"
			} else if analysis == "Bearish Rejection" {
				signalStr += "[REVERSAL-] Bearish Setup"
			}
		}

		if hasATR {

			atrThreshold := bar.Close * 0.01
			atrSignal := indicators.DetermineATRSignal(atrVal, atrThreshold)

			if signalStr != "" {
				signalStr += " | "
			}

			switch atrSignal {
			case "high volatility":
				signalStr += "⚡ High Vol"
			case "low volatility":
				signalStr += "❄️  Low Vol"
			}
		} else {
			priceRange := bar.High - bar.Low
			rangePct := (priceRange / bar.Close) * 100

			if signalStr != "" {
				signalStr += " | "
			}

			if rangePct > 0.5 {
				signalStr += "⚡ High Volatility"
			} else if rangePct < 0.1 {
				signalStr += "❄️  Low Volatility"
			} else {
				signalStr += "↔️ Medium Volatility"
			}
		}

		if signalStr == "" {
			signalStr = "-"
		}

		fmt.Printf("%-20s | %11.2f | %9.2f | %6.2f | %8d | %6s | %6s | %9s | %9s | %-25s | %-20s\n",
			displayTimestamp, bar.Close, priceChange, priceChangePercent, bar.Volume, rsiStr, atrStr, bodyToUpperStr, bodyToLowerStr, analysisStr, signalStr)
	}

	displayFinalSignal(bars, symbol, latestAnalysis, latestRSI, latestATR, "stock")

	if queries != nil {
		fmt.Println()
		displayWhaleEventsInline(symbol, queries)
	}
	displaySupportResistance(bars)

	displayPatternSignals(bars, symbol)
}

func displayPatternSignals(bars []datafeed.Bar, symbol string) {
	if len(bars) < 5 {
		return
	}

	patternDetector := detection.NewPatternDetector()
	patterns := patternDetector.DetectAllPatterns(bars)

	if len(patterns) == 0 {
		return
	}

	// Filter and select the best pattern
	var bestPattern *detection.PatternSignal
	var bullishPatterns []detection.PatternSignal
	var bearishPatterns []detection.PatternSignal
	var neutralPatterns []detection.PatternSignal

	for _, pattern := range patterns {
		if pattern.Detected {
			if pattern.Direction == "LONG" {
				bullishPatterns = append(bullishPatterns, pattern)
			} else if pattern.Direction == "SHORT" {
				bearishPatterns = append(bearishPatterns, pattern)
			} else {
				neutralPatterns = append(neutralPatterns, pattern)
			}
		}
	}

	// Select the highest confidence pattern from the dominant direction
	if len(bullishPatterns) > 0 && len(bearishPatterns) > 0 {
		// Conflicting signals - pick the highest confidence overall
		var highestConfidence float64
		for i := range bullishPatterns {
			if bullishPatterns[i].Confidence > highestConfidence {
				highestConfidence = bullishPatterns[i].Confidence
				bestPattern = &bullishPatterns[i]
			}
		}
		for i := range bearishPatterns {
			if bearishPatterns[i].Confidence > highestConfidence {
				highestConfidence = bearishPatterns[i].Confidence
				bestPattern = &bearishPatterns[i]
			}
		}
	} else if len(bullishPatterns) > 0 {
		bestPattern = &bullishPatterns[0]
		for i := range bullishPatterns {
			if bullishPatterns[i].Confidence > bestPattern.Confidence {
				bestPattern = &bullishPatterns[i]
			}
		}
	} else if len(bearishPatterns) > 0 {
		bestPattern = &bearishPatterns[0]
		for i := range bearishPatterns {
			if bearishPatterns[i].Confidence > bestPattern.Confidence {
				bestPattern = &bearishPatterns[i]
			}
		}
	} else if len(neutralPatterns) > 0 {
		bestPattern = &neutralPatterns[0]
		for i := range neutralPatterns {
			if neutralPatterns[i].Confidence > bestPattern.Confidence {
				bestPattern = &neutralPatterns[i]
			}
		}
	}

	if bestPattern == nil {
		return
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════════════════════════════")
	fmt.Println("CHART PATTERN ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")

	directionIcon := "Neutral"
	if bestPattern.Direction == "LONG" {
		directionIcon = "UP"
	} else if bestPattern.Direction == "SHORT" {
		directionIcon = "DOWN"
	}

	fmt.Printf("%s PRIMARY PATTERN: %s\n", directionIcon, bestPattern.Pattern)
	fmt.Printf("   Direction: %s | Confidence: %.1f%%\n", bestPattern.Direction, bestPattern.Confidence)
	fmt.Printf("   Support: $%.2f | Resistance: $%.2f\n", bestPattern.SupportLevel, bestPattern.ResistanceLevel)

	if bestPattern.Direction == "LONG" && bestPattern.PriceTargetUp > 0 {
		fmt.Printf("   Target: $%.2f | Stop Loss: $%.2f | R:R %.2f:1\n",
			bestPattern.PriceTargetUp, bestPattern.StopLossLevel, bestPattern.RiskRewardRatio)
	} else if bestPattern.Direction == "SHORT" && bestPattern.PriceTargetDown > 0 {
		fmt.Printf("   Target: $%.2f | Stop Loss: $%.2f | R:R %.2f:1\n",
			bestPattern.PriceTargetDown, bestPattern.StopLossLevel, bestPattern.RiskRewardRatio)
	}

	fmt.Printf("   Reasoning: %s\n", bestPattern.Reasoning)

	// Show count of other detected patterns
	otherCount := len(bullishPatterns) + len(bearishPatterns) + len(neutralPatterns) - 1
	if otherCount > 0 {
		fmt.Printf("\n   Note: %d other pattern(s) detected but showing highest confidence\n", otherCount)
		if len(bullishPatterns) > 0 && len(bearishPatterns) > 0 {
			fmt.Printf("   Mixed signals: %d bullish, %d bearish patterns found\n", len(bullishPatterns), len(bearishPatterns))
		}
	}

	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
}

func displayFinalSignal(bars []datafeed.Bar, symbol string, analysis string, rsi, atr *float64, assetType string) {
	if len(bars) == 0 {
		return
	}

	// Calculate RSI values array for divergence detection
	closes := make([]float64, len(bars))
	for i, bar := range bars {
		closes[i] = bar.Close
	}
	rsiValues, err := indicators.CalculateRSI(closes, 14)
	if err != nil {
		rsiValues = []float64{} // Use empty array if calculation fails
	}

	signal := signals.CalculateSignal(rsi, atr, bars, symbol, analysis, rsiValues)
	filter := signals.NewSignalQualityFilter()
	filter.MinConfidenceThreshold = 70.0
	filter.VerboseLogging = true

	tradeSignal := signals.ConvertToTradeSignal(signal)
	filteredResult := filter.FilterSignal(tradeSignal)

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")

	recommendationStr := signals.FormatSignal(signal)

	if filteredResult.Passed {
		fmt.Printf("[FINAL] RECOMMENDATION: %s \n", recommendationStr)
		fmt.Printf("[PASS] Signal Quality: %.1f%% - %s\n", filteredResult.QualityScore, filteredResult.RecommendedAction)
	} else {
		fmt.Printf("[WARNING] FILTERED SIGNAL: %s\n", recommendationStr)
		fmt.Printf("-X- Quality Check Failed: %s\n", filteredResult.FailureReason)
		fmt.Printf("   Recommendation: %s\n", filteredResult.RecommendedAction)
	}

	fmt.Printf("Reason: %s\n", signal.Reasoning)
	// S/R Validation
	if tradeSignal != nil && len(bars) > 0 {
		srValidator := signals.NewSupportResistanceValidator()
		srValidator.MinValidationScore = 50.0
		srValidation := srValidator.ValidateSignalWithSR(tradeSignal, bars, bars[0].Close)
		if srValidation != nil {
			fmt.Printf("\\n[S/R] Validation: Score %.0f/100", srValidation.ValidationScore)
			if srValidation.IsValidLocation {
				fmt.Print(" [VALID]\\n")
			} else {
				fmt.Print(" [WARNING]\\n")
			}
			fmt.Printf("   Support: $%.2f | Resistance: $%.2f | Current: $%.2f\\n",
				srValidation.SupportLevel, srValidation.ResistanceLevel, srValidation.CurrentPrice)
			fmt.Printf("   %s\\n", srValidation.DetailedAnalysis)
			fmt.Printf("   %s\\n", srValidation.RecommendedAction)
		}
	}
	fmt.Println("\nSignal Breakdown:")
	for _, component := range signal.Components {
		marker := "[+]"
		if component.Score < 0 {
			marker = "[-]"
		}
		fmt.Printf("  %s %-20s %+.1f (weight: %.0f%%)\n",
			marker,
			component.Name,
			component.Score,
			component.Weight*100)
	}

	if signal.DivergenceDetails != "" {
		fmt.Println("\nDIVERGENCE DETECTED:")
		fmt.Printf("   %s\n", signal.DivergenceDetails)
	}

	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
	fmt.Println(" MULTI-TIMEFRAME ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")

	multiSignal, err := FetchMultiTimeframeSignals(symbol, assetType)
	if err != nil {
		fmt.Printf("[WARNING] Could not fetch multi-timeframe data: %v\n", err)
	} else {
		// Display multi-timeframe analysis
		fmt.Print(signals.FormatMultiTimeframeSignal(*multiSignal))

		if multiSignal.IsMultiTimeframeConfirmed(true) {
			fmt.Println("[STRONG] CONFIRMATION: Multiple timeframes aligned!")
			fmt.Println("   This signal has high probability of success.")
		} else if multiSignal.IsMultiTimeframeConfirmed(false) {
			fmt.Println("[MODERATE] CONFIRMATION: Partial timeframe alignment.")
			fmt.Println("   Consider waiting for stronger confirmation.")
		} else {
			fmt.Println("[NONE] NO CONFIRMATION: Timeframes are conflicting.")
			fmt.Println("   High risk - consider skipping this trade.")
		}
	}

	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
}

func displayWhaleEventsInline(symbol string, queries *sqlc.Queries) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	whales, err := datafeed.GetRecentWhales(ctx, queries, symbol, 10)
	if err != nil {
		fmt.Printf("[WARNING] Could not fetch whale events: %v\n", err)
		return
	}

	if len(whales) == 0 {
		fmt.Println("[WHALE] Activity: No significant volume anomalies detected")
		return
	}

	fmt.Println("[WHALE] ACTIVITY DETECTED:")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
	fmt.Println("Timestamp            | Direction | Z-Score | Volume (M)  | Price    | Conviction")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")

	for _, whale := range whales {
		dirMarker := "[BUY]"
		if whale.Direction == "SELL" {
			dirMarker = "[SELL]"
		}

		tsStr := "---"
		if !whale.Timestamp.IsZero() {
			tsStr = whale.Timestamp.Format("2006-01-02 15:04:05")
		}

		volM := float64(whale.Volume) / 1_000_000

		convictionStr := whale.Conviction
		if whale.Conviction == "HIGH" {
			convictionStr = "[!] HIGH"
		} else if whale.Conviction == "MEDIUM" {
			convictionStr = "[~] MEDIUM"
		}

		fmt.Printf("%s | %s %-7s | %7s | %10.1f | %8s | %s\n",
			tsStr,
			dirMarker,
			whale.Direction,
			whale.ZScore,
			volM,
			whale.ClosePrice,
			convictionStr,
		)
	}
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
}

func displaySupportResistance(bars []datafeed.Bar) {
	if len(bars) == 0 {
		return
	}

	support := indicators.FindSupport(bars)
	resistance := indicators.FindResistance(bars)
	pivot := indicators.FindPivotPoint(bars)
	currentPrice := bars[0].Close

	distanceToSupport := indicators.DistanceToSupport(currentPrice, support)
	distanceToResistance := indicators.DistanceToResistance(currentPrice, resistance)

	isAtSupportLevel := indicators.IsAtSupport(currentPrice, support)
	isAtResistanceLevel := indicators.IsAtResistance(currentPrice, resistance)
	isBreakoutUp := indicators.IsBreakoutAboveResistance(currentPrice, resistance)
	isBreakoutDown := indicators.IsBreakoutBelowSupport(currentPrice, support)

	fmt.Println()
	fmt.Println("📊 SUPPORT & RESISTANCE LEVELS:")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("Current Price:  $%.2f\n", currentPrice)
	fmt.Printf("Support Level:  $%.2f (%.2f%% below)  ", support, distanceToSupport)
	if isAtSupportLevel {
		fmt.Printf("[BUY] AT SUPPORT - BUYING OPPORTUNITY")
	} else if isBreakoutDown {
		fmt.Printf("[SELL] BROKEN SUPPORT - POSSIBLE SELL")
	}
	fmt.Println()

	fmt.Printf("Resistance:     $%.2f (%.2f%% above)  ", resistance, distanceToResistance)
	if isAtResistanceLevel {
		fmt.Printf("[SELL] AT RESISTANCE - SELLING PRESSURE")
	} else if isBreakoutUp {
		fmt.Printf("[BUY] ABOVE RESISTANCE - BREAKOUT!")
	}
	fmt.Println()

	fmt.Printf("Pivot Point:    $%.2f\n", pivot)
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
}

func ShowTimeframeMenu() (string, error) {
	fmt.Println("Choose timeframe:")
	fmt.Println("1.  1 Minute")
	fmt.Println("2.  3 Minutes")
	fmt.Println("3.  5 Minutes")
	fmt.Println("4.  10 Minutes")
	fmt.Println("5.  30 Minutes")
	fmt.Println("6.  1 Hour")
	fmt.Println("7.  2 Hours")
	fmt.Println("8.  4 Hours")
	fmt.Println("9.  1 Day")
	fmt.Println("10. 1 Week")
	fmt.Println("11. 1 Month")
	fmt.Print("Enter choice: ")
	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Invalid input. Please enter a number between 1 and 11.")
		return "", err
	}

	switch choice {
	case 1:
		return "1Min", nil
	case 2:
		return "3Min", nil
	case 3:
		return "5Min", nil
	case 4:
		return "10Min", nil
	case 5:
		return "30Min", nil
	case 6:
		return "1Hour", nil
	case 7:
		return "2Hour", nil
	case 8:
		return "4Hour", nil
	case 9:
		return "1Day", nil
	case 10:
		return "1Week", nil
	case 11:
		return "1Month", nil
	default:
		fmt.Println("Invalid choice.")
		return "", fmt.Errorf("invalid choice")
	}
}

func ShowAssetTypeMenu() (string, error) {
	fmt.Println("\nChoose asset type:")
	fmt.Println("1. Stock")
	fmt.Println("2. Crypto")
	fmt.Print("Enter choice: ")
	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Invalid input. Please enter 1 or 2.")
		return "", err
	}

	switch choice {
	case 1:
		return "stock", nil
	case 2:
		return "crypto", nil
	default:
		fmt.Println("Invalid choice.")
		return "", fmt.Errorf("invalid choice")
	}
}

func ShowDisplayMenu() (string, error) {
	fmt.Println("\nChoose display format:")
	fmt.Println("1. Basic Data")
	fmt.Println("2. Full OHLC")
	fmt.Println("3. Analytics")
	fmt.Println("4. All Data")
	fmt.Println("5. Export Data")
	fmt.Println("6. vWAP Analysis")

	fmt.Print("Enter choice: ")
	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Invalid input. Please enter a number between 1 and 7.")
		return "", err
	}

	switch choice {
	case 1:
		return "basic", nil
	case 2:
		return "full", nil
	case 3:
		return "analytics", nil
	case 4:
		return "all", nil
	case 5:
		return "export", nil
	case 6:
		return "vwap", nil
	default:
		fmt.Println("Invalid choice.")
	}
	return "", fmt.Errorf("invalid choice")
}

func ShowTimezoneMenu() (*time.Location, error) {
	nyLocation, _ := time.LoadLocation("America/New_York")
	chicagoLocation, _ := time.LoadLocation("America/Chicago")
	laLocation, _ := time.LoadLocation("America/Los_Angeles")
	londonLocation, _ := time.LoadLocation("Europe/London")
	tokyoLocation, _ := time.LoadLocation("Asia/Tokyo")
	hkLocation, _ := time.LoadLocation("Asia/Hong_Kong")

	timezones := map[int]struct {
		name     string
		location *time.Location
	}{
		1: {"UTC", time.UTC},
		2: {"America/New_York (NYSE/NASDAQ)", nyLocation},
		3: {"America/Chicago (CME)", chicagoLocation},
		4: {"America/Los_Angeles (PST)", laLocation},
		5: {"Europe/London (LSE)", londonLocation},
		6: {"Asia/Tokyo (TSE)", tokyoLocation},
		7: {"Asia/Hong_Kong (HKEX)", hkLocation},
		8: {"Local (System Time)", time.Local},
	}

	fmt.Println("\nChoose timezone:")
	for i := 1; i <= len(timezones); i++ {
		fmt.Printf("%d. %s\n", i, timezones[i].name)
	}

	fmt.Print("Enter choice: ")
	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Invalid input. Please enter a valid number.")
		return nil, err
	}

	if tz, exists := timezones[choice]; exists {
		return tz.location, nil
	} else {
		fmt.Println("Invalid choice. Defaulting to UTC.")
		return time.UTC, nil
	}
}

func PrepareExportData(bars []datafeed.Bar, symbol string, timezone *time.Location) []export.ExportRecord {
	var records []export.ExportRecord

	var rsiMap map[string]float64
	var atrMap map[string]float64

	var startTime, endTime time.Time
	if len(bars) > 0 {
		if t, err := time.Parse(time.RFC3339, bars[0].Timestamp); err == nil {
			startTime = t
		}
		if t, err := time.Parse(time.RFC3339, bars[len(bars)-1].Timestamp); err == nil {
			endTime = t
		}
	}

	if !startTime.IsZero() && !endTime.IsZero() {
		rsiMap, _ = datafeed.FetchRSIByTimestampRange(symbol, startTime, endTime)
		atrMap, _ = datafeed.FetchATRByTimestampRange(symbol, startTime, endTime)
	} else {
		fetchLimit := len(bars) * 10
		rsiMap, _ = datafeed.FetchRSIForDisplay(symbol, fetchLimit)
		atrMap, _ = datafeed.FetchATRForDisplay(symbol, fetchLimit)
	}

	for _, bar := range bars {
		t, _ := time.Parse(time.RFC3339, bar.Timestamp)
		timestampStr := t.In(timezone).Format("2006-01-02 15:04:05")

		rsiVal, hasRSI := rsiMap[t.Format("2006-01-02 15:04:05")]
		atrVal, hasATR := atrMap[t.Format("2006-01-02 15:04:05")]

		var rsiPtr *float64
		if hasRSI {
			rsiPtr = &rsiVal
		}
		var atrPtr *float64
		if hasATR {
			atrPtr = &atrVal
		}

		candle := analyzer.Candlestick{Open: bar.Open, Close: bar.Close, High: bar.High, Low: bar.Low}
		_, results := analyzer.AnalyzeCandlestick(candle)
		analysis := results["Analysis"]

		var signals []string
		if hasRSI {
			rsiSignal := indicators.DetermineRSISignal(rsiVal)
			switch rsiSignal {
			case "overbought":
				signals = append(signals, "Overbought")
			case "oversold":
				signals = append(signals, "Oversold")
			case "neutral":
				signals = append(signals, "Neutral")
			}
		}
		if hasATR {
			atrThreshold := bar.Close * 0.01
			atrSignal := indicators.DetermineATRSignal(atrVal, atrThreshold)
			switch atrSignal {
			case "high volatility":
				signals = append(signals, "High Vol")
			case "low volatility":
				signals = append(signals, "Low Vol")
			}
		} else {
			priceRange := bar.High - bar.Low
			rangePct := (priceRange / bar.Close) * 100

			if rangePct > 0.5 {
				signals = append(signals, "High Volatility")
			} else if rangePct < 0.1 {
				signals = append(signals, "Low Volatility")
			} else {
				signals = append(signals, "Medium Volatility")
			}
		}

		record := export.ExportRecord{
			Timestamp: timestampStr,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    bar.Volume,
			RSI:       rsiPtr,
			ATR:       atrPtr,
			Analysis:  analysis,
			Signals:   signals,
		}
		records = append(records, record)
	}

	return records
}

func DisplayVWAPAnalysis(bars []datafeed.Bar, symbol string, timeframe string) {
	if len(bars) == 0 {
		fmt.Printf(" No data available for %s\n", symbol)
		return
	}

	if len(bars) < 3 {
		fmt.Printf(" Need at least 3 bars for complete vWAP analysis\n")
		return
	}

	typesBars := make([]types.Bar, len(bars))
	for i := range bars {
		typesBars[i] = types.Bar(bars[i])
	}

	vwapCalc := indicators.NewVWAPCalculator(typesBars)
	analysis := vwapCalc.AnalyzeVWAP(1.0)

	fmt.Printf("\n vWAP (Volume Weighted Average Price) Analysis for %s (%s)\n", symbol, timeframe)
	fmt.Println("==========================================")
	fmt.Println()

	fmt.Println("QUICK SUMMARY:")
	for key, value := range analysis {
		fmt.Printf("  %-18s: %v\n", key, value)
	}

	fmt.Println("\n vWAP DETAILS:")
	allVWAPValues := vwapCalc.CalculateAllValues()
	if len(allVWAPValues) > 0 {
		fmt.Printf("  Min vWAP: %.2f\n", utils.Min(allVWAPValues...))
		fmt.Printf("  Max vWAP: %.2f\n", utils.Max(allVWAPValues...))
		fmt.Printf("  Current vWAP: %.2f\n", vwapCalc.Calculate())
	}

	fmt.Println("\n vWAP BY BAR:")
	fmt.Println("Timestamp           | Close Price | vWAP       | Distance % | Trend")
	fmt.Println("--------------------|-------------|------------|------------|---------")

	for i, bar := range bars {
		vwap := vwapCalc.CalculateAt(i)
		distance := ((bar.Close - vwap) / vwap) * 100
		trend := "---"

		if bar.Close > vwap {
			trend = "Above"
		} else if bar.Close < vwap {
			trend = "Below"
		} else {
			trend = "Neutral"
		}

		fmt.Printf("%-20s | %11.2f | %10.2f | %10.2f | %s\n",
			bar.Timestamp, bar.Close, vwap, distance, trend)
	}

	fmt.Println("\n SUPPORT/RESISTANCE LEVELS:")
	currentVWAP := vwapCalc.Calculate()
	isSupport := vwapCalc.IsVWAPSupport(1.0)
	isResistance := vwapCalc.IsVWAPResistance(1.0)

	if isSupport {
		fmt.Println("   vWAP is acting as SUPPORT")
		fmt.Println("     → Price touched vWAP from above")
		fmt.Println("     → Look for bounce UP")
	} else if isResistance {
		fmt.Println("   vWAP is acting as RESISTANCE")
		fmt.Println("     → Price touched vWAP from below")
		fmt.Println("     → Look for bounce DOWN")
	} else {
		fmt.Println("   vWAP is neither support nor resistance (no recent contact)")
	}

	fmt.Println("\n BOUNCE DETECTION:")
	isBounce, bounceType := vwapCalc.GetVWAPBounce(1.0)

	if isBounce {
		fmt.Printf("  BOUNCE DETECTED: %s\n", bounceType)
		if bounceType == "bullish_bounce" {
			fmt.Println("     → Price bounced UP from vWAP")
			fmt.Println("     → Potential BUY signal")
		} else if bounceType == "bearish_bounce" {
			fmt.Println("     → Price bounced DOWN from vWAP")
			fmt.Println("     → Potential SELL signal")
		}
	} else {
		fmt.Println("  [WARNING] No bounce detected in last 3 bars")
	}

	fmt.Println("\n CURRENT TREND:")
	trend := vwapCalc.GetVWAPTrend()
	switch trend {
	case 1:
		fmt.Println("Price is ABOVE vWAP (Bullish)")
		fmt.Println("Uptrend favors buyers")
		fmt.Printf("Support level: vWAP at %.2f\n", currentVWAP)
	case -1:
		fmt.Println("Price is BELOW vWAP (Bearish)")
		fmt.Println("Downtrend favors sellers")
		fmt.Printf("Resistance level: vWAP at %.2f\n", currentVWAP)
	default:
		fmt.Println(" Price is AT vWAP (Neutral)")
		fmt.Println(" Potential decision point")
	}

	fmt.Println("\n==========================================")
}

func displayNewsForSymbol(symbol string, newsStorage *newsscraping.NewsStorage) {
	ctx := context.Background()
	articles, err := newsStorage.GetLatestNews(ctx, symbol, 5)
	if err != nil || len(articles) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
	fmt.Printf(" LATEST NEWS FOR %s\n", symbol)
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")

	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
}
