package cache

// Global singleton instance of course cache

import "sync"

var (
	instance CourseCache
	once     sync.Once
)

func GetInstance() CourseCache {
	once.Do(func() {
		instance = NewCourseCache()
	})
	return instance
}
