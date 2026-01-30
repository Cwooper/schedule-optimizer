// Package db provides database connection management.
package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Open opens a SQLite database connection.
// Uses WAL mode for better concurrent read performance.
// Schema must be applied separately via store.ApplySchema().
func Open(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "data/schedule.db"
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// Open with WAL mode, busy timeout, and foreign key enforcement
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite works best with a single connection to avoid locking issues
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("Database connection established", "path", dbPath)
	return db, nil
}
