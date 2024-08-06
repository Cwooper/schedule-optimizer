// Package gpa for processing and updating protobufs with gpas
package gpa

import (
	"schedule-optimizer/internal/models"
)

// Adds the GPA variable to each course
func GenerateGPA(courseList *[]models.Course) {
	// Store a map of each CourseKey that has been processed
	// If CourseKey is in the map, use that GPA
	// If the course isn't in the map, then process the gpa, and add course to map
	// 

}
