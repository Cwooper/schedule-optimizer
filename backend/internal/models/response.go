package models

// Response Struct that is JSON serializable
type Response struct {
	Schedules []Schedule
	Warnings  []string
	Errors    []string
}
