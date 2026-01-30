// Package testutil provides shared test helpers for database testing.
package testutil

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"schedule-optimizer/internal/store"
)

// getProjectRoot finds the project root by looking for go.mod
func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	// Walk up from internal/testutil to find project root
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find project root")
		}
		dir = parent
	}
}

// SetupTestDB creates an in-memory SQLite database with the schema applied.
// Returns the database connection and a store.Queries instance.
func SetupTestDB(t testing.TB) (*sql.DB, *store.Queries) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Read and apply the migration
	migrationPath := filepath.Join(getProjectRoot(), "migrations", "000001_initial_schema.up.sql")
	schema, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}

	return db, store.New(db)
}

// SeedTestData inserts common test data used across multiple test packages.
func SeedTestData(t testing.TB, db *sql.DB) {
	t.Helper()

	data := `
		INSERT INTO terms (code, description) VALUES
			('202520', 'Spring 2025'),
			('202510', 'Winter 2025');

		INSERT INTO sections (term, crn, subject, course_number, title, credit_hours_low, enrollment, max_enrollment, seats_available, is_open)
		VALUES
			('202520', '20001', 'CSCI', '247', 'Data Structures', 4, 25, 30, 5, 1),
			('202520', '20002', 'CSCI', '301', 'Algorithms', 4, 30, 30, 0, 0),
			('202520', '20003', 'MATH', '204', 'Linear Algebra', 5, 20, 35, 15, 1),
			('202510', '10001', 'CSCI', '247', 'Data Structures', 4, 28, 30, 2, 1);

		INSERT INTO instructors (section_id, name, email, is_primary)
		VALUES
			(1, 'Dr. Smith', 'smith@wwu.edu', 1),
			(2, 'Dr. Jones', 'jones@wwu.edu', 1),
			(3, 'Dr. Brown', 'brown@wwu.edu', 1),
			(4, 'Dr. Smith', 'smith@wwu.edu', 1);

		INSERT INTO meeting_times (section_id, start_time, end_time, building, room, monday, wednesday, friday)
		VALUES
			(1, '1000', '1050', 'CF', '105', 1, 1, 1),
			(2, '1400', '1550', 'CF', '110', 0, 0, 0),
			(3, '0900', '0950', 'BH', '220', 1, 1, 1),
			(4, '1000', '1050', 'CF', '105', 1, 1, 1);

		UPDATE meeting_times SET tuesday = 1, thursday = 1 WHERE section_id = 2;
	`

	if _, err := db.Exec(data); err != nil {
		t.Fatalf("failed to seed test data: %v", err)
	}
}
