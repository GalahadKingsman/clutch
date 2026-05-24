package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppEnv   string
	LogLevel string

	APIPublicURL     string
	MiniAppPublicURL string
	CORSOrigins      []string

	TelegramBotToken   string
	TelegramBotUser    string
	TelegramWebhookSec string

	JWTSecret string

	DatabaseURL string
	RedisURL    string

	SolanaRPCURL      string
	ClutchProgramID   string
	ClutchTreasuryKey string

	APIPort          string
	BotWebhookPort   string
}

func Load() (*Config, error) {
	c := &Config{
		AppEnv:   getEnv("APP_ENV", "production"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		APIPublicURL:     getEnv("API_PUBLIC_URL", "https://api.example.com"),
		MiniAppPublicURL: getEnv("MINIAPP_PUBLIC_URL", "https://app.example.com"),
		CORSOrigins:      splitCSV(getEnv("CORS_ORIGINS", "https://app.example.com")),

		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramBotUser:    getEnv("TELEGRAM_BOT_USERNAME", "clutch_game_bot"),
		TelegramWebhookSec: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),

		JWTSecret: os.Getenv("JWT_SECRET"),

		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    getEnv("REDIS_URL", "redis://redis:6379/0"),

		SolanaRPCURL:      getEnv("SOLANA_RPC_URL", "https://api.devnet.solana.com"),
		ClutchProgramID:   os.Getenv("CLUTCH_PROGRAM_ID"),
		ClutchTreasuryKey: os.Getenv("CLUTCH_TREASURY_PUBKEY"),

		APIPort:        getEnv("API_PORT", "8080"),
		BotWebhookPort: getEnv("BOT_WEBHOOK_PORT", "8081"),
	}

	if c.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
