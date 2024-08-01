package gpa

// Loads the gpa grade_distribution.csv and converts it to a protobuf
// after processing the data

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"schedule-optimizer/internal/utils"
)

const (
	CSV_Filename = "grade_distribution.csv"
)

var (
	GradeColumns = []string{"CNT_A", "CNT_AM", "CNT_BP", "CNT_B", "CNT_BM",
		"CNT_CP", "CNT_C", "CNT_CM", "CNT_DP", "CNT_D", "CNT_DM", "CNT_F"}
	GpaValues = []float64{4.0, 3.7, 3.3, 3.0, 2.7, 2.3, 2.0, 1.7, 1.3, 1.0, 0.7, 0.0}
)

type SubjectGPA struct {
	Subject string
	GPA     float64
}

type ProfessorGPA struct {
	Professor string
	GPA       float64
}

type CourseGPA struct {
	Subject   string
	Professor string
	GPA       float64
}

type GPAData struct {
	SubjectGPAs   []SubjectGPA
	ProfessorGPAs []ProfessorGPA
	CourseGPAs    []CourseGPA
}

// Loads the grade_distribution csv file
func loadCSV() ([][]string, error) {
	csvFile := filepath.Join(utils.DataDirectory, CSV_Filename)
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read csv: %w", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)

	// Read and discard the header row
	_, err = csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv header: %w", err)
	}

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv records: %w", err)
	}

	return records, nil
}

func GetGPAData() (GPAData, error) {
	var gpaData GPAData
	// if grade_distribution.pb exists, use that
	// otherwise, process the grade_distribution.csv

	return gpaData, nil
}

func Test() error {
	records, err := loadCSV()
	if err != nil {
		return fmt.Errorf("failed to load csv: %w", err)
	}

	fmt.Printf("Rows: %d\n", len(records))
	return nil
}
