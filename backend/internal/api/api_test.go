package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleTermsResponse = `[
    {
        "code": "202530",
        "description": "Summer 2025 (View Only)"
    },
    {
        "code": "202520",
        "description": "Spring 2025"
    }
]`

const sampleSubjectsResponse = `[
    {
        "code": "CSCI",
        "description": "Computer Science"
    },
    {
        "code": "MATH",
        "description": "Mathematics"
    }
]`

const sampleResponse = `{
    "success": true,
    "totalCount": 3,
    "data": [
        {
            "id": 370941,
            "term": "202520",
            "termDesc": "Spring 2025",
            "courseReferenceNumber": "23923",
            "partOfTerm": "1",
            "courseNumber": "497F",
            "subject": "CSCI",
            "subjectDescription": "Computer Science",
            "sequenceNumber": "0",
            "campusDescription": "Main Campus",
            "scheduleTypeDescription": "Lecture",
            "courseTitle": "Robotics",
            "creditHours": null,
            "maximumEnrollment": 24,
            "enrollment": 0,
            "seatsAvailable": 24,
            "waitCapacity": 999,
            "waitCount": 0,
            "waitAvailable": 999,
            "crossList": null,
            "crossListCapacity": null,
            "crossListCount": null,
            "crossListAvailable": null,
            "creditHourHigh": null,
            "creditHourLow": 4,
            "creditHourIndicator": null,
            "openSection": true,
            "linkIdentifier": null,
            "isSectionLinked": false,
            "subjectCourse": "CSCI497F",
            "faculty": [
                {
                    "bannerId": "82787",
                    "category": null,
                    "class": "net.hedtech.banner.student.faculty.FacultyResultDecorator",
                    "courseReferenceNumber": "23923",
                    "displayName": "Nilles, Alli",
                    "emailAddress": "nillesa2@wwu.edu",
                    "primaryIndicator": true,
                    "term": "202520"
                }
            ],
            "meetingsFaculty": [
                {
                    "category": "01",
                    "class": "net.hedtech.banner.student.schedule.SectionSessionDecorator",
                    "courseReferenceNumber": "23923",
                    "faculty": [
                    ],
                    "meetingTime": {
                        "beginTime": "1000",
                        "building": "KB",
                        "buildingDescription": "Kaiser Borsari Hall",
                        "campus": "M",
                        "campusDescription": "Main Campus",
                        "category": "01",
                        "class": "net.hedtech.banner.general.overall.MeetingTimeDecorator",
                        "courseReferenceNumber": "23923",
                        "creditHourSession": 4.0,
                        "endDate": "06/13/2025",
                        "endTime": "1140",
                        "friday": false,
                        "hoursWeek": 3.33,
                        "meetingScheduleType": "L",
                        "meetingType": "CLAS",
                        "meetingTypeDescription": "Class",
                        "monday": true,
                        "room": "307",
                        "saturday": false,
                        "startDate": "04/01/2025",
                        "sunday": false,
                        "term": "202520",
                        "thursday": false,
                        "tuesday": false,
                        "wednesday": true
                    },
                    "term": "202520"
                }
            ],
            "reservedSeatSummary": null,
            "sectionAttributes": [
                {
                    "class": "net.hedtech.banner.student.schedule.SectionDegreeProgramAttributeDecorator",
                    "code": "FTF",
                    "courseReferenceNumber": "23923",
                    "description": "DELIVERY Face-to-Face",
                    "isZTCAttribute": false,
                    "termCode": "202520"
                }
            ],
            "instructionalMethod": "FTF",
            "instructionalMethodDescription": "Face to Face Instruction"
        },
        {
            "id": 354757,
            "term": "202520",
            "termDesc": "Spring 2025",
            "courseReferenceNumber": "23433",
            "partOfTerm": "1",
            "courseNumber": "497S",
            "subject": "CSCI",
            "subjectDescription": "Computer Science",
            "sequenceNumber": "0",
            "campusDescription": "Main Campus",
            "scheduleTypeDescription": "Seminar",
            "courseTitle": "Usable Security &amp; Privacy",
            "creditHours": null,
            "maximumEnrollment": 35,
            "enrollment": 0,
            "seatsAvailable": 35,
            "waitCapacity": 999,
            "waitCount": 0,
            "waitAvailable": 999,
            "crossList": null,
            "crossListCapacity": null,
            "crossListCount": null,
            "crossListAvailable": null,
            "creditHourHigh": null,
            "creditHourLow": 4,
            "creditHourIndicator": null,
            "openSection": true,
            "linkIdentifier": null,
            "isSectionLinked": false,
            "subjectCourse": "CSCI497S",
            "faculty": [
                {
                    "bannerId": "82785",
                    "category": null,
                    "class": "net.hedtech.banner.student.faculty.FacultyResultDecorator",
                    "courseReferenceNumber": "23433",
                    "displayName": "Mare, Shri",
                    "emailAddress": "shri.mare@wwu.edu",
                    "primaryIndicator": true,
                    "term": "202520"
                }
            ],
            "meetingsFaculty": [
                {
                    "category": "01",
                    "class": "net.hedtech.banner.student.schedule.SectionSessionDecorator",
                    "courseReferenceNumber": "23433",
                    "faculty": [
                    ],
                    "meetingTime": {
                        "beginTime": "0800",
                        "building": "CF",
                        "buildingDescription": "Communication Facility",
                        "campus": "M",
                        "campusDescription": "Main Campus",
                        "category": "01",
                        "class": "net.hedtech.banner.general.overall.MeetingTimeDecorator",
                        "courseReferenceNumber": "23433",
                        "creditHourSession": 4.0,
                        "endDate": "06/13/2025",
                        "endTime": "0850",
                        "friday": true,
                        "hoursWeek": 3.33,
                        "meetingScheduleType": "S",
                        "meetingType": "CLAS",
                        "meetingTypeDescription": "Class",
                        "monday": true,
                        "room": "314",
                        "saturday": false,
                        "startDate": "04/01/2025",
                        "sunday": false,
                        "term": "202520",
                        "thursday": false,
                        "tuesday": true,
                        "wednesday": true
                    },
                    "term": "202520"
                }
            ],
            "reservedSeatSummary": null,
            "sectionAttributes": [
                {
                    "class": "net.hedtech.banner.student.schedule.SectionDegreeProgramAttributeDecorator",
                    "code": "FTF",
                    "courseReferenceNumber": "23433",
                    "description": "DELIVERY Face-to-Face",
                    "isZTCAttribute": false,
                    "termCode": "202520"
                }
            ],
            "instructionalMethod": "FTF",
            "instructionalMethodDescription": "Face to Face Instruction"
        },
        {
            "id": 358194,
            "term": "202520",
            "termDesc": "Spring 2025",
            "courseReferenceNumber": "23718",
            "partOfTerm": "1",
            "courseNumber": "497Y",
            "subject": "CSCI",
            "subjectDescription": "Computer Science",
            "sequenceNumber": "0",
            "campusDescription": "Main Campus",
            "scheduleTypeDescription": "Lecture/Lab",
            "courseTitle": "Electronic Textiles",
            "creditHours": null,
            "maximumEnrollment": 35,
            "enrollment": 0,
            "seatsAvailable": 35,
            "waitCapacity": 999,
            "waitCount": 0,
            "waitAvailable": 999,
            "crossList": null,
            "crossListCapacity": null,
            "crossListCount": null,
            "crossListAvailable": null,
            "creditHourHigh": null,
            "creditHourLow": 4,
            "creditHourIndicator": null,
            "openSection": true,
            "linkIdentifier": null,
            "isSectionLinked": false,
            "subjectCourse": "CSCI497Y",
            "faculty": [
                {
                    "bannerId": "82786",
                    "category": null,
                    "class": "net.hedtech.banner.student.faculty.FacultyResultDecorator",
                    "courseReferenceNumber": "23718",
                    "displayName": "Hardin, Caroline",
                    "emailAddress": "caroline.hardin@wwu.edu",
                    "primaryIndicator": true,
                    "term": "202520"
                }
            ],
            "meetingsFaculty": [
                {
                    "category": "01",
                    "class": "net.hedtech.banner.student.schedule.SectionSessionDecorator",
                    "courseReferenceNumber": "23718",
                    "faculty": [
                    ],
                    "meetingTime": {
                        "beginTime": "1200",
                        "building": "CF",
                        "buildingDescription": "Communication Facility",
                        "campus": "M",
                        "campusDescription": "Main Campus",
                        "category": "01",
                        "class": "net.hedtech.banner.general.overall.MeetingTimeDecorator",
                        "courseReferenceNumber": "23718",
                        "creditHourSession": 4.0,
                        "endDate": "06/13/2025",
                        "endTime": "1340",
                        "friday": false,
                        "hoursWeek": 3.33,
                        "meetingScheduleType": "C",
                        "meetingType": "CLAS",
                        "meetingTypeDescription": "Class",
                        "monday": false,
                        "room": "420",
                        "saturday": false,
                        "startDate": "04/01/2025",
                        "sunday": false,
                        "term": "202520",
                        "thursday": true,
                        "tuesday": true,
                        "wednesday": false
                    },
                    "term": "202520"
                }
            ],
            "reservedSeatSummary": null,
            "sectionAttributes": [
                {
                    "class": "net.hedtech.banner.student.schedule.SectionDegreeProgramAttributeDecorator",
                    "code": "FTF",
                    "courseReferenceNumber": "23718",
                    "description": "DELIVERY Face-to-Face",
                    "isZTCAttribute": false,
                    "termCode": "202520"
                }
            ],
            "instructionalMethod": "FTF",
            "instructionalMethodDescription": "Face to Face Instruction"
        }
    ],
    "pageOffset": 0,
    "pageMaxSize": 50,
    "sectionsFetchedCount": 3,
    "pathMode": "search",
    "searchResultsConfigs": [
        {
            "config": "courseTitle",
            "display": "Title",
            "title": "Title",
            "required": true,
            "width": "9%"
        }
    ],
	"ztcEncodedImage": ""
}`

