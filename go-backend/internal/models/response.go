package models

// Response Struct that is JSON serializable
type Response struct {
	Schedule []Schedule
	Warnings []string
	Errors	 []string
}
