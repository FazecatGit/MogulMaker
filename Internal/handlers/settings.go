package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
)

type SettingsPayload struct {
	Trading struct {
		MaxDailyLoss     float64 `json:"maxDailyLoss"`
		MaxPositionRisk  float64 `json:"maxPositionRisk"`
		MaxOpenPositions int     `json:"maxOpenPositions"`
		TradingHoursOnly bool    `json:"tradingHoursOnly"`
		AutoStopLoss     bool    `json:"autoStopLoss"`
		AutoProfitTaking bool    `json:"autoProfitTaking"`
	} `json:"trading,omitempty"`
	Notifications struct {
		EmailAlerts                 bool `json:"emailAlerts"`
		TradeExecutionNotifications bool `json:"tradeExecutionNotifications"`
		RiskAlerts                  bool `json:"riskAlerts"`
		DailySummary                bool `json:"dailySummary"`
		NewsAlerts                  bool `json:"newsAlerts"`
	} `json:"notifications,omitempty"`
	API struct {
		AlpacaKey    string `json:"alpacaKey"`
		AlpacaSecret string `json:"alpacaSecret"`
		FinnhubKey   string `json:"finnhubKey"`
	} `json:"api,omitempty"`
}

type SettingsResponse struct {
	Trading struct {
		MaxDailyLoss     float64 `json:"maxDailyLoss"`
		MaxPositionRisk  float64 `json:"maxPositionRisk"`
		MaxOpenPositions int     `json:"maxOpenPositions"`
		TradingHoursOnly bool    `json:"tradingHoursOnly"`
		AutoStopLoss     bool    `json:"autoStopLoss"`
		AutoProfitTaking bool    `json:"autoProfitTaking"`
	} `json:"trading"`
	Notifications struct {
		EmailAlerts                 bool `json:"emailAlerts"`
		TradeExecutionNotifications bool `json:"tradeExecutionNotifications"`
		RiskAlerts                  bool `json:"riskAlerts"`
		DailySummary                bool `json:"dailySummary"`
		NewsAlerts                  bool `json:"newsAlerts"`
	} `json:"notifications"`
	API struct {
		AlpacaKeyMasked    string `json:"alpacaKeyMasked"`
		AlpacaSecretMasked string `json:"alpacaSecretMasked"`
		FinnhubKeyMasked   string `json:"finnhubKeyMasked"`
	} `json:"api"`
	Message string `json:"message,omitempty"`
}

// GetSettings retrieves all user settings
func GetSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := SettingsResponse{}

		// Fetch trading settings
		response.Trading.MaxDailyLoss = getSetting(db, "max_daily_loss", 5000.0).(float64)
		response.Trading.MaxPositionRisk = getSetting(db, "max_position_risk", 1000.0).(float64)
		response.Trading.MaxOpenPositions = getSetting(db, "max_open_positions", 10).(int)
		response.Trading.TradingHoursOnly = getSetting(db, "trading_hours_only", true).(bool)
		response.Trading.AutoStopLoss = getSetting(db, "auto_stop_loss", true).(bool)
		response.Trading.AutoProfitTaking = getSetting(db, "auto_profit_taking", false).(bool)

		// Fetch notification settings
		response.Notifications.EmailAlerts = getSetting(db, "email_alerts", true).(bool)
		response.Notifications.TradeExecutionNotifications = getSetting(db, "trade_execution_notifications", true).(bool)
		response.Notifications.RiskAlerts = getSetting(db, "risk_alerts", true).(bool)
		response.Notifications.DailySummary = getSetting(db, "daily_summary", false).(bool)
		response.Notifications.NewsAlerts = getSetting(db, "news_alerts", true).(bool)

		// Fetch API settings (masked)
		response.API.AlpacaKeyMasked = maskSensitiveValue(getSetting(db, "alpaca_api_key", "").(string))
		response.API.AlpacaSecretMasked = maskSensitiveValue(getSetting(db, "alpaca_api_secret", "").(string))
		response.API.FinnhubKeyMasked = maskSensitiveValue(getSetting(db, "finnhub_api_key", "").(string))

		json.NewEncoder(w).Encode(response)
	}
}

// GetSetting retrieves a specific setting
func GetSetting(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		key := r.URL.Query().Get("key")

		if key == "" {
			http.Error(w, "Setting key is required", http.StatusBadRequest)
			return
		}

		value := getSetting(db, key, nil)
		if value == nil {
			http.Error(w, "Setting not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"key":   key,
			"value": value,
		})
	}
}

