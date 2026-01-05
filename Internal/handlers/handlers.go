package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	datafeed "github.com/fazecat/mongelmaker/Internal/database"
	database "github.com/fazecat/mongelmaker/Internal/database/sqlc"
	"github.com/fazecat/mongelmaker/Internal/database/watchlist"
	newsscraping "github.com/fazecat/mongelmaker/Internal/news_scraping"
	"github.com/fazecat/mongelmaker/Internal/strategy"
	"github.com/fazecat/mongelmaker/Internal/utils/config"
	"github.com/fazecat/mongelmaker/Internal/utils/scanner"
	"github.com/fazecat/mongelmaker/Internal/utils/scoring"
	"github.com/fazecat/mongelmaker/interactive"
)

// clears any remaining input from stdin
func ClearInputBuffer() {
	reader := bufio.NewReader(os.Stdin)
	for {
		r, _, err := reader.ReadRune()
		if err != nil || r == '\n' {
			break
		}
	}
}

func HandleScan(ctx context.Context, cfg *config.Config, q *database.Queries) {
	if len(cfg.Profiles) == 0 {
		fmt.Println("❌ No profiles configured")
		return
	}

	fmt.Println("\n📋 Available Profiles:")
	profiles := make([]string, 0)
	for name := range cfg.Profiles {
		profiles = append(profiles, name)
	}

	for i, profileName := range profiles {
		profile := cfg.Profiles[profileName]
		shortSignalsAvail := "❌"
		if cfg.Features.EnableShortSignals {
			shortSignalsAvail = "✅"
		}
		fmt.Printf("%d. %s (scan: %d days, short signals: %s)\n", i+1, profileName, profile.ScanIntervalDays, shortSignalsAvail)
	}

	fmt.Print("Select profile (number): ")
	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(profiles) {
		fmt.Println("❌ Invalid selection")
		return
	}

	selectedProfile := profiles[choice-1]

	fmt.Printf("🔄 Scanning profile: %s\n", selectedProfile)
	scannedCount, err := scanner.PerformScan(ctx, selectedProfile, cfg, q)
	if err != nil {
		fmt.Printf("❌ Scan failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Scan complete! Updated %d symbols\n", scannedCount)
}

func HandleAnalyzeSingle(ctx context.Context, assetType string, q *database.Queries) {
	if assetType == "" {
		assetType = "stock" // default
	}
	ClearInputBuffer()

	var symbolExample string
	if assetType == "crypto" {
		symbolExample = "e.g., BTC/USD"
	} else {
		symbolExample = "e.g., AAPL"
	}

	fmt.Printf("Enter symbol (%s): ", symbolExample)
	var symbol string
	_, err := fmt.Scanln(&symbol)
	if err != nil || symbol == "" {
		fmt.Println("❌ Invalid symbol")
		return
	}

	timeframe, err := interactive.ShowTimeframeMenu()
	if err != nil {
		fmt.Println("❌ Invalid timeframe")
		return
	}

	fmt.Print("Enter number of bars (default 100): ")
	var numBars int
	_, err = fmt.Scanln(&numBars)
	if err != nil || numBars < 14 {
		numBars = 100
	}

	bars, err := interactive.FetchMarketDataWithType(symbol, timeframe, numBars, "", assetType)
	if err != nil {
		fmt.Printf("❌ Failed to fetch data: %v\n", err)
		return
	}

	err = datafeed.CalculateAndStoreRSI(symbol, bars)
	if err != nil {
		fmt.Printf("❌ Failed to calculate and store RSI: %v\n", err)
		return
	}

	err = datafeed.CalculateAndStoreATR(symbol, bars)
	if err != nil {
		fmt.Printf("❌ Failed to calculate and store ATR: %v\n", err)
		return
	}

	displayChoice, _ := interactive.ShowDisplayMenu()
	ClearInputBuffer()

	switch displayChoice {
	case "basic":
		interactive.DisplayBasicData(bars, symbol, timeframe)
	case "full":
		interactive.DisplayAdvancedData(bars, symbol, timeframe)
	case "analytics":
		tz, _ := interactive.ShowTimezoneMenu()
		ClearInputBuffer()
		interactive.DisplayAnalyticsData(bars, symbol, timeframe, tz, q)
		fmt.Println("\n--- Press Enter to continue ---")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	case "vwap":
		interactive.DisplayVWAPAnalysis(bars, symbol, timeframe)
	default:
		interactive.DisplayBasicData(bars, symbol, timeframe)
	}
}

func HandleScreener(ctx context.Context, cfg *config.Config, q *database.Queries) {
	assetType, err := interactive.ShowAssetTypeMenu()
	if err != nil {
		fmt.Println("❌ Invalid asset type")
		return
	}

	var symbols []string
	if assetType == "crypto" {
		fmt.Println("\n📝 Enter crypto symbols (comma-separated, e.g., BTC/USD):")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("❌ No symbols entered")
			return
		}
		for _, sym := range strings.Split(input, ",") {
			symbols = append(symbols, strings.TrimSpace(sym))
		}
	} else {
		symbols = strategy.GetPopularStocks()
	}

	if len(symbols) == 0 {
		fmt.Println("❌ Could not get symbols")
		return
	}

	criteria := strategy.DefaultScreenerCriteria()

	fmt.Printf("🔍 Screening %s (%d symbols)...\n", assetType, len(symbols))
	results, err := strategy.ScreenStocksWithType(symbols, "1Day", 100, criteria, nil, assetType)
	if err != nil {
		fmt.Printf("❌ Screener failed: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("📭 No symbols matched criteria")
		return
	}

	fmt.Printf("\n📊 Screening Results (%d total):\n", len(results))
	fmt.Println("==========================================")
	fmt.Println("# | Symbol | Score  | RSI    | ATR    | Signals                    | Analysis")
	fmt.Println("--|--------|--------|--------|--------|----------------------------|----------------------")

	for i, stock := range results {
		rsiStr := "  -   "
		if stock.RSI != nil {
			rsiStr = fmt.Sprintf("%6.2f", *stock.RSI)
		}

		atrStr := "  -   "
		if stock.ATR != nil {
			atrStr = fmt.Sprintf("%6.2f", *stock.ATR)
		}

		signalsStr := ""
		if len(stock.Signals) > 0 {
			for j, sig := range stock.Signals {
				if j > 0 {
					signalsStr += ", "
				}
				signalsStr += sig
			}
		} else {
			signalsStr = "-"
		}

		if len(signalsStr) > 26 {
			signalsStr = signalsStr[:23] + "..."
		}

		analysis := "---"
		if stock.RSI != nil {
			if *stock.RSI > 70 {
				analysis = "🔴 Overbought"
			} else if *stock.RSI < 30 {
				analysis = "🟢 Oversold"
			} else if *stock.RSI > 50 {
				analysis = "📈 Bullish"
			} else {
				analysis = "📉 Bearish"
			}
		}

		fmt.Printf("%2d| %s | %.2f | %s | %s | %-26s | %s\n",
			i+1, stock.Symbol, stock.Score, rsiStr, atrStr, signalsStr, analysis)
	}

	fmt.Print("\nSelect stock for details (or press Enter to skip): ")
	var choice int
	_, err = fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(results) {
		return
	}

	selectedStock := results[choice-1]

	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("📊 Detailed Analysis: %s\n", selectedStock.Symbol)
	fmt.Printf(strings.Repeat("=", 80) + "\n\n")

	fmt.Printf("🎯 Score: %.2f\n", selectedStock.Score)

	if selectedStock.RSI != nil {
		fmt.Printf("📈 RSI (14): %.2f", *selectedStock.RSI)
		if *selectedStock.RSI > 70 {
			fmt.Print(" 🔴 Overbought")
		} else if *selectedStock.RSI < 30 {
			fmt.Print(" 🟢 Oversold")
		}
		fmt.Println()
	}

	if selectedStock.LongSignal != nil {
		fmt.Printf("\n📈 LONG Signal: %s (Confidence: %.1f%%)\n", selectedStock.LongSignal.Direction, selectedStock.LongSignal.Confidence)
		fmt.Printf("   Reason: %s\n", selectedStock.LongSignal.Reasoning)
	}

	if selectedStock.ShortSignal != nil {
		fmt.Printf("\n📉 SHORT Signal: %s (Confidence: %.1f%%)\n", selectedStock.ShortSignal.Direction, selectedStock.ShortSignal.Confidence)
		fmt.Printf("   Reason: %s\n", selectedStock.ShortSignal.Reasoning)
	}

	if selectedStock.ATR != nil {
		fmt.Printf("📊 ATR: %.2f", *selectedStock.ATR)
		if *selectedStock.ATR > 1.0 {
			fmt.Print(" ⚠️ High Volatility")
		}
		fmt.Println()
	}

	if len(selectedStock.Signals) > 0 {
		fmt.Println("\n🔔 Signals:")
		for _, sig := range selectedStock.Signals {
			fmt.Printf("   • %s\n", sig)
		}
	}

	if selectedStock.Recommendation != "" {
		fmt.Printf("\n📝 Recommendation: %s\n", selectedStock.Recommendation)
	}

	fmt.Println("\n📰 Fetching recent news...")
	finnhubClient := newsscraping.NewFinnhubClient()
	newsArticles, err := finnhubClient.FetchNews(selectedStock.Symbol, 5)
	if err != nil {
		fmt.Printf("⚠️ Could not fetch news: %v\n", err)
	} else if len(newsArticles) > 0 {
		fmt.Printf("\n📰 Recent News (%d articles):\n", len(newsArticles))
		fmt.Println(strings.Repeat("-", 80))
		for i, article := range newsArticles {
			sentimentIcon := "⚪"
			switch article.Sentiment {
			case newsscraping.Positive:
				sentimentIcon = "🟢"
			case newsscraping.Negative:
				sentimentIcon = "🔴"
			}

			catalystIcon := ""
			if article.CatalystType != newsscraping.NoCatalyst {
				catalystIcon = fmt.Sprintf(" [%s]", article.CatalystType)
			}

			fmt.Printf("\n%d. %s %s%s\n", i+1, sentimentIcon, article.Headline, catalystIcon)
			fmt.Printf("   🔗 %s\n", article.URL)
			fmt.Printf("   📅 %s\n", article.PublishedAt.Format("Jan 02, 2006 15:04"))
		}
		fmt.Println()
	} else {
		fmt.Println("📭 No recent news found")
	}

	fmt.Print("\n➕ Add to watchlist? (y/n): ")
	var addChoice string
	fmt.Scanln(&addChoice)

	if strings.ToLower(addChoice) == "y" {
		reason := "Added from screener"
		if selectedStock.Recommendation != "" {
			reason = fmt.Sprintf("Added from screener - %s", selectedStock.Recommendation)
			if len(reason) > 200 {
				reason = reason[:200]
			}
		}
		_, err = watchlist.AddToWatchlist(ctx, q, selectedStock.Symbol, "stock", selectedStock.Score, reason)
		if err != nil {
			fmt.Printf("❌ Failed to add to watchlist: %v\n", err)
			return
		}
		fmt.Printf("✅ Added %s to watchlist (Score: %.2f)\n", selectedStock.Symbol, selectedStock.Score)
	}

	fmt.Println("\n--- Press Enter to continue ---")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func HandleWatchlist(ctx context.Context, q *database.Queries) {
	watchlist, err := q.GetWatchlist(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to fetch watchlist: %v\n", err)
		return
	}

	if len(watchlist) == 0 {
		fmt.Println("📭 Watchlist is empty")
		return
	}

	fmt.Println("\n📊 Current Watchlist:")
	fmt.Println("Symbol | Score | Added Date | Last Updated | Category")
	fmt.Println("-------|-------|------------|--------------|---------")
	for _, item := range watchlist {
		addedStr := "N/A"
		if item.AddedDate.Valid {
			addedStr = item.AddedDate.Time.Format("2006-01-02")
		}
		updatedStr := "N/A"
		if item.LastUpdated.Valid {
			updatedStr = item.LastUpdated.Time.Format("2006-01-02")
		}
		fmt.Printf("%s | %.2f | %s | %s | %s\n", item.Symbol, item.Score, addedStr, updatedStr, scoring.ScoreCategory(float64(item.Score)))
	}
}

func HandleScout(ctx context.Context, cfg *config.Config, q *database.Queries) {
	if len(cfg.Profiles) == 0 {
		fmt.Println("❌ No profiles configured")
		return
	}

	profiles := make([]string, 0)
	for name := range cfg.Profiles {
		profiles = append(profiles, name)
	}

	fmt.Println("\n📋 Available Profiles:")
	for i, profileName := range profiles {
		profile := cfg.Profiles[profileName]
		fmt.Printf("%d. %s (scan interval: %d days, default threshold: %.1f)\n", i+1, profileName, profile.ScanIntervalDays, profile.Threshold)
	}

	fmt.Print("Select profile (number): ")
	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(profiles) {
		fmt.Println("❌ Invalid selection")
		return
	}

	selectedProfile := profiles[choice-1]

	var minScore float64
	fmt.Print("Enter score threshold (0.0 - 10.0): ")
	_, err = fmt.Scanln(&minScore)
	if err != nil || minScore < 0 || minScore > 10 {
		fmt.Println("❌ Invalid threshold. Must be between 0.0 and 10.0")
		return
	}

	fmt.Printf("\n✅ Using %s profile with threshold: %.1f\n", selectedProfile, minScore)

	// Asset type selection
	assetType := "stock"
	if cfg.Features.CryptoSupport {
		fmt.Println("\n🪙 Asset Type:")
		fmt.Println("1. Stocks")
		fmt.Println("2. Crypto")
		fmt.Print("Select asset type (1-2): ")
		var typeChoice int
		_, err := fmt.Scanln(&typeChoice)
		if err == nil && typeChoice == 2 {
			assetType = "crypto"
		}
	}
	fmt.Printf("🔍 Scanning %s assets\n", assetType)

	var batchSize int
	fmt.Print("Review every N symbols (50 or 100): ")
	fmt.Scanln(&batchSize)
	if batchSize != 50 && batchSize != 100 {
		batchSize = 50 // default
	}

	offset := 0
	batchNum := 1

	for {
		fmt.Printf("\n🔄 Scanning batch %d (evaluating %d symbols)...\n", batchNum, batchSize)
		candidates, totalSymbols, err := scanner.PerformProfileScan(ctx, selectedProfile, minScore, offset, batchSize, cfg)
		if err != nil {
			fmt.Printf("❌ Scout scan failed: %v\n", err)
			return
		}

		if offset >= totalSymbols {
			fmt.Println("✅ Scout scan complete - all symbols evaluated")
			break
		}

		if len(candidates) == 0 {
			fmt.Printf("📭 No candidates found in this batch (evaluated %d-%d of %d symbols)\n", offset+1, offset+batchSize, totalSymbols)
		} else {
			fmt.Printf("\n📊 Batch %d candidates (%d of %d total symbols evaluated):\n", batchNum, offset+batchSize, totalSymbols)

			for _, candidate := range candidates {
				fmt.Printf("\n   %s\n", candidate.Symbol)
				fmt.Printf("      Score: %.2f | Pattern: %s\n", candidate.Score, candidate.Analysis)

				for {
					fmt.Print("      (e)xpand / (y)es / (n)o / (i)gnore: ")
					var choice string
					fmt.Scanln(&choice)
					choice = strings.ToLower(choice)

					if choice == "e" {
						tz, _ := interactive.ShowTimezoneMenu()
						ClearInputBuffer()
						interactive.DisplayAnalyticsData(candidate.Bars, candidate.Symbol, "1Day", tz, q)
						continue
					}

					if choice == "y" {
						fmt.Printf("      Adding %s to watchlist...\n", candidate.Symbol)
						reason := fmt.Sprintf("Scouted - Pattern: %s", candidate.Analysis)
						_, err := watchlist.AddToWatchlist(ctx, q, candidate.Symbol, "stock", candidate.Score, reason)
						if err != nil {
							fmt.Printf("      ❌ Failed to add: %v\n", err)
						} else {
							fmt.Printf("      ✅ Added %s to watchlist (Score: %.2f)\n", candidate.Symbol, candidate.Score)
						}
						break
					}

					if choice == "i" {
						err := q.AddToScoutSkipList(ctx, database.AddToScoutSkipListParams{
							Symbol:      candidate.Symbol,
							ProfileName: selectedProfile,
							AssetType:   "stock",
							Reason: sql.NullString{
								String: "User ignored during scout",
								Valid:  true,
							},
						})
						if err != nil {
							fmt.Printf("      ❌ Failed to ignore: %v\n", err)
						} else {
							fmt.Printf("      ⏭️ Skipping %s for 2 days\n", candidate.Symbol)
						}
						break
					}

					if choice == "n" {
						fmt.Printf("      ⏭️ Skipped %s\n", candidate.Symbol)
						break
					}

					fmt.Println("      ❌ Invalid choice. Try again.")
				}
			}
		}

		nextOffset := offset + batchSize
		if nextOffset < totalSymbols {
			ClearInputBuffer()
			fmt.Print("\n⏸️  Continue scanning next batch? (y to continue, or press Enter to stop): ")
			var continueChoice string
			fmt.Scanln(&continueChoice)
			continueChoice = strings.ToLower(continueChoice)
			if continueChoice != "y" {
				fmt.Println("⏹️ Scout review stopped")
				break
			}
		}

		offset = nextOffset
		batchNum++
	}

	fmt.Println("\n--- Press Enter to continue ---")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// displays trade signals and executes manual trades
func HandleExecuteTrades(ctx context.Context, cfg *config.Config, q *database.Queries, client *alpaca.Client) {
	fmt.Println("\n❌ No recent signals loaded. Run 'Analyze' first to detect trade opportunities.")
	fmt.Println("(Manual trade feature requires signal data from screener)")
	fmt.Println("\nNote: To execute trades, run analysis to generate LONG/SHORT signals,")
	fmt.Println("then ExecuteTradesFromSignals() will handle the interactive trade selection.")
}

// displays trade history and statistics
func HandleTradeHistory(ctx context.Context, cfg *config.Config, q *database.Queries) {
	fmt.Println("\n=== Trade History ===")
	var symbol string
	fmt.Print("Enter symbol (or 'all' for all trades): ")
	fmt.Scanln(&symbol)

	// Get trades
	if symbol == "" || symbol == "all" {
		trades, err := datafeed.GetOpenTrades(ctx)
		if err != nil {
			fmt.Printf("❌ Error retrieving open trades: %v\n", err)
			return
		}

		if len(trades) > 0 {
			fmt.Println("\n📊 Open Trades:")
			for _, trade := range trades {
				fmt.Printf("  %s | %s x %s @ %s | Status: %s\n",
					trade.Symbol, trade.Side, trade.Quantity, trade.Price, trade.Status)
			}
		}
	} else {
		trades, err := datafeed.GetTradeHistory(ctx, symbol, 50)
		if err != nil {
			fmt.Printf("❌ Error retrieving trades: %v\n", err)
			return
		}

		if len(trades) == 0 {
			fmt.Println("No trades found")
			return
		}

		fmt.Println("\n📋 Trade History for " + symbol + ":")
		for _, trade := range trades {
			fmt.Printf("  %s x %s @ %s | Total: %s | Status: %s\n",
				trade.Side, trade.Quantity, trade.Price, trade.TotalValue, trade.Status)
		}
	}
}

func HandlePaperTrade(ctx context.Context, client *alpaca.Client, queries *database.Queries, cfg *config.Config) error {
	ClearInputBuffer()

	// Get symbol from user
	fmt.Print("Enter symbol to paper trade (e.g., AAPL): ")
	var symbol string
	_, err := fmt.Scanln(&symbol)
	if err != nil || symbol == "" {
		fmt.Println("❌ Invalid symbol")
		return fmt.Errorf("invalid symbol")
	}

	// Get quantity
	fmt.Print("Enter quantity (default 1): ")
	var quantity int64
	_, err = fmt.Scanln(&quantity)
	if err != nil || quantity < 1 {
		quantity = 1
	}

	// Fetch bars
	fmt.Println("📊 Fetching market data...")
	bars, err := interactive.FetchMarketDataWithType(symbol, "1Day", 100, "", "stock")
	if err != nil {
		fmt.Printf("❌ Failed to fetch data: %v\n", err)
		return err
	}

	if len(bars) == 0 {
		fmt.Println("❌ No data returned")
		return fmt.Errorf("no bars fetched")
	}

	// Get latest bar
	bar := bars[len(bars)-1]

	// Convert bars to closes for RSI
	closes := make([]float64, len(bars))
	atrBars := make([]strategy.ATRBar, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
		atrBars[i] = strategy.ATRBar{
			High:  b.High,
			Low:   b.Low,
			Close: b.Close,
		}
	}

	// Calculate indicators
	rsiValues, err := strategy.CalculateRSI(closes, 14)
	if err != nil || len(rsiValues) == 0 {
		fmt.Println("❌ Could not calculate RSI")
		return fmt.Errorf("RSI calculation failed")
	}

	atrValues, err := strategy.CalculateATR(atrBars, 14)
	if err != nil || len(atrValues) == 0 {
		fmt.Println("❌ Could not calculate ATR")
		return fmt.Errorf("ATR calculation failed")
	}

	// Get latest indicator values
	latestRSI := rsiValues[len(rsiValues)-1]
	latestATR := atrValues[len(atrValues)-1]

	// Get criteria from config
	profile := cfg.Profiles["balanced"] // use default or let user select
	criteria := strategy.ScreenerCriteria{
		MinOversoldRSI: profile.Indicators.RSI.MinOversold,
		MaxRSI:         profile.Indicators.RSI.MaxOverbought,
		MinATR:         profile.Indicators.ATR.MinVolatility,
	}

	// Analyze both directions
	longSignal := strategy.AnalyzeForLongs(bar, &latestRSI, &latestATR, criteria)
	shortSignal := strategy.AnalyzeForShorts(bar, &latestRSI, &latestATR, criteria)

	// Pick the better signal
	var signal *strategy.TradeSignal
	if longSignal != nil && shortSignal != nil {
		if longSignal.Confidence >= shortSignal.Confidence {
			signal = longSignal
		} else {
			signal = shortSignal
		}
	} else if longSignal != nil {
		signal = longSignal
	} else if shortSignal != nil {
		signal = shortSignal
	}

	if signal == nil {
		fmt.Printf("❌ No trade signal found (RSI: %.2f, ATR: %.2f)\n", latestRSI, latestATR)
		return fmt.Errorf("no valid signal")
	}

	// Execute the trade
	fmt.Printf("📈 Placing %s order: %s x %d @ %.2f%% confidence\n",
		signal.Direction, symbol, quantity, signal.Confidence)
	fmt.Printf("   Reason: %s\n", signal.Reasoning)

	err = strategy.ExecuteTrade(ctx, client, symbol, quantity, signal)
	if err != nil {
		fmt.Printf("❌ Trade execution failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Paper trade executed!")
	return nil
}

func HandleWatchlistMenu(ctx context.Context, cfg *config.Config, q *database.Queries) {
	for {
		fmt.Println("\n--- Watchlist Menu ---")
		fmt.Println("1. Scan Watchlist")
		fmt.Println("2. View Watchlist")
		fmt.Println("3. Back")
		fmt.Print("Enter choice (1-3): ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("Invalid input. Try again.")
			continue
		}

		switch choice {
		case 1:
			HandleScan(ctx, cfg, q)
		case 2:
			HandleWatchlist(ctx, q)
		case 3:
			return
		default:
			fmt.Println("Invalid choice. Try again.")
		}
	}
}

func HandleAnalyzeAssetType(ctx context.Context, cfg *config.Config, q *database.Queries) {
	for {
		fmt.Println("\n🔬 Analyze:")
		fmt.Println("1. Stock")
		if cfg.Features.CryptoSupport {
			fmt.Println("2. Crypto")
			fmt.Println("3. Back")
		} else {
			fmt.Println("2. Back")
		}
		fmt.Print("Enter choice: ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("❌ Invalid input")
			continue
		}

		if choice == 1 {
			HandleAnalyzeSingle(ctx, "stock", datafeed.Queries)
			ClearInputBuffer()
		} else if choice == 2 && cfg.Features.CryptoSupport {
			HandleAnalyzeSingle(ctx, "crypto", datafeed.Queries)
			ClearInputBuffer()
		} else if (choice == 2 && !cfg.Features.CryptoSupport) || (choice == 3 && cfg.Features.CryptoSupport) {
			return
		} else {
			fmt.Println("❌ Invalid choice")
		}
	}
}
