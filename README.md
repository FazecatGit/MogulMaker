## MogulMaker

# Description 

A trading bot analyzing stocks/crypto via Alpaca API. Provides bullish/bearish signals, volatility scores across timeframes, and a screener for assets. Interactive CLI menu with risk management and trade monitoring.

# Motivation
To undestand how GO works at its limit as a beginner and delving deep into its fundamentals. Whilst also trying out Machine learning with Python to see when trades are bad and adjust the scoring that I have to fit the
threshold of profit.

I would love to develop it further with C++ knowledge of AI Agents creating a robust analytical bot that'll suggest what trades to put in and give you indicators of when to sell off. But that'll require me developing it as
another project/repo.

# Quick Start // Installation

git clone https://github.com/FazecatGit/MongelMaker.git
cd MongelMaker
cp .env.example .env
# Add your APCA_API_KEY_ID, APCA_API_SECRET_KEY, and FINNHUB_API_KEY to .env
# Also create a Postgre profile to store your data
go mod tidy
go build -o mongelmaker

Populate your .env file:

text
APCA_API_KEY_ID=your_alpaca_key
APCA_API_SECRET_KEY=your_alpaca_secret
FINNHUB_API_KEY=your_finnhub_key (optional)

# Usage

go run . / go run main.go; all the features will be there from the display menu on the CLI
I will implement usage utlizing the CLI commands later when I get the chance.

## Contributing

Solo Project currently but will accept contributions later.
