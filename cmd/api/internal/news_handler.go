package internal

import (
	"log"
	"net/http"

	newsscraping "github.com/fazecat/mogulmaker/Internal/news_scraping"
)

func (api *API) HandleGetNews(w http.ResponseWriter, r *http.Request) {
	positions, err := api.AlpacaClient.GetPositions()
	if err != nil {
		log.Printf("Error fetching positions: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch positions")
		return
	}

	// Get watchlist from database
	ctx := r.Context()
	watchlist, err := api.Queries.GetWatchlist(ctx)
	if err != nil {
		log.Printf("Warning: Could not fetch watchlist: %v", err)
		// Continue anyway - we can still get news for positions
	}

	// Create a map of symbols that the user cares about
	symbols := make(map[string]bool)

	// Add all position symbols
	for _, pos := range positions {
		symbols[pos.Symbol] = true
	}

	// Add all watchlist symbols
	for _, item := range watchlist {
		symbols[item.Symbol] = true
	}

	if len(symbols) == 0 {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"news":  []interface{}{},
			"count": 0,
		})
		return
	}

	// Fetch news for each symbol
	finnhubClient := newsscraping.NewFinnhubClient()
	var allNews []map[string]interface{}
	newsCount := 0

	// Track seen URLs to avoid duplicates
	seenURLs := make(map[string]bool)

	for symbol := range symbols {
		articles, err := finnhubClient.FetchNews(symbol, 5) // 5 articles per symbol
		if err != nil {
			log.Printf("Warning: Failed to fetch news for %s: %v", symbol, err)
			continue
		}

		//format
		for _, article := range articles {
			// Skip duplicate articles by URL
			if seenURLs[article.URL] {
				continue
			}
			seenURLs[article.URL] = true

			newsCount++
			news := map[string]interface{}{
				"id":           article.ID,
				"symbol":       article.Symbol,
				"headline":     article.Headline,
				"url":          article.URL,
				"published_at": article.PublishedAt.Format("2006-01-02T15:04:05Z"),
				"source":       article.Source,
				"sentiment":    article.Sentiment,
				"catalyst":     article.CatalystType,
				"impact":       article.Impact,
			}
			allNews = append(allNews, news)
		}
	}

	response := map[string]interface{}{
		"news":            allNews,
		"count":           newsCount,
		"symbols_tracked": len(symbols),
	}

	WriteJSON(w, http.StatusOK, response)
}
