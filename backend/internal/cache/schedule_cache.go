// Package cache provides in-memory caching for schedule generation.
// Only "active" terms (current + upcoming) are loaded for fast schedule building.
// Historical/search queries go directly to SQLite via the search service.
package cache

import (
	"context"
	"database/sql"
	"log/slog"
	"slices"
	"sync"
	"time"

	"schedule-optimizer/internal/store"
)

// Course represents a fully hydrated course section for schedule generation.
// Includes all meeting times needed for conflict detection.
type Course struct {
	ID                  int64         `json:"id"`
	Term                string        `json:"term"`
	CRN                 string        `json:"crn"`
	Subject             string        `json:"subject"`
	SubjectDescription  string        `json:"subjectDescription,omitempty"`
	CourseNumber        string        `json:"courseNumber"`
	Title               string        `json:"title"`
	Credits             int           `json:"credits"`
	Instructor          string        `json:"instructor,omitempty"`
	InstructorEmail     string        `json:"instructorEmail,omitempty"`
	Enrollment          int           `json:"enrollment"`
	MaxEnrollment       int           `json:"maxEnrollment"`
	SeatsAvailable      int           `json:"seatsAvailable"`
	WaitCount           int           `json:"waitCount"`
	IsOpen              bool          `json:"isOpen"`
	InstructionalMethod string        `json:"instructionalMethod,omitempty"`
	MeetingTimes        []MeetingTime `json:"meetingTimes"`
}

// MeetingTime represents when and where a course meets.
type MeetingTime struct {
	Days      [7]bool `json:"days"` // [Sun, Mon, Tue, Wed, Thu, Fri, Sat]
	StartTime string  `json:"startTime,omitempty"`
	EndTime   string  `json:"endTime,omitempty"`
	Building  string  `json:"building,omitempty"`
	Room      string  `json:"room,omitempty"`
}

// TermData holds all courses for a single term, indexed for fast access.
type TermData struct {
	Term         string
	LoadedAt     time.Time
	Courses      map[string]*Course   // CRN -> Course
	BySubject    map[string][]*Course // Subject -> Courses
	ByCourseCode map[string][]*Course // "CSCI:247" -> Courses (all sections)
}

// ScheduleCache holds course data for active terms used in schedule generation.
type ScheduleCache struct {
	mu          sync.RWMutex
	terms       map[string]*TermData
	activeTerms []string
	queries     *store.Queries
}

// NewScheduleCache creates a new schedule cache.
func NewScheduleCache(queries *store.Queries) *ScheduleCache {
	return &ScheduleCache{
		terms:   make(map[string]*TermData),
		queries: queries,
	}
}

// LoadTerm loads all course data for a term into memory using sqlc queries.
func (c *ScheduleCache) LoadTerm(ctx context.Context, term string) error {
	start := time.Now()

	sections, err := c.queries.GetSectionsWithInstructorByTerm(ctx, term)
	if err != nil {
		return err
	}

	meetingTimes, err := c.queries.GetMeetingTimesByTerm(ctx, term)
	if err != nil {
		return err
	}

	// Index by section ID to avoid N+1 queries when building courses
	meetingsBySection := make(map[int64][]*store.GetMeetingTimesByTermRow)
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
			MeetingTimes:        []MeetingTime{},
		}

		// Add meeting times from pre-fetched data
		for _, m := range meetingsBySection[s.ID] {
			mt := MeetingTime{
				Days: [7]bool{
					nullInt(m.Sunday) == 1,
					nullInt(m.Monday) == 1,
					nullInt(m.Tuesday) == 1,
					nullInt(m.Wednesday) == 1,
					nullInt(m.Thursday) == 1,
					nullInt(m.Friday) == 1,
					nullInt(m.Saturday) == 1,
				},
				StartTime: nullString(m.StartTime),
				EndTime:   nullString(m.EndTime),
				Building:  nullString(m.Building),
				Room:      nullString(m.Room),
			}
			course.MeetingTimes = append(course.MeetingTimes, mt)
		}

		// Index by CRN, subject, and course code
		termData.Courses[course.CRN] = course
		termData.BySubject[course.Subject] = append(termData.BySubject[course.Subject], course)
		courseCode := course.Subject + ":" + course.CourseNumber
		termData.ByCourseCode[courseCode] = append(termData.ByCourseCode[courseCode], course)
	}

	c.mu.Lock()
	c.terms[term] = termData
	// Track active terms
	found := slices.Contains(c.activeTerms, term)
	if !found {
		c.activeTerms = append(c.activeTerms, term)
	}
	c.mu.Unlock()

	slog.Info("Loaded term into schedule cache",
		"term", term,
		"courses", len(termData.Courses),
		"subjects", len(termData.BySubject),
		"duration", time.Since(start),
	)

	return nil
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

// GetActiveTerms returns the list of currently loaded terms.
func (c *ScheduleCache) GetActiveTerms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.activeTerms))
	copy(result, c.activeTerms)
	return result
}

// UnloadTerm removes a term from the cache to free memory.
func (c *ScheduleCache) UnloadTerm(term string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.terms, term)
	for i, t := range c.activeTerms {
		if t == term {
			c.activeTerms = append(c.activeTerms[:i], c.activeTerms[i+1:]...)
			break
		}
	}
	slog.Info("Unloaded term from schedule cache", "term", term)
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
