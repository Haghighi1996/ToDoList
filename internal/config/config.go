package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (Config, error) {
	config := Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if config.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	return config, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
