package utils

import (
	"testing"
	"time"

	"github.com/fazecat/mogulmaker/Internal/utils/config"
)

// Create config once for all tests
var testCfg = &config.Config{
	Global: struct {
		MarketHours struct {
			RegularOpen    string `yaml:"regular_open"`
			RegularClose   string `yaml:"regular_close"`
			PremarketOpen  string `yaml:"premarket_open"`
			AfterhourClose string `yaml:"afterhours_close"`
			Timezone       string `yaml:"timezone"`
		} `yaml:"market_hours"`
		LiquidityMinimumUSD int `yaml:"liquidity_minimum_usd"`
	}{
		MarketHours: struct {
			RegularOpen    string `yaml:"regular_open"`
			RegularClose   string `yaml:"regular_close"`
			PremarketOpen  string `yaml:"premarket_open"`
			AfterhourClose string `yaml:"afterhours_close"`
			Timezone       string `yaml:"timezone"`
		}{
			PremarketOpen:  "04:00",
			RegularOpen:    "09:30",
			RegularClose:   "16:00",
			AfterhourClose: "20:00",
			Timezone:       "EST",
		},
	},
}

func TestMondayRegularHours(t *testing.T) {
	// Create time in EST timezone directly
	estLoc, _ := time.LoadLocation("America/New_York")
	monday := time.Date(2023, 3, 6, 10, 0, 0, 0, estLoc)
	result, isOpen := CheckMarketStatus(monday, testCfg)
	if result != "REGULAR" || !isOpen {
		t.Errorf("Expected REGULAR/true, got %s/%v", result, isOpen)
	}
}
func TestSaturdayClosed(t *testing.T) {
	saturday := time.Date(2023, 3, 4, 10, 0, 0, 0, time.UTC)
	result, isOpen := CheckMarketStatus(saturday, testCfg)
	if result != "CLOSED" || isOpen {
		t.Errorf("Expected CLOSED/false, got %s/%v", result, isOpen)
	}
}
