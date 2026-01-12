## MongelMaker

# Description 

This is a trading bot where it allows you to analyse stocks or crypto using the Alpaca API where it grabs the data from and showcases it into the Program with an analysis of each bar with different time frames.
With that data you can see if that Stock/Crypto is good to invest with indication showcasing if its bullish/bearish and its Volitality level, by using caluclations of formulas. Granted the formulas might be skewed
but it can be adjusted with the users theory of trading by adjusting the config or if you want to go the code and change it. 

It also has a Screener where it will scan all the Stocks or Crpyto that Alapaca has provided and with the defualt scores I've done it will go through all of it in chunks and you as the user can look it up to see if
meets your desire to invest your money towards it.

Its still a working progress as this project was just meant to be for my understanding with GO and Python(still need to do it). 

# Motivation
To be able to achieve understanding with how programming and the flow of indie developer, with its struggles and workarounds, espeically during the time of AI.
I've learnt a bit with this as I've always get stuck at the start before AI was introduced, but with that little boost it has made me motivated to try and see what I can make with my ideas in my head.

# Quick Start
Make sure you have an Alpaca account where there will be instructions for you to follow to grab the API Code of your account making it linked.

After you have your Alpaca account make sure you put it in a .env file (and if you dont have it just name a file .env), paste the details and you should be ready to start!

Alas, if you want to include Finnhubb (newscraping API) for the Stock you have selected, make sure you also include it to your .env with the API Token provided for you.

# Usage

-v is turn on verbose mode 

## Contributing

### Clone the repo

```bash
git clone https://github.com/fazecatGit/MongelMaker.git
cd MongelMaker

# Install Go 1.19+ if you haven't already
# Set up your environment variables (create a .env file with your API keys)
cp .env.example .env

Build the application
go build -o mongelmaker

Once on the Root of the file
go run . OR go run main.go
