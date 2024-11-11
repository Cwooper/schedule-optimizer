package models

import (
	"log"
	"strconv"
	"strings"

	"github.com/cwooper/schedule-optimizer/internal/utils"
)

type Professors map[string]float64                    // Average GPA per professor
type Subjects map[string]float64                      // Average GPA per subject (e.g. "CSCI 301")
type CourseGPAs map[string]float64                    // Average GPA per professor with subject
type ProfessorSubjects map[string]map[string]struct{} // Professor Last Name with Subject as a backup

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
	Subjects          Subjects
	Professors        Professors
	CourseGPAs        CourseGPAs
	ProfessorSubjects ProfessorSubjects
	LastNameIndex     map[string][]string
}

func NewGPAData() *GPAData {
	return &GPAData{
		Subjects:          make(Subjects),
		Professors:        make(Professors),
		CourseGPAs:        make(CourseGPAs),
		ProfessorSubjects: make(ProfessorSubjects),
		LastNameIndex:     make(map[string][]string),
	}
}

// Helper function to get last name from full name
func getLastName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[len(parts)-1])
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
			log.Printf("failed to parse: %v\n", record[col])
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

	// Create temporary map for professor-subject relationships
	for _, record := range records {
		if record[utils.CNT_A_COL] == "" ||
			record[utils.GRADE_COUNT_COL] == "0" ||
			(record[utils.GRADE_COUNT_COL] == record[utils.CNT_W_COL]) {
			continue
		}

		professor := simplifyName(record[utils.PROFESSOR_COL])
		if professor == "" {
			continue
		}

		subject := record[utils.CODE_COL]
		if subject == "" {
			continue
		}

		// Extract subject prefix (e.g., "CSCI" from "CSCI 347")
		matches := utils.SubjectRegexp.FindStringSubmatch(subject)
		if len(matches) < 2 {
			continue
		}
		subjectPrefix := matches[1] // First capture group contains the subject prefix

		// Build professor-subject relationship
		if data.ProfessorSubjects[professor] == nil {
			data.ProfessorSubjects[professor] = make(map[string]struct{})
		}
		data.ProfessorSubjects[professor][subjectPrefix] = struct{}{}

		// Build last name index
		lastName := getLastName(professor)
		if lastName != "" {
			data.LastNameIndex[lastName] = append(data.LastNameIndex[lastName], professor)
		}

		count, err := strconv.Atoi(record[utils.GRADE_COUNT_COL])
		if err != nil {
			log.Printf("failed to parse grade count: %v", err)
			continue
		}

		w_count, err := strconv.Atoi(record[utils.CNT_W_COL])
		if err != nil {
			log.Printf("failed to parse w count: %v", err)
			continue
		}

		count -= w_count
		total := totalGPA(record)
		if total == 0.0 {
			continue
		}

		courseKey := CourseKey(subject, professor)

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
