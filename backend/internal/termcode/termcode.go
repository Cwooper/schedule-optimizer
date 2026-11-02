// Package termcode handles WWU Banner term codes: YYYYQQ, where QQ is 10
// (Winter), 20 (Spring), 30 (Summer), or 40 (Fall). Winter 2026 is "202610".
package termcode

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Winter = 10
	Spring = 20
	Summer = 30
	Fall   = 40
)

func Parse(code string) (year, quarter int, err error) {
	if len(code) != 6 {
		return 0, 0, fmt.Errorf("invalid term code length: %q", code)
	}
	year, err = strconv.Atoi(code[:4])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid year in term code: %q", code)
	}
	quarter, err = strconv.Atoi(code[4:])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid quarter in term code: %q", code)
	}
	switch quarter {
	case Winter, Spring, Summer, Fall:
		return year, quarter, nil
	}
	return 0, 0, fmt.Errorf("invalid quarter %d in term code: %q", quarter, code)
}

func Make(year, quarter int) string {
	return fmt.Sprintf("%d%02d", year, quarter)
}

// FromName builds a code from a quarter name ("Winter", "fall") and year.
// Returns "" for an unknown quarter name.
func FromName(quarter string, year int) string {
	var q int
	switch strings.ToLower(quarter) {
	case "winter":
		q = Winter
	case "spring":
		q = Spring
	case "summer":
		q = Summer
	case "fall":
		q = Fall
	default:
		return ""
	}
	return Make(year, q)
}

// FromDescription converts a label like "Winter 2026" to "202610", or ""
// if it is not of that form.
func FromDescription(desc string) string {
	parts := strings.Fields(desc)
	if len(parts) != 2 {
		return ""
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return ""
	}
	return FromName(parts[0], year)
}

// Ordinal numbers terms consecutively (Winter 2025 + 1 = Spring 2025), so
// the difference of two ordinals counts the terms between them.
func Ordinal(code string) (int, error) {
	year, quarter, err := Parse(code)
	if err != nil {
		return 0, err
	}
	return year*4 + quarter/10 - 1, nil
}

// Current returns the term a date most likely falls in, by month alone.
// Use the calendar package where real term dates matter.
func Current(now time.Time) string {
	var quarter int
	switch month := now.Month(); {
	case month >= time.September:
		quarter = Fall
	case month >= time.June:
		quarter = Summer
	case month >= time.April:
		quarter = Spring
	default:
		quarter = Winter
	}
	return Make(now.Year(), quarter)
}

// AcademicYear returns the four terms of academic year N, which runs from
// Fall of N-1 through Summer of N.
func AcademicYear(year int) []string {
	return []string{
		Make(year-1, Fall),
		Make(year, Winter),
		Make(year, Spring),
		Make(year, Summer),
	}
}