const emptyResponse = `{
    "success": true,
    "totalCount": 0,
    "data": [],
    "pageOffset": 0,
    "pageMaxSize": 500,
    "pathMode": "search"
}`

const initSessionResponse = `{"success": true}`
const termSelectResponse = `{"success": true}`

// mock client that can handle different endpoints
type mockHTTPClient struct {
	// For sequential responses (old style)
	responses []string
	calls     int

	// For endpoint-specific responses (new style)
	endpointResponses map[string]string
	endpointCalls     map[string]int
}

func newMockHTTPClient() *mockHTTPClient {
	return &mockHTTPClient{
		endpointResponses: make(map[string]string),
		endpointCalls:     make(map[string]int),
	}
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	// Increment total calls
	m.calls++

	// Track endpoint-specific calls if using that mode
	if m.endpointResponses != nil {
		endpoint := req.URL.Path
		m.endpointCalls[endpoint]++
	}

	var respBody string

	// If using endpoint-specific responses
	if len(m.endpointResponses) > 0 {
		if response, ok := m.endpointResponses[req.URL.Path]; ok {
			respBody = response
		} else {
			respBody = emptyResponse
		}
	} else {
		// Using sequential responses (old style)
		if m.calls <= len(m.responses) {
			respBody = m.responses[m.calls-1]
		} else {
			respBody = emptyResponse
		}
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(respBody)),
		Header:     make(http.Header),
	}, nil
}

