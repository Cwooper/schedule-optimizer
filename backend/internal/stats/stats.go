package stats

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Stats holds all statistics for the schedule optimizer
type Stats struct {
	// Core statistics
	ScheduleRequests   uint64            `json:"schedule_requests"`
	TotalSchedules     uint64            `json:"total_schedules"`
	SearchRequests     uint64            `json:"search_requests"`
	SubjectCounts      map[string]uint64 `json:"subject_counts"`
	LastCourseUpdate   time.Time         `json:"last_course_update"`
	ServerCreationDate time.Time         `json:"server_creation_date"`

	// Derived statistics
	AverageSchedulesPerRequest float64 `json:"avg_schedules_per_request"`

	// Internal fields
	mutex    sync.RWMutex `json:"-"`
	filePath string       `json:"-"`
}

// New creates a new Stats instance
func New(filePath string) *Stats {
	s := &Stats{
		SubjectCounts:      make(map[string]uint64),
		ServerCreationDate: time.Now(),
		filePath:           filePath,
	}

	// Try to load existing stats from disk
	if err := s.Load(); err != nil {
		log.Printf("No existing stats found or error loading stats: %v", err)
	}

	// Start periodic saving
	go s.startPeriodicSave()

	return s
}

// IncrementScheduleRequest increments the schedule request counter and updates averages
func (s *Stats) IncrementScheduleRequest(schedulesGenerated uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.ScheduleRequests++
	s.TotalSchedules += schedulesGenerated
	s.AverageSchedulesPerRequest = float64(s.TotalSchedules) / float64(s.ScheduleRequests)
}

// IncrementSearchRequest increments the search request counter
func (s *Stats) IncrementSearchRequest() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.SearchRequests++
}

// IncrementSubject increments the counter for a specific subject
func (s *Stats) IncrementSubject(subject string) {
	// Note that some courses have spaces in them
	lastSpace := strings.LastIndex(subject, " ")
	if lastSpace != - 1 {
		subject = subject[:lastSpace]
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.SubjectCounts[subject]++
}

// UpdateLastCourseUpdate sets the last course update time
func (s *Stats) UpdateLastCourseUpdate() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.LastCourseUpdate = time.Now()
}

// GetStats returns a copy of the current statistics as JSON
func (s *Stats) GetStats() Stats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy without the mutex and filepath
	return Stats{
		ScheduleRequests:           s.ScheduleRequests,
		TotalSchedules:             s.TotalSchedules,
		SearchRequests:             s.SearchRequests,
		SubjectCounts:              s.SubjectCounts,
		LastCourseUpdate:           s.LastCourseUpdate,
		ServerCreationDate:         s.ServerCreationDate,
		AverageSchedulesPerRequest: s.AverageSchedulesPerRequest,
	}
}

// Save persists the current stats to disk
func (s *Stats) Save() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(s)
}

// Load reads stats from disk
func (s *Stats) Load() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(s)
}

// startPeriodicSave starts a goroutine that saves stats to disk periodically
func (s *Stats) startPeriodicSave() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		if err := s.Save(); err != nil {
			log.Printf("Error saving stats: %v", err)
		}
	}
}

// Close saves the stats one final time and cleans up
func (s *Stats) Close() error {
	return s.Save()
}

func GetStatsFile() (string, error) {
    workingDir, err := os.Getwd()
    if err != nil {
        return "", fmt.Errorf("failed to get working directory: %v", err)
    }

    absolutePath := filepath.Join(workingDir, "../data/server_stats.json")
    absolutePath, err = filepath.Abs(absolutePath)
    if err != nil {
        return "", fmt.Errorf("failed to get absolute path: %v", err)
    }

    return absolutePath, nil
}
