package interactive

import (
	"context"
	"fmt"
	"time"

	datafeed "github.com/fazecat/mongelmaker/Internal/database"
	sqlc "github.com/fazecat/mongelmaker/Internal/database/sqlc"
	"github.com/fazecat/mongelmaker/Internal/export"
	"github.com/fazecat/mongelmaker/Internal/strategy"
	"github.com/fazecat/mongelmaker/Internal/types"
	"github.com/fazecat/mongelmaker/Internal/utils"
	"github.com/fazecat/mongelmaker/Internal/utils/analyzer"
	"github.com/fazecat/mongelmaker/Internal/utils/scoring"
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

func ShowMainMenu() (string, error) {
	fmt.Println("Welcome to MongelMaker Interactive!")
	fmt.Println("1. Analyze Single Stock")
	fmt.Println("2. Screen Multiple Stocks")
	fmt.Print("Enter choice: ")
	var choice int
	_, err := fmt.Scan(&choice)

	if err != nil {
		fmt.Println("Invalid input. Please enter a valid number.")
		return "", err
	}

	switch choice {
	case 1:
		return "single", nil
	case 2:
		return "screener", nil
	default:
		fmt.Println("Invalid choice.")
		return "", fmt.Errorf("invalid choice")
	}
}

func PickStockFromResults(results []strategy.StockScore) (string, error) {
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
	fmt.Printf("\n📊 Basic Data for %s (%s)\n", symbol, timeframe)
	fmt.Println("Timestamp           | Close Price | Volume")
	fmt.Println("--------------------|-------------|----------")

	for _, bar := range bars {
		fmt.Printf("%-20s | %11.2f | %8d\n", bar.Timestamp, bar.Close, bar.Volume)

	}
}

func DisplayAdvancedData(bars []datafeed.Bar, symbol string, timeframe string) {
	fmt.Printf("\n📊 Advanced Data for %s (%s)\n", symbol, timeframe)
	fmt.Println("Timestamp           | Open Price | High Price | Low Price | Close Price | Volume")
	fmt.Println("--------------------|------------|------------|-----------|-------------|----------")

	for _, bar := range bars {
		fmt.Printf("%-20s | %11.2f | %11.2f | %9.2f | %11.2f | %8d\n",
			bar.Timestamp, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
	}
}

func DisplayAnalyticsData(bars []datafeed.Bar, symbol string, timeframe string, tz *time.Location, queries *sqlc.Queries) {
	fmt.Printf("\n📈 Analytics Data for %s (%s) - Timezone: %s\n", symbol, timeframe, tz.String())

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

		rsiValues, err := strategy.CalculateRSI(closes, 14)
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
		// Store same ATR for all recent bars (ATR is calculated for the full period)
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
			rsiSignal := strategy.DetermineRSISignal(rsiVal)
			switch rsiSignal {
			case "overbought":
				signalStr += "📈 Overbought"
			case "oversold":
				signalStr += "📉 Oversold"
			case "neutral":
				signalStr += "➡️  Neutral"
			}
		} else {

			analysis := results["Analysis"]
			if analysis == "Strong Bullish" || analysis == "Bullish" {
				signalStr += "📈 Bullish Signal"
			} else if analysis == "Strong Bearish" || analysis == "Bearish" {
				signalStr += "📉 Bearish Signal"
			} else if analysis == "Doji (indecision)" {
				signalStr += "➡️  Wait Signal"
			} else if analysis == "Bullish Rejection" {
				signalStr += "📈 Reversal Setup"
			} else if analysis == "Bearish Rejection" {
				signalStr += "📉 Reversal Setup"
			}
		}

		if hasATR {

			atrThreshold := bar.Close * 0.01
			atrSignal := strategy.DetermineATRSignal(atrVal, atrThreshold)

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

	// Display final signal recommendation (before whale events)
	displayFinalSignal(bars, symbol, latestAnalysis, latestRSI, latestATR)

	// Display whale events if database available
	if queries != nil {
		fmt.Println()
		displayWhaleEventsInline(symbol, queries)
	}

	// Display support/resistance levels
	displaySupportResistance(bars)
}

func displayFinalSignal(bars []datafeed.Bar, symbol string, analysis string, rsi, atr *float64) {
	if len(bars) == 0 {
		return
	}

	signal := strategy.CalculateSignal(rsi, atr, bars, symbol, analysis)

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")

	recommendationStr := strategy.FormatSignal(signal)
	fmt.Printf("🎯 FINAL RECOMMENDATION: %s\n", recommendationStr)

	fmt.Printf("Reason: %s\n", signal.Reasoning)
	fmt.Println("\nSignal Breakdown:")
	for _, component := range signal.Components {
		emoji := "🟢"
		if component.Score < 0 {
			emoji = "🔴"
		}
		fmt.Printf("  %s %-20s %+.1f (weight: %.0f%%)\n",
			emoji,
			component.Name,
			component.Score,
			component.Weight*100)
	}
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
}

func displayWhaleEventsInline(symbol string, queries *sqlc.Queries) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch recent whale events
	whales, err := datafeed.GetRecentWhales(ctx, queries, symbol, 10)
	if err != nil {
		fmt.Printf("⚠️  Could not fetch whale events: %v\n", err)
		return
	}

	if len(whales) == 0 {
		fmt.Println("🐋 Whale Activity: No significant volume anomalies detected")
		return
	}

	fmt.Println("🐋 WHALE ACTIVITY DETECTED:")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
	fmt.Println("Timestamp            | Direction | Z-Score | Volume (M)  | Price    | Conviction")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")

	for _, whale := range whales {
		emoji := "🟢"
		if whale.Direction == "SELL" {
			emoji = "🔴"
		}

		tsStr := "---"
		if !whale.Timestamp.IsZero() {
			tsStr = whale.Timestamp.Format("2006-01-02 15:04:05")
		}

		volM := float64(whale.Volume) / 1_000_000

		convictionStr := whale.Conviction
		if whale.Conviction == "HIGH" {
			convictionStr = "🚨 HIGH"
		} else if whale.Conviction == "MEDIUM" {
			convictionStr = "⚠️  MEDIUM"
		}

		fmt.Printf("%s | %s %-7s | %7s | %10.1f | %8s | %s\n",
			tsStr,
			emoji,
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

	support := strategy.FindSupport(bars)
	resistance := strategy.FindResistance(bars)
	pivot := strategy.FindPivotPoint(bars)
	currentPrice := bars[0].Close

	distanceToSupport := strategy.DistanceToSupport(currentPrice, support)
	distanceToResistance := strategy.DistanceToResistance(currentPrice, resistance)

	isAtSupportLevel := strategy.IsAtSupport(currentPrice, support)
	isAtResistanceLevel := strategy.IsAtResistance(currentPrice, resistance)
	isBreakoutUp := strategy.IsBreakoutAboveResistance(currentPrice, resistance)
	isBreakoutDown := strategy.IsBreakoutBelowSupport(currentPrice, support)

	fmt.Println()
	fmt.Println("📊 SUPPORT & RESISTANCE LEVELS:")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("Current Price:  $%.2f\n", currentPrice)
	fmt.Printf("Support Level:  $%.2f (%.2f%% below)  ", support, distanceToSupport)
	if isAtSupportLevel {
		fmt.Printf("🟢 AT SUPPORT - BUYING OPPORTUNITY")
	} else if isBreakoutDown {
		fmt.Printf("🔴 BROKEN SUPPORT - POSSIBLE SELL")
	}
	fmt.Println()

	fmt.Printf("Resistance:     $%.2f (%.2f%% above)  ", resistance, distanceToResistance)
	if isAtResistanceLevel {
		fmt.Printf("🔴 AT RESISTANCE - SELLING PRESSURE")
	} else if isBreakoutUp {
		fmt.Printf("🟢 ABOVE RESISTANCE - BREAKOUT!")
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
			rsiSignal := strategy.DetermineRSISignal(rsiVal)
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
			atrSignal := strategy.DetermineATRSignal(atrVal, atrThreshold)
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
		fmt.Printf("⚠️  No data available for %s\n", symbol)
		return
	}

	if len(bars) < 3 {
		fmt.Printf("⚠️  Need at least 3 bars for complete vWAP analysis\n")
		return
	}

	typesBars := make([]types.Bar, len(bars))
	for i := range bars {
		typesBars[i] = types.Bar(bars[i])
	}

	vwapCalc := strategy.NewVWAPCalculator(typesBars)
	analysis := vwapCalc.AnalyzeVWAP(1.0)

	fmt.Printf("\n💰 vWAP (Volume Weighted Average Price) Analysis for %s (%s)\n", symbol, timeframe)
	fmt.Println("==========================================")
	fmt.Println()

	fmt.Println("📊 QUICK SUMMARY:")
	for key, value := range analysis {
		fmt.Printf("  %-18s: %v\n", key, value)
	}

	fmt.Println("\n📈 vWAP DETAILS:")
	allVWAPValues := vwapCalc.CalculateAllValues()
	if len(allVWAPValues) > 0 {
		fmt.Printf("  Min vWAP: %.2f\n", utils.Min(allVWAPValues...))
		fmt.Printf("  Max vWAP: %.2f\n", utils.Max(allVWAPValues...))
		fmt.Printf("  Current vWAP: %.2f\n", vwapCalc.Calculate())
	}

	fmt.Println("\n📊 vWAP BY BAR:")
	fmt.Println("Timestamp           | Close Price | vWAP       | Distance % | Trend")
	fmt.Println("--------------------|-------------|------------|------------|---------")

	for i, bar := range bars {
		vwap := vwapCalc.CalculateAt(i)
		distance := ((bar.Close - vwap) / vwap) * 100
		trend := "---"

		if bar.Close > vwap {
			trend = "📈 Above"
		} else if bar.Close < vwap {
			trend = "📉 Below"
		} else {
			trend = "➡️  At"
		}

		fmt.Printf("%-20s | %11.2f | %10.2f | %10.2f | %s\n",
			bar.Timestamp, bar.Close, vwap, distance, trend)
	}

	fmt.Println("\n🎯 SUPPORT/RESISTANCE LEVELS:")
	currentVWAP := vwapCalc.Calculate()
	isSupport := vwapCalc.IsVWAPSupport(1.0)
	isResistance := vwapCalc.IsVWAPResistance(1.0)

	if isSupport {
		fmt.Println("  ✅ vWAP is acting as SUPPORT")
		fmt.Println("     → Price touched vWAP from above")
		fmt.Println("     → Look for bounce UP")
	} else if isResistance {
		fmt.Println("  ✅ vWAP is acting as RESISTANCE")
		fmt.Println("     → Price touched vWAP from below")
		fmt.Println("     → Look for bounce DOWN")
	} else {
		fmt.Println("  ⚠️  vWAP is neither support nor resistance (no recent contact)")
	}

	fmt.Println("\n🔄 BOUNCE DETECTION:")
	isBounce, bounceType := vwapCalc.GetVWAPBounce(1.0)

	if isBounce {
		fmt.Printf("  ✅ BOUNCE DETECTED: %s\n", bounceType)
		if bounceType == "bullish_bounce" {
			fmt.Println("     → Price bounced UP from vWAP")
			fmt.Println("     → Potential BUY signal 🟢")
		} else if bounceType == "bearish_bounce" {
			fmt.Println("     → Price bounced DOWN from vWAP")
			fmt.Println("     → Potential SELL signal 🔴")
		}
	} else {
		fmt.Println("  ⚠️  No bounce detected in last 3 bars")
	}

	fmt.Println("\n📉 CURRENT TREND:")
	trend := vwapCalc.GetVWAPTrend()
	switch trend {
	case 1:
		fmt.Println("  📈 Price is ABOVE vWAP (Bullish)")
		fmt.Println("     → Uptrend favors buyers")
		fmt.Printf("     → Support level: vWAP at %.2f\n", currentVWAP)
	case -1:
		fmt.Println("  📉 Price is BELOW vWAP (Bearish)")
		fmt.Println("     → Downtrend favors sellers")
		fmt.Printf("     → Resistance level: vWAP at %.2f\n", currentVWAP)
	default:
		fmt.Println("  ➡️  Price is AT vWAP (Neutral)")
		fmt.Println("     → Potential decision point")
	}

	fmt.Println("\n==========================================")
}
