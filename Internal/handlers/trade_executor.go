package handlers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"

	datafeed "github.com/fazecat/mongelmaker/Internal/database"
	"github.com/fazecat/mongelmaker/Internal/strategy"
	"github.com/fazecat/mongelmaker/Internal/utils/config"
)

// displays detected signals and lets user execute trades
func ExecuteTradesFromSignals(ctx context.Context, cfg *config.Config, scores []strategy.StockScore, client *alpaca.Client) {
	if client == nil {
		fmt.Println("❌ Alpaca client not initialized")
		return
	}

	// Filter scores with trade signals
	var signalsAvailable []strategy.StockScore
	for _, score := range scores {
		if score.LongSignal != nil || score.ShortSignal != nil {
			signalsAvailable = append(signalsAvailable, score)
		}
	}

	if len(signalsAvailable) == 0 {
		fmt.Println("✅ No trade signals detected")
		return
	}

	fmt.Printf("\n📊 Found %d symbols with trade signals\n", len(signalsAvailable))

	// Display all signals
	for i, score := range signalsAvailable {
		fmt.Printf("\n[%d] %s\n", i+1, score.Symbol)

		if score.LongSignal != nil {
			fmt.Printf("    🟢 LONG  | Confidence: %.2f%% | %s\n",
				score.LongSignal.Confidence, score.LongSignal.Reasoning)
		}

		if score.ShortSignal != nil {
			fmt.Printf("    🔴 SHORT | Confidence: %.2f%% | %s\n",
				score.ShortSignal.Confidence, score.ShortSignal.Reasoning)
		}
	}

	// Execute selected trades
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nEnter symbol number to trade (or 'done' to finish): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "done" {
			break
		}

		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(signalsAvailable) {
			fmt.Println("❌ Invalid selection")
			continue
		}

		score := signalsAvailable[idx-1]
		executeTradeForSymbol(ctx, cfg, score, client)
	}

	fmt.Println("\n✅ Trade execution complete")
}

func executeTradeForSymbol(ctx context.Context, cfg *config.Config, score strategy.StockScore, client *alpaca.Client) {
	fmt.Printf("\n💼 Trading %s:\n", score.Symbol)

	// Show options
	if score.LongSignal != nil && score.ShortSignal != nil {
		fmt.Println("1. LONG  (Buy)")
		fmt.Println("2. SHORT (Sell)")
		fmt.Println("3. Skip")
		fmt.Print("Enter choice (1-3): ")

		var choice int
		_, _ = fmt.Scanln(&choice)

		switch choice {
		case 1:
			executeLongTrade(ctx, cfg, score, client)
		case 2:
			executeShortTrade(ctx, cfg, score, client)
		default:
			fmt.Println("⏭️  Skipped")
		}
	} else if score.LongSignal != nil {
		fmt.Printf("Confidence: %.2f%%\n", score.LongSignal.Confidence)
		fmt.Print("Execute LONG order? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "y" {
			executeLongTrade(ctx, cfg, score, client)
		}
	} else if score.ShortSignal != nil {
		fmt.Printf("Confidence: %.2f%%\n", score.ShortSignal.Confidence)
		fmt.Print("Execute SHORT order? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "y" {
			executeShortTrade(ctx, cfg, score, client)
		}
	}
}

func executeLongTrade(ctx context.Context, cfg *config.Config, score strategy.StockScore, client *alpaca.Client) {
	fmt.Print("Enter quantity to buy: ")
	var qty int64
	_, err := fmt.Scanln(&qty)
	if err != nil || qty <= 0 {
		fmt.Println("❌ Invalid quantity")
		return
	}

	err = strategy.ExecuteTrade(ctx, client, score.Symbol, qty, score.LongSignal)
	if err != nil {
		fmt.Printf("❌ Trade failed: %v\n", err)
	} else {
		fmt.Printf("✅ LONG order executed: %s x%d\n", score.Symbol, qty)
	}
}

func executeShortTrade(ctx context.Context, cfg *config.Config, score strategy.StockScore, client *alpaca.Client) {
	fmt.Print("Enter quantity to short: ")
	var qty int64
	_, err := fmt.Scanln(&qty)
	if err != nil || qty <= 0 {
		fmt.Println("❌ Invalid quantity")
		return
	}

	err = strategy.ExecuteTrade(ctx, client, score.Symbol, qty, score.ShortSignal)
	if err != nil {
		fmt.Printf("❌ Trade failed: %v\n", err)
	} else {
		fmt.Printf("✅ SHORT order executed: %s x%d\n", score.Symbol, qty)
	}
}

// displays past trades
func ViewTradeHistory(ctx context.Context, symbol string) {
	trades, err := datafeed.GetTradeHistory(ctx, symbol, 50)
	if err != nil {
		fmt.Printf("❌ Failed to fetch trade history: %v\n", err)
		return
	}

	if len(trades) == 0 {
		fmt.Printf("ℹ️  No trades found for %s\n", symbol)
		return
	}

	fmt.Printf("\n📈 Trade History for %s (Last 50):\n", symbol)
	fmt.Println("─────────────────────────────────────────────────────────────")
	for _, trade := range trades {
		status := "PENDING"
		if trade.Status.Valid {
			status = trade.Status.String
		}

		createdAt := "N/A"
		if trade.CreatedAt.Valid {
			createdAt = trade.CreatedAt.Time.Format("2006-01-02 15:04")
		}

		fmt.Printf("%s | %s | Qty: %s | Price: %s | Total: %s | %s\n",
			trade.Side, createdAt, trade.Quantity, trade.Price, trade.TotalValue, status)
	}
}

// shows all open positions
func ViewOpenTrades(ctx context.Context) {
	trades, err := datafeed.GetOpenTrades(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to fetch open trades: %v\n", err)
		return
	}

	if len(trades) == 0 {
		fmt.Println("ℹ️  No open trades")
		return
	}

	fmt.Printf("\n📊 Open Trades (%d):\n", len(trades))
	fmt.Println("─────────────────────────────────────────────────────────────")
	for _, trade := range trades {
		status := "PENDING"
		if trade.Status.Valid {
			status = trade.Status.String
		}

		createdAt := "N/A"
		if trade.CreatedAt.Valid {
			createdAt = trade.CreatedAt.Time.Format("2006-01-02 15:04")
		}

		fmt.Printf("%s %s x%s @ %s | %s | %s\n",
			trade.Side, trade.Symbol, trade.Quantity, trade.Price, createdAt, status)
	}
}
