package grades

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"schedule-optimizer/internal/store"
)

// aggregateKey builds a map key for an aggregate bucket.
type aggregateKey struct {
	level        string
	subject      string
	courseNumber string
	instructor   string
}

// computeAggregates loads all grade_rows and pre-computes aggregates at four levels.
func computeAggregates(ctx context.Context, db *sql.DB, queries *store.Queries) error {
	// Read grade rows outside the write transaction
	rows, err := queries.GetAllGradeRows(ctx)
	if err != nil {
		return err
	}

	buckets := make(map[aggregateKey]*GradeAggregate)

	for _, row := range rows {
		prof := row.Professor

		// Course+Professor level
		if prof != "" {
			key := aggregateKey{"course_professor", row.Subject, row.CourseNumber, prof}
			getBucket(buckets, key).addRow(row)
		}

		// Course level (all professors combined)
		courseKey := aggregateKey{"course", row.Subject, row.CourseNumber, ""}
		getBucket(buckets, courseKey).addRow(row)

		// Professor level (all courses combined)
		if prof != "" {
			profKey := aggregateKey{"professor", "", "", prof}
			getBucket(buckets, profKey).addRow(row)
		}

		// Subject level
		subjectKey := aggregateKey{"subject", row.Subject, "", ""}
		getBucket(buckets, subjectKey).addRow(row)
	}

	// Delete + insert atomically in a single transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	txQueries := store.New(tx)

	if err := txQueries.DeleteAllGradeAggregates(ctx); err != nil {
		return err
	}

	var count int
	for _, agg := range buckets {
		agg.finish()

		var passRate sql.NullFloat64
		if agg.PassRate != nil {
			passRate = sql.NullFloat64{Float64: *agg.PassRate, Valid: true}
		}

		if err := txQueries.InsertGradeAggregate(ctx, store.InsertGradeAggregateParams{
			Level:         agg.Level,
			Subject:       agg.Subject,
			CourseNumber:  agg.CourseNumber,
			Instructor:    agg.Instructor,
			SectionsCount: int64(agg.Sections),
			StudentsCount: int64(agg.Students),
			CntA:          int64(agg.Counts[GradeA]),
			CntAm:         int64(agg.Counts[GradeAMinus]),
			CntBp:         int64(agg.Counts[GradeBPlus]),
			CntB:          int64(agg.Counts[GradeB]),
			CntBm:         int64(agg.Counts[GradeBMinus]),
			CntCp:         int64(agg.Counts[GradeCPlus]),
			CntC:          int64(agg.Counts[GradeC]),
			CntCm:         int64(agg.Counts[GradeCMinus]),
			CntDp:         int64(agg.Counts[GradeDPlus]),
			CntD:          int64(agg.Counts[GradeD]),
			CntDm:         int64(agg.Counts[GradeDMinus]),
			CntF:          int64(agg.Counts[GradeF]),
			CntW:          int64(agg.Counts[GradeW]),
			CntP:          int64(agg.Counts[GradeP]),
			CntNp:         int64(agg.Counts[GradeNP]),
			CntS:          int64(agg.Counts[GradeS]),
			CntU:          int64(agg.Counts[GradeU]),
			Gpa:           agg.GPA,
			PassRate:      passRate,
		}); err != nil {
			return err
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("Grade aggregates computed",
		"aggregates", count,
		"from_rows", len(rows),
	)
	return nil
}

func getBucket(buckets map[aggregateKey]*GradeAggregate, key aggregateKey) *GradeAggregate {
	if buckets[key] == nil {
		buckets[key] = &GradeAggregate{
			Level:        key.level,
			Subject:      key.subject,
			CourseNumber: key.courseNumber,
			Instructor:   key.instructor,
		}
	}
	return buckets[key]
}
