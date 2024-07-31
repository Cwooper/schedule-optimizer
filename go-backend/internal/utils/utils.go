// Package utils for "modular" hard-coding and re-used functions
package utils

import "math"

const (
	ROUND       = 2    // Round to x decimal places
	MINS_IN_DAY = 1440 // Number of minutes in a day (60 * 24)

	MIN_COURSES = 1 // Minimum courses in a single schedule
	MAX_COURSES = 7	// Maximum courses in a single schedule
)

// Helper function to round float64 to n decimal places
func Round(x float64) float64 {
	pow := math.Pow(10, float64(ROUND))
	return math.Round(x*pow) / pow
}
