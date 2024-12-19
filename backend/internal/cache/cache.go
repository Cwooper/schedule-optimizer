// This package is intended for caching course data for less file I/O
// Mostly a consequence of virtualizing the server, causing slow I/O
// Also implements Course Caching by CRN
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
	GetCourseList(term string) ([]models.Course, error)
	GetProto(term string) (*pb.CourseList, error)
	Invalidate()
}

// courseCacheManager implements the CourseCache interface
type courseCacheManager struct {
	mu            sync.RWMutex
	protoCache    map[string]*pb.CourseList  // Cache for protocol buffers
	courseCache   map[string][]models.Course // Cache for course lists
	coursesByTerm map[string]map[int]int     // Index map of CRN to position in courseCache for O(1) lookups
}

// NewCourseCache creates a new CourseCache instance
func NewCourseCache() CourseCache {
	return &courseCacheManager{
		protoCache:    make(map[string]*pb.CourseList),
		courseCache:   make(map[string][]models.Course),
		coursesByTerm: make(map[string]map[int]int),
	}
}

// GetCourseList returns a slice of courses for the given term, using cache if available
func (cm *courseCacheManager) GetCourseList(term string) ([]models.Course, error) {
	// First check if we have it in the course cache
	cm.mu.RLock()
	if courses, ok := cm.courseCache[term]; ok {
		cm.mu.RUnlock()
		return courses, nil
	}
	cm.mu.RUnlock()

	// If not in cache, get the proto and convert it
	proto, err := cm.GetProto(term)
	if err != nil {
		return nil, fmt.Errorf("failed to get proto for term %s: %w", term, err)
	}

	courses := protoutils.ProtoToCourses(proto)

	// Cache the courses and build the index
	cm.mu.Lock()
	cm.courseCache[term] = courses

	// Build index for O(1) lookups by CRN
	crnIndex := make(map[int]int)
	for i, course := range courses {
		crnIndex[course.CRN] = i
	}
	cm.coursesByTerm[term] = crnIndex
	cm.mu.Unlock()

	return courses, nil
}

// GetProto returns the protocol buffer for the given term, using cache if available
func (cm *courseCacheManager) GetProto(term string) (*pb.CourseList, error) {
	// Check if we have it in the proto cache
	cm.mu.RLock()
	if proto, ok := cm.protoCache[term]; ok {
		cm.mu.RUnlock()
		return proto, nil
	}
	cm.mu.RUnlock()

	// Create the full file path
	filename := filepath.Join(utils.DataDirectory, term+".pb")

	// If not in cache, load from disk
	proto, err := utils.LoadCoursesProtobuf(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load protobuf for term %s: %w", term, err)
	}

	// Cache the proto
	cm.mu.Lock()
	cm.protoCache[term] = proto
	cm.mu.Unlock()

	return proto, nil
}

// Invalidate clears all caches, called after course data is updated by scraper
func (cm *courseCacheManager) Invalidate() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.protoCache = make(map[string]*pb.CourseList)
	cm.courseCache = make(map[string][]models.Course)
	cm.coursesByTerm = make(map[string]map[int]int)
}
