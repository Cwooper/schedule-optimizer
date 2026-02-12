package generator

import "schedule-optimizer/internal/cache"

// CourseInfo contains course-level data sent once per unique course code.
type CourseInfo struct {
	Subject      string  `json:"subject"`
	CourseNumber string  `json:"courseNumber"`
	Title        string  `json:"title"`
	Credits      int     `json:"credits"`
	GPA          float64 `json:"gpa,omitempty"`
}

// SectionInfo contains section-level data sent once per unique CRN.
type SectionInfo struct {
	CRN            string              `json:"crn"`
	Term           string              `json:"term"`
	CourseKey      string              `json:"courseKey"`
	Instructor     string              `json:"instructor,omitempty"`
	Enrollment     int                 `json:"enrollment"`
	MaxEnrollment  int                 `json:"maxEnrollment"`
	SeatsAvailable int                 `json:"seatsAvailable"`
	WaitCount      int                 `json:"waitCount"`
	IsOpen         bool                `json:"isOpen"`
	MeetingTimes   []cache.MeetingTime `json:"meetingTimes"`
	GPA            float64             `json:"gpa,omitempty"`
	GPASource      string              `json:"gpaSource,omitempty"`
}

// ScheduleRef contains CRN references for a single schedule.
type ScheduleRef struct {
	CRNs    []string `json:"crns"`
	Score   float64  `json:"score"`
	Weights []Weight `json:"weights"`
}

// Response is the wire format for schedule generation results.
type Response struct {
	Courses       map[string]CourseInfo  `json:"courses"`
	Sections      map[string]SectionInfo `json:"sections"`
	Schedules     []ScheduleRef          `json:"schedules"`
	Asyncs        []string               `json:"asyncs"`
	CourseResults []CourseResult         `json:"courseResults"`
	Stats         GenerateStats          `json:"stats"`
}

// ToResponse transforms the internal GenerateResponse to the deduplicated wire format.
func (r *GenerateResponse) ToResponse() *Response {
	courses := make(map[string]CourseInfo)
	sections := make(map[string]SectionInfo)
	schedules := make([]ScheduleRef, 0, len(r.Schedules))
	asyncs := make([]string, 0, len(r.Asyncs))

	// Process all schedules, collecting unique courses and sections
	for _, sched := range r.Schedules {
		crns := make([]string, 0, len(sched.Courses))

		for _, course := range sched.Courses {
			courseKey := course.Subject + ":" + course.CourseNumber

			// Add course if not seen
			if _, exists := courses[courseKey]; !exists {
				courses[courseKey] = CourseInfo{
					Subject:      course.Subject,
					CourseNumber: course.CourseNumber,
					Title:        course.Title,
					Credits:      course.Credits,
				}
			}

			// Add section if not seen
			if _, exists := sections[course.CRN]; !exists {
				sections[course.CRN] = SectionInfo{
					CRN:            course.CRN,
					Term:           course.Term,
					CourseKey:      courseKey,
					Instructor:     course.Instructor,
					Enrollment:     course.Enrollment,
					MaxEnrollment:  course.MaxEnrollment,
					SeatsAvailable: course.SeatsAvailable,
					WaitCount:      course.WaitCount,
					IsOpen:         course.IsOpen,
					MeetingTimes:   course.MeetingTimes,
					GPA:            course.GPA,
					GPASource:      course.GPASource,
				}
			}

			crns = append(crns, course.CRN)
		}

		schedules = append(schedules, ScheduleRef{
			CRNs:    crns,
			Score:   sched.Score,
			Weights: sched.Weights,
		})
	}

	// Process async sections
	for _, course := range r.Asyncs {
		courseKey := course.Subject + ":" + course.CourseNumber

		// Add course if not seen
		if _, exists := courses[courseKey]; !exists {
			courses[courseKey] = CourseInfo{
				Subject:      course.Subject,
				CourseNumber: course.CourseNumber,
				Title:        course.Title,
				Credits:      course.Credits,
			}
		}

		// Add section if not seen
		if _, exists := sections[course.CRN]; !exists {
			sections[course.CRN] = SectionInfo{
				CRN:            course.CRN,
				Term:           course.Term,
				CourseKey:      courseKey,
				Instructor:     course.Instructor,
				Enrollment:     course.Enrollment,
				MaxEnrollment:  course.MaxEnrollment,
				SeatsAvailable: course.SeatsAvailable,
				WaitCount:      course.WaitCount,
				IsOpen:         course.IsOpen,
				MeetingTimes:   course.MeetingTimes,
				GPA:            course.GPA,
				GPASource:      course.GPASource,
			}
		}

		asyncs = append(asyncs, course.CRN)
	}

	// Compute course-level GPA as average of all section GPAs for each course
	type gpaAcc struct {
		total float64
		count int
	}
	courseGPAs := make(map[string]*gpaAcc)
	for _, sec := range sections {
		if sec.GPA > 0 {
			acc, ok := courseGPAs[sec.CourseKey]
			if !ok {
				acc = &gpaAcc{}
				courseGPAs[sec.CourseKey] = acc
			}
			acc.total += sec.GPA
			acc.count++
		}
	}
	for key, acc := range courseGPAs {
		ci := courses[key]
		ci.GPA = acc.total / float64(acc.count)
		courses[key] = ci
	}

	return &Response{
		Courses:       courses,
		Sections:      sections,
		Schedules:     schedules,
		Asyncs:        asyncs,
		CourseResults: r.CourseResults,
		Stats:         r.Stats,
	}
}
