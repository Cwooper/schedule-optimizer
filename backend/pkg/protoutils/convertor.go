// Package protoutils for converting courses to and from protobufs
package protoutils

import (
	"github.com/cwooper/schedule-optimizer/internal/models"
	pb "github.com/cwooper/schedule-optimizer/internal/proto/generated"
)

// CoursesToProto converts an array of Course structs to a CourseList protobuf message
func CoursesToProto(courses []models.Course) *pb.CourseList {
	pbCourses := make([]*pb.Course, len(courses))
	for i, course := range courses {
		pbCourses[i] = CourseToProto(course)
	}
	return &pb.CourseList{
		Courses: pbCourses,
	}
}

// ProtoToCourses converts a CourseList protobuf message to an array of Course structs
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
		Instructor:     course.Instructor,
		Sessions:       sessionsToProto(course.Sessions),
		Gpa:            course.GPA,
		Capacity:       int32(course.Capacity),
		Enrolled:       int32(course.Enrolled),
		AvailableSeats: int32(course.AvailableSeats),
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
		Instructor:     pbCourse.Instructor,
		Sessions:       protoToSessions(pbCourse.Sessions),
		GPA:            pbCourse.Gpa,
		Capacity:       int(pbCourse.Capacity),
		Enrolled:       int(pbCourse.Enrolled),
		AvailableSeats: int(pbCourse.AvailableSeats),
		CourseString:   pbCourse.CourseString,
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
			IsAsync:   session.IsAsync,
			IsTimeTbd: session.IsTimeTBD,
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
			IsAsync:   pbSession.IsAsync,
			IsTimeTBD: pbSession.IsTimeTbd,
		}
	}
	return sessions
}

// GPADataToProto converts a GPAData struct to a GPAData protobuf message
func GPADataToProto(gpaData models.GPAData) *pb.GPAData {
	// Convert ProfessorSubjects to proto format
	professorSubjects := make(map[string]*pb.GPAData_SubjectSet)
	for prof, subjects := range gpaData.ProfessorSubjects {
		subjectList := make([]string, 0, len(subjects))
		for subject := range subjects {
			subjectList = append(subjectList, subject)
		}
		professorSubjects[prof] = &pb.GPAData_SubjectSet{
			Subjects: subjectList,
		}
	}

	// Convert LastNameIndex to proto format
	lastNameIndex := make(map[string]*pb.GPAData_ProfessorList)
	for lastName, professors := range gpaData.LastNameIndex {
		lastNameIndex[lastName] = &pb.GPAData_ProfessorList{
			Names: professors,
		}
	}

	return &pb.GPAData{
		Subjects:          gpaData.Subjects,
		Professors:        gpaData.Professors,
		CourseGpas:        gpaData.CourseGPAs,
		ProfessorSubjects: professorSubjects,
		LastNameIndex:     lastNameIndex,
	}
}

// ProtoToGPAData converts a GPAData protobuf message to a GPAData struct
func ProtoToGPAData(pbGPAData *pb.GPAData) models.GPAData {
	// Convert ProfessorSubjects from proto format
	professorSubjects := make(models.ProfessorSubjects)
	for prof, subjectSet := range pbGPAData.ProfessorSubjects {
		professorSubjects[prof] = make(map[string]struct{})
		for _, subject := range subjectSet.Subjects {
			professorSubjects[prof][subject] = struct{}{}
		}
	}

	// Convert LastNameIndex from proto format
	lastNameIndex := make(map[string][]string)
	for lastName, profList := range pbGPAData.LastNameIndex {
		lastNameIndex[lastName] = profList.Names
	}

	return models.GPAData{
		Subjects:          pbGPAData.Subjects,
		Professors:        pbGPAData.Professors,
		CourseGPAs:        pbGPAData.CourseGpas,
		ProfessorSubjects: professorSubjects,
		LastNameIndex:     lastNameIndex,
	}
}
