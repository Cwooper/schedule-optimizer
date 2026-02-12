package jobs

import (
	"context"
	"log/slog"
	"time"

	"schedule-optimizer/internal/stats/grades"
)

// GradeImportJob imports grade data and computes mappings/aggregates on startup.
// Runs once on boot. All steps are idempotent.
type GradeImportJob struct {
	gradeService *grades.Service
	hasRun       bool
}

func NewGradeImportJob(gradeService *grades.Service) *GradeImportJob {
	return &GradeImportJob{gradeService: gradeService}
}

func (j *GradeImportJob) Name() string { return "grade-import" }

func (j *GradeImportJob) ShouldRun(_ time.Time) bool {
	return !j.hasRun
}

func (j *GradeImportJob) Run(ctx context.Context, _ time.Time) error {
	j.hasRun = true

	if err := j.gradeService.ImportAndCompute(ctx); err != nil {
		slog.Error("Grade import failed", "error", err)
		return err
	}

	return nil
}
