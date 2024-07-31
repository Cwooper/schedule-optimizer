package models

type ScheduleRequest struct {
	Courses []string // Course Names
	Forced  []string // Forced Courses
	Term    string   // Term to find Courses
	Min     int      // Minimum Courses per Schedule
	Max     int      // Maximum Courses per Schedul
}
