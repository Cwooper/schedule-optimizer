package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	// Use a temp directory for test databases
	tempDir := t.TempDir()

	t.Run("creates database file and directory", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "subdir", "test.db")

		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open(%q) failed: %v", dbPath, err)
		}
		defer db.Close()

		// Check directory was created
		if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
			t.Error("directory was not created")
		}

		// Check database file was created
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("database file was not created")
		}

		// Verify we can ping the database
		if err := db.Ping(); err != nil {
			t.Errorf("Ping() failed: %v", err)
		}
	})

	t.Run("uses default path when empty", func(t *testing.T) {
		// Change to temp dir so default path doesn't pollute working dir
		originalWd, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(originalWd)

		db, err := Open("")
		if err != nil {
			t.Fatalf("Open(\"\") failed: %v", err)
		}
		defer db.Close()

		expectedPath := filepath.Join(tempDir, "data", "schedule.db")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected database at %q, but it doesn't exist", expectedPath)
		}
	})

	t.Run("enables WAL mode", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "wal_test.db")

		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open(%q) failed: %v", dbPath, err)
		}
		defer db.Close()

		var journalMode string
		err = db.QueryRow("PRAGMA journal_mode;").Scan(&journalMode)
		if err != nil {
			t.Fatalf("failed to query journal_mode: %v", err)
		}

		if journalMode != "wal" {
			t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
		}
	})

	t.Run("sets busy timeout", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "timeout_test.db")

		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open(%q) failed: %v", dbPath, err)
		}
		defer db.Close()

		var busyTimeout int
		err = db.QueryRow("PRAGMA busy_timeout;").Scan(&busyTimeout)
		if err != nil {
			t.Fatalf("failed to query busy_timeout: %v", err)
		}

		if busyTimeout != 5000 {
			t.Errorf("busy_timeout = %d, want %d", busyTimeout, 5000)
		}
	})

	t.Run("can execute basic queries", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "query_test.db")

		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open(%q) failed: %v", dbPath, err)
		}
		defer db.Close()

		// Create a test table
		_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)")
		if err != nil {
			t.Fatalf("CREATE TABLE failed: %v", err)
		}

		// Insert data
		_, err = db.Exec("INSERT INTO test (value) VALUES (?)", "hello")
		if err != nil {
			t.Fatalf("INSERT failed: %v", err)
		}

		// Query data
		var value string
		err = db.QueryRow("SELECT value FROM test WHERE id = 1").Scan(&value)
		if err != nil {
			t.Fatalf("SELECT failed: %v", err)
		}
		if value != "hello" {
			t.Errorf("value = %q, want %q", value, "hello")
		}
	})
}

func BenchmarkOpen(b *testing.B) {
	tempDir := b.TempDir()

	for b.Loop() {
		dbPath := filepath.Join(tempDir, "bench.db")
		db, err := Open(dbPath)
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		db.Close()
		os.Remove(dbPath)
	}
}