// UpdateSettings updates user settings
func UpdateSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var payload SettingsPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Update trading settings
		if payload.Trading.MaxDailyLoss > 0 {
			setSetting(db, "max_daily_loss", payload.Trading.MaxDailyLoss)
		}
		if payload.Trading.MaxPositionRisk > 0 {
			setSetting(db, "max_position_risk", payload.Trading.MaxPositionRisk)
		}
		if payload.Trading.MaxOpenPositions > 0 {
			setSetting(db, "max_open_positions", payload.Trading.MaxOpenPositions)
		}
		setSetting(db, "trading_hours_only", payload.Trading.TradingHoursOnly)
		setSetting(db, "auto_stop_loss", payload.Trading.AutoStopLoss)
		setSetting(db, "auto_profit_taking", payload.Trading.AutoProfitTaking)

		// Update notification settings
		setSetting(db, "email_alerts", payload.Notifications.EmailAlerts)
		setSetting(db, "trade_execution_notifications", payload.Notifications.TradeExecutionNotifications)
		setSetting(db, "risk_alerts", payload.Notifications.RiskAlerts)
		setSetting(db, "daily_summary", payload.Notifications.DailySummary)
		setSetting(db, "news_alerts", payload.Notifications.NewsAlerts)

		// Update API settings (only update if provided)
		if payload.API.AlpacaKey != "" {
			setSetting(db, "alpaca_api_key", payload.API.AlpacaKey)
			os.Setenv("ALPACA_API_KEY", payload.API.AlpacaKey)
		}
		if payload.API.AlpacaSecret != "" {
			setSetting(db, "alpaca_api_secret", payload.API.AlpacaSecret)
			os.Setenv("ALPACA_API_SECRET", payload.API.AlpacaSecret)
		}
		if payload.API.FinnhubKey != "" {
			setSetting(db, "finnhub_api_key", payload.API.FinnhubKey)
			os.Setenv("FINNHUB_API_KEY", payload.API.FinnhubKey)
		}

		response := SettingsResponse{
			Message: "Settings updated successfully",
		}
		json.NewEncoder(w).Encode(response)
	}
}

// ValidateCredentials validates API credentials
func ValidateCredentials(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var payload struct {
			AlpacaKey    string `json:"alpacaKey"`
			AlpacaSecret string `json:"alpacaSecret"`
			FinnhubKey   string `json:"finnhubKey"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// TODO: Add actual validation logic for Alpaca and Finnhub credentials
		// For now, just check if they're not empty
		valid := payload.AlpacaKey != "" && payload.AlpacaSecret != ""

		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   valid,
			"message": "Credentials validation check completed",
		})
	}
}

// Helper functions

func getSetting(db *sql.DB, key string, defaultValue interface{}) interface{} {
	var value string
	var settingType string

	err := db.QueryRow(
		"SELECT setting_value, setting_type FROM settings WHERE setting_key = $1",
		key,
	).Scan(&value, &settingType)

	if err != nil {
		if err == sql.ErrNoRows {
			return defaultValue
		}
		return defaultValue
	}

	// Convert to appropriate type
	switch settingType {
	case "number":
		var floatVal float64
		json.Unmarshal([]byte(value), &floatVal)
		return floatVal
	case "boolean":
		var boolVal bool
		json.Unmarshal([]byte(value), &boolVal)
		return boolVal
	case "json":
		var jsonVal interface{}
		json.Unmarshal([]byte(value), &jsonVal)
		return jsonVal
	default:
		return value
	}
}

func setSetting(db *sql.DB, key string, value interface{}) error {
	var valueStr string
	settingType := "string"

	switch v := value.(type) {
	case bool:
		settingType = "boolean"
		if v {
			valueStr = "true"
		} else {
			valueStr = "false"
		}
	case float64:
		settingType = "number"
		valueStr = json.Number(string(rune(int(v)))).String()
		// Better approach for float64
		bytes, _ := json.Marshal(v)
		valueStr = string(bytes)
	case int:
		settingType = "number"
		bytes, _ := json.Marshal(v)
		valueStr = string(bytes)
	default:
		valueStr = value.(string)
	}

	_, err := db.Exec(
		"UPDATE settings SET setting_value = $1, setting_type = $2, updated_at = CURRENT_TIMESTAMP WHERE setting_key = $3",
		valueStr, settingType, key,
	)
	return err
}

func maskSensitiveValue(value string) string {
	if value == "" {
		return "Not set"
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:4] + "****...****"
}
