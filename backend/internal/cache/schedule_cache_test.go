package cache

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/testutil"
)

func TestNewScheduleCache(t *testing.T) {
	_, queries := testutil.SetupTestDB(t)

	cache := NewScheduleCache(queries, nil)

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

	cache := NewScheduleCache(queries, nil)
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

	cache := NewScheduleCache(queries, nil)
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

	cache := NewScheduleCache(queries, nil)
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

	cache := NewScheduleCache(queries, nil)
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

	cache := NewScheduleCache(queries, nil)
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

	cache := NewScheduleCache(queries, nil)
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

func TestInvalidateReloads(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	cache := NewScheduleCache(queries, nil)
	ctx := context.Background()

	before, err := cache.Term(ctx, "202520")
	if err != nil {
		t.Fatal(err)
	}
	// A scrape updates seats and the calendar job publishes finals.
	if _, err := db.Exec("UPDATE sections SET seats_available = 1 WHERE crn = '20001'"); err != nil {
		t.Fatal(err)
	}
	testutil.SeedSpring2025Calendar(t, db)

	if cached, _ := cache.Term(ctx, "202520"); cached != before {
		t.Fatal("Term reloaded without an invalidation")
	}
	cache.Invalidate("202520")
	after, err := cache.Term(ctx, "202520")
	if err != nil {
		t.Fatal(err)
	}
	course := after.Courses["20001"]
	if course.SeatsAvailable != 1 || course.Final == nil || course.Final.Date != "2025-06-10" {
		t.Errorf("after invalidation: seats %d, final %+v; want 1 seat and the Jun 10 final", course.SeatsAvailable, course.Final)
	}
	// A request holding the old snapshot still reads a whole term.
	if old := before.Courses["20001"]; old.SeatsAvailable != 5 || old.Final != nil || len(before.Courses) != 3 {
		t.Errorf("old snapshot changed: seats %d, final %+v, %d courses", old.SeatsAvailable, old.Final, len(before.Courses))
	}
}

func TestLoadBeforeInvalidationIsNotCached(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	cache := NewScheduleCache(queries, nil)
	ctx := context.Background()

	// A load reads the database, a scrape rewrites it and invalidates, and
	// then the load finishes.
	started := cache.generation
	cache.Invalidate("202520")
	td, err := cache.load(ctx, "202520", started)
	if err != nil || len(td.Courses) != 3 {
		t.Fatalf("load = %d courses, %v; want the whole term returned", len(td.Courses), err)
	}
	if cache.IsTermLoaded("202520") {
		t.Error("a load that started before an invalidation was cached")
	}

	cache.LoadTerm(ctx, "202520")
	cache.LoadTerm(ctx, "202510")
	cache.InvalidateAll()
	if terms := cache.GetActiveTerms(); len(terms) != 0 {
		t.Errorf("after InvalidateAll, cached terms = %v", terms)
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

	cache := NewScheduleCache(queries, nil)
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

	cache := NewScheduleCache(queries, nil)
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
				cache.Invalidate(term)
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentLoadSameTerm(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)

	cache := NewScheduleCache(queries, nil)
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

func TestTerm(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	testutil.SeedSpring2025Calendar(t, db)
	cache := NewScheduleCache(queries, nil)
	ctx := context.Background()

	first, err := cache.Term(ctx, "202520")
	if err != nil {
		t.Fatal(err)
	}
	if second, _ := cache.Term(ctx, "202520"); second != first {
		t.Error("a cached term was loaded again")
	}
	// CRN 20001 meets MWF 10:00, the seeded slot's range.
	want := calendar.Final{Date: "2025-06-10", Start: "1030", End: "1230", Certain: true}
	if got := first.Courses["20001"].Final; got == nil || *got != want {
		t.Errorf("CRN 20001 final = %+v, want %+v", got, want)
	}

	// The section's own type decides: a lab-only section has no final.
	if _, err := db.Exec("UPDATE sections SET schedule_type = 'Lab' WHERE crn = '20001'"); err != nil {
		t.Fatal(err)
	}
	cache.Invalidate("202520")
	lab, err := cache.Term(ctx, "202520")
	if err != nil {
		t.Fatal(err)
	}
	if got := lab.Courses["20001"].Final; got != nil {
		t.Errorf("lab CRN 20001 final = %+v, want none", got)
	}
}

// Requests racing invalidations each get a whole term, never a partial one
// or none.
func TestTermConcurrentWithInvalidation(t *testing.T) {
	db, queries := testutil.SetupTestDB(t)
	testutil.SeedTestData(t, db)
	cache := NewScheduleCache(queries, nil)
	ctx := context.Background()

	const requests = 50
	var wg sync.WaitGroup
	for range requests {
		wg.Go(func() {
			td, err := cache.Term(ctx, "202520")
			if err != nil || len(td.Courses) != 3 {
				t.Errorf("Term = %v, %v; want 3 courses", td, err)
			}
		})
	}
	wg.Go(func() {
		for range requests {
			cache.Invalidate("202520")
		}
	})
	wg.Wait()
}

// Benchmarks

func BenchmarkLoadTerm(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	seedBenchmarkData(b, db, 100)

	ctx := context.Background()

	for b.Loop() {
		cache := NewScheduleCache(queries, nil)
		if err := cache.LoadTerm(ctx, "202520"); err != nil {
			b.Fatalf("LoadTerm failed: %v", err)
		}
	}
}

func BenchmarkGetCourse(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	seedBenchmarkData(b, db, 100)

	cache := NewScheduleCache(queries, nil)
	ctx := context.Background()
	cache.LoadTerm(ctx, "202520")

	for b.Loop() {
		cache.GetCourse("202520", "20050")
	}
}

func BenchmarkGetCoursesByCRNs(b *testing.B) {
	db, queries := testutil.SetupTestDB(b)
	seedBenchmarkData(b, db, 100)

	cache := NewScheduleCache(queries, nil)
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

	cache := NewScheduleCache(queries, nil)
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