// TestGetTerms tests the term retrieval functionality
func TestGetTerms(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		expectError   bool
		expectTerms   int
		expectedCodes []string
	}{
		{
			name:          "Success",
			response:      sampleTermsResponse,
			expectError:   false,
			expectTerms:   2,
			expectedCodes: []string{"202530", "202520"},
		},
		{
			name:        "Empty response",
			response:    "[]",
			expectError: false,
			expectTerms: 0,
		},
		{
			name:        "Invalid JSON",
			response:    "{invalid}",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := newMockHTTPClient()
			mockTransport.endpointResponses["/StudentRegistrationSsb/ssb/classSearch/getTerms"] = tt.response

			client := &Client{
				httpClient: &http.Client{
					Transport: mockTransport,
				},
			}

			terms, err := client.GetTerms()

			if tt.expectError != (err != nil) {
				t.Errorf("expected error=%v but got error=%v", tt.expectError, err)
			}

			if !tt.expectError {
				if len(terms) != tt.expectTerms {
					t.Errorf("expected %d terms, got %d", tt.expectTerms, len(terms))
				}

				if tt.expectedCodes != nil {
					for i, expectedCode := range tt.expectedCodes {
						if terms[i].Code != expectedCode {
							t.Errorf("expected term code %s, got %s", expectedCode, terms[i].Code)
						}
					}
				}
			}
		})
	}
}

