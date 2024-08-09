package models

// Response Struct that is JSON serializable
type Response struct {
	Schedules []Schedule
	Asyncs    []Course
	Warnings  []string
	Errors    []string
}
