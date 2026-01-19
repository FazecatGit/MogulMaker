package datafeed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/fazecat/mogulmaker/Internal/types"
	"github.com/fazecat/mogulmaker/Internal/utils"
)

type Bar = types.Bar

func GetAlpacaBars(symbol string, timeframe string, limit int, startDate string) ([]Bar, error) {
	return GetAlpacaBarsWithType(symbol, timeframe, limit, startDate, "stock")
}

func GetAlpacaBarsWithType(symbol string, timeframe string, limit int, startDate string, assetType string) ([]Bar, error) {
	apiKey := os.Getenv("ALPACA_API_KEY")
	secretKey := os.Getenv("ALPACA_API_SECRET")

	if startDate == "" {
		now := time.Now().UTC()

		timeframeToDur := func(tf string) time.Duration {
			switch tf {
			case "1Min":
				return time.Minute
			case "3Min":
				return 3 * time.Minute
			case "5Min":
				return 5 * time.Minute
			case "10Min":
				return 10 * time.Minute
			case "30Min":
				return 30 * time.Minute
			case "1Hour":
				return time.Hour
			case "2Hour":
				return 2 * time.Hour
			case "4Hour":
				return 4 * time.Hour
			case "1Day":
				return 24 * time.Hour
			case "1Week":
				return 7 * 24 * time.Hour
			case "1Month":
				return 30 * 24 * time.Hour
			default:
				return 24 * time.Hour
			}
		}

		barDur := timeframeToDur(timeframe)
		totalDur := barDur * time.Duration(limit+2)
		start := now.Add(-totalDur)
		startDate = start.Format(time.RFC3339)
	}

	var apiURL string
	if assetType == "crypto" {
		apiURL = fmt.Sprintf(
			"https://data.alpaca.markets/v1beta3/crypto/us/bars?symbols=%s&timeframe=%s&limit=%d&start=%s",
			url.QueryEscape(symbol), timeframe, limit, startDate,
		)
	} else {
		apiURL = fmt.Sprintf(
			"https://data.alpaca.markets/v2/stocks/%s/bars?timeframe=%s&limit=%d&start=%s",
			symbol, timeframe, limit, startDate,
		)
	}

	fmt.Printf("🔗 API Request: %s\n", apiURL)

	var bars []Bar
	retryConfig := utils.DefaultRetryConfig()

	err := utils.RetryWithBackoff(func() error {
		req, _ := http.NewRequest("GET", apiURL, nil)
		req.Header.Set("APCA-API-KEY-ID", apiKey)
		req.Header.Set("APCA-API-SECRET-KEY", secretKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		fmt.Printf("📡 API Response Status: %s\n", resp.Status)

		if resp.StatusCode == 403 {
			fmt.Printf("⚠️  403 Forbidden - Your account may not have access to %s data\n", timeframe)
			bars = []Bar{}
			return nil
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		// Handle different response structures for stock vs crypto
		if assetType == "crypto" {
			type CryptoResponse struct {
				Bars map[string][]types.CryptoBar `json:"bars"`
			}
			var r CryptoResponse
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				return err
			}
			// Extract bars for the requested symbol and convert to standard Bar format
			for _, barSlice := range r.Bars {
				for _, cb := range barSlice {
					bars = append(bars, Bar{
						Timestamp: cb.Timestamp,
						Open:      cb.Open,
						High:      cb.High,
						Low:       cb.Low,
						Close:     cb.Close,
						Volume:    int64(cb.Volume), // Convert float to int64
					})
				}
				break
			}
		} else {
			// v2 stock endpoint returns flat structure with int volumes
			type StockResponse struct {
				Bars []Bar `json:"bars"`
			}
			var r StockResponse
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				return err
			}
			bars = r.Bars
		}
		return nil
	}, retryConfig)

	if err != nil {
		return nil, err
	}

	fmt.Printf("📊 Received %d bars\n", len(bars))

	// Reverse bars to latest-first (most recent data first)
	for i, j := 0, len(bars)-1; i < j; i, j = i+1, j-1 {
		bars[i], bars[j] = bars[j], bars[i]
	}

	return bars, nil
}

type LastQuote struct {
	Price float64 `json:"ap"`
}

func GetLastQuote(symbol string) (*LastQuote, error) {
	apiKey := os.Getenv("ALPACA_API_KEY")
	secretKey := os.Getenv("ALPACA_API_SECRET")

	url := fmt.Sprintf("https://data.alpaca.markets/v2/stocks/%s/quotes/latest", url.PathEscape(symbol))

	var quote *LastQuote
	retryConfig := utils.DefaultRetryConfig()

	err := utils.RetryWithBackoff(func() error {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("APCA-API-KEY-ID", apiKey)
		req.Header.Set("APCA-API-SECRET-KEY", secretKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get last quote: %s", resp.Status)
		}

		type Response struct {
			Quote LastQuote `json:"quote"`
		}

		var r Response
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return err
		}

		quote = &r.Quote
		return nil
	}, retryConfig)

	return quote, err
}

func GetLastTrade(symbol string) (*Bar, error) {
	apiKey := os.Getenv("ALPACA_API_KEY")
	secretKey := os.Getenv("ALPACA_API_SECRET")

	url := fmt.Sprintf("https://data.alpaca.markets/v2/stocks/%s/trades/latest", url.PathEscape(symbol))

	var trade *Bar
	retryConfig := utils.DefaultRetryConfig()

	err := utils.RetryWithBackoff(func() error {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("APCA-API-KEY-ID", apiKey)
		req.Header.Set("APCA-API-SECRET-KEY", secretKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get last trade: %s", resp.Status)
		}

		var r Bar
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return err
		}

		trade = &r
		return nil
	}, retryConfig)

	return trade, err
}

var alpacaClient *alpaca.Client

func InitAlpacaClient() error {
	apiKey := os.Getenv("ALPACA_API_KEY")
	secretKey := os.Getenv("ALPACA_API_SECRET")

	if apiKey == "" || secretKey == "" {
		return fmt.Errorf("ALPACA_API_KEY or ALPACA_API_SECRET not set")
	}

	alpacaClient = alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    apiKey,
		APISecret: secretKey,
		BaseURL:   "https://paper-api.alpaca.markets",
	})

	return nil
}

func GetAlpacaClient() *alpaca.Client {
	return alpacaClient
}
