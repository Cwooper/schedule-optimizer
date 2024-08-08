package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwooper/schedule-optimizer/internal/utils"
)

type Professors map[string]float64
type Subjects map[string]float64
type CourseGPAs map[string]float64

func CourseKey(subject, professor string) string {
	return subject + "|" + professor
}

func simplifyName(name string) string {
	parts := strings.Fields(name)
	if len(parts) < 2 {
		return name // Return original name if it can't be simplified
	}
	return parts[0] + " " + parts[len(parts)-1] // First name + Last name
}

type GPAData struct {
	Subjects   Subjects
	Professors Professors
	CourseGPAs CourseGPAs
}

func NewGPAData() *GPAData {
	return &GPAData{}
}

// Calculates the averages of the maps
func calculateAverages(typeCount map[string]int, typeTotal map[string]float64) map[string]float64 {
	averages := make(map[string]float64)

	for key, count := range typeCount {
		if total, ok := typeTotal[key]; ok && count != 0 {
			averages[key] = total / float64(count)
		}
	}

	return averages
}

// Totals up the gpa count in the record
func totalGPA(record []string) float64 {
	total := 0.0
	for col := utils.CNT_A_COL; col <= utils.CNT_F_COL; col++ {
		count, err := strconv.Atoi(record[col])
		if err != nil {
			fmt.Printf("failed to parse: %v\n", record[col])
			return 0.0
		}
		total += (utils.GpaMap[col] * float64(count))
	}
	return total
}

// Processes all of the records intoGPAData
func (data *GPAData) ProcessRecords(records [][]string) {
	professorCount := make(map[string]int)
	subjectCount := make(map[string]int)
	professorTotal := make(map[string]float64)
	subjectTotal := make(map[string]float64)
	courseCount := make(map[string]int)
	courseTotal := make(map[string]float64)

	for _, record := range records {
		if record[utils.CNT_A_COL] == "" || record[utils.GRADE_COUNT_COL] == "0" {
			continue // Skip records with no data
		}

		count, err := strconv.Atoi(record[utils.GRADE_COUNT_COL])
		if err != nil {
			fmt.Printf("failed to parse grade count: %v", err)
			continue
		}
		total := totalGPA(record)
		if total == 0.0 {
			continue
		}

		professor := simplifyName(record[utils.PROFESSOR_COL])
		subject := record[utils.CODE_COL]
		courseKey := CourseKey(professor, subject)

		if professor != "" {
			professorCount[professor] += count
			professorTotal[professor] += total
		}
		if subject != "" {
			subjectCount[subject] += count
			subjectTotal[subject] += total
		}
		if professor != "" && subject != "" {
			courseCount[courseKey] += count
			courseTotal[courseKey] += total
		}
	}

	data.Professors = calculateAverages(professorCount, professorTotal)
	data.Subjects = calculateAverages(subjectCount, subjectTotal)
	data.CourseGPAs = calculateAverages(courseCount, courseTotal)
}
