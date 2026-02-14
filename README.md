# MogulMaker

MogulMaker is a real-time trading platform with a Go backend and Next.js frontend, built for automated and quantitative strategies.”
That helps visitors immediately know what this is without scrolling.

## Description 

A comprehensive trading platform with Go backend and Next.js frontend. Features real-time market analysis, automated trading strategies, risk management, technical indicators (RSI, ATR, Volume), news sentiment analysis, and a powerful stock screener. All controlled through an intuitive web interface.

## Motivation

To understand how Go works at its limit as a beginner and delve deep into its fundamentals. Also exploring how to integrate a modern web stack (Next.js, TypeScript, React) with a high-performance Go backend for real-time trading applications.

Future goals include developing AI agents with C++ or Python to create a robust analytical bot that suggests optimal trades and provides real-time sell indicators.

---

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+** ([Download](https://golang.org/dl/))
- **Node.js 20+** ([Download](https://nodejs.org/))
- **PostgreSQL 14+** ([Download](https://www.postgresql.org/download/))
- **Git** ([Download](https://git-scm.com/downloads))

### Required Accounts

- **Alpaca Trading Account** ([Sign up](https://alpaca.markets)) - Required to execute trades
- **Finnhub API Key** ([Sign up](https://finnhub.io)) - Required for news and sentiment analysis

---

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/FazecatGit/MogulMaker.git
cd MogulMaker
```

### 2. Set Up PostgreSQL Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create the database
CREATE DATABASE mogulmaker;

# Exit psql
\q
```

### 3. Configure Environment Variables

Create a `.env` file in the project root:

```bash
# Alpaca Trading API (Get from: https://alpaca.markets)
ALPACA_API_KEY=your_alpaca_key_here
ALPACA_API_SECRET=your_alpaca_secret_here

# Finnhub News API (Get from: https://finnhub.io)
FINNHUB_API_KEY=your_finnhub_key_here

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_postgres_password
DB_NAME=mogulmaker
DB_SSLMODE=disable

# Security - AES-256 encryption key (generate with: go run scripts/generate_key.go)
SETTINGS_ENCRYPTION_KEY=your_encryption_key_here
```


### 4. Install Dependencies

**Backend (Go):**
```bash
go mod tidy
```

**Frontend (Next.js):**
```bash
cd frontend
npm install
cd ..
```

**API Gateway:**
```bash
cd api-gateway
npm install
cd ..
```

---

## Running the Application

### Start Everything at Once

```powershell
# Windows PowerShell
./start-all.ps1
```


This starts:
- **Go Backend** on `http://localhost:8080`
- **API Gateway** on `http://localhost:3000`
- **Frontend** on `http://localhost:3001`


---

## First-Time Setup

### 1. Open the Web Interface

Navigate to: **http://localhost:3001**

### 2. Configure API Keys

1. Go to **Settings** page
2. Enter your API keys:
   - Alpaca API Key
   - Alpaca API Secret
   - Finnhub API Key (optional, for news)
3. Click **Save Settings**

** That's it!** Your keys are now:
- Encrypted with AES-256
- Stored securely in PostgreSQL
- Automatically loaded on startup

---

## Features

### Dashboard
- Portfolio overview
- Real-time P&L tracking
- Market statistics
- Recent trades summary

### Analyzer
- Multi-timeframe analysis (5m, 15m, 1h, 4h, 1d)
- Technical indicators (RSI, ATR, Volume)
- Price charts with indicators
- Signal strength analysis

### Scouter
- Filter by score, volatility, volume
- Multi-timeframe scanning
- Watchlist integration
- Skip list management

### News Feed
- Real-time market news (Finnhub)
- Sentiment analysis
- RSS feeds integration
- Catalyst detection

### Positions & Trades
- Active position monitoring
- Trade history
- P&L analysis
- Performance metrics

### Risk Management
- Portfolio risk metrics
- Drawdown monitoring
- Position sizing
- Risk alerts

### Watchlist
- Save symbols for monitoring
- Bulk scanning
- Score tracking
- Historical analysis

### Settings
- Secure API key management
- Trading preferences
- Auto stop-loss/profit taking
- Encrypted storage

---

### Testing

**Frontend:**
```bash
cd frontend
npm run test         
npm run test:e2e     
npm run lighthouse    
```

**Backend:**
```bash
go test ./...
```

### Building for Production

**Frontend:**
```bash
cd frontend
npm run build
npm run start
```

**Backend:**
```bash
cd cmd/api
go build -o mogulmaker.exe main.go
./mogulmaker.exe
```

---

## Troubleshooting

### "Database connection failed"
- Check PostgreSQL is running: `pg_ctl status`
- Verify database exists: `psql -U postgres -l`
- Check `.env` credentials

### "SETTINGS_ENCRYPTION_KEY not set"
- Add encryption key to `.env` file
- Generate new key: `go run scripts/generate_key.go`

### API keys not persisting
- Check backend logs for errors
- Verify encryption key is valid
- Re-enter keys through Settings UI

---

## Contributing

Solo project currently, but contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

