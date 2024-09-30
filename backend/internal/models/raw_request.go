package models

type RawRequest struct {
	// Used by all
	Term    string   // Term to find Courses

	// Used by scheduler
	Courses []string // Course Names
	Forced  []string // Forced Courses
	Min     int      // Minimum Courses per Schedule
	Max     int      // Maximum Courses per Schedule

	// Used by Search
	SearchTerm string
}
