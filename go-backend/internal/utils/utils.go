// Package utils for "modular" hard-coding and re-used functions
package utils

import (
	"math"
	"os"
	"path/filepath"
)

const (
	ROUND       = 2    // Round to x decimal places
	MINS_IN_DAY = 1440 // Number of minutes in a day (60 * 24)

	MIN_COURSES = 1 // Minimum courses in a single schedule
	MAX_COURSES = 7 // Maximum courses in a single schedule

	MAX_SUBJECT_WAIT = 30 // Days
	MAX_COURSE_WAIT  = 2  // Days
	MAX_TERM_WAIT    = 10 // Days

	TIME_FORMAT = "2006-01-02 15:04:05"

	URL        = "https://web4u.banner.wwu.edu/pls/wwis/wwskcfnd.TimeTable"
	SUBJECT_ID = "subj" // HTML select id name
	TERM_ID    = "term" // HTML term id name
)

var DataDirectory string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	DataDirectory = filepath.Join(wd, "data")

	// Ensure that the directory exists
	if err := os.MkdirAll(DataDirectory, os.ModePerm); err != nil {
		panic(err)
	}
}

// Helper function to round float64 to n decimal places
func Round(x float64) float64 {
	pow := math.Pow(10, float64(ROUND))
	return math.Round(x*pow) / pow
}
