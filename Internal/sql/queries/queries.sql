-- name: GetClosingPrices :many
SELECT close_price, timestamp
FROM historical_bars
WHERE symbol = $1 
  AND timeframe = $2
ORDER BY timestamp ASC
LIMIT $3;

-- name: SaveRSI :exec
INSERT INTO rsi_calculation (symbol, calculation_timestamp, rsi_value)
VALUES ($1, $2, $3)
ON CONFLICT (symbol, calculation_timestamp)
DO UPDATE SET rsi_value = EXCLUDED.rsi_value;

-- name: GetLatestRSI :one
SELECT rsi_value, calculation_timestamp
FROM rsi_calculation
WHERE symbol = $1
ORDER BY calculation_timestamp DESC
LIMIT 1;

-- name: GetATRPrices :many
SELECT high_price, low_price, close_price, timestamp
FROM historical_bars
WHERE symbol = $1 
  AND timeframe = $2
ORDER BY timestamp ASC
LIMIT $3;

-- name: GetATR :one
SELECT atr_value, calculation_timestamp
FROM atr_calculation
WHERE symbol = $1
ORDER BY calculation_timestamp DESC
LIMIT 1;

-- name: SaveATR :exec
INSERT INTO atr_calculation (symbol, calculation_timestamp, atr_value)
VALUES ($1, $2, $3)
ON CONFLICT (symbol, calculation_timestamp)
DO UPDATE SET atr_value = EXCLUDED.atr_value;

-- name: GetRSIForDateRange :many
SELECT calculation_timestamp, rsi_value
FROM rsi_calculation
WHERE symbol = $1
ORDER BY calculation_timestamp DESC
LIMIT $2;

-- name: GetATRForDateRange :many
SELECT calculation_timestamp, atr_value
FROM atr_calculation
WHERE symbol = $1
ORDER BY calculation_timestamp DESC
LIMIT $2;

-- name: GetRSIByTimestampRange :many
SELECT calculation_timestamp, rsi_value
FROM rsi_calculation
WHERE symbol = $1
  AND calculation_timestamp >= $2
  AND calculation_timestamp <= $3
ORDER BY calculation_timestamp ASC;

-- name: GetATRByTimestampRange :many
SELECT calculation_timestamp, atr_value
FROM atr_calculation
WHERE symbol = $1
  AND calculation_timestamp >= $2
  AND calculation_timestamp <= $3
ORDER BY calculation_timestamp ASC;

-- name: SaveNewsArticle :exec
INSERT INTO news_articles (symbol, headline, url, published_at, source, sentiment)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (url) DO NOTHING;

-- name: GetLatestNews :many
SELECT id, symbol, headline, url, published_at, source, sentiment, created_at
FROM news_articles
WHERE symbol = $1
ORDER BY published_at DESC
LIMIT $2;

-- name: GetNewsForScreener :many
SELECT id, symbol, headline, url, published_at, source, sentiment, created_at
FROM news_articles
WHERE symbol = ANY($1::text[])
ORDER BY published_at DESC;

-- name: GetNewsBySymbol :many
SELECT id, symbol, headline, url, published_at, source, sentiment, created_at
FROM news_articles
WHERE symbol = $1
AND published_at > NOW() - INTERVAL '7 days'
ORDER BY published_at DESC;

-- name: GetWhaleEventsBySymbol :many
SELECT * FROM whale_events
WHERE symbol = $1 AND timestamp > NOW() - INTERVAL '7 days'
ORDER BY timestamp DESC
LIMIT $2;

-- name: GetHighConvictionWhales :many
SELECT * FROM whale_events
WHERE symbol = $1 AND conviction = 'HIGH'
ORDER BY z_score DESC
LIMIT 10;

-- name: CreateWhaleEvent :exec
INSERT INTO whale_events (
    symbol, timestamp, direction, volume, z_score, close_price, price_change, conviction
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: AddToWatchlist :one
-- Add a new candidate to watchlist and return the ID
INSERT INTO watchlist (symbol, asset_type, score, reason, added_date, last_updated, status)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'active')
RETURNING id;

