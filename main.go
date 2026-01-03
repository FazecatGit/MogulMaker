package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	datafeed "github.com/fazecat/mongelmaker/Internal/database"
	database "github.com/fazecat/mongelmaker/Internal/database/sqlc"
	"github.com/fazecat/mongelmaker/Internal/handlers"
	newsscraping "github.com/fazecat/mongelmaker/Internal/news_scraping"
	"github.com/fazecat/mongelmaker/Internal/utils"
	"github.com/fazecat/mongelmaker/Internal/utils/config"
	"github.com/fazecat/mongelmaker/Internal/utils/scanner"
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
	// "github.com/fazecat/mongelmaker/Internal/utils"

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
	fmt.Printf("📊 Market Status: %s (Open: %v)\n\n", status, isOpen)

	// for the scouting feature
	err = datafeed.InitAlpacaClient()
	if err != nil {
		log.Printf("Warning: Alpaca client initialization failed: %v\n", err)
	}

	finnhubClient := newsscraping.NewFinnhubClient()
	_ = finnhubClient

	ctx := context.Background()
	go startBackgroundScanner(ctx, cfg)

	for {
		fmt.Println("\n--- MongelMaker Menu ---")
		fmt.Println("1. Watchlist")
		fmt.Println("2. Analyze (Stock/Crypto)")
		fmt.Println("3. Scout Symbols")
		fmt.Println("4. Execute Trades")
		fmt.Println("5. Trade History")
		fmt.Println("6. Configure Settings")
		fmt.Println("7. Exit")
		fmt.Print("Enter choice (1-7): ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("Invalid input. Try again.")
			continue
		}

		switch choice {
		case 1:
			handleWatchlistMenu(ctx, cfg, datafeed.Queries)
		case 2:
			handleAnalyzeAssetType(ctx, cfg, datafeed.Queries)
		case 3:
			handlers.HandleScout(ctx, cfg, datafeed.Queries)
		case 4:
			handlers.HandleExecuteTrades(ctx, cfg, datafeed.Queries, alpclient)
		case 5:
			handlers.HandleTradeHistory(ctx, cfg, datafeed.Queries)
		case 6:
			config.ConfigureInteractive(cfg)
		case 7:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Try again.")
		}
	}
}

func handleWatchlistMenu(ctx context.Context, cfg *config.Config, q *database.Queries) {
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
			handlers.HandleScan(ctx, cfg, q)
		case 2:
			handlers.HandleWatchlist(ctx, q)
		case 3:
			return
		default:
			fmt.Println("Invalid choice. Try again.")
		}
	}
}

func handleAnalyzeAssetType(ctx context.Context, cfg *config.Config, q *database.Queries) {
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
			handlers.HandleAnalyzeSingle(ctx, "stock", datafeed.Queries)
			handlers.ClearInputBuffer()
		} else if choice == 2 && cfg.Features.CryptoSupport {
			handlers.HandleAnalyzeSingle(ctx, "crypto", datafeed.Queries)
			handlers.ClearInputBuffer()
		} else if (choice == 2 && !cfg.Features.CryptoSupport) || (choice == 3 && cfg.Features.CryptoSupport) {
			return
		} else {
			fmt.Println("❌ Invalid choice")
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
