package main

import (
	"fmt"
	"strings"
	"time"

	newscraping "github.com/fazecat/mogulmaker/Internal/news_scraping"
)

func main() {
	// 1. Try to fetch news
	fmt.Println("📰 Fetching news from RSS...")
	rss := newscraping.NewRSSClinet()
	articles, err := rss.FetchNews("AAPL", 5)
	if err != nil {
		fmt.Printf("⚠️  RSS fetch failed: %v\n", err)
		fmt.Println("📝 Creating test articles instead...")
		// Create test articles to demonstrate sentiment and catalyst detection
		articles = []newscraping.NewsArticle{
			{
				Symbol:      "AAPL",
				Headline:    "Apple Reports Record Profit, Earnings Beat Expectations",
				URL:         "https://example.com/1",
				PublishedAt: time.Now(),
				Source:      "Test",
				CreatedAt:   time.Now(),
			},
			{
				Symbol:      "AAPL",
				Headline:    "Apple Stock Surges on Strong Q4 Revenue Growth",
				URL:         "https://example.com/2",
				PublishedAt: time.Now(),
				Source:      "Test",
				CreatedAt:   time.Now(),
			},
			{
				Symbol:      "AAPL",
				Headline:    "Apple Faces FDA Investigation Over Privacy Concerns",
				URL:         "https://example.com/3",
				PublishedAt: time.Now(),
				Source:      "Test",
				CreatedAt:   time.Now(),
			},
			{
				Symbol:      "AAPL",
				Headline:    "Apple Stock Plunges After Missing Analyst Expectations",
				URL:         "https://example.com/4",
				PublishedAt: time.Now(),
				Source:      "Test",
				CreatedAt:   time.Now(),
			},
		}
	}

	// 2. Analyze sentiment and catalysts
	sentiment := newscraping.NewSentimentAnalyzer()
	catalyst := newscraping.NewCatalystDetector()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SENTIMENT & CATALYST ANALYSIS")
	fmt.Println(strings.Repeat("=", 80))

	for _, article := range articles {
		// Sentiment
		sent, score := sentiment.Analyze(article.Headline)

		// Catalyst
		catalystType := catalyst.Detect(article.Headline)
		impact := catalyst.GetImpact(catalystType)

		// Display
		fmt.Printf("\n📝 %s\n", article.Headline)
		fmt.Printf("   URL: %s\n", article.URL)
		fmt.Printf("   📊 Sentiment: %s (Score: %.2f)\n", sent, score)
		fmt.Printf("   🎯 Catalyst: %s (Impact: %.0f%%)\n", catalystType, impact*100)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ Sentiment & Catalyst Analysis Complete!")
	fmt.Println(strings.Repeat("=", 80))
}