-- name: GetWatchlist :many
-- Get all active watchlist items, ordered by score
SELECT id, symbol, asset_type, score, reason, added_date, last_updated
FROM watchlist
WHERE status = 'active'
ORDER BY score DESC;

-- name: GetWatchlistBySymbol :one
-- Get a watchlist item by symbol
SELECT id, symbol, asset_type, score, reason, added_date, last_updated
FROM watchlist
WHERE symbol = $1 AND status = 'active';

-- name: UpdateWatchlistScore :exec
-- Update score and add history entry
UPDATE watchlist
SET score = $1, last_updated = CURRENT_TIMESTAMP
WHERE symbol = $2;

-- name: AddWatchlistHistory :exec
-- Log score change with full analysis data (as JSON)
INSERT INTO watchlist_history (watchlist_id, old_score, new_score, analysis_data, timestamp)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP);

-- name: ArchiveOldWatchlist :exec
-- Archive symbols with unchanged score for 30+ days
UPDATE watchlist
SET status = 'archived'
WHERE id IN (
  SELECT w.id FROM watchlist w
  WHERE w.status = 'active'
  AND datetime(w.last_updated) < datetime('now', '-30 days')
);

-- name: SkipSymbol :exec
-- Add to skip backlog (recheck in 30 days)
INSERT INTO skip_backlog (symbol, asset_type, reason, timestamp, recheck_after)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP, datetime('now', '+30 days'));

-- name: GetRecheckableSymbols :many
-- Get symbols from skip backlog that are ready for reconsideration
SELECT symbol, asset_type, reason FROM skip_backlog
WHERE recheck_after <= CURRENT_TIMESTAMP;

-- name: RemoveFromSkipBacklog :exec
-- Remove symbol from skip backlog after rechecking
DELETE FROM skip_backlog WHERE symbol = $1;

-- Scan Log Queries

-- name: GetScanLog :one
-- Get the latest scan log entry for a profile
SELECT id, profile_name, last_scan_timestamp, next_scan_due, symbols_scanned
FROM scan_log
WHERE profile_name = $1;

-- name: UpsertScanLog :exec
-- Insert or update scan log entry
INSERT INTO scan_log (profile_name, last_scan_timestamp, next_scan_due, symbols_scanned)
VALUES ($1, $2, $3, $4)
ON CONFLICT (profile_name) DO UPDATE SET
    last_scan_timestamp = EXCLUDED.last_scan_timestamp,
    next_scan_due = EXCLUDED.next_scan_due,
    symbols_scanned = EXCLUDED.symbols_scanned,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetAllScanLogs :many
-- Get all scan log entries
SELECT id, profile_name, last_scan_timestamp, next_scan_due, symbols_scanned
FROM scan_log
ORDER BY profile_name;

-- name: AddToScoutSkipList :exec
INSERT INTO scout_skip_list (symbol, profile_name, asset_type, reason, recheck_after)
VALUES ($1, $2, $3, $4, NOW() + INTERVAL '2 days');

-- name: IsSymbolSkipped :one
SELECT COUNT(*) > 0 as is_skipped
FROM scout_skip_list
WHERE symbol = $1 
  AND profile_name = $2 
  AND recheck_after > NOW();

-- name: LogTrade :exec
INSERT INTO trades (symbol, side, quantity, price, total_value, alpaca_order_id, status, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW());

-- name: GetTradeHistory :many
SELECT id, symbol, side, quantity, price, total_value, alpaca_order_id, status, created_at, filled_at
FROM trades
WHERE symbol = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetAllOpenTrades :many
SELECT id, symbol, side, quantity, price, total_value, alpaca_order_id, status, created_at
FROM trades
WHERE status = 'PENDING' OR status = 'FILLED'
ORDER BY created_at DESC;

-- name: UpdateTradeStatus :exec
UPDATE trades
SET status = $1, filled_at = NOW()
WHERE alpaca_order_id = $2;