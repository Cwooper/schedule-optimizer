// Package models for re-usable objects
package models

import "strings"

// Holds one session worth of data (repeatable time range over days)
type Session struct {
	Days      string // Days of the week in format (e.g. "MTWRF")
	StartTime int    // Start time in 24-hour format (0000-2359)
	EndTime   int    // End time in 24-hour format (0000-2359)
	Location  string // Room or location
}

// Holds all of a single courses data
type Course struct {
	Subject        string    // Course Subject
	Title          string    // Course Title
	Instructor     string    // Course Instructor
	Credits        int       // Number of credits
	CRN            int       // Course Record Number
	Sessions       []Session // All sessions for this course (nil if asnyc)
	GPA            float64   // Average GPA for this course
	Capacity       int       // Maximum enrollment capacity
	Enrollment     int       // Number of currently enrolled students
	AvailableSeats int       // Number of seats available
	WaitlistCount  int       // Number of students on waitlist
	AdditionalFees string    // Any noted additional fees
	Restrictions   string    // Any noted restrictions
	Attributes     string    // Any noted attributes
	Prerequisites  string    // Any noted prerequisites
}

// Returns whether the courses conflict with one another
func (c Course) Conflicts(other Course) bool {
	// Conflicts on the same subject
	if c.Subject == other.Subject {
		return true
	}

	for _, s1 := range c.Sessions {
		for _, s2 := range other.Sessions {
			if sessionsConflict(s1, s2) {
				return true
			}
		}
	}

	return false
}

// sessionsConflict checks if two sessions conflict
func sessionsConflict(s1, s2 Session) bool {
	// Check if s2 constains the day in s1. If so, checks time overlap
	for _, day1 := range s1.Days {
		if strings.ContainsRune(s2.Days, day1) {
			if timesConflict(s1.StartTime, s1.EndTime, s2.StartTime, s2.EndTime) {
				return true
			}
		}
	}
	return false
}

// timesConflict checks if two time ranges overlap
func timesConflict(start1, end1, start2, end2 int) bool {
	return (((start1 >= start2) && (start1 <= end2)) || ((end1 >= start2) && (end1 <= end2)))
}
