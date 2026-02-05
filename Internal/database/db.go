package datafeed

import (
	"database/sql"
	"fmt"
	"os"

	database "github.com/fazecat/mogulmaker/Internal/database/sqlc"
	_ "github.com/lib/pq"
)

var Queries *database.Queries
var DB *sql.DB

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func InitDatabase() error {
	config := DatabaseConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "postgres"),
		Password: os.Getenv("DB_PASSWORD"), // Required - no default
		DBName:   getEnvOrDefault("DB_NAME", "mongelmaker"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	Queries = database.New(DB)

	fmt.Println("Database connected successfully!")
	return nil
}

// initializeSchema creates watchlist tables if they don't exist
func initializeSchema() error {
	schemaSQL := `
	CREATE TABLE IF NOT EXISTS watchlist (
		id SERIAL PRIMARY KEY,
		symbol TEXT NOT NULL UNIQUE,
		asset_type TEXT NOT NULL,
		score REAL NOT NULL,
		reason TEXT,
		added_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active'
	);

	CREATE TABLE IF NOT EXISTS watchlist_history (
		id SERIAL PRIMARY KEY,
		watchlist_id INTEGER NOT NULL,
		old_score REAL,
		new_score REAL NOT NULL,
		analysis_data TEXT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(watchlist_id) REFERENCES watchlist(id)
	);

	CREATE TABLE IF NOT EXISTS skip_backlog (
		id SERIAL PRIMARY KEY,
		symbol TEXT NOT NULL UNIQUE,
		asset_type TEXT NOT NULL,
		reason TEXT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		recheck_after TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_watchlist_symbol ON watchlist(symbol);
	CREATE INDEX IF NOT EXISTS idx_watchlist_status ON watchlist(status);
	`

	_, err := DB.Exec(schemaSQL)
	return err
}

func CloseDatabase() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return DB.Ping()
}
