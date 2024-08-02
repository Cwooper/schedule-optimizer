package gpa

// Loads the gpa grade_distribution.csv and converts it to a protobuf
// after processing the data

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"schedule-optimizer/internal/models"
	pb "schedule-optimizer/internal/proto/generated"
	"schedule-optimizer/internal/utils"
	"schedule-optimizer/pkg/protoutils"

	"google.golang.org/protobuf/proto"
)

const (
	CSV_FILENAME = "grade_distribution.csv"
	PB_FILENAME  = "grade_distribution.pb"
)

// Loads the grade_distribution csv file
func loadCSV() ([][]string, error) {
	csvFile := filepath.Join(utils.DataDirectory, CSV_FILENAME)
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

// Gets the GPAData from the grade_distribution file
func getFromCSV() (models.GPAData, error) {
	gpaData := models.NewGPAData()
	records, err := loadCSV()
	if err != nil {
		return *gpaData, fmt.Errorf("failed to load csv: %w", err)
	}

	gpaData.ProcessRecords(records)

	return *gpaData, nil
}

// Saves the GPAData to the protobuf
func SaveGpaDataAsPB(gpaData models.GPAData) error {
	pbGPAData := protoutils.GPADataToProto(gpaData)
	data, err := proto.Marshal(pbGPAData)
	if err != nil {
		return fmt.Errorf("failed to marshal GPAData: %w", err)
	}

	pbFilename := filepath.Join(utils.DataDirectory, PB_FILENAME)
	err = os.WriteFile(pbFilename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write GPAData to file: %w", err)
	}

	return nil
}

// Gets the GPAData from the protobuf file
func GetGpaDataFromPB(file *os.File) (models.GPAData, error) {
	data, err := os.ReadFile(file.Name())
	if err != nil {
		return models.GPAData{}, fmt.Errorf("failed to read GPAData file: %w", err)
	}

	var pbGPAData pb.GPAData
	err = proto.Unmarshal(data, &pbGPAData)
	if err != nil {
		return models.GPAData{}, fmt.Errorf("failed to unmarshal GPAData: %w", err)
	}

	return protoutils.ProtoToGPAData(&pbGPAData), nil
}

// Returns the GPAData structure
func GetGPAData() (models.GPAData, error) {
	var gpaData models.GPAData
	pbFilename := filepath.Join(utils.DataDirectory, PB_FILENAME)

	// Get from the protobuf if it exists
	if file, err := os.Open(pbFilename); err == nil {
		gpaData, err = GetGpaDataFromPB(file)
		if err != nil {
			return gpaData, fmt.Errorf("failed to get gpadata from protobuf: %w", err)
		}
		return gpaData, nil
	}

	// Otherwise get and process from the csv
	gpaData, err := getFromCSV()
	if err != nil {
		return gpaData, fmt.Errorf("failed to get gpadata from csv: %w", err)
	}

	err = SaveGpaDataAsPB(gpaData)
	if err != nil {
		return gpaData, fmt.Errorf("failed to save gpaData to protobuf: %w", err)
	}
	fmt.Printf("Saved GPA Data to %v\n", pbFilename)

	return gpaData, nil
}
