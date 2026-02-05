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

	// Pagination
	Limit  int `form:"limit"`  // Max results (default 50, max 200)
	Offset int `form:"offset"` // Skip N results
}

// SearchResponse contains search results with pagination metadata.
type SearchResponse struct {
	Sections []*SectionResult `json:"sections"`
	Total    int              `json:"total"`             // Number of results returned
	HasMore  bool             `json:"hasMore"`           // More results available
	Warning  string           `json:"warning,omitempty"` // Warning for broad queries
}

// SectionResult represents a single section in search results.
type SectionResult struct {
	ID             int64   `json:"id"`
	CRN            string  `json:"crn"`
	Term           string  `json:"term"`
	Subject        string  `json:"subject"`
	CourseNumber   string  `json:"courseNumber"`
	Title          string  `json:"title"`
	Credits        int     `json:"credits"`
	CreditsHigh    int     `json:"creditsHigh,omitempty"` // For variable credit courses
	Enrollment     int     `json:"enrollment"`
	MaxEnrollment  int     `json:"maxEnrollment"`
	SeatsAvailable int     `json:"seatsAvailable"`
	WaitCount      int     `json:"waitCount"`
	IsOpen         bool    `json:"isOpen"`
	Campus         string  `json:"campus,omitempty"`
	ScheduleType   string  `json:"scheduleType,omitempty"`

	// Instructor info (primary instructor only)
	Instructor      string `json:"instructor,omitempty"`
	InstructorEmail string `json:"instructorEmail,omitempty"`

	// Meeting times (fetched separately)
	MeetingTimes []MeetingTimeInfo `json:"meetingTimes,omitempty"`

	// Relevance score (only for all-time searches)
	RelevanceScore float64 `json:"relevanceScore,omitempty"`
}

// MeetingTimeInfo represents a single meeting time for a section.
type MeetingTimeInfo struct {
	Days      []bool `json:"days"`      // [Sun, Mon, Tue, Wed, Thu, Fri, Sat]
	StartTime string `json:"startTime"` // "0900" format
	EndTime   string `json:"endTime"`   // "0950" format
	Building  string `json:"building,omitempty"`
	Room      string `json:"room,omitempty"`
}

// Constants for validation
const (
	DefaultLimit   = 50
	MaxLimit       = 200
	MaxTokens      = 3
	WarningMessage = "Showing maximum results. Try narrowing your search for better results."
)
