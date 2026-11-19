package grades

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"schedule-optimizer/internal/store"
)

func joinRow(term, crn, subject, number, professor, bannerSubject, bannerNumber, bannerInstructor string, c Counts) *store.GetGradeRowsWithBannerRow {
	null := func(s string) sql.NullString { return sql.NullString{String: s, Valid: s != ""} }
	return &store.GetGradeRowsWithBannerRow{
		GradeRow: store.GradeRow{
			Term: term, Crn: crn, Subject: subject, CourseNumber: number, Title: subject + " " + number,
			Professor: professor, StudentsEnrolled: 30,
			CntA: int64(c[GradeA]), CntB: int64(c[GradeB]), CntW: int64(c[GradeW]),
		},
		BannerSubject:      null(bannerSubject),
		BannerCourseNumber: null(bannerNumber),
		BannerInstructor:   null(bannerInstructor),
	}
}

func TestCourseKeys(t *testing.T) {
	rows := []*store.GetGradeRowsWithBannerRow{
		// LANG is a department: its 101 is several Banner courses.
		joinRow("202340", "1", "LANG", "101", "", "SPAN", "101", "", Counts{}),
		joinRow("202340", "2", "LANG", "101", "", "FREN", "101", "", Counts{}),
		joinRow("202340", "3", "HHD", "104", "", "PE", "104", "", Counts{}),
		// No Banner section: HHD 104 always joined PE 104, LANG 101 is ambiguous.
		joinRow("201540", "4", "HHD", "104", "", "", "", "", Counts{}),
		joinRow("201540", "5", "LANG", "101", "", "", "", "", Counts{}),
		joinRow("201540", "6", "CSD", "680", "", "", "", "", Counts{}),
	}
	want := []CourseKey{
		{"SPAN", "101"}, {"FREN", "101"}, {"PE", "104"},
		{"PE", "104"}, {"LANG", "101"}, {"CSD", "680"},
	}
	got := courseKeys(rows)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestProfessorNames(t *testing.T) {
	var rows []*store.GetGradeRowsWithBannerRow
	join := func(grade, banner string, times int) {
		for range times {
			rows = append(rows, joinRow("202340", "1", "CSCI", "305", grade, "CSCI", "305", banner, Counts{}))
		}
	}
	// Two spellings of one person both join the same Banner name.
	join("Grace Dunbar Miller", "Dunbar Miller, Grace", 7)
	join("Grace Dunbar-Miller", "Dunbar Miller, Grace", 4)
	join("Grace Dunbar-Miller", "Lee, Kara", 1)
	// A name change: no shared surname, but joined often.
	join("Melanie Aceves", "Williams, Melanie", 115)
	join("Melanie Williams", "Williams, Melanie", 1)
	// A co-taught section joined the other instructor once.
	join("Alex Nguyen", "Lee, Kara", 1)
	join("Ana Ortiz", "", 1)

	want := map[string]string{
		"Grace Dunbar Miller": "Dunbar Miller, Grace",
		"Grace Dunbar-Miller": "Dunbar Miller, Grace",
		"Melanie Aceves":      "Williams, Melanie",
		"Melanie Williams":    "Williams, Melanie",
		"Alex Nguyen":         "Alex Nguyen",
		"Ana Ortiz":           "Ana Ortiz",
	}
	got := professorNames(rows)
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for grade, name := range want {
		if got[grade] != name {
			t.Errorf("%q -> %q, want %q", grade, got[grade], name)
		}
	}
}

func TestBuildIndex(t *testing.T) {
	a10 := Counts{GradeA: 10}
	idx := buildIndex(indexSource{
		rows: []*store.GetGradeRowsWithBannerRow{
			joinRow("202340", "100", "CSCI", "305", "Jordan Smith", "CSCI", "305", "Smith, Jordan", a10),
			joinRow("202440", "200", "CSCI", "305", "", "CSCI", "305", "Smith, Jordan", a10),
			joinRow("202440", "201", "CSCI", "305", "", "CSCI", "305", "New, Hire", a10),
			joinRow("202440", "202", "CSCI", "305", "", "CSCI", "305", "", a10),
			joinRow("202340", "300", "CSCI", "330", "Jordan Smith", "CSCI", "330", "", Counts{GradeB: 10}),
		},
		titles:   []*store.GetLatestCourseTitlesRow{{Subject: "CSCI", CourseNumber: "305", Title: "Data Structures and Algorithms I"}},
		subjects: []*store.GetLatestSubjectDescriptionsRow{{Subject: "CSCI", SubjectDescription: sql.NullString{String: "Computer Science", Valid: true}}},
	})

	c := idx.Course(CourseKey{"CSCI", "305"})
	if c == nil {
		t.Fatal("CSCI 305 missing")
	}
	if c.Title != "Data Structures and Algorithms I" {
		t.Errorf("title = %q, want Banner's", c.Title)
	}
	if got := idx.Course(CourseKey{"CSCI", "330"}).Title; got != "CSCI 330" {
		t.Errorf("title without a Banner title = %q, want the grade file's", got)
	}
	// Newest first; a blank grade name takes the Banner instructor.
	wantProfs := []struct{ crn, prof string }{
		{"200", "Smith, Jordan"}, {"201", "New, Hire"}, {"202", ""}, {"100", "Smith, Jordan"},
	}
	for i, w := range wantProfs {
		s := c.Sections[i]
		if s.CRN != w.crn || s.Professor != w.prof {
			t.Errorf("section %d = %s/%q, want %s/%q", i, s.CRN, s.Professor, w.crn, w.prof)
		}
	}
	level, ok := idx.Level(CourseKey{"CSCI", "399"})
	if !ok || level[GradeA] != 40 || level[GradeB] != 10 {
		t.Errorf("CSCI 3xx level = %v, %v", level, ok)
	}
	if subs := idx.Subjects(); len(subs) != 1 || subs[0].Name != "Computer Science" {
		t.Errorf("subjects = %v", subs)
	}
}

func BenchmarkLoadIndex(b *testing.B) {
	path := os.Getenv("TEST_DB_PATH")
	if path == "" {
		path = "../../../data/schedule.db"
	}
	db, err := sql.Open("sqlite3", "file:"+path+"?mode=ro")
	if err != nil {
		b.Skipf("open database: %v", err)
	}
	defer db.Close()
	queries := store.New(db)
	svc := NewService(db, queries, "")
	ctx := context.Background()
	if n, err := queries.GetGradeRowCount(ctx); err != nil || n == 0 {
		b.Skipf("no grade data: %v", err)
	}
	for b.Loop() {
		if _, err := svc.loadIndex(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
