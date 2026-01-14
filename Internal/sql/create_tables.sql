-- MongelMaker Trading Bot Database Schema - UP ONLY

-- Historical OHLCV data table
CREATE TABLE historical_bars (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    timeframe VARCHAR(10) NOT NULL, -- '1Min', '5Min', '1Day', etc.
    timestamp TIMESTAMP NOT NULL,
    open_price DECIMAL(10, 4) NOT NULL,
    high_price DECIMAL(10, 4) NOT NULL,
    low_price DECIMAL(10, 4) NOT NULL,
    close_price DECIMAL(10, 4) NOT NULL,
    volume BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure no duplicate bars for same symbol/timeframe/timestamp
    UNIQUE(symbol, timeframe, timestamp)
);

-- Trading signals table
CREATE TABLE signals (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    signal_type VARCHAR(10) NOT NULL, -- 'BUY', 'SELL', 'HOLD'
    current_price DECIMAL(10, 4) NOT NULL,
    sma_value DECIMAL(10, 4),
    confidence DECIMAL(3, 2), -- 0.00 to 1.00
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    executed BOOLEAN DEFAULT FALSE
);

-- Executed trades table
CREATE TABLE trades (
    id SERIAL PRIMARY KEY,
    signal_id INTEGER REFERENCES signals(id),
    symbol VARCHAR(10) NOT NULL,
    side VARCHAR(4) NOT NULL, -- 'BUY' or 'SELL'
    quantity DECIMAL(10, 4) NOT NULL,
    price DECIMAL(10, 4) NOT NULL,
    total_value DECIMAL(12, 4) NOT NULL,
    commission DECIMAL(8, 4) DEFAULT 0,
    alpaca_order_id VARCHAR(50), -- Alpaca's order ID
    status VARCHAR(20) DEFAULT 'PENDING', -- 'PENDING', 'FILLED', 'CANCELLED'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    filled_at TIMESTAMP
);

-- Current positions table
CREATE TABLE positions (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) UNIQUE NOT NULL,
    quantity DECIMAL(10, 4) NOT NULL,
    avg_entry_price DECIMAL(10, 4) NOT NULL,
    current_price DECIMAL(10, 4),
    market_value DECIMAL(12, 4),
    unrealized_pnl DECIMAL(12, 4),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Portfolio performance history
CREATE TABLE portfolio_history (
    id SERIAL PRIMARY KEY,
    total_equity DECIMAL(12, 4) NOT NULL,
    cash_balance DECIMAL(12, 4) NOT NULL,
    positions_value DECIMAL(12, 4) NOT NULL,
    day_change DECIMAL(12, 4),
    total_return DECIMAL(12, 4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Risk management and monitoring tables

-- Risk events log (Phase 4.3)
CREATE TABLE risk_events (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- 'MAX_DAILY_LOSS_HIT', 'MAX_POSITIONS_HIT', 'POSITION_SIZE_EXCEEDED', etc.
    severity VARCHAR(20) NOT NULL, -- 'CRITICAL', 'WARNING', 'INFO'
    symbol VARCHAR(10),
    details TEXT,
    account_value DECIMAL(12, 4),
    daily_loss DECIMAL(12, 4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Trade records with completion details (Phase 4.3)
CREATE TABLE trade_records (
    id VARCHAR(50) PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    entry_time TIMESTAMP NOT NULL,
    exit_time TIMESTAMP,
    direction VARCHAR(10) NOT NULL, -- 'LONG', 'SHORT'
    entry_price DECIMAL(10, 4) NOT NULL,
    exit_price DECIMAL(10, 4),
    quantity BIGINT NOT NULL,
    entry_reason TEXT,
    exit_reason TEXT,
    realized_pnl DECIMAL(12, 4),
    realized_pnl_percent DECIMAL(5, 2),
    commission DECIMAL(8, 4) DEFAULT 0,
    duration INTERVAL,
    status VARCHAR(20) DEFAULT 'OPEN', -- 'COMPLETED', 'CANCELLED', 'PARTIAL_EXIT'
    tags VARCHAR(255)[], -- Array of tags
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Portfolio stats snapshot for analytics (Phase 4.3)
CREATE TABLE portfolio_stats_history (
    id SERIAL PRIMARY KEY,
    snapshot_date DATE NOT NULL,
    total_trades INTEGER,
    winning_trades INTEGER,
    losing_trades INTEGER,
    breakeven_trades INTEGER,
    win_rate DECIMAL(5, 2),
    total_profit DECIMAL(12, 4),
    total_loss DECIMAL(12, 4),
    net_profit DECIMAL(12, 4),
    profit_factor DECIMAL(5, 2),
    avg_trade_length INTERVAL,
    max_consecutive_wins INTEGER,
    max_consecutive_losses INTEGER,
    max_drawdown DECIMAL(12, 4),
    max_drawdown_percent DECIMAL(5, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(snapshot_date)
);

-- Daily risk summary for monitoring (Phase 4.3)
CREATE TABLE daily_risk_summary (
    id SERIAL PRIMARY KEY,
    summary_date DATE NOT NULL,
    account_balance DECIMAL(12, 4),
    open_positions INTEGER,
    daily_loss DECIMAL(12, 4),
    daily_loss_percent DECIMAL(5, 2),
    portfolio_risk DECIMAL(12, 4),
    portfolio_risk_percent DECIMAL(5, 2),
    health_status VARCHAR(50), -- 'HEALTHY', 'WARNING', 'CRITICAL'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(summary_date)
);

-- Indexes for better query performance
CREATE INDEX idx_historical_bars_symbol_timeframe ON historical_bars(symbol, timeframe);
CREATE INDEX idx_historical_bars_timestamp ON historical_bars(timestamp);
CREATE INDEX idx_signals_symbol_created ON signals(symbol, created_at);
CREATE INDEX idx_trades_symbol_created ON trades(symbol, created_at);
CREATE INDEX idx_positions_symbol ON positions(symbol);
CREATE INDEX idx_risk_events_timestamp ON risk_events(timestamp);
CREATE INDEX idx_risk_events_symbol ON risk_events(symbol);
CREATE INDEX idx_trade_records_symbol ON trade_records(symbol);
CREATE INDEX idx_trade_records_entry_time ON trade_records(entry_time);
CREATE INDEX idx_portfolio_stats_history_date ON portfolio_stats_history(snapshot_date);
CREATE INDEX idx_daily_risk_summary_date ON daily_risk_summary(summary_date);