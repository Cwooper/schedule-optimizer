// Package cache provides in-memory caching for schedule generation.
// Only terms schedules are generated for are loaded.
// Historical/search queries go directly to SQLite via the search service.
package cache

import (
	"context"
	"database/sql"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"schedule-optimizer/internal/calendar"
	"schedule-optimizer/internal/meeting"
	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/store"
)

// Course represents a fully hydrated course section for schedule generation.
// Includes all meeting times needed for conflict detection.
type Course struct {
	ID                  int64           `json:"id"`
	Term                string          `json:"term"`
	CRN                 string          `json:"crn"`
	Subject             string          `json:"subject"`
	SubjectDescription  string          `json:"subjectDescription,omitempty"`
	CourseNumber        string          `json:"courseNumber"`
	Title               string          `json:"title"`
	Credits             int             `json:"credits"`
	Instructor          string          `json:"instructor,omitempty"`
	InstructorEmail     string          `json:"instructorEmail,omitempty"`
	Enrollment          int             `json:"enrollment"`
	MaxEnrollment       int             `json:"maxEnrollment"`
	SeatsAvailable      int             `json:"seatsAvailable"`
	WaitCount           int             `json:"waitCount"`
	IsOpen              bool            `json:"isOpen"`
	InstructionalMethod string          `json:"instructionalMethod,omitempty"`
	MeetingTimes        []meeting.Time  `json:"meetingTimes"`
	Final               *calendar.Final `json:"final,omitempty"`
	GPA                 float64         `json:"gpa,omitempty"`
	GPASource           string          `json:"gpaSource,omitempty"` // "course_professor", "course", ""
	PassRate            *float64        `json:"passRate,omitempty"`
}

// TermData holds all courses for a single term, indexed for fast access. It
// is never modified once loaded: a reload builds a new one, so a request that
// holds one reads a whole term throughout.
type TermData struct {
	Term         string
	LoadedAt     time.Time
	Courses      map[string]*Course   // CRN -> Course
	BySubject    map[string][]*Course // Subject -> Courses
	ByCourseCode map[string][]*Course // "CSCI:247" -> Courses (all sections)
}

// ScheduleCache holds course data for the terms schedules are generated for.
// A term loads on its first request and stays until a job rewrites its data
// and invalidates it.
type ScheduleCache struct {
	mu    sync.RWMutex
	terms map[string]*TermData
	// generation counts invalidations. A load installs its result only if
	// none happened since it started, since it may have read the rows the
	// invalidation replaced.
	generation   uint64
	queries      *store.Queries
	gradeService *grades.Service
	loadGroup    singleflight.Group // Deduplicates concurrent loads
}

// NewScheduleCache creates a new schedule cache.
func NewScheduleCache(queries *store.Queries, gradeService *grades.Service) *ScheduleCache {
	return &ScheduleCache{
		terms:        make(map[string]*TermData),
		queries:      queries,
		gradeService: gradeService,
	}
}

// Term returns a term's data, loading it if it is not cached. Concurrent
// requests for a term share one load.
func (c *ScheduleCache) Term(ctx context.Context, term string) (*TermData, error) {
	c.mu.RLock()
	td, generation := c.terms[term], c.generation
	c.mu.RUnlock()
	if td != nil {
		return td, nil
	}
	// Keyed by generation so a request after an invalidation does not join a
	// load that started before it. The load outlives a canceled request
	// because other requests may be waiting on it.
	key := term + "@" + strconv.FormatUint(generation, 10)
	v, err, _ := c.loadGroup.Do(key, func() (any, error) {
		return c.load(context.WithoutCancel(ctx), term, generation)
	})
	if err != nil {
		return nil, err
	}
	return v.(*TermData), nil
}

// LoadTerm loads a term into the cache, replacing any cached copy.
func (c *ScheduleCache) LoadTerm(ctx context.Context, term string) error {
	c.mu.RLock()
	generation := c.generation
	c.mu.RUnlock()
	_, err := c.load(ctx, term, generation)
	return err
}

// Invalidate drops a term so its next request reloads it from the database.
func (c *ScheduleCache) Invalidate(term string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.generation++
	delete(c.terms, term)
}

// InvalidateAll drops every cached term.
func (c *ScheduleCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.generation++
	clear(c.terms)
}

