package cache

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"schedule-optimizer/internal/testutil"
)

func TestNewScheduleCache(t *testing.T) {
	_, queries := testutil.SetupTestDB(t)

	cache := NewScheduleCache(queries)

	if cache == nil {
		t.Fatal("NewScheduleCache returned nil")
	}
	if cache.terms == nil {
		t.Error("terms map is nil")
	}
	if cache.queries != queries {
		t.Error("queries not set correctly")
	}
}

func TestLoadTerm(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()

	t.Run("loads term successfully", func(t *testing.T) {
		err := cache.LoadTerm(ctx, "202520")
		if err != nil {
			t.Fatalf("LoadTerm failed: %v", err)
		}

		if !cache.IsTermLoaded("202520") {
			t.Error("term not reported as loaded")
		}
	})

	t.Run("term data is indexed correctly", func(t *testing.T) {
		course, ok := cache.GetCourse("202520", "20001")
		if !ok {
			t.Fatal("course 20001 not found")
		}
		if course.Subject != "CSCI" {
			t.Errorf("Subject = %q, want CSCI", course.Subject)
		}
		if course.Instructor != "Dr. Smith" {
			t.Errorf("Instructor = %q, want Dr. Smith", course.Instructor)
		}
	})

	t.Run("meeting times loaded", func(t *testing.T) {
		course, _ := cache.GetCourse("202520", "20001")
		if len(course.MeetingTimes) != 1 {
			t.Fatalf("expected 1 meeting time, got %d", len(course.MeetingTimes))
		}

		mt := course.MeetingTimes[0]
		if mt.StartTime != "1000" {
			t.Errorf("StartTime = %q, want 1000", mt.StartTime)
		}
		if !mt.Days[1] { // Monday
			t.Error("Monday should be true")
		}
		if mt.Days[2] { // Tuesday
			t.Error("Tuesday should be false")
		}
	})

	t.Run("empty term loads without error", func(t *testing.T) {
		err := cache.LoadTerm(ctx, "999999")
		if err != nil {
			t.Fatalf("LoadTerm for empty term failed: %v", err)
		}

		courses := cache.GetAllCourses("999999")
		if len(courses) != 0 {
			t.Errorf("expected 0 courses, got %d", len(courses))
		}
	})
}

func TestGetCourse(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	t.Run("returns existing course", func(t *testing.T) {
		course, ok := cache.GetCourse("202520", "20001")
		if !ok {
			t.Fatal("course not found")
		}
		if course.CRN != "20001" {
			t.Errorf("CRN = %q, want 20001", course.CRN)
		}
	})

	t.Run("returns false for nonexistent CRN", func(t *testing.T) {
		_, ok := cache.GetCourse("202520", "99999")
		if ok {
			t.Error("expected false for nonexistent CRN")
		}
	})

	t.Run("returns false for unloaded term", func(t *testing.T) {
		_, ok := cache.GetCourse("202510", "10001")
		if ok {
			t.Error("expected false for unloaded term")
		}
	})
}

func TestGetCoursesByCRNs(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	t.Run("returns multiple courses", func(t *testing.T) {
		courses := cache.GetCoursesByCRNs("202520", []string{"20001", "20002"})
		if len(courses) != 2 {
			t.Errorf("expected 2 courses, got %d", len(courses))
		}
	})

	t.Run("filters out nonexistent CRNs", func(t *testing.T) {
		courses := cache.GetCoursesByCRNs("202520", []string{"20001", "99999"})
		if len(courses) != 1 {
			t.Errorf("expected 1 course, got %d", len(courses))
		}
	})

	t.Run("returns nil for unloaded term", func(t *testing.T) {
		courses := cache.GetCoursesByCRNs("202510", []string{"10001"})
		if courses != nil {
			t.Errorf("expected nil, got %v", courses)
		}
	})
}

func TestGetCoursesBySubject(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	t.Run("returns courses for subject", func(t *testing.T) {
		courses := cache.GetCoursesBySubject("202520", "CSCI")
		if len(courses) != 2 {
			t.Errorf("expected 2 CSCI courses, got %d", len(courses))
		}
	})

	t.Run("returns nil for nonexistent subject", func(t *testing.T) {
		courses := cache.GetCoursesBySubject("202520", "PHYS")
		if courses != nil {
			t.Errorf("expected nil, got %v", courses)
		}
	})
}

func TestGetAllCourses(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	courses := cache.GetAllCourses("202520")
	if len(courses) != 3 {
		t.Errorf("expected 3 courses, got %d", len(courses))
	}
}

func TestGetActiveTerms(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()

	// Initially empty
	terms := cache.GetActiveTerms()
	if len(terms) != 0 {
		t.Errorf("expected 0 active terms, got %d", len(terms))
	}

	// Load terms
	cache.LoadTerm(ctx, "202520")
	cache.LoadTerm(ctx, "202510")

	terms = cache.GetActiveTerms()
	if len(terms) != 2 {
		t.Errorf("expected 2 active terms, got %d", len(terms))
	}
}

