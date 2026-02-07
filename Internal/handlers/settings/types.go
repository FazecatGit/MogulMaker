package settings

type SettingsPayload struct {
	Trading       *TradeSettings        `json:"trading,omitempty"`
	Notifications *NotificationSettings `json:"notifications,omitempty"`
	API           *APISettings          `json:"api,omitempty"`
}

type TradeSettings struct {
	MaxDailyLoss     float64 `json:"maxDailyLoss"`
	MaxPositionRisk  float64 `json:"maxPositionRisk"`
	MaxOpenPositions int     `json:"maxOpenPositions"`
	TradingHoursOnly bool    `json:"tradingHoursOnly"`
	AutoStopLoss     bool    `json:"autoStopLoss"`
	AutoProfitTaking bool    `json:"autoProfitTaking"`
}

type NotificationSettings struct {
	EmailAlerts                 bool `json:"emailAlerts"`
	TradeExecutionNotifications bool `json:"tradeExecutionNotifications"`
	RiskAlerts                  bool `json:"riskAlerts"`
	DailySummary                bool `json:"dailySummary"`
	NewsAlerts                  bool `json:"newsAlerts"`
}

type APISettings struct {
	AlpacaKey    string `json:"alpacaKey"`
	AlpacaSecret string `json:"alpacaSecret"`
	FinnhubKey   string `json:"finnhubKey"`
}

type SettingsResponse struct {
	Trading       TradeSettings        `json:"trading"`
	Notifications NotificationSettings `json:"notifications"`
	API           map[string]string    `json:"api"`
	Message       string               `json:"message,omitempty"`
}
