package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/fazecat/mogulmaker/Internal/utils/config"
)

func CheckMarketStatus(t time.Time, cfg *config.Config) (status string, isOpen bool) {
	timeInEST, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "CLOSED", false
	}
	estTime := t.In(timeInEST)

	// Get total minutes since midnight
	hour, min, _ := estTime.Clock()
	totalMinutes := hour*60 + min

	weekday := estTime.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return "CLOSED", false
	}
	premktOpen, err := parseTimeToMinutes(cfg.Global.MarketHours.PremarketOpen)
	if err != nil {
		return "CLOSED", false
	}
	regularOpen, err := parseTimeToMinutes(cfg.Global.MarketHours.RegularOpen)
	if err != nil {
		return "CLOSED", false
	}
	regularClose, err := parseTimeToMinutes(cfg.Global.MarketHours.RegularClose)
	if err != nil {
		return "CLOSED", false
	}
	afterClose, err := parseTimeToMinutes(cfg.Global.MarketHours.AfterhourClose)
	if err != nil {
		return "CLOSED", false
	}

	// Determine market status
	if totalMinutes >= premktOpen && totalMinutes < regularOpen {
		return "PREMARKET", true
	} else if totalMinutes >= regularOpen && totalMinutes <= regularClose {
		return "REGULAR", true
	} else if totalMinutes > regularClose && totalMinutes <= afterClose {
		return "AFTERHOURS", true
	}

	return "CLOSED", false
}

func parseTimeToMinutes(timeStr string) (int, error) {
	if timeStr == "" {
		return -1, errors.New("invalid time string")
	}
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return -1, errors.New("invalid time format")
	}
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	return hour*60 + minute, nil
}
