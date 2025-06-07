package config

import (
	"os"
	"time"
)

type Config struct {
	Port          string
	LlamaURL      string
	TokenExpiry   time.Duration
	SessionExpiry time.Duration
	BufferSize    int
	LogLevel      string
	Environment   string
}

func New() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		LlamaURL:      getEnv("LLAMA_URL", "http://localhost:11434/api/generate"),
		TokenExpiry:   24 * time.Hour,
		SessionExpiry: 30 * time.Minute,
		BufferSize:    256,
		LogLevel:      getEnv("LOG_LEVEL", "INFO"),
		Environment:   getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
