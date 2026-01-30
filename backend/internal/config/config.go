package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Port               string
	Environment        string
	CORSAllowedOrigins []string
	DatabasePath       string
}

// Load reads environment variables and returns a Config struct.
// Normalizes environment to lowercase.
func Load() *Config {
	port := getEnv("PORT", "8080")
	environment := strings.ToLower(getEnv("ENVIRONMENT", "development"))
	corsOrigins := parseCORSOrigins(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"))
	databasePath := getEnv("DATABASE_PATH", "data/schedule.db")

	slog.Info("Configuration loaded",
		"port", port,
		"environment", environment,
		"cors_origins", corsOrigins,
		"database_path", databasePath,
	)

	return &Config{
		Port:               port,
		Environment:        environment,
		CORSAllowedOrigins: corsOrigins,
		DatabasePath:       databasePath,
	}
}

// Retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseCORSOrigins splits a comma-separated string of origins into a slice
func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}

	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))

	for _, origin := range parts {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

