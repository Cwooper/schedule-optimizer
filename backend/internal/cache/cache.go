// This package is intended for caching course data for less file I/O
// Mostly a consequence of virtualizing the server, causing slow I/O
// Also implements Course Caching by Subject for O(1) lookups
package cache

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/cwooper/schedule-optimizer/internal/models"
	pb "github.com/cwooper/schedule-optimizer/internal/proto/generated"
	"github.com/cwooper/schedule-optimizer/internal/utils"
	"github.com/cwooper/schedule-optimizer/pkg/protoutils"
)

// CourseCache interface defines the methods for interacting with the course cache
type CourseCache interface {
	// GetGlobalCourses returns courses for a given subject across all terms.
	GetGlobalCourses(subject string) ([]models.Course, error)
	// GetCourses returns courses for a given term and subject.
	GetCourses(term, subject string) ([]models.Course, error)
	// GetCourseList returns the full list of courses for a given term.
	GetCourseList(term string) ([]models.Course, error)
	// PreloadCache preloads courses for a list of terms.
	PreloadCache(terms []string) error
	// Invalidate clears all caches.
	Invalidate()
}

// courseCacheManager implements the CourseCache interface
type courseCacheManager struct {
	mu                 sync.RWMutex
	protoCache         map[string]*pb.CourseList             // term -> protobuf data
	termCache          map[string][]models.Course            // term -> courses list
	termSubjectCache   map[string]map[string][]models.Course // term -> subject -> courses
	globalSubjectCache map[string][]models.Course            // subject -> courses across all terms
}

// NewCourseCache creates a new CourseCache instance
func NewCourseCache() CourseCache {
	return &courseCacheManager{
		protoCache:         make(map[string]*pb.CourseList),
		termCache:          make(map[string][]models.Course),
		termSubjectCache:   make(map[string]map[string][]models.Course),
		globalSubjectCache: make(map[string][]models.Course),
	}
}

// Helper to build subject indexes for a term once courses are loaded
func (cm *courseCacheManager) buildSubjectIndexes(term string, courses []models.Course) {
	termIndex := make(map[string][]models.Course)
	for _, course := range courses {
		// Update term-specific index
		termIndex[course.Subject] = append(termIndex[course.Subject], course)
		// Update global index
		cm.globalSubjectCache[course.Subject] = append(cm.globalSubjectCache[course.Subject], course)
	}
	cm.termSubjectCache[term] = termIndex
}

// GetGlobalCourses returns all courses for the given subject, using cache if available
// Returns every course found in all terms
func (cm *courseCacheManager) GetGlobalCourses(subject string) ([]models.Course, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	courses, exists := cm.globalSubjectCache[subject]
	if !exists || len(courses) == 0 {
		return nil, fmt.Errorf("no courses found for subject: %s", subject)
	}

	return courses, nil
}

// GetCourses returns all courses for the given term and subject, using cache if available
// Returns every course found in the given term
func (cm *courseCacheManager) GetCourses(term, subject string) ([]models.Course, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	termIndex, exists := cm.termSubjectCache[term]
	if !exists {
		return nil, fmt.Errorf("no data for term: %s", term)
	}

	courses, exists := termIndex[subject]
	if !exists || len(courses) == 0 {
		return nil, fmt.Errorf("no courses found for subject: %s in term: %s", subject, term)
	}

	return courses, nil
}

// GetCourseList returns a slice of courses for the given term, using cache if available
func (cm *courseCacheManager) GetCourseList(term string) ([]models.Course, error) {
	cm.mu.RLock()
	if courses, ok := cm.termCache[term]; ok {
		cm.mu.RUnlock()
		return courses, nil
	}
	cm.mu.RUnlock()

	proto, err := cm.GetProto(term)
	if err != nil {
		return nil, fmt.Errorf("failed to get proto for term %s: %w", term, err)
	}

	courses := protoutils.ProtoToCourses(proto)

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.termCache[term] = courses
	cm.buildSubjectIndexes(term, courses)
	return courses, nil
}

// GetProto returns the protocol buffer for the given term, using cache if available
func (cm *courseCacheManager) GetProto(term string) (*pb.CourseList, error) {
	cm.mu.RLock()
	if proto, ok := cm.protoCache[term]; ok {
		cm.mu.RUnlock()
		return proto, nil
	}
	cm.mu.RUnlock()

	filename := filepath.Join(utils.DataDirectory, term+".pb")
	proto, err := utils.LoadCoursesProtobuf(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load protobuf for term %s: %w", term, err)
	}

	cm.mu.Lock()
	cm.protoCache[term] = proto
	cm.mu.Unlock()
	return proto, nil
}

// PreloadCache loads all terms into the cache.
func (cm *courseCacheManager) PreloadCache(terms []string) error {
	for _, term := range terms {
		if _, err := cm.GetCourseList(term); err != nil {
			return fmt.Errorf("failed to preload term %s: %w", term, err)
		}
	}
	return nil
}

// Invalidate clears all caches and indexes.
func (cm *courseCacheManager) Invalidate() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.protoCache = make(map[string]*pb.CourseList)
	cm.termCache = make(map[string][]models.Course)
	cm.termSubjectCache = make(map[string]map[string][]models.Course)
	cm.globalSubjectCache = make(map[string][]models.Course)
}
