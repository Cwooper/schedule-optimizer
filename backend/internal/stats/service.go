// Package stats composes sub-services for statistics and analytics.
package stats

import "schedule-optimizer/internal/stats/grades"

// Service coordinates statistics sub-services.
type Service struct {
	Grades *grades.Service
}

// NewService creates a stats service with the provided sub-services.
func NewService(gradeService *grades.Service) *Service {
	return &Service{Grades: gradeService}
}
