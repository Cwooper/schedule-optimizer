package grades

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"schedule-optimizer/internal/store"
)

// computeMappings joins grade_rows with sections+instructors on (term, CRN)
// to auto-discover subject and instructor name mappings.
func computeMappings(ctx context.Context, db *sql.DB, queries *store.Queries) error {
	joinRows, err := queries.GetGradeBannerJoinData(ctx)
	if err != nil {
		return err
	}

	subjectCounts := make(map[string]map[string]int)    // banner_subject -> grade_subject -> count
	instructorCounts := make(map[string]map[string]int)  // banner_name -> grade_name -> count

	for _, row := range joinRows {
		bannerSubject := row.BannerSubject
		gradeSubject := row.GradeSubject

		// Subject mapping (always record, including identity and empty-professor rows)
		if subjectCounts[bannerSubject] == nil {
			subjectCounts[bannerSubject] = make(map[string]int)
		}
		subjectCounts[bannerSubject][gradeSubject]++

		// Instructor mapping (skip rows with no professor data)
		gradeName := strings.TrimSpace(row.GradeProfessor)
		if gradeName == "" {
			continue
		}
		if !row.BannerInstructor.Valid || row.BannerInstructor.String == "" {
			continue
		}
		bannerName := html.UnescapeString(row.BannerInstructor.String)

		if instructorCounts[bannerName] == nil {
			instructorCounts[bannerName] = make(map[string]int)
		}
		instructorCounts[bannerName][gradeName]++
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	txQueries := store.New(tx)

	// Upsert subject mappings: pick the grade_subject with the highest count
	var subjectMappings int
	for bannerSubject, candidates := range subjectCounts {
		bestSubject, bestCount := pickBest(candidates)
		if err := txQueries.UpsertSubjectMapping(ctx, store.UpsertSubjectMappingParams{
			BannerSubject: bannerSubject,
			GradeSubject:  bestSubject,
			MatchCount:    int64(bestCount),
		}); err != nil {
			return err
		}
		subjectMappings++
	}

	// Upsert instructor mappings: pick the grade_name with the highest count
	var instructorMappings int
	for bannerName, candidates := range instructorCounts {
		bestName, bestCount := pickBest(candidates)
		if err := txQueries.UpsertInstructorMapping(ctx, store.UpsertInstructorMappingParams{
			BannerName: bannerName,
			GradeName:  bestName,
			MatchCount: int64(bestCount),
		}); err != nil {
			return err
		}
		instructorMappings++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("Grade mappings computed",
		"subject_mappings", subjectMappings,
		"instructor_mappings", instructorMappings,
		"crn_joins", len(joinRows),
	)
	return nil
}

// pickBest returns the key with the highest count from a candidate map.
func pickBest(candidates map[string]int) (string, int) {
	var best string
	var bestCount int
	for k, v := range candidates {
		if v > bestCount {
			best = k
			bestCount = v
		}
	}
	return best, bestCount
}
