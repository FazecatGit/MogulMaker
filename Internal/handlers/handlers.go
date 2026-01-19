package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	datafeed "github.com/fazecat/mongelmaker/Internal/database"
	database "github.com/fazecat/mongelmaker/Internal/database/sqlc"
	"github.com/fazecat/mongelmaker/Internal/database/watchlist"
	"github.com/fazecat/mongelmaker/Internal/handlers/risk"
	newsscraping "github.com/fazecat/mongelmaker/Internal/news_scraping"
	"github.com/fazecat/mongelmaker/Internal/strategy"
	positionPkg "github.com/fazecat/mongelmaker/Internal/strategy/position"
	"github.com/fazecat/mongelmaker/Internal/types"
	"github.com/fazecat/mongelmaker/Internal/utils/config"
	"github.com/fazecat/mongelmaker/Internal/utils/scanner"
	"github.com/fazecat/mongelmaker/Internal/utils/scoring"
	"github.com/fazecat/mongelmaker/interactive"
	"github.com/shopspring/decimal"
)

// Global position manager for tracking open trades
var (
	globalPosManager *positionPkg.PositionManager
	posManagerMutex  sync.RWMutex
)

// stores the position manager for alerts
func SetGlobalPositionManager(pm *positionPkg.PositionManager) {
	posManagerMutex.Lock()
	defer posManagerMutex.Unlock()
	globalPosManager = pm
}

// retrieves the position manager
func GetGlobalPositionManager() *positionPkg.PositionManager {
	posManagerMutex.RLock()
	defer posManagerMutex.RUnlock()
	return globalPosManager
}

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

