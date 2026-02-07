package settings

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
)

// Handler handles all settings-related operations
type Handler struct {
	DB *sql.DB
}

// NewHandler creates a new settings handler
func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error JSON response
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// HandleGetSettings returns all settings
func (h *Handler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := SettingsResponse{
		Trading: TradeSettings{
			MaxDailyLoss:     GetSetting(h.DB, "max_daily_loss", 5000.0).(float64),
			MaxPositionRisk:  GetSetting(h.DB, "max_position_risk", 1000.0).(float64),
			MaxOpenPositions: int(GetSetting(h.DB, "max_open_positions", 10.0).(float64)),
			TradingHoursOnly: GetSetting(h.DB, "trading_hours_only", true).(bool),
			AutoStopLoss:     GetSetting(h.DB, "auto_stop_loss", true).(bool),
			AutoProfitTaking: GetSetting(h.DB, "auto_profit_taking", false).(bool),
		},
		Notifications: NotificationSettings{
			EmailAlerts:                 GetSetting(h.DB, "email_alerts", true).(bool),
			TradeExecutionNotifications: GetSetting(h.DB, "trade_execution_notifications", true).(bool),
			RiskAlerts:                  GetSetting(h.DB, "risk_alerts", true).(bool),
			DailySummary:                GetSetting(h.DB, "daily_summary", false).(bool),
			NewsAlerts:                  GetSetting(h.DB, "news_alerts", true).(bool),
		},
		API: map[string]string{
			"alpacaKeyMasked":    MaskSensitiveValue(GetSetting(h.DB, "alpaca_api_key", "").(string)),
			"alpacaSecretMasked": MaskSensitiveValue(GetSetting(h.DB, "alpaca_api_secret", "").(string)),
			"finnhubKeyMasked":   MaskSensitiveValue(GetSetting(h.DB, "finnhub_api_key", "").(string)),
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleUpdateSettings updates user settings
func (h *Handler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var payload SettingsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update trading settings
	if payload.Trading != nil {
		if payload.Trading.MaxDailyLoss > 0 {
			SetSetting(h.DB, "max_daily_loss", payload.Trading.MaxDailyLoss)
		}
		if payload.Trading.MaxPositionRisk > 0 {
			SetSetting(h.DB, "max_position_risk", payload.Trading.MaxPositionRisk)
		}
		if payload.Trading.MaxOpenPositions > 0 {
			SetSetting(h.DB, "max_open_positions", float64(payload.Trading.MaxOpenPositions))
		}
		SetSetting(h.DB, "trading_hours_only", payload.Trading.TradingHoursOnly)
		SetSetting(h.DB, "auto_stop_loss", payload.Trading.AutoStopLoss)
		SetSetting(h.DB, "auto_profit_taking", payload.Trading.AutoProfitTaking)
	}

	// Update notification settings
	if payload.Notifications != nil {
		SetSetting(h.DB, "email_alerts", payload.Notifications.EmailAlerts)
		SetSetting(h.DB, "trade_execution_notifications", payload.Notifications.TradeExecutionNotifications)
		SetSetting(h.DB, "risk_alerts", payload.Notifications.RiskAlerts)
		SetSetting(h.DB, "daily_summary", payload.Notifications.DailySummary)
		SetSetting(h.DB, "news_alerts", payload.Notifications.NewsAlerts)
	}

	// Update API settings
	if payload.API != nil {
		if payload.API.AlpacaKey != "" {
			SetSetting(h.DB, "alpaca_api_key", payload.API.AlpacaKey)
			os.Setenv("ALPACA_API_KEY", payload.API.AlpacaKey)
		}
		if payload.API.AlpacaSecret != "" {
			SetSetting(h.DB, "alpaca_api_secret", payload.API.AlpacaSecret)
			os.Setenv("ALPACA_API_SECRET", payload.API.AlpacaSecret)
		}
		if payload.API.FinnhubKey != "" {
			SetSetting(h.DB, "finnhub_api_key", payload.API.FinnhubKey)
			os.Setenv("FINNHUB_API_KEY", payload.API.FinnhubKey)
		}
	}

	response := SettingsResponse{
		Message: "Settings updated successfully",
	}

	writeJSON(w, http.StatusOK, response)
}
