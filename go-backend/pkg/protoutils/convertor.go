// Package protoutils for converting courses to and from protobufs
package protoutils

import (
	"schedule-optimizer/internal/models"
	pb "schedule-optimizer/internal/proto/generated"
)

// CoursesToProto converts an array of Course structs to a CourseArray protobuf message
func CoursesToProto(courses []models.Course) *pb.CourseArray {
	pbCourses := make([]*pb.Course, len(courses))
	for i, course := range courses {
		pbCourses[i] = CourseToProto(course)
	}
	return &pb.CourseArray{
		Courses: pbCourses,
	}
}

// ProtoToCourses converts a CourseArray protobuf message to an array of Course structs
func ProtoToCourses(pbCourseArray *pb.CourseArray) []models.Course {
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
		Instructor:     course.Instructor,
		Credits:        int32(course.Credits),
		Crn:            int32(course.CRN),
		Sessions:       sessionsToProto(course.Sessions),
		Gpa:            course.GPA,
		Capacity:       int32(course.Capacity),
		Enrollment:     int32(course.Enrollment),
		AvailableSeats: int32(course.AvailableSeats),
		WaitlistCount:  int32(course.WaitlistCount),
		AdditionalFees: course.AdditionalFees,
		Restrictions:   course.Restrictions,
		Attributes:     course.Attributes,
		Prerequisites:  course.Prerequisites,
	}
}

// ProtoToCourse converts a single Course protobuf message to a Course struct
func ProtoToCourse(pbCourse *pb.Course) models.Course {
	return models.Course{
		Subject:        pbCourse.Subject,
		Instructor:     pbCourse.Instructor,
		Credits:        int(pbCourse.Credits),
		CRN:            int(pbCourse.Crn),
		Sessions:       protoToSessions(pbCourse.Sessions),
		GPA:            pbCourse.Gpa,
		Capacity:       int(pbCourse.Capacity),
		Enrollment:     int(pbCourse.Enrollment),
		AvailableSeats: int(pbCourse.AvailableSeats),
		WaitlistCount:  int(pbCourse.WaitlistCount),
		AdditionalFees: pbCourse.AdditionalFees,
		Restrictions:   pbCourse.Restrictions,
		Attributes:     pbCourse.Attributes,
		Prerequisites:  pbCourse.Prerequisites,
	}
}

// sessionsToProto converts an array of Session structs to an array of Session protobuf messages
func sessionsToProto(sessions []models.Session) []*pb.Session {
	pbSessions := make([]*pb.Session, len(sessions))
	for i, session := range sessions {
		pbSessions[i] = &pb.Session{
			Days:      session.Days,
			StartTime: int32(session.StartTime),
			EndTime:   int32(session.EndTime),
			Location:  session.Location,
		}
	}
	return pbSessions
}

// protoToSessions converts an array of Session protobuf messages to an array of Session structs
func protoToSessions(pbSessions []*pb.Session) []models.Session {
	sessions := make([]models.Session, len(pbSessions))
	for i, pbSession := range pbSessions {
		sessions[i] = models.Session{
			Days:      pbSession.Days,
			StartTime: int(pbSession.StartTime),
			EndTime:   int(pbSession.EndTime),
			Location:  pbSession.Location,
		}
	}
	return sessions
}