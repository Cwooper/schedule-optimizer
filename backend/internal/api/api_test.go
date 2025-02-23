package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

// mockHTTPClient creates a test HTTP client that returns predefined responses
type mockHTTPClient struct {
	responses []string
	calls     int
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	// Increment call counter
	m.calls++

	// Determine which response to return
	var respBody string
	if m.calls <= len(m.responses) {
		respBody = m.responses[m.calls-1]
	} else {
		respBody = emptyResponse
	}

	// Create response
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(respBody)),
		Header:     make(http.Header),
	}, nil
}

func TestGetCourses(t *testing.T) {
	tests := []struct {
		name           string
		responses     []string
		expectedCalls int
		expectError   bool
		expectCourses int
	}{
		{
			name:           "Success single page",
			responses:     []string{sampleResponse, sampleResponse, emptyResponse},
			expectedCalls: 3, // Initial session setup + term selection + course request
			expectError:   false,
			expectCourses: 1,
		},
		{
			name:           "Empty response",
			responses:     []string{sampleResponse, sampleResponse, emptyResponse},
			expectedCalls: 3,
			expectError:   false,
			expectCourses: 1,
		},
		{
			name:           "Multiple pages",
			responses:     []string{sampleResponse, sampleResponse, sampleResponse, emptyResponse},
			expectedCalls: 4,
			expectError:   false,
			expectCourses: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockTransport := &mockHTTPClient{
				responses: tt.responses,
			}

			client := &Client{
				httpClient: &http.Client{
					Transport: mockTransport,
				},
			}

			// Test the GetCourses method
			courses, err := client.GetCourses("202520", "CSCI", "%")

			// Check error
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check number of courses
			if len(courses) != tt.expectCourses {
				t.Errorf("expected %d courses, got %d", tt.expectCourses, len(courses))
			}

			// Check number of API calls
			if mockTransport.calls != tt.expectedCalls {
				t.Errorf("expected %d API calls, got %d", tt.expectedCalls, mockTransport.calls)
			}
		})
	}
}

func TestCourseConversion(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return sample response for all requests
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(sampleResponse))
	}))
	defer ts.Close()

	// Create client using test server URL
	client := &Client{
		httpClient: ts.Client(),
	}

	// Get courses
	courses, err := client.GetCourses("202520", "CSCI", "%")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(courses) == 0 {
		t.Fatal("expected at least one course")
	}

	// Check the first course
	course := courses[0]

	// Verify course details
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Subject", course.Subject, "CSCI 497F"},
		{"Title", course.Title, "Robotics"},
		{"Instructor", course.Instructor, "Nilles, Alli"},
		{"Days", course.Sessions[0].Days, "MW"},
		{"Location", course.Sessions[0].Location, "KB 307 (Kaiser Borsari Hall)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, tt.got)
			}
		})
	}

	// Verify numeric fields
	if len(course.Sessions) == 0 {
		t.Fatal("expected at least one session")
	}

	session := course.Sessions[0]
	if session.StartTime != 1000 {
		t.Errorf("expected session start time 1000, got %d", session.StartTime)
	}
	if session.EndTime != 1140 {
		t.Errorf("expected session end time 1140, got %d", session.EndTime)
	}
	if course.Capacity != 24 {
		t.Errorf("expected capacity 24, got %d", course.Capacity)
	}
}

func TestSessionCreation(t *testing.T) {
	// Test specific session creation
	meetingFaculty := MeetingFacultyData{
		Category: "01",
		MeetingTime: MeetingTime{
			BeginTime:           "1000",
			EndTime:            "1140",
			Building:           "KB",
			Room:              "307",
			BuildingDescription: "Kaiser Borsari Hall",
			Monday:             true,
			Wednesday:          true,
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
		name       string
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
