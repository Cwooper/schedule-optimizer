package scraper

// TermResponse represents a term from the Banner API /classSearch/getTerms endpoint.
type TermResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// APIResponse is the top-level response from /searchResults/searchResults.
type APIResponse struct {
	Success    bool         `json:"success"`
	TotalCount int          `json:"totalCount"`
	Data       []CourseData `json:"data"`
}

// CourseData represents a single course/section from the Banner API.
type CourseData struct {
	ID                       int64              `json:"id"`
	Term                     string             `json:"term"`
	TermDesc                 string             `json:"termDesc"`
	CourseReferenceNumber    string             `json:"courseReferenceNumber"`
	PartOfTerm               string             `json:"partOfTerm"`
	CourseNumber             string             `json:"courseNumber"`
	Subject                  string             `json:"subject"`
	SubjectDescription       string             `json:"subjectDescription"`
	SequenceNumber           string             `json:"sequenceNumber"`
	CampusDescription        string             `json:"campusDescription"`
	ScheduleTypeDescription  string             `json:"scheduleTypeDescription"`
	CourseTitle              string             `json:"courseTitle"`
	CreditHours              *float64           `json:"creditHours"`
	CreditHourHigh           *float64           `json:"creditHourHigh"`
	CreditHourLow            *float64           `json:"creditHourLow"`
	MaximumEnrollment        int                `json:"maximumEnrollment"`
	Enrollment               int                `json:"enrollment"`
	SeatsAvailable           int                `json:"seatsAvailable"`
	WaitCapacity             int                `json:"waitCapacity"`
	WaitCount                int                `json:"waitCount"`
	OpenSection              bool               `json:"openSection"`
	InstructionalMethod      string             `json:"instructionalMethod"`
	InstructionalMethodDescr string             `json:"instructionalMethodDescription"`
	Faculty                  []FacultyData      `json:"faculty"`
	MeetingsFaculty          []MeetingsFaculty  `json:"meetingsFaculty"`
	SectionAttributes        []SectionAttribute `json:"sectionAttributes"`
}

// FacultyData represents an instructor for a section.
type FacultyData struct {
	BannerID         string `json:"bannerId"`
	DisplayName      string `json:"displayName"`
	EmailAddress     string `json:"emailAddress"`
	PrimaryIndicator bool   `json:"primaryIndicator"`
}

// MeetingsFaculty wraps meeting time data from Banner API.
type MeetingsFaculty struct {
	MeetingTime MeetingTimeData `json:"meetingTime"`
}

// MeetingTimeData represents a meeting time for a section.
type MeetingTimeData struct {
	BeginTime              string   `json:"beginTime"`
	EndTime                string   `json:"endTime"`
	StartDate              string   `json:"startDate"`
	EndDate                string   `json:"endDate"`
	Building               string   `json:"building"`
	BuildingDescription    string   `json:"buildingDescription"`
	Room                   string   `json:"room"`
	Monday                 bool     `json:"monday"`
	Tuesday                bool     `json:"tuesday"`
	Wednesday              bool     `json:"wednesday"`
	Thursday               bool     `json:"thursday"`
	Friday                 bool     `json:"friday"`
	Saturday               bool     `json:"saturday"`
	Sunday                 bool     `json:"sunday"`
	MeetingScheduleType    string   `json:"meetingScheduleType"`
	MeetingType            string   `json:"meetingType"`
	MeetingTypeDescription string   `json:"meetingTypeDescription"`
	CreditHourSession      *float64 `json:"creditHourSession"`
	HoursWeek              *float64 `json:"hoursWeek"`
}

// SectionAttribute represents an attribute (GUR, etc.) for a section.
type SectionAttribute struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// PageResult holds the result of fetching a single page.
type PageResult struct {
	Courses    []CourseData
	TotalCount int
	Offset     int
	Error      error
}
