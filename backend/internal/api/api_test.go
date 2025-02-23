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
            "courseNumber": "497F",
            "subject": "CSCI",
            "subjectDescription": "Computer Science",
            "courseTitle": "Robotics",
            "creditHourLow": 4,
            "maximumEnrollment": 24,
            "enrollment": 0,
            "seatsAvailable": 24,
            "faculty": [
                {
                    "bannerId": "82787",
                    "displayName": "Nilles, Alli",
                    "emailAddress": "nillesa2@wwu.edu",
                    "primaryIndicator": true,
                    "term": "202520"
                }
            ],
            "meetingsFaculty": [
                {
                    "category": "01",
                    "meetingTime": {
                        "beginTime": "1000",
                        "building": "KB",
                        "buildingDescription": "Kaiser Borsari Hall",
                        "campus": "M",
                        "category": "01",
                        "endTime": "1140",
                        "friday": false,
                        "monday": true,
                        "room": "307",
                        "wednesday": true
                    }
                }
            ],
            "sectionAttributes": [
                {
                    "code": "FTF",
                    "description": "DELIVERY Face-to-Face",
                    "isZTCAttribute": false,
                    "termCode": "202520"
                }
            ]
        }
    ],
    "pageOffset": 0,
    "pageMaxSize": 500,
    "pathMode": "search"
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
		{"Instructor", course.Sessions[0].Instructor, "Nilles, Alli"},
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