func TestGetSubjects(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		term           string
		expectError    bool
		expectSubjects int
		expectedCodes  []string
	}{
		{
			name:           "Success",
			response:       sampleSubjectsResponse,
			term:           "202520",
			expectError:    false,
			expectSubjects: 2,
			expectedCodes:  []string{"CSCI", "MATH"},
		},
		{
			name:           "Empty response",
			response:       "[]",
			term:           "202520",
			expectError:    false,
			expectSubjects: 0,
		},
		{
			name:        "Invalid JSON",
			response:    "{invalid}",
			term:        "202520",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := newMockHTTPClient()
			mockTransport.endpointResponses["/StudentRegistrationSsb/ssb/classSearch/get_subject"] = tt.response

			client := &Client{
				httpClient: &http.Client{
					Transport: mockTransport,
				},
			}

			subjects, err := client.GetSubjects(tt.term)

			if tt.expectError != (err != nil) {
				t.Errorf("expected error=%v but got error=%v", tt.expectError, err)
			}

			if !tt.expectError {
				if len(subjects) != tt.expectSubjects {
					t.Errorf("expected %d subjects, got %d", tt.expectSubjects, len(subjects))
				}

				// Check specific subject codes if provided
				if tt.expectedCodes != nil {
					for i, expectedCode := range tt.expectedCodes {
						if subjects[i].Code != expectedCode {
							t.Errorf("expected subject code %s, got %s", expectedCode, subjects[i].Code)
						}
					}
				}
			}
		})
	}
}

func TestCourseConversion(t *testing.T) {
	// Create test server that handles all endpoints
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return appropriate response based on the endpoint
		switch {
		case strings.Contains(r.URL.Path, "termSelection"):
			w.Write([]byte(initSessionResponse))
		case strings.Contains(r.URL.Path, "term/search"):
			w.Write([]byte(termSelectResponse))
		case strings.Contains(r.URL.Path, "searchResults"):
			w.Write([]byte(sampleResponse))
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	client := &Client{
		httpClient: ts.Client(),
		baseURL:    ts.URL,
	}

	courses, err := client.GetCourses("202520", "CSCI", "%")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(courses) == 0 {
		t.Fatal("expected at least one course")
	}

	// Test all courses from the sample response
	expectedCourses := []struct {
		subject    string
		title      string
		instructor string
		days       string
		location   string
		crn        int
		capacity   int
	}{
		{
			subject:    "CSCI 497F",
			title:      "Robotics",
			instructor: "Nilles, Alli",
			days:       "MW",
			location:   "KB 307 (Kaiser Borsari Hall)",
			crn:        23923,
			capacity:   24,
		},
		{
			subject:    "CSCI 497S",
			title:      "Usable Security & Privacy",
			instructor: "Mare, Shri",
			days:       "MTWF",
			location:   "CF 314 (Communication Facility)",
			crn:        23433,
			capacity:   35,
		},
		{
			subject:    "CSCI 497Y",
			title:      "Electronic Textiles",
			instructor: "Hardin, Caroline",
			days:       "TR",
			location:   "CF 420 (Communication Facility)",
			crn:        23718,
			capacity:   35,
		},
	}

	for i, expected := range expectedCourses {
		course := courses[i]
		t.Run(fmt.Sprintf("Course %d", i), func(t *testing.T) {
			if course.Subject != expected.subject {
				t.Errorf("Subject: got %q, want %q", course.Subject, expected.subject)
			}
			if course.Title != expected.title {
				t.Errorf("Title: got %q, want %q", course.Title, expected.title)
			}
			if course.Instructor != expected.instructor {
				t.Errorf("Instructor: got %q, want %q", course.Instructor, expected.instructor)
			}
			if course.Sessions[0].Days != expected.days {
				t.Errorf("Days: got %q, want %q", course.Sessions[0].Days, expected.days)
			}
			if course.Sessions[0].Location != expected.location {
				t.Errorf("Location: got %q, want %q", course.Sessions[0].Location, expected.location)
			}
			if course.CRN != expected.crn {
				t.Errorf("CRN: got %d, want %d", course.CRN, expected.crn)
			}
			if course.Capacity != expected.capacity {
				t.Errorf("Capacity: got %d, want %d", course.Capacity, expected.capacity)
			}
		})
	}
}

// TestGetCourses tests the course retrieval functionality
func TestGetCourses(t *testing.T) {
	tests := []struct {
		name           string
		courseResponse string
		expectError    bool
		expectCourses  int
	}{
		{
			name:           "Success single page",
			courseResponse: sampleResponse,
			expectError:    false,
			expectCourses:  3,
		},
		{
			name:           "Empty response",
			courseResponse: emptyResponse,
			expectError:    false,
			expectCourses:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server that handles all endpoints
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				switch {
				case strings.Contains(r.URL.Path, "termSelection"):
					w.Write([]byte(initSessionResponse))
				case strings.Contains(r.URL.Path, "term/search"):
					w.Write([]byte(termSelectResponse))
				case strings.Contains(r.URL.Path, "searchResults"):
					w.Write([]byte(tt.courseResponse))
				default:
					t.Errorf("unexpected request to %s", r.URL.Path)
					http.Error(w, "unexpected request", http.StatusBadRequest)
				}
			}))
			defer ts.Close()

			client := &Client{
				httpClient: ts.Client(),
				baseURL:    ts.URL,
			}

			courses, err := client.GetCourses("202520", "CSCI", "%")

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(courses) != tt.expectCourses {
				t.Errorf("expected %d courses, got %d", tt.expectCourses, len(courses))
			}
		})
	}
}

