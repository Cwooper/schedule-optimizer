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
	cache        CacheInvalidator
	hasRun       bool
}

func NewGradeImportJob(gradeService *grades.Service, cache CacheInvalidator) *GradeImportJob {
	return &GradeImportJob{gradeService: gradeService, cache: cache}
}

func (j *GradeImportJob) Name() string { return "grade-import" }

func (j *GradeImportJob) ShouldRun(context.Context, time.Time) bool {
	return !j.hasRun
}

func (j *GradeImportJob) Run(ctx context.Context, _ time.Time) error {
	j.hasRun = true

	if err := j.gradeService.ImportAndCompute(ctx); err != nil {
		slog.Error("Grade import failed", "error", err)
		return err
	}
	// A term cached before the import has no GPAs.
	j.cache.InvalidateAll()

	return nil
}
