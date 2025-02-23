package api

// Many of the attribute sin this file are going to be unused, but they are
// included for completeness.

// APIResponse represents the top-level API response structure
type APIResponse struct {
	Success         bool                 `json:"success"`
	TotalCount      int                  `json:"totalCount"`
	Data            []CourseData         `json:"data"`
	PageOffset      int                  `json:"pageOffset"`
	PageMaxSize     int                  `json:"pageMaxSize"`
	SectionsFetched int                  `json:"sectionsFetchedCount"`
	PathMode        string               `json:"pathMode"`
	SearchConfigs   []SearchResultConfig `json:"searchResultsConfigs"`
}

// CourseData represents a single course entry from the API
type CourseData struct {
	ID                             int                  `json:"id"`
	Term                           string               `json:"term"`
	TermDesc                       string               `json:"termDesc"`
	CourseReferenceNumber          string               `json:"courseReferenceNumber"`
	PartOfTerm                     string               `json:"partOfTerm"`
	CourseNumber                   string               `json:"courseNumber"`
	Subject                        string               `json:"subject"`
	SubjectDescription             string               `json:"subjectDescription"`
	SequenceNumber                 string               `json:"sequenceNumber"`
	CampusDescription              string               `json:"campusDescription"`
	ScheduleTypeDescription        string               `json:"scheduleTypeDescription"`
	CourseTitle                    string               `json:"courseTitle"`
	CreditHours                    *float64             `json:"creditHours"`
	MaximumEnrollment              int                  `json:"maximumEnrollment"`
	Enrollment                     int                  `json:"enrollment"`
	SeatsAvailable                 int                  `json:"seatsAvailable"`
	WaitCapacity                   int                  `json:"waitCapacity"`
	WaitCount                      int                  `json:"waitCount"`
	WaitAvailable                  int                  `json:"waitAvailable"`
	CrossList                      *string              `json:"crossList"`
	CrossListCapacity              *int                 `json:"crossListCapacity"`
	CrossListCount                 *int                 `json:"crossListCount"`
	CrossListAvailable             *int                 `json:"crossListAvailable"`
	CreditHourHigh                 *float64             `json:"creditHourHigh"`
	CreditHourLow                  float64              `json:"creditHourLow"`
	OpenSection                    bool                 `json:"openSection"`
	LinkIdentifier                 *string              `json:"linkIdentifier"`
	IsSectionLinked                bool                 `json:"isSectionLinked"`
	SubjectCourse                  string               `json:"subjectCourse"`
	Faculty                        []FacultyData        `json:"faculty"`
	MeetingsFaculty                []MeetingFacultyData `json:"meetingsFaculty"`
	ReservedSeatSummary            *string              `json:"reservedSeatSummary"`
	SectionAttributes              []SectionAttribute   `json:"sectionAttributes"`
	InstructionalMethod            string               `json:"instructionalMethod"`
	InstructionalMethodDescription string               `json:"instructionalMethodDescription"`
}

// FacultyData represents a faculty member's information
type FacultyData struct {
	BannerID         string  `json:"bannerId"`
	Category         *string `json:"category"`
	DisplayName      string  `json:"displayName"`
	EmailAddress     string  `json:"emailAddress"`
	PrimaryIndicator bool    `json:"primaryIndicator"`
	Term             string  `json:"term"`
}

// MeetingFacultyData represents meeting time and faculty information
type MeetingFacultyData struct {
	Category    string      `json:"category"`
	MeetingTime MeetingTime `json:"meetingTime"`
	Term        string      `json:"term"`
}

// MeetingTime represents the details of when a class meets
type MeetingTime struct {
	BeginTime              string  `json:"beginTime"`
	Building               string  `json:"building"`
	BuildingDescription    string  `json:"buildingDescription"`
	Campus                 string  `json:"campus"`
	CampusDescription      string  `json:"campusDescription"`
	Category               string  `json:"category"`
	CreditHourSession      float64 `json:"creditHourSession"`
	EndDate                string  `json:"endDate"`
	EndTime                string  `json:"endTime"`
	Friday                 bool    `json:"friday"`
	HoursWeek              float64 `json:"hoursWeek"`
	MeetingScheduleType    string  `json:"meetingScheduleType"`
	MeetingType            string  `json:"meetingType"`
	MeetingTypeDescription string  `json:"meetingTypeDescription"`
	Monday                 bool    `json:"monday"`
	Room                   string  `json:"room"`
	Saturday               bool    `json:"saturday"`
	StartDate              string  `json:"startDate"`
	Sunday                 bool    `json:"sunday"`
	Term                   string  `json:"term"`
	Thursday               bool    `json:"thursday"`
	Tuesday                bool    `json:"tuesday"`
	Wednesday              bool    `json:"wednesday"`
}

// SectionAttribute represents attributes of a course section
type SectionAttribute struct {
	Code           string `json:"code"`
	Description    string `json:"description"`
	IsZTCAttribute bool   `json:"isZTCAttribute"`
	TermCode       string `json:"termCode"`
}

// SearchResultConfig represents the configuration for search results display
type SearchResultConfig struct {
	Config   string `json:"config"`
	Display  string `json:"display"`
	Title    string `json:"title"`
	Required bool   `json:"required"`
	Width    string `json:"width"`
}

// TermResponse represents a single term from the API
type TermResponse struct {
    Code        string `json:"code"`
    Description string `json:"description"`
}

// SubjectResponse represents a single subject from the API
type SubjectResponse struct {
    Code        string `json:"code"`
    Description string `json:"description"`
}
