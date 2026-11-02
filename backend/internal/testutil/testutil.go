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

// migrationsDir resolves from this file's path so tests in any package
// directory find it.
func migrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

// SetupTestDB returns an in-memory SQLite database with every migration
// applied, closed when the test ends.
func SetupTestDB(t testing.TB) (*sql.DB, *store.Queries) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	// Each connection to :memory: is a separate, empty database, so a second
	// pooled connection (a transaction, say) would not see the schema.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	migrations, err := filepath.Glob(filepath.Join(migrationsDir(), "*.up.sql"))
	if err != nil || len(migrations) == 0 {
		t.Fatalf("no migrations found: %v", err)
	}
	for _, path := range migrations { // Glob sorts, so this is version order
		schema, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read migration: %v", err)
		}
		if _, err := db.Exec(string(schema)); err != nil {
			t.Fatalf("failed to apply %s: %v", filepath.Base(path), err)
		}
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

// SeedSpring2025Calendar gives SeedTestData's Spring 2025 term published
// dates, a Phase I row, a holiday, a deadline, and one finals slot matching
// CRN 20001 (MWF 10:00).
func SeedSpring2025Calendar(t testing.TB, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO term_dates (term_code, start_date, end_date) VALUES ('202520', '2025-03-31', '2025-06-13');
		INSERT INTO important_dates (term_code, date, end_date, description, category, opens_registration) VALUES
			('202520', '2025-02-25', '2025-03-10', 'Phase I - registration', 'registration', 1),
			('202520', '2025-05-26', NULL, 'Memorial Day', 'holiday', 0),
			('202520', '2025-05-11', NULL, 'Last day to withdraw from a class', 'deadline', 0),
			('202520', '2025-06-09', '2025-06-13', 'Final Exams', 'academic', 0);
		INSERT INTO finals_mappings (term_code, time_range_start, time_range_end, has_tuth, exam_date, exam_start_time, exam_end_time)
			VALUES ('202520', '0930', '1029', 0, '2025-06-10', '1030', '1230');
	`)
	if err != nil {
		t.Fatal(err)
	}
}
