// Package grades provides grade data import, mapping, aggregation, and runtime GPA lookup.
// Grade data comes from a static PRR Excel file. Subject and instructor name mappings
// are auto-discovered by joining grade data with scraped Banner data on (term, CRN).
package grades

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"

	"schedule-optimizer/internal/store"
)

// Service manages grade data lifecycle and provides in-memory GPA lookups.
type Service struct {
	db       *sql.DB
	queries  *store.Queries
	dataPath string

	mu            sync.RWMutex
	subjectMap    map[string]string          // banner_subject → grade_subject
	instructorMap map[string]string          // banner_name → grade_name
	courseProfAgg map[string]*GradeAggregate // "SUBJECT:NUM:INSTRUCTOR"
	courseAgg     map[string]*GradeAggregate // "SUBJECT:NUM"
	professorAgg  map[string]*GradeAggregate // "INSTRUCTOR"
	subjectAgg    map[string]*GradeAggregate // "SUBJECT"
	loaded        bool
}

// NewService creates a new grade service.
func NewService(db *sql.DB, queries *store.Queries, dataPath string) *Service {
	return &Service{
		db:       db,
		queries:  queries,
		dataPath: dataPath,
	}
}

// ImportAndCompute runs the full pipeline: import Excel → compute mappings → aggregate → load into memory.
// Each step is idempotent: if data already exists in DB, the step is skipped.
func (s *Service) ImportAndCompute(ctx context.Context) error {
	// Step 1: Import Excel if needed
	count, err := s.queries.GetGradeRowCount(ctx)
	if err != nil {
		return fmt.Errorf("check grade row count: %w", err)
	}
	if count == 0 {
		slog.Info("Importing grade data from Excel", "path", s.dataPath)
		if err := importExcel(ctx, s.db, s.dataPath); err != nil {
			return fmt.Errorf("import excel: %w", err)
		}
	} else {
		slog.Info("Grade data already imported", "rows", count)
	}

	// Step 2: Compute mappings if needed
	mappingCount, err := s.queries.GetSubjectMappingCount(ctx)
	if err != nil {
		return fmt.Errorf("check mapping count: %w", err)
	}
	if mappingCount == 0 {
		slog.Info("Computing grade-to-Banner mappings via CRN join")
		if err := computeMappings(ctx, s.db, s.queries); err != nil {
			return fmt.Errorf("compute mappings: %w", err)
		}
	} else {
		slog.Info("Grade mappings already computed", "subject_mappings", mappingCount)
	}

	// Step 3: Compute aggregates if needed
	aggCount, err := s.queries.GetGradeAggregateCount(ctx)
	if err != nil {
		return fmt.Errorf("check aggregate count: %w", err)
	}
	if aggCount == 0 {
		slog.Info("Computing grade aggregates")
		if err := computeAggregates(ctx, s.db, s.queries); err != nil {
			return fmt.Errorf("compute aggregates: %w", err)
		}
	} else {
		slog.Info("Grade aggregates already computed", "aggregates", aggCount)
	}

	// Step 4: Load into memory
	return s.LoadFromDB(ctx)
}

