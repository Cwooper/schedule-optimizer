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

// accumulator collects grade counts for one aggregate bucket.
type accumulator struct {
	sections int
	students int
	a, am, bp, b, bm int
	cp, c, cm         int
	dp, d, dm, f       int
	w, p, np, s, u     int
}

func (acc *accumulator) addRow(row *store.GradeRow) {
	acc.sections++
	acc.students += int(row.GradeCount)
	acc.a += int(row.CntA)
	acc.am += int(row.CntAm)
	acc.bp += int(row.CntBp)
	acc.b += int(row.CntB)
	acc.bm += int(row.CntBm)
	acc.cp += int(row.CntCp)
	acc.c += int(row.CntC)
	acc.cm += int(row.CntCm)
	acc.dp += int(row.CntDp)
	acc.d += int(row.CntD)
	acc.dm += int(row.CntDm)
	acc.f += int(row.CntF)
	acc.w += int(row.CntW)
	acc.p += int(row.CntP)
	acc.np += int(row.CntNp)
	acc.s += int(row.CntS)
	acc.u += int(row.CntU)
}

func (acc *accumulator) toAggregate(key aggregateKey) *GradeAggregate {
	return &GradeAggregate{
		Level:        key.level,
		Subject:      key.subject,
		CourseNumber: key.courseNumber,
		Instructor:   key.instructor,
		Sections:     acc.sections,
		Students:     acc.students,
		CntA:         acc.a,
		CntAM:        acc.am,
		CntBP:        acc.bp,
		CntB:         acc.b,
		CntBM:        acc.bm,
		CntCP:        acc.cp,
		CntC:         acc.c,
		CntCM:        acc.cm,
		CntDP:        acc.dp,
		CntD:         acc.d,
		CntDM:        acc.dm,
		CntF:         acc.f,
		CntW:         acc.w,
		CntP:         acc.p,
		CntNP:        acc.np,
		CntS:         acc.s,
		CntU:         acc.u,
		GPA:          computeGPA(acc.a, acc.am, acc.bp, acc.b, acc.bm, acc.cp, acc.c, acc.cm, acc.dp, acc.d, acc.dm, acc.f),
		PassRate:     computePassRate(acc.s, acc.u, acc.p, acc.np),
	}
}

// computeAggregates loads all grade_rows and pre-computes aggregates at four levels.
func computeAggregates(ctx context.Context, db *sql.DB, queries *store.Queries) error {
	// Read grade rows outside the write transaction
	rows, err := queries.GetAllGradeRows(ctx)
	if err != nil {
		return err
	}

	buckets := make(map[aggregateKey]*accumulator)

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
	for key, acc := range buckets {
		agg := acc.toAggregate(key)

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
			CntA:          int64(agg.CntA),
			CntAm:         int64(agg.CntAM),
			CntBp:         int64(agg.CntBP),
			CntB:          int64(agg.CntB),
			CntBm:         int64(agg.CntBM),
			CntCp:         int64(agg.CntCP),
			CntC:          int64(agg.CntC),
			CntCm:         int64(agg.CntCM),
			CntDp:         int64(agg.CntDP),
			CntD:          int64(agg.CntD),
			CntDm:         int64(agg.CntDM),
			CntF:          int64(agg.CntF),
			CntW:          int64(agg.CntW),
			CntP:          int64(agg.CntP),
			CntNp:         int64(agg.CntNP),
			CntS:          int64(agg.CntS),
			CntU:          int64(agg.CntU),
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

func getBucket(buckets map[aggregateKey]*accumulator, key aggregateKey) *accumulator {
	if buckets[key] == nil {
		buckets[key] = &accumulator{}
	}
	return buckets[key]
}
