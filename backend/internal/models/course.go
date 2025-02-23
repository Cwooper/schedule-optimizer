// Package models for re-usable objects
package models

import (
	"fmt"
	"strings"
)

// Holds one session worth of data (repeatable time range over days)
type Session struct {
	Days       string // Days of the week in format (e.g. "MTWRF")
	StartTime  int    // Start time in 24-hour format (0000-2359)
	EndTime    int    // End time in 24-hour format (0000-2359)
	Instructor string // Session Instructor
	Location   string // Room or location
	IsAsync    bool   // True if the course is asynchronous
	IsTimeTBD  bool   // True if times or days are tbd
}

// Holds all of a single courses data
type Course struct {
	Subject        string    // Course Subject
	Title          string    // Course Title
	Credits        string    // Number of credits
	CRN            int       // Course Record Number
	Sessions       []Session // All sessions for this course
	GPA            float64   // Average GPA for this course
	Capacity       int       // Maximum enrollment capacity
	Enrolled       int       // Number of currently enrolled students
	AvailableSeats int       // Number of seats available
	AdditionalFees string    // Any noted additional fees
	Restrictions   string    // Any noted restrictions
	Attributes     string    // Any noted attributes
	Prerequisites  string    // Any noted prerequisites
	CourseString   string    // The course as a string for searching
}

// String provides a debugging string representation of a Session
func (s Session) String() string {
	return fmt.Sprintf("Days: %s, Start: %04d, End: %04d, Instructor: %s, Location: %s, IsAsync: %v, IsTimeTBD: %v",
		s.Days, s.StartTime, s.EndTime, s.Instructor, s.Location, s.IsAsync, s.IsTimeTBD)
}

// String provides a debugging string representation of a Course
func (c Course) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Subject: %s\n", c.Subject)
	fmt.Fprintf(&b, "Title: %s\n", c.Title)
	fmt.Fprintf(&b, "Credits: %s\n", c.Credits)
	fmt.Fprintf(&b, "CRN: %d\n", c.CRN)
	fmt.Fprintf(&b, "GPA: %.2f\n", c.GPA)
	fmt.Fprintf(&b, "Capacity: %d\n", c.Capacity)
	fmt.Fprintf(&b, "Enrolled: %d\n", c.Enrolled)
	fmt.Fprintf(&b, "AvailableSeats: %d\n", c.AvailableSeats)
	fmt.Fprintf(&b, "AdditionalFees: %s\n", c.AdditionalFees)
	fmt.Fprintf(&b, "Restrictions: %s\n", c.Restrictions)
	fmt.Fprintf(&b, "Attributes: %s\n", c.Attributes)
	fmt.Fprintf(&b, "Prerequisites: %s\n", c.Prerequisites)
	fmt.Fprintf(&b, "CourseString: %s\n", c.CourseString)
	fmt.Fprintf(&b, "Sessions:\n")
	for _, s := range c.Sessions {
		fmt.Fprintf(&b, "  %s\n", s)
	}

	return b.String()
}

// Initialize fully empty course
func NewCourse() *Course {
	return &Course{
		Subject:        "",
		Title:          "",
		Credits:        "",
		CRN:            0,
		Sessions:       nil,
		GPA:            0.0,
		Capacity:       0,
		Enrolled:       0,
		AvailableSeats: 0,
		AdditionalFees: "",
		Restrictions:   "",
		Attributes:     "",
		Prerequisites:  "",
		CourseString:   "",
	}
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
	// Two courses can't be known to conflict if they are online or tbd
	if s1.IsAsync || s1.IsTimeTBD || s2.IsAsync || s2.IsTimeTBD {
		return true
	}

	// Check if s2 contains the day in s1. If so, checks time overlap
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
	return start1 < end2 && start2 < end1
}

// Returns true if a course has any tbd or async times.
func HasAsyncOrTBD(course Course) bool {
	for _, session := range course.Sessions {
		if session.IsAsync || session.IsTimeTBD {
			return true
		}
	}
	return false
}
