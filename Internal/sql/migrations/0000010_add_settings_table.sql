-- +goose Up
-- Settings table for storing user configuration
CREATE TABLE IF NOT EXISTS settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(255) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL,
    setting_type VARCHAR(50) DEFAULT 'string', -- 'string', 'number', 'boolean', 'json'
    is_encrypted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(setting_key);

-- Insert default settings
INSERT INTO settings (setting_key, setting_value, setting_type, is_encrypted) 
VALUES 
  ('alpaca_api_key', '', 'string', TRUE),
  ('alpaca_api_secret', '', 'string', TRUE),
  ('finnhub_api_key', '', 'string', TRUE),
  ('auto_stop_loss', 'true', 'boolean', FALSE),
  ('auto_profit_taking', 'false', 'boolean', FALSE)
ON CONFLICT (setting_key) DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS idx_settings_key;
DROP TABLE IF EXISTS settings;
