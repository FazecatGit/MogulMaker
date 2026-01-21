package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	datafeed "github.com/fazecat/mogulmaker/Internal/database"
	"github.com/fazecat/mogulmaker/Internal/handlers"
	"github.com/fazecat/mogulmaker/Internal/handlers/monitoring"
	"github.com/fazecat/mogulmaker/Internal/handlers/risk"
	newsscraping "github.com/fazecat/mogulmaker/Internal/news_scraping"
	"github.com/fazecat/mogulmaker/Internal/strategy"
	"github.com/fazecat/mogulmaker/Internal/strategy/position"
	"github.com/fazecat/mogulmaker/Internal/utils"
	"github.com/fazecat/mogulmaker/Internal/utils/config"
	"github.com/fazecat/mogulmaker/Internal/utils/scanner"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	err = datafeed.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer datafeed.CloseDatabase()

	// Test the retry logic
	// utils.TestRetryLogic()
	// "github.com/fazecat/mogulmaker/Internal/utils"

	apiKey := os.Getenv("ALPACA_API_KEY")
	secretKey := os.Getenv("ALPACA_API_SECRET")

	alpclient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    apiKey,
		APISecret: secretKey,
		BaseURL:   "https://paper-api.alpaca.markets",
	})

	req, _ := http.NewRequest("GET", "https://paper-api.alpaca.markets/v2/account", nil)
	req.Header.Set("APCA-API-KEY-ID", apiKey)
	req.Header.Set("APCA-API-SECRET-KEY", secretKey)

	_, err = alpclient.GetAccount()
	if err != nil {
		log.Fatalln(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	cfg, _ := config.LoadConfig()
	status, isOpen := utils.CheckMarketStatus(time.Now(), cfg)
	fmt.Printf("Market Status: %s (Open: %v)\n\n", status, isOpen)

	account, err := alpclient.GetAccount()
	if err != nil {
		log.Printf("Warning: Could not fetch account for risk manager: %v\n", err)
	}

	var riskMgr *risk.Manager
	if account != nil {
		accountEquity, _ := account.Equity.Float64()
		riskMgr = risk.NewManager(alpclient, accountEquity)
		log.Println("Risk Manager initialized")
	} else {
		log.Println("Risk Manager could not be initialized - account data unavailable")
	}

	orderConfig := &strategy.OrderConfig{
		MaxOpenPositions:      5,
		MaxPortfolioPercent:   20.0,
		StopLossPercent:       2.0,
		TakeProfitPercent:     5.0,
		SafeBailPercent:       3.0,
		MaxDailyLossPercent:   -2.0,
		PartialExitPercentage: 0.5,
	}
	posManager := position.NewPositionManager(alpclient, orderConfig)

	tradeMon := monitoring.NewMonitor(posManager, riskMgr, datafeed.Queries)
	log.Println("Trade Monitor initialized")

	log.Println("Previous trades loaded from database")

	err = datafeed.InitAlpacaClient()
	if err != nil {
		log.Printf("Warning: Alpaca client initialization failed: %v\n", err)
	}

	finnhubClient := newsscraping.NewFinnhubClient()
	newsStorage := newsscraping.NewNewsStorage(datafeed.Queries)
	log.Println("News scraping initialized")

	ctx := context.Background()
	go startBackgroundScanner(ctx, cfg)

	for {
		if pm := handlers.GetGlobalPositionManager(); pm != nil {
			pm.CheckMenuAlerts()
		}

		fmt.Println("\n--- MongelMaker Menu ---")
		fmt.Println("1. Watchlist")
		fmt.Println("2. Analyze (Stock/Crypto)")
		fmt.Println("3. Scout Symbols")
		fmt.Println("4. Execute Trades")
		fmt.Println("5. Trade History")
		fmt.Println("6. Configure Settings")
		fmt.Println("7. Close/Sell Position")
		fmt.Println("8. Risk Manager Dashboard")
		fmt.Println("9. Trade Monitor")
		fmt.Println("10. Exit")
		fmt.Print("Enter choice (1-10): ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("Invalid input. Try again.")
			continue
		}

		switch choice {
		case 1:
			handlers.HandleWatchlistMenu(ctx, cfg, datafeed.Queries)
		case 2:
			handlers.HandleAnalyzeAssetType(ctx, cfg, datafeed.Queries, newsStorage, finnhubClient)
		case 3:
			handlers.HandleScout(ctx, cfg, datafeed.Queries, newsStorage, finnhubClient)
		case 4:
			handlers.HandleExecuteTrades(ctx, cfg, datafeed.Queries, alpclient)
		case 5:
			handlers.HandleTradeHistory(ctx, cfg, datafeed.Queries)
		case 6:
			config.ConfigureInteractive(cfg)
		case 7:
			handlers.HandleClosePosition(ctx, alpclient, cfg)
		case 8:
			handlers.HandleDisplayRiskManager(riskMgr, posManager)
		case 9:
			handlers.HandleDisplayTradeMonitor(tradeMon)
		case 10:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Try again.")
		}
	}
}

func startBackgroundScanner(ctx context.Context, cfg *config.Config) {
	log.Println("Background scanner started...")
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			log.Println("Background scanner stopped")
			return
		default:
			log.Println("Background scanner tick - checking for scans...")
			_, err := scanner.PerformScan(ctx, "default", cfg, datafeed.Queries)
			if err != nil {
				log.Printf("Background scan error: %v", err)
			} else {
				log.Println("Background scan completed successfully")
			}
			scanner.PerformScan(ctx, "default", cfg, datafeed.Queries)

		}
	}
}