func HandleAnalyzeSingle(ctx context.Context, assetType string, q *database.Queries, newsStorage *newsscraping.NewsStorage, finnhubClient *newsscraping.FinnhubClient) {
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

	// Fetch and store news for stocks (not crypto)
	if assetType == "stock" && finnhubClient != nil && newsStorage != nil {
		fmt.Println("📰 Fetching latest news...")
		newsArticles, err := finnhubClient.FetchNews(symbol, 5)
		if err == nil && len(newsArticles) > 0 {
			for _, article := range newsArticles {
				_ = newsStorage.SaveArticle(ctx, article)
			}
			log.Printf("✅ Saved %d news articles for %s", len(newsArticles), symbol)
		} else if err != nil {
			log.Printf("⚠️  Could not fetch news: %v", err)
		}
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
		interactive.DisplayAnalyticsData(bars, symbol, timeframe, tz, q, newsStorage)
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

func HandleScout(ctx context.Context, cfg *config.Config, q *database.Queries, newsStorage *newsscraping.NewsStorage, finnhubClient *newsscraping.FinnhubClient) {
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
						newsStorage := newsscraping.NewNewsStorage(q)
						interactive.DisplayAnalyticsData(candidate.Bars, candidate.Symbol, "1Day", tz, q, newsStorage)
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
	ClearInputBuffer()

	separator := "============================================================"
	fmt.Println("\n" + separator)
	fmt.Println("🚀 LIVE TRADE EXECUTION")
	fmt.Println(separator)

	// Get account info
	account, err := client.GetAccount()
	if err != nil {
		fmt.Printf("❌ Failed to get account info: %v\n", err)
		return
	}

	accountValueFloat, _ := account.Equity.Float64()
	accountValue := accountValueFloat
	fmt.Printf("💰 Account Balance: $%.2f\n", accountValue)

	// Create position manager with safety limits
	orderConfig := &strategy.OrderConfig{
		MaxPortfolioPercent:   20.0, // Max 20% of account per trade
		MaxOpenPositions:      5,    // Max 5 concurrent trades
		StopLossPercent:       2.0,  // 2% stop loss
		TakeProfitPercent:     5.0,  // 5% take profit
		SafeBailPercent:       3.0,  // 3% safe bail for partial exit
		MaxDailyLossPercent:   -2.0, // Stop if daily loss exceeds -2%
		PartialExitPercentage: 0.5,  // Exit 50% at take profit
	}

	posManager := positionPkg.NewPositionManager(client, orderConfig)

	// Store globally so menu can access alerts
	SetGlobalPositionManager(posManager)

	// Get user input
	fmt.Print("\nEnter symbol to trade (e.g., AAPL): ")
	var symbol string
	_, err = fmt.Scanln(&symbol)
	if err != nil || symbol == "" {
		fmt.Println("❌ Invalid symbol")
		return
	}

	fmt.Print("Enter direction (LONG/SHORT): ")
	var direction string
	_, err = fmt.Scanln(&direction)
	if err != nil {
		fmt.Println("❌ Invalid direction")
		return
	}

	if direction != "LONG" && direction != "SHORT" {
		fmt.Println("❌ Direction must be LONG or SHORT")
		return
	}

	fmt.Print("Enter quantity (or 0 to auto-calculate): ")
	var quantity int64
	_, err = fmt.Scanln(&quantity)
	if err != nil || quantity < 0 {
		fmt.Println("❌ Invalid quantity")
		return
	}

	// Fetch current market data
	fmt.Println("\n📊 Fetching market data...")
	bars, err := interactive.FetchMarketDataWithType(symbol, "1Day", 100, "", "stock")
	if err != nil {
		fmt.Printf("❌ Failed to fetch data: %v\n", err)
		return
	}

	if len(bars) == 0 {
		fmt.Println("❌ No market data available")
		return
	}

	bar := bars[len(bars)-1]
	entryPrice := bar.Close

	// Calculate price targets
	stopLoss, takeProfit := strategy.CalculatePriceTargets(entryPrice, direction, orderConfig)
	safeBail := 0.0
	if direction == "LONG" {
		safeBail = entryPrice * (1 + (orderConfig.SafeBailPercent / 100))
	} else {
		safeBail = entryPrice * (1 - (orderConfig.SafeBailPercent / 100))
	}

	// Auto-calculate quantity if needed
	if quantity == 0 {
		quantity = strategy.CalculatePositionSize(accountValue, entryPrice, stopLoss, orderConfig.MaxPortfolioPercent, orderConfig)
		fmt.Printf("📐 Auto-calculated quantity: %d shares\n", quantity)
	}

	// Create order request
	orderReq := &strategy.OrderRequest{
		Symbol:           symbol,
		Quantity:         quantity,
		Direction:        direction,
		SignalConfidence: 75.0, // Default confidence for manual trades
		TradeReason:      "Manual execution from HandleExecuteTrades",
		StopLossPrice:    stopLoss,
		TakeProfitPrice:  takeProfit,
		EntryPrice:       entryPrice,
		UseStopOrder:     true,
		UseLimitOrder:    false,
	}

	// Validate order
	openPositions := posManager.CountOpenPositions()
	dailyLoss := posManager.GetDailyLoss()

	validation := strategy.ValidateOrder(orderReq, orderConfig, accountValue, openPositions, dailyLoss)

	if !validation.IsValid {
		fmt.Println("❌ ORDER VALIDATION FAILED:")
		for _, issue := range validation.Issues {
			fmt.Printf("   • %s\n", issue)
		}
		return
	}

	// Display order preview
	fmt.Println("\n" + separator)
	fmt.Println("📋 ORDER PREVIEW")
	fmt.Println(separator)
	fmt.Printf("Symbol:              %s\n", orderReq.Symbol)
	fmt.Printf("Direction:           %s\n", orderReq.Direction)
	fmt.Printf("Quantity:            %d shares\n", orderReq.Quantity)
	fmt.Printf("Entry Price:         $%.2f\n", orderReq.EntryPrice)
	fmt.Printf("Stop Loss:           $%.2f (%.2f%% below entry)\n", stopLoss, orderConfig.StopLossPercent)
	fmt.Printf("Take Profit:         $%.2f (%.2f%% above entry)\n", takeProfit, orderConfig.TakeProfitPercent)
	fmt.Printf("Safe Bail:           $%.2f\n", safeBail)
	fmt.Printf("Max Risk:            $%.2f (%.2f%% of portfolio)\n", validation.RiskAmount, validation.PortfolioRisk)
	fmt.Printf("Potential Gain:      $%.2f\n", validation.PotentialGain)
	fmt.Printf("Risk/Reward Ratio:   1:%.2f\n", validation.PotentialGain/validation.RiskAmount)
	fmt.Println(separator)

	// Confirm trade
	fmt.Print("\n⚠️  CONFIRM TRADE? (yes/no): ")
	var confirm string
	_, err = fmt.Scanln(&confirm)
	if err != nil || (confirm != "yes" && confirm != "y") {
		fmt.Println("❌ Trade cancelled")
		return
	}

	// Build Alpaca order
	alpacaOrder, err := strategy.BuildPlaceOrderRequest(orderReq)
	if err != nil {
		fmt.Printf("❌ Failed to build order: %v\n", err)
		return
	}

	// Execute trade
	fmt.Println("\n⏳ Executing trade...")
	order, err := client.PlaceOrder(*alpacaOrder)
	if err != nil {
		fmt.Printf("❌ Trade execution failed: %v\n", err)
		return
	}

	// Add to position manager
	signal := &types.TradeSignal{
		Direction:  direction,
		Confidence: orderReq.SignalConfidence,
		Reasoning:  orderReq.TradeReason,
	}

	posManager.AddPosition(order, signal, entryPrice, stopLoss, takeProfit, safeBail)

	// Log execution
	strategy.LogOrderExecution(orderReq, validation, order.ID)

	// Log trade to database for persistent storage
	err = datafeed.LogTradeExecution(ctx, order.Symbol, direction, orderReq.Quantity,
		decimal.NewFromFloat(entryPrice), order.ID, order.Status)
	if err != nil {
		log.Printf("⚠️  Warning: Could not log trade to database: %v\n", err)
	}

	fmt.Println("\n✅ TRADE EXECUTED SUCCESSFULLY!")
	fmt.Printf("Order ID: %s | Status: %s\n", order.ID, order.Status)
	fmt.Println("\n📡 Position monitoring enabled in background")
	fmt.Println("   View it anytime via: Trade Monitor (Option 9)")

	// Start monitoring position in background
	go posManager.MonitorPositions(ctx, 5*time.Second)
}

// HandleClosePosition closes/sells an open position
func HandleClosePosition(ctx context.Context, client *alpaca.Client, cfg *config.Config) {
	ClearInputBuffer()

	separator := "============================================================"
	fmt.Println("\n" + separator)
	fmt.Println(" CLOSE/SELL POSITION")
	fmt.Println(separator)

	// Simple approach - ask user for symbol to close
	fmt.Print("\nEnter symbol to close (e.g., AAPL): ")
	var symbol string
	_, err := fmt.Scanln(&symbol)
	if err != nil || symbol == "" {
		fmt.Println(" Invalid symbol")
		return
	}

	fmt.Printf("\n  Close all positions for %s? (yes/no): ", symbol)
	var confirm string
	_, err = fmt.Scanln(&confirm)
	if err != nil || (confirm != "yes" && confirm != "y") {
		fmt.Println(" Close cancelled")
		return
	}

	fmt.Println("\n⏳ Closing position...")
	order, err := client.ClosePosition(symbol, alpaca.ClosePositionRequest{})
	if err != nil {
		fmt.Printf("❌ Failed to close position: %v\n", err)
		return
	}

	fmt.Println("\n POSITION CLOSED SUCCESSFULLY!")
	fmt.Printf("Symbol: %s\n", order.Symbol)
	fmt.Printf("Order ID: %s | Status: %s\n", order.ID, order.Status)
	if order.FilledAvgPrice != nil {
		avgPrice, _ := order.FilledAvgPrice.Float64()
		fmt.Printf("Filled Avg Price: $%.2f\n", avgPrice)
	}
	fmt.Println(separator)
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
			displayCount := 10 // Start with 10 trades
			totalTrades := len(trades)

			for {
				// Get trades to display (up to displayCount)
				endIndex := displayCount
				if endIndex > totalTrades {
					endIndex = totalTrades
				}

				fmt.Printf("\n📊 Open Trades (Showing %d of %d):\n", endIndex, totalTrades)
				for i := 0; i < endIndex; i++ {
					trade := trades[i]
					fmt.Printf("  %s | %s x %s @ %s | Status: %s\n",
						trade.Symbol, trade.Side, trade.Quantity, trade.Price, trade.Status)
				}

				// Show pagination options
				if endIndex < totalTrades {
					fmt.Printf("\n💡 Showing %d of %d trades\n", endIndex, totalTrades)
					fmt.Print("Press Enter to load 10 more, or type 'q' to quit: ")
					var input string
					fmt.Scanln(&input)

					if input == "q" || input == "Q" {
						break
					}

					displayCount += 10
				} else {
					fmt.Printf("\n✅ All %d trades displayed\n", totalTrades)
					break
				}
			}
		} else {
			fmt.Println("\n📝 No trades found in database")

			// Debug: Check ALL trades regardless of status
			allTrades, debugErr := datafeed.GetAllTradesDebug(ctx)
			if debugErr == nil && len(allTrades) > 0 {
				fmt.Println("\n🔍 DEBUG: Trades exist but with different statuses:")
				for _, trade := range allTrades {
					fmt.Printf("  %s | %s x %s | Status: %s\n",
						trade.Symbol, trade.Side, trade.Quantity, trade.Status.String)
				}
			}

			fmt.Println("\n💡 Tip: Trades are logged to the database when executed through MongelMaker")
			fmt.Println("   Your open positions in Alpaca can be viewed in the Trade Monitor (Option 9)")
		}
	} else {
		trades, err := datafeed.GetTradeHistory(ctx, symbol, 100) // Increased from 50 to 100
		if err != nil {
			fmt.Printf("❌ Error retrieving trades: %v\n", err)
			return
		}

		if len(trades) == 0 {
			fmt.Println("\n📝 No trades found for " + symbol)
			fmt.Println("💡 Tip: Trades are logged to the database when executed through MongelMaker")
			fmt.Println("   Your open positions in Alpaca can be viewed in the Trade Monitor (Option 9)")
			return
		}

		totalTrades := len(trades)
		displayCount := 10

		for {
			// Get trades to display (up to displayCount)
			endIndex := displayCount
			if endIndex > totalTrades {
				endIndex = totalTrades
			}

			fmt.Printf("\n📋 Trade History for %s (Showing %d of %d):\n", symbol, endIndex, totalTrades)
			for i := 0; i < endIndex; i++ {
				trade := trades[i]
				fmt.Printf("  %s x %s @ %s | Total: %s | Status: %s\n",
					trade.Side, trade.Quantity, trade.Price, trade.TotalValue, trade.Status)
			}

			// Show pagination options
			if endIndex < totalTrades {
				fmt.Printf("\n💡 Showing %d of %d trades\n", endIndex, totalTrades)
				fmt.Print("Press Enter to load 10 more, or type 'q' to quit: ")
				var input string
				fmt.Scanln(&input)

				if input == "q" || input == "Q" {
					break
				}

				displayCount += 10
			} else {
				fmt.Printf("\n✅ All %d trades displayed\n", totalTrades)
				break
			}
		}
	}
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

func HandleAnalyzeAssetType(ctx context.Context, cfg *config.Config, q *database.Queries, newsStorage *newsscraping.NewsStorage, finnhubClient *newsscraping.FinnhubClient) {
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
			HandleAnalyzeSingle(ctx, "stock", datafeed.Queries, newsStorage, finnhubClient)
			ClearInputBuffer()
		} else if choice == 2 && cfg.Features.CryptoSupport {
			HandleAnalyzeSingle(ctx, "crypto", datafeed.Queries, newsStorage, finnhubClient)
			ClearInputBuffer()
		} else if (choice == 2 && !cfg.Features.CryptoSupport) || (choice == 3 && cfg.Features.CryptoSupport) {
			return
		} else {
			fmt.Println("❌ Invalid choice")
		}
	}
}

