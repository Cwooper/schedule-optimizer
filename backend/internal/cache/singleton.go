package cache

import "sync"

var (
	instance CourseCache
	once     sync.Once
	mu       sync.RWMutex // Additional mutex for safe instance replacement
)

// GetInstance returns the current CourseCache instance.
// If no instance exists, it creates a new one using NewCourseCache.
func GetInstance() CourseCache {
	once.Do(func() {
		instance = NewCourseCache()
	})
	
	mu.RLock()
	defer mu.RUnlock()
	return instance
}

// SetInstance allows injection of a CourseCache instance for testing.
// This should only be used in tests.
func SetInstance(cache CourseCache) {
	mu.Lock()
	defer mu.Unlock()
	instance = cache
}

// ResetInstance clears the singleton instance and allows a new one to be created.
// This should only be used in tests after SetInstance was called.
func ResetInstance() {
	mu.Lock()
	defer mu.Unlock()
	instance = nil
	once = sync.Once{} // Reset the once flag
}
