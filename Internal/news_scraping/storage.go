package newsscraping

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/fazecat/mogulmaker/Internal/database/sqlc"
)

type NewsStorage struct {
	queries *db.Queries
}

func NewNewsStorage(queries *db.Queries) *NewsStorage {
	return &NewsStorage{queries: queries}
}

func (ns *NewsStorage) SaveArticle(ctx context.Context, article NewsArticle) error {
	err := ns.queries.SaveNewsArticle(ctx, db.SaveNewsArticleParams{
		Symbol:      article.Symbol,
		Headline:    article.Headline,
		Url:         article.URL,
		PublishedAt: article.PublishedAt,
		Source:      sql.NullString{String: article.Source, Valid: article.Source != ""},
		Sentiment:   sql.NullString{String: string(article.Sentiment), Valid: article.Sentiment != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to save article: %w", err)
	}
	return nil
}

func (ns *NewsStorage) GetLatestNews(ctx context.Context, symbol string, limit int32) ([]NewsArticle, error) {
	rows, err := ns.queries.GetLatestNews(ctx, db.GetLatestNewsParams{
		Symbol: symbol,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}

	var articles []NewsArticle
	for _, row := range rows {
		articles = append(articles, NewsArticle{
			ID:          int64(row.ID),
			Symbol:      row.Symbol,
			Headline:    row.Headline,
			URL:         row.Url,
			PublishedAt: row.PublishedAt,
			Source:      row.Source.String,
			Sentiment:   SentimentScore(row.Sentiment.String),
			CreatedAt:   row.CreatedAt.Time,
		})
	}
	return articles, nil
}

func (ns *NewsStorage) GetNewsForScreener(ctx context.Context, symbols []string) ([]NewsArticle, error) {
	rows, err := ns.queries.GetNewsForScreener(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch screener news: %w", err)
	}

	var articles []NewsArticle
	for _, row := range rows {
		articles = append(articles, NewsArticle{
			ID:          int64(row.ID),
			Symbol:      row.Symbol,
			Headline:    row.Headline,
			URL:         row.Url,
			PublishedAt: row.PublishedAt,
			Source:      row.Source.String,
			Sentiment:   SentimentScore(row.Sentiment.String),
			CreatedAt:   row.CreatedAt.Time,
		})
	}
	return articles, nil
}
