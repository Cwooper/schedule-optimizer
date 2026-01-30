package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		want         string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_VAR_NOT_SET",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_VAR_SET",
			defaultValue: "default",
			envValue:     "custom",
			setEnv:       true,
			want:         "custom",
		},
		{
			name:         "returns default for empty env value",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing env var
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.want)
			}
		})
	}
}

func TestParseCORSOrigins(t *testing.T) {
	tests := []struct {
		name    string
		origins string
		want    []string
	}{
		{
			name:    "empty string",
			origins: "",
			want:    []string{},
		},
		{
			name:    "single origin",
			origins: "http://localhost:3000",
			want:    []string{"http://localhost:3000"},
		},
		{
			name:    "multiple origins",
			origins: "http://localhost:3000,http://localhost:5173",
			want:    []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:    "origins with spaces",
			origins: "http://localhost:3000 , http://localhost:5173 ",
			want:    []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:    "empty entries filtered",
			origins: "http://localhost:3000,,http://localhost:5173",
			want:    []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:    "whitespace-only entries filtered",
			origins: "http://localhost:3000,   ,http://localhost:5173",
			want:    []string{"http://localhost:3000", "http://localhost:5173"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCORSOrigins(tt.origins)
			if len(got) != len(tt.want) {
				t.Errorf("parseCORSOrigins(%q) returned %d items, want %d", tt.origins, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseCORSOrigins(%q)[%d] = %q, want %q", tt.origins, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Save original env vars
	originalPort := os.Getenv("PORT")
	originalEnv := os.Getenv("ENVIRONMENT")
	originalCORS := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalDB := os.Getenv("DATABASE_PATH")

	// Restore after test
	defer func() {
		restoreEnv("PORT", originalPort)
		restoreEnv("ENVIRONMENT", originalEnv)
		restoreEnv("CORS_ALLOWED_ORIGINS", originalCORS)
		restoreEnv("DATABASE_PATH", originalDB)
	}()

	t.Run("loads defaults when env not set", func(t *testing.T) {
		os.Unsetenv("PORT")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
		os.Unsetenv("DATABASE_PATH")

		cfg := Load()

		if cfg.Port != "8080" {
			t.Errorf("Port = %q, want %q", cfg.Port, "8080")
		}
		if cfg.Environment != "development" {
			t.Errorf("Environment = %q, want %q", cfg.Environment, "development")
		}
		if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:3000" {
			t.Errorf("CORSAllowedOrigins = %v, want [http://localhost:3000]", cfg.CORSAllowedOrigins)
		}
		if cfg.DatabasePath != "data/schedule.db" {
			t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, "data/schedule.db")
		}
	})

	t.Run("loads custom values from env", func(t *testing.T) {
		os.Setenv("PORT", "9000")
		os.Setenv("ENVIRONMENT", "production")
		os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://api.example.com")
		os.Setenv("DATABASE_PATH", "/var/data/app.db")

		cfg := Load()

		if cfg.Port != "9000" {
			t.Errorf("Port = %q, want %q", cfg.Port, "9000")
		}
		if cfg.Environment != "production" {
			t.Errorf("Environment = %q, want %q", cfg.Environment, "production")
		}
		if len(cfg.CORSAllowedOrigins) != 2 {
			t.Errorf("CORSAllowedOrigins has %d items, want 2", len(cfg.CORSAllowedOrigins))
		}
		if cfg.DatabasePath != "/var/data/app.db" {
			t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, "/var/data/app.db")
		}
	})
}

func restoreEnv(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}
