package search

// SearchRequest contains all search parameters from the API request.
type SearchRequest struct {
	// Scope parameters (at most one should be set; if neither, searches all terms)
	Term string `form:"term"` // Specific term code (e.g., "202540")
	Year int    `form:"year"` // Academic year (e.g., 2025 = Fall 2024 through Summer 2025)

	// Filters (at least one required)
	Subject      string `form:"subject"`      // Exact match (e.g., "CSCI")
	CourseNumber string `form:"courseNumber"` // Supports wildcards: "2*", "201"
	Title        string `form:"title"`        // Token-based search (max 3 tokens)
	Instructor   string `form:"instructor"`   // Token-based search (max 3 tokens)

	// Additional filters
	OpenSeats  bool `form:"openSeats"`  // Only sections with available seats
	MinCredits *int `form:"minCredits"` // Minimum credit hours
	MaxCredits *int `form:"maxCredits"` // Maximum credit hours
}

// CourseInfo contains course-level data sent once per unique course code.
type CourseInfo struct {
	Subject      string `json:"subject"`
	CourseNumber string `json:"courseNumber"`
	Title        string `json:"title"`
	Credits      int    `json:"credits"`
	CreditsHigh  int    `json:"creditsHigh,omitempty"`
}

// MeetingTimeInfo represents a single meeting time for a section.
type MeetingTimeInfo struct {
	Days      []bool `json:"days"` // [Sun, Mon, Tue, Wed, Thu, Fri, Sat]
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Building  string `json:"building"`
	Room      string `json:"room"`
}

// SectionInfo contains section-level data sent once per unique term:CRN combination.
type SectionInfo struct {
	CRN             string            `json:"crn"`
	Term            string            `json:"term"`
	CourseKey       string            `json:"courseKey"`
	Instructor      string            `json:"instructor,omitempty"`
	InstructorEmail string            `json:"instructorEmail,omitempty"`
	Enrollment      int               `json:"enrollment"`
	MaxEnrollment   int               `json:"maxEnrollment"`
	SeatsAvailable  int               `json:"seatsAvailable"`
	WaitCount       int               `json:"waitCount"`
	IsOpen          bool              `json:"isOpen"`
	Campus          string            `json:"campus,omitempty"`
	ScheduleType    string            `json:"scheduleType,omitempty"`
	MeetingTimes    []MeetingTimeInfo `json:"meetingTimes"`
}

// CourseRef groups sections of a course with relevance score.
// SectionKeys are in format "term:crn" to uniquely identify sections across terms.
type CourseRef struct {
	CourseKey      string   `json:"courseKey"`
	SectionKeys    []string `json:"sectionKeys"`
	RelevanceScore float64  `json:"relevanceScore,omitempty"`
}

// SearchResponse is the normalized wire format for search results.
type SearchResponse struct {
	Courses  map[string]CourseInfo  `json:"courses"`
	Sections map[string]SectionInfo `json:"sections"`
	Results  []CourseRef            `json:"results"`
	Total    int                    `json:"total"`             // Number of courses returned
	Warning  string                 `json:"warning,omitempty"` // Set if results truncated
	Stats    SearchStats            `json:"stats"`
}

// SearchStats contains timing and count information about the search.
type SearchStats struct {
	TotalSections   int     `json:"totalSections"`             // Number of sections found
	TotalCourses    int     `json:"totalCourses"`              // Number of unique courses
	TimeMs          float64 `json:"timeMs"`                    // Search duration in milliseconds
	SectionLimitHit bool    `json:"sectionLimitHit,omitempty"` // True if section fetch limit was reached
}

// sectionRow is the internal representation for processing section data before grouping.
type sectionRow struct {
	ID              int64
	CRN             string
	Term            string
	Subject         string
	CourseNumber    string
	Title           string
	Credits         int
	CreditsHigh     int
	Enrollment      int
	MaxEnrollment   int
	SeatsAvailable  int
	WaitCount       int
	IsOpen          bool
	Campus          string
	ScheduleType    string
	Instructor      string
	InstructorEmail string
	RelevanceScore  float64
}

// Constants for validation
const (
	MaxCourseResults      = 200
	MaxSectionFetch       = 2000 // Safety limit for SQL query
	MaxTokens             = 3
	CourseWarningMessage  = "Showing maximum results. Try narrowing your search for better results."
	SectionWarningMessage = "Showing maximum sections. Try narrowing your search for better results."
)
