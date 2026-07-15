package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	HTTPAddr    string
	DBPath      string
	JWTSecret   string
	ServiceToken string

	// Bot
	BotToken       string
	BotMode        string // "polling" | "webhook"
	WebhookURL     string
	WebhookSecret  string
	BotAPIBaseURL  string // optional custom TG API base

	// Telegram Mini App
	WebAppURL string // public base URL of the Mini App (e.g. https://app.example.com)

	// Scheduler
	SchedulerInterval time.Duration

	// API base URL used by the bot adapter to reach the backend.
	APIBaseURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DBPath:      getEnv("DB_PATH", "data/crm.db"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-prod"),
		ServiceToken: getEnv("SERVICE_TOKEN", "local-dev-service-token"),

		BotToken:      getEnv("BOT_TOKEN", ""),
		BotMode:       getEnv("BOT_MODE", "polling"),
		WebhookURL:    getEnv("WEBHOOK_URL", ""),
		WebhookSecret: getEnv("WEBHOOK_SECRET", ""),
		BotAPIBaseURL: getEnv("BOT_API_BASE_URL", ""),

		WebAppURL: getEnv("WEBAPP_URL", ""),

		SchedulerInterval: getEnvDuration("SCHEDULER_INTERVAL", 30*time.Second),
		APIBaseURL:        getEnv("API_BASE_URL", "http://localhost:8080"),
	}

	if cfg.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}
	if cfg.BotMode == "webhook" && cfg.WebhookURL == "" {
		return nil, fmt.Errorf("WEBHOOK_URL is required when BOT_MODE=webhook")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
