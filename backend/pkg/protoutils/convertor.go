// Package protoutils for converting courses to and from protobufs
package protoutils

import (
	"github.com/cwooper/schedule-optimizer/internal/models"
	pb "github.com/cwooper/schedule-optimizer/internal/proto/generated"
)

// CoursesToProto converts an array of Course structs to a CourseArray protobuf message
func CoursesToProto(courses []models.Course) *pb.CourseList {
	pbCourses := make([]*pb.Course, len(courses))
	for i, course := range courses {
		pbCourses[i] = CourseToProto(course)
	}
	return &pb.CourseList{
		Courses: pbCourses,
	}
}

// ProtoToCourses converts a CourseArray protobuf message to an array of Course structs
func ProtoToCourses(pbCourseArray *pb.CourseList) []models.Course {
	courses := make([]models.Course, len(pbCourseArray.Courses))
	for i, pbCourse := range pbCourseArray.Courses {
		courses[i] = ProtoToCourse(pbCourse)
	}
	return courses
}

// CourseToProto converts a single Course struct to a Course protobuf message
func CourseToProto(course models.Course) *pb.Course {
	return &pb.Course{
		Subject:        course.Subject,
		Title:          course.Title,
		Credits:        course.Credits,
		Crn:            int32(course.CRN),
		Sessions:       sessionsToProto(course.Sessions),
		Gpa:            course.GPA,
		Capacity:       int32(course.Capacity),
		Enrolled:       int32(course.Enrolled),
		AvailableSeats: int32(course.AvailableSeats),
		AdditionalFees: course.AdditionalFees,
		Restrictions:   course.Restrictions,
		Attributes:     course.Attributes,
		Prerequisites:  course.Prerequisites,
		CourseString:   course.CourseString,
	}
}

// ProtoToCourse converts a single Course protobuf message to a Course struct
func ProtoToCourse(pbCourse *pb.Course) models.Course {
	return models.Course{
		Subject:        pbCourse.Subject,
		Title:          pbCourse.Title,
		Credits:        pbCourse.Credits,
		CRN:            int(pbCourse.Crn),
		Sessions:       protoToSessions(pbCourse.Sessions),
		GPA:            pbCourse.Gpa,
		Capacity:       int(pbCourse.Capacity),
		Enrolled:       int(pbCourse.Enrolled),
		AvailableSeats: int(pbCourse.AvailableSeats),
		AdditionalFees: pbCourse.AdditionalFees,
		Restrictions:   pbCourse.Restrictions,
		Attributes:     pbCourse.Attributes,
		Prerequisites:  pbCourse.Prerequisites,
		CourseString:   pbCourse.CourseString,
	}
}

// sessionsToProto converts an array of Session structs to an array of Session protobuf messages
func sessionsToProto(sessions []models.Session) []*pb.Session {
	pbSessions := make([]*pb.Session, len(sessions))
	for i, session := range sessions {
		pbSessions[i] = &pb.Session{
			Days:       session.Days,
			StartTime:  int32(session.StartTime),
			EndTime:    int32(session.EndTime),
			Instructor: session.Instructor,
			Location:   session.Location,
			IsAsync:    session.IsAsync,
			IsTimeTbd:  session.IsTimeTBD,
		}
	}
	return pbSessions
}

// protoToSessions converts an array of Session protobuf messages to an array of Session structs
func protoToSessions(pbSessions []*pb.Session) []models.Session {
	sessions := make([]models.Session, len(pbSessions))
	for i, pbSession := range pbSessions {
		sessions[i] = models.Session{
			Days:       pbSession.Days,
			StartTime:  int(pbSession.StartTime),
			EndTime:    int(pbSession.EndTime),
			Instructor: pbSession.Instructor,
			Location:   pbSession.Location,
			IsAsync:    pbSession.IsAsync,
			IsTimeTBD:  pbSession.IsTimeTbd,
		}
	}
	return sessions
}

// GPADataToProto converts a GPAData struct to a GPAData protobuf message
func GPADataToProto(gpaData models.GPAData) *pb.GPAData {
	return &pb.GPAData{
		Subjects:   gpaData.Subjects,
		Professors: gpaData.Professors,
		CourseGpas: gpaData.CourseGPAs,
	}
}

// ProtoToGPAData converts a GPAData protobuf message to a GPAData struct
func ProtoToGPAData(pbGPAData *pb.GPAData) models.GPAData {
	return models.GPAData{
		Subjects:   pbGPAData.Subjects,
		Professors: pbGPAData.Professors,
		CourseGPAs: pbGPAData.CourseGpas,
	}
}
