// Package generator provides schedule generation with bitmask-based conflict detection.
package generator

import "schedule-optimizer/internal/cache"

// Constants for schedule generation limits.
const (
	MaxSchedulesToReturn   = 2000  // Limit on schedules returned to client
	MaxSchedulesToGenerate = 20000 // Safety limit to prevent runaway generation
	MaxInputCourses        = 13
	DefaultMinCourses      = 1
	DefaultMaxCourses      = 8
)

// GenerateRequest contains the parameters for schedule generation.
type GenerateRequest struct {
	Term         string        `json:"term" binding:"required"`
	Courses      []string      `json:"courses" binding:"required"`
	ForcedCRNs   []string      `json:"forcedCrns,omitempty"`
	BlockedTimes []BlockedTime `json:"blockedTimes,omitempty"`
	MinCourses   int           `json:"minCourses"`
	MaxCourses   int           `json:"maxCourses"`
}

// BlockedTime represents a single time block the user cannot attend.
type BlockedTime struct {
	Day       int    `json:"day"`       // 0=Mon, 1=Tue, 2=Wed, 3=Thu, 4=Fri
	StartTime string `json:"startTime"` // "0900" format
	EndTime   string `json:"endTime"`   // "1700" format
}

// GenerateResponse contains the results of schedule generation.
type GenerateResponse struct {
	Schedules     []Schedule      `json:"schedules"`
	Asyncs        []*cache.Course `json:"asyncs,omitempty"`
	CourseResults []CourseResult  `json:"courseResults"`
	Stats         GenerateStats   `json:"stats"`
}

// CourseResult reports what happened to each requested course.
type CourseResult struct {
	Name   string       `json:"name"`
	Status CourseStatus `json:"status"`
	Count  int          `json:"count,omitempty"`
}

// CourseStatus indicates the outcome of looking up a requested course.
type CourseStatus string

const (
	StatusFound      CourseStatus = "found"       // Has scheduleable sections
	StatusAsyncOnly  CourseStatus = "async_only"  // Only async/TBD sections exist
	StatusBlocked    CourseStatus = "blocked"     // All sections filtered by user's blocked times
	StatusNotOffered CourseStatus = "not_offered" // Valid course, not offered this term
	StatusNotExists  CourseStatus = "not_exists"  // Course code doesn't exist at all
)

// Schedule represents a valid combination of course sections.
type Schedule struct {
	Courses []*cache.Course `json:"courses"`
	Score   float64         `json:"score"`
	Weights []Weight        `json:"weights"`
}

// Weight represents a single scoring component.
type Weight struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// GenerateStats contains timing and count information about the generation.
type GenerateStats struct {
	TotalGenerated int     `json:"totalGenerated"`
	TimeMs         float64 `json:"timeMs"`
}

// courseGroup holds all scheduleable sections for a single course.
type courseGroup struct {
	courseKey string
	sections  []*sectionData
}

// sectionData pairs a course with its precomputed time mask.
type sectionData struct {
	course *cache.Course
	mask   TimeMask
}