func TestSessionCreation(t *testing.T) {
	// Test specific session creation
	meetingFaculty := MeetingFacultyData{
		Category: "01",
		MeetingTime: MeetingTime{
			BeginTime:           "1000",
			EndTime:             "1140",
			Building:            "KB",
			Room:                "307",
			BuildingDescription: "Kaiser Borsari Hall",
			Monday:              true,
			Wednesday:           true,
		},
	}

	session, err := createSession(meetingFaculty)
	if err != nil {
		t.Fatalf("unexpected error creating session: %v", err)
	}

	// Check session details
	if session.Days != "MW" {
		t.Errorf("expected days MW, got %s", session.Days)
	}
	if session.StartTime != 1000 {
		t.Errorf("expected start time 1000, got %d", session.StartTime)
	}
	if session.EndTime != 1140 {
		t.Errorf("expected end time 1140, got %d", session.EndTime)
	}
	if !strings.Contains(session.Location, "Kaiser Borsari Hall") {
		t.Errorf("expected location to contain building description, got %s", session.Location)
	}
}

func TestAsyncAndTBDSessions(t *testing.T) {
	tests := []struct {
		name        string
		meetingTime MeetingTime
		expectAsync bool
		expectTBD   bool
	}{
		{
			name: "Normal session",
			meetingTime: MeetingTime{
				BeginTime: "1000",
				EndTime:   "1140",
			},
			expectAsync: false,
			expectTBD:   false,
		},
		{
			name: "Async session",
			meetingTime: MeetingTime{
				BeginTime: "",
				EndTime:   "",
			},
			expectAsync: true,
			expectTBD:   false,
		},
		{
			name: "TBD session",
			meetingTime: MeetingTime{
				BeginTime: "TBD",
				EndTime:   "TBD",
			},
			expectAsync: false,
			expectTBD:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meetingFaculty := MeetingFacultyData{
				MeetingTime: tt.meetingTime,
			}

			session, err := createSession(meetingFaculty)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if session.IsAsync != tt.expectAsync {
				t.Errorf("expected IsAsync=%v, got %v", tt.expectAsync, session.IsAsync)
			}
			if session.IsTimeTBD != tt.expectTBD {
				t.Errorf("expected IsTimeTBD=%v, got %v", tt.expectTBD, session.IsTimeTBD)
			}
		})
	}
}
