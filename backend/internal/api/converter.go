package api

import (
	"fmt"
	"html"
	"strings"

	"github.com/cwooper/schedule-optimizer/internal/models"
)

// apiToCourse converts a CourseData into our internal Course model
func apiToCourse(data CourseData) (*models.Course, error) {
	// Get primary instructor
	var instructor string
	for _, faculty := range data.Faculty {
		if faculty.PrimaryIndicator {
			instructor = html.UnescapeString(faculty.DisplayName)
			break
		}
	}

	// Create base course with HTML-decoded strings
	course := &models.Course{
		Subject:        fmt.Sprintf("%s %s", data.Subject, data.CourseNumber),
		Title:          html.UnescapeString(data.CourseTitle),
		Credits:        formatCredits(data.CreditHourLow, data.CreditHourHigh),
		Instructor:     instructor,
		Capacity:       data.MaximumEnrollment,
		Enrolled:       data.Enrollment,
		AvailableSeats: data.SeatsAvailable,
		GPA:            0, // Will be calculated separately
	}

	// Parse CRN
	crn, err := parseCRN(data.CourseReferenceNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid CRN %s: %w", data.CourseReferenceNumber, err)
	}
	course.CRN = crn

	// Process each meeting time into a session
	for _, meetingFaculty := range data.MeetingsFaculty {
		session, err := createSession(meetingFaculty)
		if err != nil {
			return nil, fmt.Errorf("error creating session: %w", err)
		}
		course.Sessions = append(course.Sessions, *session)
	}

	// Build the course string for searching
	course.CourseString = buildCourseString(course.Subject, course.Title, instructor)

	return course, nil
}

// formatCredits handles both fixed and variable credit hours
func formatCredits(low float64, high *float64) string {
	if high == nil || *high == low {
		return fmt.Sprintf("%.0f", low)
	}
	return fmt.Sprintf("%.0f-%.0f", low, *high)
}

// parseCRN converts the CRN string to an integer
func parseCRN(crn string) (int, error) {
	var result int
	_, err := fmt.Sscanf(crn, "%d", &result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// createSession converts meeting time data into a Session
func createSession(meetingFaculty MeetingFacultyData) (*models.Session, error) {
	mt := meetingFaculty.MeetingTime

	// Check if this is an async or TBD session
	isAsync := mt.MeetingType == "ASYNC" || mt.BeginTime == ""
	isTBD := mt.BeginTime == "TBD" || mt.EndTime == "TBD"

	// Convert times
	startTime, err := parseTime(mt.BeginTime)
	if err != nil {
		startTime = 0 // Reset on error
	}

	endTime, err := parseTime(mt.EndTime)
	if err != nil {
		endTime = 0 // Reset on error
	}

	// Build the days string
	days := buildDaysString(mt)

	// Build the location string
	location := buildLocationString(mt.Building, mt.Room)

	return &models.Session{
		Days:      days,
		StartTime: startTime,
		EndTime:   endTime,
		Location:  location,
		IsAsync:   isAsync,
		IsTimeTBD: isTBD,
	}, nil
}

func parseTime(timeStr string) (int, error) {
	if timeStr == "" || timeStr == "TBD" {
		return 0, fmt.Errorf("invalid time format")
	}

	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%02d%02d", &hour, &minute)
	if err != nil {
		return 0, err
	}

	return hour*100 + minute, nil
}

// buildDaysString creates the days string (e.g., "MWF")
func buildDaysString(mt MeetingTime) string {
	var days strings.Builder

	if mt.Monday {
		days.WriteString("M")
	}
	if mt.Tuesday {
		days.WriteString("T")
	}
	if mt.Wednesday {
		days.WriteString("W")
	}
	if mt.Thursday {
		days.WriteString("R")
	}
	if mt.Friday {
		days.WriteString("F")
	}

	return days.String()
}

func buildLocationString(building, room string) string {
	var location strings.Builder

	if building != "" {
		location.WriteString(building)
	}
	if room != "" {
		if location.Len() > 0 {
			location.WriteString(" ")
		}
		location.WriteString(room)
	}

	return location.String()
}

func buildCourseString(subject, title string, instructor string) string {
	parts := []string{
		strings.ToUpper(subject),
		strings.ToUpper(title),
		strings.ToUpper(instructor),
	}
	return strings.Join(parts, "|")
}
