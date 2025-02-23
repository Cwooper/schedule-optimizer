package models

import (
	"fmt"
	"strings"
)

// Session represents a meeting time for a course
type Session struct {
	Days      string // Days of the week in format (e.g. "MTWRF")
	StartTime int    // Start time in 24-hour format (0000-2359)
	EndTime   int    // End time in 24-hour format (0000-2359)
	Location  string // Room or location
	IsAsync   bool   // True if the course is asynchronous
	IsTimeTBD bool   // True if times or days are tbd
}

// Course represents a single course offering
type Course struct {
	Subject        string    // Course Subject (e.g., "CSCI 141")
	Title          string    // Course Title
	Credits        string    // Number of credits
	CRN            int       // Course Record Number
	Instructor     string    // Primary instructor
	Sessions       []Session // All sessions for this course
	GPA            float64   // Average GPA for this course
	Capacity       int       // Maximum enrollment capacity
	Enrolled       int       // Number of currently enrolled students
	AvailableSeats int       // Number of seats available
	CourseString   string    // The course as a string for searching
}

// Initialize fully empty course
func NewCourse() *Course {
	return &Course{
		Subject:        "",
		Title:          "",
		Credits:        "",
		CRN:            0,
		Instructor:     "",
		Sessions:       nil,
		GPA:            0.0,
		Capacity:       0,
		Enrolled:       0,
		AvailableSeats: 0,
		CourseString:   "",
	}
}

// String provides a debugging string representation of a Session
func (s Session) String() string {
	return fmt.Sprintf("Days: %s, Start: %04d, End: %04d, Location: %s, IsAsync: %v, IsTimeTBD: %v",
		s.Days, s.StartTime, s.EndTime, s.Location, s.IsAsync, s.IsTimeTBD)
}

// String provides a debugging string representation of a Course
func (c Course) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Subject: %s\n", c.Subject)
	fmt.Fprintf(&b, "Title: %s\n", c.Title)
	fmt.Fprintf(&b, "Credits: %s\n", c.Credits)
	fmt.Fprintf(&b, "CRN: %d\n", c.CRN)
	fmt.Fprintf(&b, "Instructor: %s\n", c.Instructor)
	fmt.Fprintf(&b, "GPA: %.2f\n", c.GPA)
	fmt.Fprintf(&b, "Capacity: %d\n", c.Capacity)
	fmt.Fprintf(&b, "Enrolled: %d\n", c.Enrolled)
	fmt.Fprintf(&b, "AvailableSeats: %d\n", c.AvailableSeats)
	fmt.Fprintf(&b, "CourseString: %s\n", c.CourseString)
	fmt.Fprintf(&b, "Sessions:\n")
	for _, s := range c.Sessions {
		fmt.Fprintf(&b, "  %s\n", s)
	}

	return b.String()
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
