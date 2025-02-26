package utils

const (
	MIN_COURSES       = 1  // Minimum courses in a single schedule
	MAX_COURSES       = 8  // Maximum courses in a single schedule
	MAX_INPUT_COURSES = 13 // Maximum courses a user an input to be combined

	MAX_SUBJECT_WAIT    = 30 // Days
	MAX_TERM_WAIT       = 10 // Days
	MAX_NEW_COURSE_WAIT = 1  // Days, How often to update current/new course data
	MAX_OLD_COURSE_WAIT = 30 // Days, How often to update old course data

	UPDATE_HOUR = "3" // 0-23 hour to update course data, your time zone (e.g. 18 is 6pm)
	UPDATE_MIN  = "0" // 0-59 minute to update your course data, (e.g. 0 is on the hour)

	// One Schedule is about 2KB. 2000 Schedules is 4MB.
	MAX_OUTPUT_SCHEDULES = 2000
	SERVER_TIMEOUT_SECS  = 2  // Server will timeout after x seconds
	MAX_OUTPUT_COURSES   = 20 // Courses to send back upong a fuzzy search

	MAX_THREADS_SCRAPING = 3 // Maximum threads used by the scrapers wait group

	MAX_OUTPUT_SEARCH_COURSES = 25 // Maximum output searched courses
)
