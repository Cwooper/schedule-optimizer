package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	Environment        string
	CORSAllowedOrigins []string
	DatabasePath       string
	ScraperConcurrency int

	// Jobs scheduler config
	JobsEnabled       bool
	ActiveScrapeHours int // Hours between active registration term scrapes
	DailyScrapeHour   int // Hour of day (0-23) for daily scrapes
	LogRetentionDays  int // Days to keep logs before pruning
	PastTermYears     int // Years of past terms to scrape
}

// Load reads environment variables and returns a Config struct.
// Normalizes environment to lowercase.
func Load() *Config {
	port := getEnv("PORT", "8080")
	environment := strings.ToLower(getEnv("ENVIRONMENT", "development"))
	corsOrigins := parseCORSOrigins(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"))
	databasePath := getEnv("DATABASE_PATH", "data/schedule.db")
	scraperConcurrency := getEnvInt("SCRAPER_CONCURRENCY", 4)

	// Jobs scheduler config
	jobsEnabled := getEnvBool("JOBS_ENABLED", true)
	activeScrapeHours := getEnvInt("JOBS_ACTIVE_SCRAPE_HOURS", 8)
	dailyScrapeHour := getEnvInt("JOBS_DAILY_SCRAPE_HOUR", 3)
	logRetentionDays := getEnvInt("JOBS_LOG_RETENTION_DAYS", 90)
	pastTermYears := getEnvInt("JOBS_PAST_TERM_YEARS", 5)

	slog.Info("Configuration loaded",
		"port", port,
		"environment", environment,
		"cors_origins", corsOrigins,
		"database_path", databasePath,
		"scraper_concurrency", scraperConcurrency,
		"jobs_enabled", jobsEnabled,
		"active_scrape_hours", activeScrapeHours,
		"daily_scrape_hour", dailyScrapeHour,
		"log_retention_days", logRetentionDays,
		"past_term_years", pastTermYears,
	)

	return &Config{
		Port:               port,
		Environment:        environment,
		CORSAllowedOrigins: corsOrigins,
		DatabasePath:       databasePath,
		ScraperConcurrency: scraperConcurrency,
		JobsEnabled:        jobsEnabled,
		ActiveScrapeHours:  activeScrapeHours,
		DailyScrapeHour:    dailyScrapeHour,
		LogRetentionDays:   logRetentionDays,
		PastTermYears:      pastTermYears,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
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