// load reads a term from the database and caches it if no invalidation
// happened since generation and its finals loaded. It returns the data either
// way: it is a whole term, only possibly older or missing finals.
func (c *ScheduleCache) load(ctx context.Context, term string, generation uint64) (*TermData, error) {
	start := time.Now()

	sections, err := c.queries.GetSectionsWithInstructorByTerm(ctx, term)
	if err != nil {
		return nil, err
	}

	meetingTimes, err := c.queries.GetMeetingTimesByTerm(ctx, term)
	if err != nil {
		return nil, err
	}

	finals, finalsOK := calendar.SectionFinals(ctx, c.queries, []string{term})

	// Index by section ID to avoid N+1 queries when building courses
	meetingsBySection := make(map[int64][]*store.MeetingTime)
	for _, m := range meetingTimes {
		meetingsBySection[m.SectionID] = append(meetingsBySection[m.SectionID], m)
	}

	termData := &TermData{
		Term:         term,
		LoadedAt:     time.Now(),
		Courses:      make(map[string]*Course, len(sections)),
		BySubject:    make(map[string][]*Course),
		ByCourseCode: make(map[string][]*Course),
	}

	for _, s := range sections {
		course := &Course{
			ID:                  s.ID,
			Term:                s.Term,
			CRN:                 s.Crn,
			Subject:             s.Subject,
			SubjectDescription:  nullString(s.SubjectDescription),
			CourseNumber:        s.CourseNumber,
			Title:               s.Title,
			Credits:             int(nullInt(s.CreditHoursLow)),
			Instructor:          nullString(s.InstructorName),
			InstructorEmail:     nullString(s.InstructorEmail),
			Enrollment:          int(nullInt(s.Enrollment)),
			MaxEnrollment:       int(nullInt(s.MaxEnrollment)),
			SeatsAvailable:      int(nullInt(s.SeatsAvailable)),
			WaitCount:           int(nullInt(s.WaitCount)),
			IsOpen:              nullInt(s.IsOpen) == 1,
			InstructionalMethod: nullString(s.InstructionalMethod),
			MeetingTimes:        meeting.FromStoreAll(meetingsBySection[s.ID]),
		}
		course.Final = finals[term].FinalFor(s.ScheduleType.String, course.MeetingTimes)

		// Look up GPA and pass rate from grade data
		if c.gradeService != nil && c.gradeService.IsLoaded() {
			course.GPA, course.PassRate, course.GPASource = c.gradeService.LookupSectionGPA(
				course.Subject, course.CourseNumber, course.Instructor,
			)
		}

		// Index by CRN, subject, and course code
		termData.Courses[course.CRN] = course
		termData.BySubject[course.Subject] = append(termData.BySubject[course.Subject], course)
		courseCode := course.Subject + ":" + course.CourseNumber
		termData.ByCourseCode[courseCode] = append(termData.ByCourseCode[courseCode], course)
	}

	// Without its finals the term is served but not kept, so the next
	// request retries rather than going without finals until an invalidation.
	if finalsOK {
		c.mu.Lock()
		if c.generation == generation {
			c.terms[term] = termData
		}
		c.mu.Unlock()
	}

	slog.Info("Loaded term into schedule cache",
		"term", term,
		"courses", len(termData.Courses),
		"subjects", len(termData.BySubject),
		"duration", time.Since(start),
	)

	return termData, nil
}

// GetCourse returns a course by CRN for a specific term.
func (c *ScheduleCache) GetCourse(term, crn string) (*Course, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	termData, ok := c.terms[term]
	if !ok {
		return nil, false
	}
	course, ok := termData.Courses[crn]
	return course, ok
}

// GetCoursesByCRNs returns multiple courses by their CRNs for schedule generation.
func (c *ScheduleCache) GetCoursesByCRNs(term string, crns []string) []*Course {
	c.mu.RLock()
	defer c.mu.RUnlock()

	termData, ok := c.terms[term]
	if !ok {
		return nil
	}

	courses := make([]*Course, 0, len(crns))
	for _, crn := range crns {
		if course, ok := termData.Courses[crn]; ok {
			courses = append(courses, course)
		}
	}
	return courses
}

// GetCoursesBySubject returns all courses for a subject in a term.
func (c *ScheduleCache) GetCoursesBySubject(term, subject string) []*Course {
	c.mu.RLock()
	defer c.mu.RUnlock()

	termData, ok := c.terms[term]
	if !ok {
		return nil
	}
	return termData.BySubject[subject]
}

// GetCoursesByCourseCode returns all sections for a course code (e.g., "CSCI:241").
func (c *ScheduleCache) GetCoursesByCourseCode(term, courseCode string) []*Course {
	c.mu.RLock()
	defer c.mu.RUnlock()

	termData, ok := c.terms[term]
	if !ok {
		return nil
	}
	return termData.ByCourseCode[courseCode]
}

// GetAllCourses returns all courses for a term.
func (c *ScheduleCache) GetAllCourses(term string) []*Course {
	c.mu.RLock()
	defer c.mu.RUnlock()

	termData, ok := c.terms[term]
	if !ok {
		return nil
	}

	courses := make([]*Course, 0, len(termData.Courses))
	for _, course := range termData.Courses {
		courses = append(courses, course)
	}
	return courses
}

// IsTermLoaded checks if a term is loaded in the cache.
func (c *ScheduleCache) IsTermLoaded(term string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.terms[term]
	return ok
}

// GetActiveTerms returns the cached terms, sorted.
func (c *ScheduleCache) GetActiveTerms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Sorted(maps.Keys(c.terms))
}

// helpers for sql.Null types
func nullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullInt(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}
