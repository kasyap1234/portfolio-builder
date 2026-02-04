package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	DatabaseURL        string
	TelegramBotToken   string
	TelegramChatID     int64
	MonthlySIPAmount   float64
	DebtReserve        float64
	PETriggerThreshold float64
}

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	_ = godotenv.Load() // Ignore error if .env doesn't exist
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	LoadEnv()

	chatID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)
	sipAmount, _ := strconv.ParseFloat(os.Getenv("MONTHLY_SIP_AMOUNT"), 64)
	if sipAmount == 0 {
		sipAmount = 35000 // Default ₹35k
	}
	debtReserve, _ := strconv.ParseFloat(os.Getenv("DEBT_RESERVE"), 64)
	peThreshold, _ := strconv.ParseFloat(os.Getenv("PE_TRIGGER_THRESHOLD"), 64)
	if peThreshold == 0 {
		peThreshold = 19.0 // Default to 19.0
	}

	return &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:     chatID,
		MonthlySIPAmount:   sipAmount,
		DebtReserve:        debtReserve,
		PETriggerThreshold: peThreshold,
	}
}
