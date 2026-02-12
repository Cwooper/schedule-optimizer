package grades

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"

	"schedule-optimizer/internal/store"
)

var codeRegex = regexp.MustCompile(`^(.+)\s+(\d+[A-Za-z]?)$`)

const sheetName = "Export Worksheet"

// Column indices in the Excel file (0-indexed).
const (
	colTerm             = 0
	colCRN              = 1
	colCode             = 2
	colTitle            = 3
	colProfessor        = 4
	colStudentsEnrolled = 5
	colGradeCount       = 6
	colCntA             = 7
	colCntAM            = 8
	colCntBP            = 9
	colCntB             = 10
	colCntBM            = 11
	colCntCP            = 12
	colCntC             = 13
	colCntCM            = 14
	colCntDP            = 15
	colCntD             = 16
	colCntDM            = 17
	colCntF             = 18
	colCntW             = 19
	colCntP             = 20
	colCntNP            = 21
	colCntS             = 22
	colCntU             = 23
)

// importExcel reads the PRR Excel file and inserts rows into grade_rows.
func importExcel(ctx context.Context, db *sql.DB, path string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return fmt.Errorf("open excel file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("read sheet %q: %w", sheetName, err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("sheet %q has no data rows", sheetName)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	txQueries := store.New(tx)

	var imported, skipped int
	for i, row := range rows[1:] { // skip header
		if len(row) < colCntU+1 {
			skipped++
			continue
		}

		code := strings.TrimSpace(row[colCode])
		matches := codeRegex.FindStringSubmatch(code)
		if matches == nil {
			slog.Warn("Skipping grade row: code regex didn't match",
				"row", i+2, "code", code)
			skipped++
			continue
		}

		professor := strings.TrimSpace(row[colProfessor])

		params := store.InsertGradeRowParams{
			Term:             strings.TrimSpace(row[colTerm]),
			Crn:              strings.TrimSpace(row[colCRN]),
			Subject:          matches[1],
			CourseNumber:     matches[2],
			Title:            strings.TrimSpace(row[colTitle]),
			Professor:        professor,
			StudentsEnrolled: parseInt(row[colStudentsEnrolled]),
			GradeCount:       parseInt(row[colGradeCount]),
			CntA:             parseInt(row[colCntA]),
			CntAm:            parseInt(row[colCntAM]),
			CntBp:            parseInt(row[colCntBP]),
			CntB:             parseInt(row[colCntB]),
			CntBm:            parseInt(row[colCntBM]),
			CntCp:            parseInt(row[colCntCP]),
			CntC:             parseInt(row[colCntC]),
			CntCm:            parseInt(row[colCntCM]),
			CntDp:            parseInt(row[colCntDP]),
			CntD:             parseInt(row[colCntD]),
			CntDm:            parseInt(row[colCntDM]),
			CntF:             parseInt(row[colCntF]),
			CntW:             parseInt(row[colCntW]),
			CntP:             parseInt(row[colCntP]),
			CntNp:            parseInt(row[colCntNP]),
			CntS:             parseInt(row[colCntS]),
			CntU:             parseInt(row[colCntU]),
		}

		if err := txQueries.InsertGradeRow(ctx, params); err != nil {
			return fmt.Errorf("insert grade row %d: %w", i+2, err)
		}
		imported++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("Grade data import complete", "imported", imported, "skipped", skipped)
	return nil
}

func parseInt(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}