// LoadFromDB loads pre-computed mappings and aggregates from DB into memory.
func (s *Service) LoadFromDB(ctx context.Context) error {
	// Load subject mappings
	subjectMappings, err := s.queries.GetSubjectMappings(ctx)
	if err != nil {
		return fmt.Errorf("load subject mappings: %w", err)
	}
	subjectMap := make(map[string]string, len(subjectMappings))
	for _, m := range subjectMappings {
		subjectMap[m.BannerSubject] = m.GradeSubject
	}

	// Load instructor mappings
	instructorMappings, err := s.queries.GetInstructorMappings(ctx)
	if err != nil {
		return fmt.Errorf("load instructor mappings: %w", err)
	}
	instructorMap := make(map[string]string, len(instructorMappings))
	for _, m := range instructorMappings {
		instructorMap[m.BannerName] = m.GradeName
	}

	// Load aggregates
	dbAggs, err := s.queries.GetAllGradeAggregates(ctx)
	if err != nil {
		return fmt.Errorf("load aggregates: %w", err)
	}

	courseProfAgg := make(map[string]*GradeAggregate)
	courseAgg := make(map[string]*GradeAggregate)
	professorAgg := make(map[string]*GradeAggregate)
	subjectAgg := make(map[string]*GradeAggregate)

	for _, row := range dbAggs {
		agg := &GradeAggregate{
			Level:        row.Level,
			Subject:      row.Subject,
			CourseNumber: row.CourseNumber,
			Instructor:   row.Instructor,
			Sections:     int(row.SectionsCount),
			Students:     int(row.StudentsCount),
			CntA:         int(row.CntA),
			CntAM:        int(row.CntAm),
			CntBP:        int(row.CntBp),
			CntB:         int(row.CntB),
			CntBM:        int(row.CntBm),
			CntCP:        int(row.CntCp),
			CntC:         int(row.CntC),
			CntCM:        int(row.CntCm),
			CntDP:        int(row.CntDp),
			CntD:         int(row.CntD),
			CntDM:        int(row.CntDm),
			CntF:         int(row.CntF),
			CntW:         int(row.CntW),
			CntP:         int(row.CntP),
			CntNP:        int(row.CntNp),
			CntS:         int(row.CntS),
			CntU:         int(row.CntU),
			GPA:          row.Gpa,
		}
		if row.PassRate.Valid {
			pr := row.PassRate.Float64
			agg.PassRate = &pr
		}

		switch row.Level {
		case "course_professor":
			courseProfAgg[row.Subject+":"+row.CourseNumber+":"+row.Instructor] = agg
		case "course":
			courseAgg[row.Subject+":"+row.CourseNumber] = agg
		case "professor":
			professorAgg[row.Instructor] = agg
		case "subject":
			subjectAgg[row.Subject] = agg
		}
	}

	s.mu.Lock()
	s.subjectMap = subjectMap
	s.instructorMap = instructorMap
	s.courseProfAgg = courseProfAgg
	s.courseAgg = courseAgg
	s.professorAgg = professorAgg
	s.subjectAgg = subjectAgg
	s.loaded = true
	s.mu.Unlock()

	slog.Info("Grade data loaded into memory",
		"subject_mappings", len(subjectMap),
		"instructor_mappings", len(instructorMap),
		"course_prof_aggs", len(courseProfAgg),
		"course_aggs", len(courseAgg),
		"professor_aggs", len(professorAgg),
		"subject_aggs", len(subjectAgg),
	)
	return nil
}

// IsLoaded returns true if grade data is loaded into memory.
func (s *Service) IsLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loaded
}

// LookupCourseGPA returns the course-level average GPA for a Banner course.
func (s *Service) LookupCourseGPA(bannerSubject, courseNumber string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gradeSubject := s.mapSubject(bannerSubject)
	key := gradeSubject + ":" + courseNumber
	if agg, ok := s.courseAgg[key]; ok {
		return agg.GPA, true
	}
	return 0, false
}

// LookupSectionGPA returns the GPA for a specific section (course+instructor combo).
// Falls back to course-level average if no course+prof aggregate exists.
// Returns the source: "course_professor", "course", or "" (no data).
func (s *Service) LookupSectionGPA(bannerSubject, courseNumber, bannerInstructor string) (float64, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gradeSubject := s.mapSubject(bannerSubject)
	gradeInstructor := s.mapInstructor(bannerInstructor)

	// Try course+professor
	if gradeInstructor != "" {
		key := gradeSubject + ":" + courseNumber + ":" + gradeInstructor
		if agg, ok := s.courseProfAgg[key]; ok {
			return agg.GPA, "course_professor"
		}
	}

	// Fallback to course average
	key := gradeSubject + ":" + courseNumber
	if agg, ok := s.courseAgg[key]; ok {
		return agg.GPA, "course"
	}

	return 0, ""
}

// GetAggregate returns a specific aggregate by level and key fields.
func (s *Service) GetAggregate(level, subject, courseNumber, instructor string) *GradeAggregate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch level {
	case "course_professor":
		return s.courseProfAgg[subject+":"+courseNumber+":"+instructor]
	case "course":
		return s.courseAgg[subject+":"+courseNumber]
	case "professor":
		return s.professorAgg[instructor]
	case "subject":
		return s.subjectAgg[subject]
	}
	return nil
}

// mapSubject translates a Banner subject to the grade-data subject.
// Returns the input unchanged if no mapping exists.
func (s *Service) mapSubject(bannerSubject string) string {
	if mapped, ok := s.subjectMap[bannerSubject]; ok {
		return mapped
	}
	return bannerSubject
}

// mapInstructor translates a Banner instructor name to the grade-data name.
// Returns empty string if no mapping exists.
func (s *Service) mapInstructor(bannerName string) string {
	if mapped, ok := s.instructorMap[bannerName]; ok {
		return mapped
	}
	return ""
}