// HandleDisplayRiskManager shows the portfolio risk report
func HandleDisplayRiskManager(riskManager interface{}, positionManager interface{}) {
	// Use type assertion directly instead of reflection
	rm, ok := riskManager.(*risk.Manager)
	if !ok || rm == nil {
		fmt.Println("⚠️  Risk Manager not available yet")
		return
	}

	// Get open positions from position manager
	var positions []*positionPkg.OpenPosition
	if pm, ok := positionManager.(*positionPkg.PositionManager); ok && pm != nil {
		// Sync with Alpaca first
		ctx := context.Background()
		if err := pm.SyncFromAlpaca(ctx); err != nil {
			fmt.Printf("Warning: Could not sync positions: %v\n", err)
		}
		positions = pm.GetOpenPositions()
	}

	// Generate and display report
	report := rm.GenerateRiskReport(positions)
	report.Print()
}

// HandleDisplayTradeMonitor shows the trade statistics and open positions
func HandleDisplayTradeMonitor(tradeMonitor interface{}) {
	type Monitor interface {
		PrintStatsReport()
		PrintOpenPositions()
		PrintRiskEvents()
		PrintTradeHistory()
	}

	if tm, ok := tradeMonitor.(Monitor); ok {
		// Show menu for what to display
		fmt.Println("\n📊 Trade Monitor Menu:")
		fmt.Println("1. Open Positions")
		fmt.Println("2. Trade Statistics")
		fmt.Println("3. Trade History")
		fmt.Println("4. Risk Events")
		fmt.Println("5. All")
		fmt.Println("6. Back")
		fmt.Print("Enter choice (1-6): ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil || choice < 1 || choice > 6 {
			fmt.Println("❌ Invalid choice")
			return
		}

		switch choice {
		case 1:
			tm.PrintOpenPositions()
		case 2:
			tm.PrintStatsReport()
		case 3:
			tm.PrintTradeHistory()
		case 4:
			tm.PrintRiskEvents()
		case 5:
			tm.PrintOpenPositions()
			tm.PrintStatsReport()
			tm.PrintTradeHistory()
			tm.PrintRiskEvents()
		case 6:
			return
		}
	} else {
		fmt.Println("⚠️  Trade Monitor not available yet")
	}
}
