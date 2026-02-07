package settings

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
)

// GetSetting retrieves a setting from the database with type conversion
func GetSetting(db *sql.DB, key string, defaultValue interface{}) interface{} {
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
		if value == "true" {
			boolVal = true
		} else {
			boolVal = false
		}
		return boolVal
	case "json":
		var jsonVal interface{}
		json.Unmarshal([]byte(value), &jsonVal)
		return jsonVal
	default:
		return value
	}
}

// SetSetting updates a setting in the database
func SetSetting(db *sql.DB, key string, value interface{}) error {
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

// LoadSettingsFromDatabase loads API keys from database and sets environment variables
func LoadSettingsFromDatabase(db *sql.DB) {
	// Load API keys from database and set as environment variables
	alpacaKey := GetSetting(db, "alpaca_api_key", "").(string)
	if alpacaKey != "" {
		os.Setenv("ALPACA_API_KEY", alpacaKey)
		log.Println("Loaded ALPACA_API_KEY from database")
	}

	alpacaSecret := GetSetting(db, "alpaca_api_secret", "").(string)
	if alpacaSecret != "" {
		os.Setenv("ALPACA_API_SECRET", alpacaSecret)
		log.Println("Loaded ALPACA_API_SECRET from database")
	}

	finnhubKey := GetSetting(db, "finnhub_api_key", "").(string)
	if finnhubKey != "" {
		os.Setenv("FINNHUB_API_KEY", finnhubKey)
		log.Println("Loaded FINNHUB_API_KEY from database")
	}

	log.Println("Settings loaded from database on startup")
}

// MaskSensitiveValue masks API keys for display
func MaskSensitiveValue(value string) string {
	if value == "" {
		return "Not set"
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:4] + "****...****"
}