func TestUnloadTerm(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	if !cache.IsTermLoaded("202520") {
		t.Fatal("term should be loaded")
	}

	cache.UnloadTerm("202520")

	if cache.IsTermLoaded("202520") {
		t.Error("term should be unloaded")
	}

	terms := cache.GetActiveTerms()
	if len(terms) != 0 {
		t.Errorf("expected 0 active terms, got %d", len(terms))
	}
}

func TestNullHelpers(t *testing.T) {
	t.Run("nullString", func(t *testing.T) {
		valid := sql.NullString{String: "test", Valid: true}
		if got := nullString(valid); got != "test" {
			t.Errorf("nullString(valid) = %q, want test", got)
		}

		invalid := sql.NullString{Valid: false}
		if got := nullString(invalid); got != "" {
			t.Errorf("nullString(invalid) = %q, want empty", got)
		}
	})

	t.Run("nullInt", func(t *testing.T) {
		valid := sql.NullInt64{Int64: 42, Valid: true}
		if got := nullInt(valid); got != 42 {
			t.Errorf("nullInt(valid) = %d, want 42", got)
		}

		invalid := sql.NullInt64{Valid: false}
		if got := nullInt(invalid); got != 0 {
			t.Errorf("nullInt(invalid) = %d, want 0", got)
		}
	})
}

// Concurrent access tests

func TestConcurrentReads(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent reads should not race
	for range numGoroutines {
		go func() {
			defer wg.Done()
			cache.GetCourse("202520", "20001")
			cache.GetAllCourses("202520")
			cache.GetCoursesBySubject("202520", "CSCI")
			cache.IsTermLoaded("202520")
			cache.GetActiveTerms()
		}()
	}

	wg.Wait()
}

func TestConcurrentReadWrite(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()

	const numReaders = 50
	const numWriters = 10
	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	// Start readers
	for range numReaders {
		go func() {
			defer wg.Done()
			for range 100 {
				cache.GetCourse("202520", "20001")
				cache.IsTermLoaded("202520")
				cache.GetActiveTerms()
			}
		}()
	}

	// Start writers (load/unload terms)
	for i := range numWriters {
		go func(id int) {
			defer wg.Done()
			term := "202520"
			if id%2 == 0 {
				term = "202510"
			}
			for range 10 {
				cache.LoadTerm(ctx, term)
				cache.UnloadTerm(term)
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentLoadSameTerm(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries)
	ctx := context.Background()

	const numGoroutines = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// All goroutines try to load the same term simultaneously
	for range numGoroutines {
		go func() {
			defer wg.Done()
			cache.LoadTerm(ctx, "202520")
		}()
	}

	wg.Wait()

	// Verify term is loaded correctly
	if !cache.IsTermLoaded("202520") {
		t.Error("term should be loaded")
	}

	courses := cache.GetAllCourses("202520")
	if len(courses) != 3 {
		t.Errorf("expected 3 courses, got %d", len(courses))
	}
}

// Benchmarks

func BenchmarkLoadTerm(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	seedBenchmarkData(b, db, 100)

	ctx := context.Background()

	for b.Loop() {
		cache := NewScheduleCache(queries)
		if err := cache.LoadTerm(ctx, "202520"); err != nil {
			b.Fatalf("LoadTerm failed: %v", err)
		}
	}
}

func BenchmarkGetCourse(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	seedBenchmarkData(b, db, 100)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	for b.Loop() {
		cache.GetCourse("202520", "20050")
	}
}

func BenchmarkGetCoursesByCRNs(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	seedBenchmarkData(b, db, 100)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	crns := []string{"20001", "20010", "20020", "20030", "20040"}

	for b.Loop() {
		cache.GetCoursesByCRNs("202520", crns)
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	seedBenchmarkData(b, db, 100)

	cache := NewScheduleCache(queries)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.GetCourse("202520", "20050")
		}
	})
}

func seedBenchmarkData(b *testing.B, db *sql.DB, count int) {
	b.Helper()

	for i := 1; i <= count; i++ {
		_, err := db.Exec(`
			INSERT INTO sections (term, crn, subject, course_number, title, credit_hours_low, is_open)
			VALUES ('202520', ?, 'CSCI', ?, 'Test Course', 4, 1)
		`, 20000+i, 100+i)
		if err != nil {
			b.Fatalf("failed to insert section: %v", err)
		}

		_, err = db.Exec(`
			INSERT INTO instructors (section_id, name, email, is_primary)
			VALUES (?, 'Dr. Test', 'test@wwu.edu', 1)
		`, i)
		if err != nil {
			b.Fatalf("failed to insert instructor: %v", err)
		}

		_, err = db.Exec(`
			INSERT INTO meeting_times (section_id, start_time, end_time, monday, wednesday, friday)
			VALUES (?, '1000', '1050', 1, 1, 1)
		`, i)
		if err != nil {
			b.Fatalf("failed to insert meeting time: %v", err)
		}
	}
}
