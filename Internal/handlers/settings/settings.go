package settings

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
)

type Handler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

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
			if err := SetSetting(h.DB, "max_daily_loss", payload.Trading.MaxDailyLoss); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to save max_daily_loss setting")
				return
			}
		}
		if payload.Trading.MaxPositionRisk > 0 {
			if err := SetSetting(h.DB, "max_position_risk", payload.Trading.MaxPositionRisk); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to save max_position_risk setting")
				return
			}
		}
		if payload.Trading.MaxOpenPositions > 0 {
			if err := SetSetting(h.DB, "max_open_positions", float64(payload.Trading.MaxOpenPositions)); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to save max_open_positions setting")
				return
			}
		}
		if err := SetSetting(h.DB, "trading_hours_only", payload.Trading.TradingHoursOnly); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save trading_hours_only setting")
			return
		}
		if err := SetSetting(h.DB, "auto_stop_loss", payload.Trading.AutoStopLoss); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save auto_stop_loss setting")
			return
		}
		if err := SetSetting(h.DB, "auto_profit_taking", payload.Trading.AutoProfitTaking); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save auto_profit_taking setting")
			return
		}
	}

	// Update notification settings
	if payload.Notifications != nil {
		if err := SetSetting(h.DB, "email_alerts", payload.Notifications.EmailAlerts); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save email_alerts setting")
			return
		}
		if err := SetSetting(h.DB, "trade_execution_notifications", payload.Notifications.TradeExecutionNotifications); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save trade_execution_notifications setting")
			return
		}
		if err := SetSetting(h.DB, "risk_alerts", payload.Notifications.RiskAlerts); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save risk_alerts setting")
			return
		}
		if err := SetSetting(h.DB, "daily_summary", payload.Notifications.DailySummary); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save daily_summary setting")
			return
		}
		if err := SetSetting(h.DB, "news_alerts", payload.Notifications.NewsAlerts); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save news_alerts setting")
			return
		}
	}

	// Update API settings
	if payload.API != nil {
		if payload.API.AlpacaKey != "" {
			if err := SetSetting(h.DB, "alpaca_api_key", payload.API.AlpacaKey); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to save Alpaca API key")
				return
			}
			os.Setenv("ALPACA_API_KEY", payload.API.AlpacaKey)
		}
		if payload.API.AlpacaSecret != "" {
			if err := SetSetting(h.DB, "alpaca_api_secret", payload.API.AlpacaSecret); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to save Alpaca API secret")
				return
			}
			os.Setenv("ALPACA_API_SECRET", payload.API.AlpacaSecret)
		}
		if payload.API.FinnhubKey != "" {
			if err := SetSetting(h.DB, "finnhub_api_key", payload.API.FinnhubKey); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to save Finnhub API key")
				return
			}
			os.Setenv("FINNHUB_API_KEY", payload.API.FinnhubKey)
		}
	}

	response := SettingsResponse{
		Message: "Settings updated successfully",
	}

	writeJSON(w, http.StatusOK, response)
}
