package models

// Response Struct that is JSON serializable
type Response struct {
	// Always used
	Warnings []string // Something went wrong
	Errors   []string // Fatal error

	// Generating schedules
	Schedules []Schedule
	Asyncs    []Course

	// Searching for courses
	Courses []Course
}
