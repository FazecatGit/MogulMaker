package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	datafeed "github.com/fazecat/mogulmaker/Internal/database"
	"github.com/fazecat/mogulmaker/Internal/handlers/monitoring"
	"github.com/fazecat/mogulmaker/Internal/handlers/risk"
	"github.com/fazecat/mogulmaker/Internal/strategy"
	"github.com/fazecat/mogulmaker/Internal/strategy/position"
	"github.com/fazecat/mogulmaker/cmd/api/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

// duped from main.go will change later to use less code
func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")
	err := datafeed.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer datafeed.CloseDatabase()
	apiKey := os.Getenv("ALPACA_API_KEY")
	secretKey := os.Getenv("ALPACA_API_SECRET")

	alpclient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    apiKey,
		APISecret: secretKey,
		BaseURL:   "https://paper-api.alpaca.markets"})

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

	// Initialize JWT manager
	jwtManager := internal.NewJWTManager()

	apiServer := &internal.API{
		PositionManager: posManager,
		RiskManager:     riskMgr,
		Queries:         datafeed.Queries,
		TradeMonitor:    tradeMon,
		AlpacaClient:    alpclient,
		JWTManager:      jwtManager,
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(internal.CorsMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    "healthy",
		})
	})

	// Public routes
	r.Get("/api/positions", apiServer.HandleGetPositions)
	r.Get("/api/positions/{symbol}", apiServer.HandleGetPositionBySymbol)
	r.Get("/api/risk", apiServer.HandleGetRiskStatus)
	r.Get("/api/stats", apiServer.HandleGetStats)
	r.Get("/api/trades", apiServer.HandleGetTrades)
	r.Get("/api/trades/statistics", apiServer.HandleTradeStatistics)
	r.Post("/api/token", apiServer.HandleGenerateToken)

	//Analytics & Monitoring
	r.Get("/api/portfolio-summary", apiServer.HandlePortfolioSummary)
	r.Get("/api/risk-adjustments", apiServer.HandleRiskAdjustments)
	r.Get("/api/performance-metrics", apiServer.HandlePerformanceMetrics)
	r.Get("/api/risk-alerts", apiServer.HandleRiskAlerts)

	// News
	r.Get("/api/news", apiServer.HandleGetNews)

	//Backtesting & Analysis
	r.Get("/api/backtest", apiServer.HandleBacktest)
	r.Get("/api/backtest/results", apiServer.HandleBacktestResults)
	r.Get("/api/backtest/status", apiServer.HandleBacktestStatus)
	r.Get("/api/analysis/symbol", apiServer.HandleSymbolAnalysis)
	r.Get("/api/analysis/report", apiServer.HandleAnalysisReport)

	// Watchlist & Scanner
	r.Get("/api/watchlist", apiServer.HandleGetWatchlist)
	r.Post("/api/watchlist", apiServer.HandleAddToWatchlist)
	r.Delete("/api/watchlist", apiServer.HandleRemoveFromWatchlist)
	r.Get("/api/watchlist/analyze", apiServer.HandleAnalyzeSymbol)
	r.Get("/api/scout", apiServer.HandleScoutStocks)

	// Trade Execution
	r.Post("/api/execute-trade", apiServer.HandleExecuteTrade)
	r.Post("/api/trades", apiServer.HandleExecuteTrade)
	r.Post("/api/trades/sell-all", apiServer.HandleSellAllTrades)
	r.Delete("/api/positions/{symbol}", apiServer.HandleClosePosition)

	log.Println("Starting API server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
