package newsscraping

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type FinnhubClient struct {
	apiKey     string
	httpClient *http.Client
}

type finnhubNewsItem struct {
	Headline string `json:"headline"`
	URL      string `json:"url"`
	DateTime int64  `json:"datetime"`
}

func NewFinnhubClient() *FinnhubClient {
	return &FinnhubClient{
		apiKey: os.Getenv("FINNHUB_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *FinnhubClient) FetchNews(symbol string, limit int) ([]NewsArticle, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("FINNHUB_API_KEY not set in environment")
	}

	dateFrom := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	dateTo := time.Now().Format("2006-01-02")
	url := fmt.Sprintf(
		"https://finnhub.io/api/v1/company-news?symbol=%s&from=%s&to=%s&token=%s",
		symbol, dateFrom, dateTo, c.apiKey,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var newsItems []finnhubNewsItem
	if err := json.NewDecoder(resp.Body).Decode(&newsItems); err != nil {
		return nil, fmt.Errorf("failed to parse news: %v", err)
	}

	var articles []NewsArticle

	sentimentAnalyzer := NewSentimentAnalyzer()
	catalystDetector := NewCatalystDetector()

	for i, item := range newsItems {
		if i >= limit {
			break
		}

		// Analyze sentiment and detect catalysts
		sentiment, _ := sentimentAnalyzer.Analyze(item.Headline)
		catalystType := catalystDetector.Detect(item.Headline)
		impact := catalystDetector.GetImpact(catalystType)

		articles = append(articles, NewsArticle{
			Symbol:       symbol,
			Headline:     item.Headline,
			URL:          item.URL,
			PublishedAt:  time.Unix(item.DateTime, 0),
			Source:       "Finnhub",
			Sentiment:    sentiment,
			CatalystType: catalystType,
			Impact:       impact,
			CreatedAt:    time.Now(),
		})
	}

	return articles, nil
}

func (c *FinnhubClient) Name() string {
	return "Finnhub"
}
