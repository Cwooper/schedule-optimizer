// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.21.12
// source: course.proto

package generated

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Session struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Days      string `protobuf:"bytes,1,opt,name=days,proto3" json:"days,omitempty"`
	StartTime int32  `protobuf:"varint,2,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime   int32  `protobuf:"varint,3,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
	Location  string `protobuf:"bytes,4,opt,name=location,proto3" json:"location,omitempty"`
}

func (x *Session) Reset() {
	*x = Session{}
	if protoimpl.UnsafeEnabled {
		mi := &file_course_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Session) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Session) ProtoMessage() {}

func (x *Session) ProtoReflect() protoreflect.Message {
	mi := &file_course_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Session.ProtoReflect.Descriptor instead.
func (*Session) Descriptor() ([]byte, []int) {
	return file_course_proto_rawDescGZIP(), []int{0}
}

func (x *Session) GetDays() string {
	if x != nil {
		return x.Days
	}
	return ""
}

func (x *Session) GetStartTime() int32 {
	if x != nil {
		return x.StartTime
	}
	return 0
}

func (x *Session) GetEndTime() int32 {
	if x != nil {
		return x.EndTime
	}
	return 0
}

func (x *Session) GetLocation() string {
	if x != nil {
		return x.Location
	}
	return ""
}

type Course struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Subject        string     `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	Instructor     string     `protobuf:"bytes,2,opt,name=instructor,proto3" json:"instructor,omitempty"`
	Credits        int32      `protobuf:"varint,3,opt,name=credits,proto3" json:"credits,omitempty"`
	Crn            int32      `protobuf:"varint,4,opt,name=crn,proto3" json:"crn,omitempty"`
	Sessions       []*Session `protobuf:"bytes,5,rep,name=sessions,proto3" json:"sessions,omitempty"`
	Gpa            float64    `protobuf:"fixed64,6,opt,name=gpa,proto3" json:"gpa,omitempty"`
	Capacity       int32      `protobuf:"varint,7,opt,name=capacity,proto3" json:"capacity,omitempty"`
	Enrollment     int32      `protobuf:"varint,8,opt,name=enrollment,proto3" json:"enrollment,omitempty"`
	AvailableSeats int32      `protobuf:"varint,9,opt,name=available_seats,json=availableSeats,proto3" json:"available_seats,omitempty"`
	WaitlistCount  int32      `protobuf:"varint,10,opt,name=waitlist_count,json=waitlistCount,proto3" json:"waitlist_count,omitempty"`
	AdditionalFees string     `protobuf:"bytes,11,opt,name=additional_fees,json=additionalFees,proto3" json:"additional_fees,omitempty"`
	Restrictions   string     `protobuf:"bytes,12,opt,name=restrictions,proto3" json:"restrictions,omitempty"`
	Attributes     string     `protobuf:"bytes,13,opt,name=attributes,proto3" json:"attributes,omitempty"`
	Prerequisites  string     `protobuf:"bytes,14,opt,name=prerequisites,proto3" json:"prerequisites,omitempty"`
}

func (x *Course) Reset() {
	*x = Course{}
	if protoimpl.UnsafeEnabled {
		mi := &file_course_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Course) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Course) ProtoMessage() {}

func (x *Course) ProtoReflect() protoreflect.Message {
	mi := &file_course_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Course.ProtoReflect.Descriptor instead.
func (*Course) Descriptor() ([]byte, []int) {
	return file_course_proto_rawDescGZIP(), []int{1}
}

func (x *Course) GetSubject() string {
	if x != nil {
		return x.Subject
	}
	return ""
}

func (x *Course) GetInstructor() string {
	if x != nil {
		return x.Instructor
	}
	return ""
}

func (x *Course) GetCredits() int32 {
	if x != nil {
		return x.Credits
	}
	return 0
}

func (x *Course) GetCrn() int32 {
	if x != nil {
		return x.Crn
	}
	return 0
}

func (x *Course) GetSessions() []*Session {
	if x != nil {
		return x.Sessions
	}
	return nil
}

func (x *Course) GetGpa() float64 {
	if x != nil {
		return x.Gpa
	}
	return 0
}

func (x *Course) GetCapacity() int32 {
	if x != nil {
		return x.Capacity
	}
	return 0
}

func (x *Course) GetEnrollment() int32 {
	if x != nil {
		return x.Enrollment
	}
	return 0
}

func (x *Course) GetAvailableSeats() int32 {
	if x != nil {
		return x.AvailableSeats
	}
	return 0
}

func (x *Course) GetWaitlistCount() int32 {
	if x != nil {
		return x.WaitlistCount
	}
	return 0
}

func (x *Course) GetAdditionalFees() string {
	if x != nil {
		return x.AdditionalFees
	}
	return ""
}

func (x *Course) GetRestrictions() string {
	if x != nil {
		return x.Restrictions
	}
	return ""
}

func (x *Course) GetAttributes() string {
	if x != nil {
		return x.Attributes
	}
	return ""
}

func (x *Course) GetPrerequisites() string {
	if x != nil {
		return x.Prerequisites
	}
	return ""
}

type CourseArray struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Courses []*Course `protobuf:"bytes,1,rep,name=courses,proto3" json:"courses,omitempty"`
}

func (x *CourseArray) Reset() {
	*x = CourseArray{}
	if protoimpl.UnsafeEnabled {
		mi := &file_course_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CourseArray) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CourseArray) ProtoMessage() {}

func (x *CourseArray) ProtoReflect() protoreflect.Message {
	mi := &file_course_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CourseArray.ProtoReflect.Descriptor instead.
func (*CourseArray) Descriptor() ([]byte, []int) {
	return file_course_proto_rawDescGZIP(), []int{2}
}

func (x *CourseArray) GetCourses() []*Course {
	if x != nil {
		return x.Courses
	}
	return nil
}

var File_course_proto protoreflect.FileDescriptor

var file_course_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x22, 0x73, 0x0a, 0x07, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x79, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x64, 0x61, 0x79, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x65, 0x6e, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xcc, 0x03, 0x0a, 0x06,
	0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x6f, 0x72,
	0x12, 0x18, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x64, 0x69, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x07, 0x63, 0x72, 0x65, 0x64, 0x69, 0x74, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x72,
	0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x63, 0x72, 0x6e, 0x12, 0x2b, 0x0a, 0x08,
	0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52,
	0x08, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x67, 0x70, 0x61,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x67, 0x70, 0x61, 0x12, 0x1a, 0x0a, 0x08, 0x63,
	0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x63,
	0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x6e, 0x72, 0x6f, 0x6c,
	0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x65, 0x6e, 0x72,
	0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x76, 0x61, 0x69, 0x6c,
	0x61, 0x62, 0x6c, 0x65, 0x5f, 0x73, 0x65, 0x61, 0x74, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0e, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x53, 0x65, 0x61, 0x74, 0x73,
	0x12, 0x25, 0x0a, 0x0e, 0x77, 0x61, 0x69, 0x74, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x77, 0x61, 0x69, 0x74, 0x6c, 0x69,
	0x73, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x64, 0x64, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x66, 0x65, 0x65, 0x73, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0e, 0x61, 0x64, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x46, 0x65, 0x65, 0x73,
	0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x73, 0x74, 0x72, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x74, 0x72, 0x69, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x73, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62,
	0x75, 0x74, 0x65, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x70, 0x72, 0x65, 0x72, 0x65, 0x71, 0x75, 0x69,
	0x73, 0x69, 0x74, 0x65, 0x73, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x70, 0x72, 0x65,
	0x72, 0x65, 0x71, 0x75, 0x69, 0x73, 0x69, 0x74, 0x65, 0x73, 0x22, 0x37, 0x0a, 0x0b, 0x43, 0x6f,
	0x75, 0x72, 0x73, 0x65, 0x41, 0x72, 0x72, 0x61, 0x79, 0x12, 0x28, 0x0a, 0x07, 0x63, 0x6f, 0x75,
	0x72, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x52, 0x07, 0x63, 0x6f, 0x75, 0x72,
	0x73, 0x65, 0x73, 0x42, 0x2d, 0x5a, 0x2b, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x2d,
	0x6f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x72, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x65, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_course_proto_rawDescOnce sync.Once
	file_course_proto_rawDescData = file_course_proto_rawDesc
)

func file_course_proto_rawDescGZIP() []byte {
	file_course_proto_rawDescOnce.Do(func() {
		file_course_proto_rawDescData = protoimpl.X.CompressGZIP(file_course_proto_rawDescData)
	})
	return file_course_proto_rawDescData
}

var file_course_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_course_proto_goTypes = []any{
	(*Session)(nil),     // 0: models.Session
	(*Course)(nil),      // 1: models.Course
	(*CourseArray)(nil), // 2: models.CourseArray
}
var file_course_proto_depIdxs = []int32{
	0, // 0: models.Course.sessions:type_name -> models.Session
	1, // 1: models.CourseArray.courses:type_name -> models.Course
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_course_proto_init() }
func file_course_proto_init() {
	if File_course_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_course_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Session); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_course_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Course); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_course_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*CourseArray); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_course_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_course_proto_goTypes,
		DependencyIndexes: file_course_proto_depIdxs,
		MessageInfos:      file_course_proto_msgTypes,
	}.Build()
	File_course_proto = out.File
	file_course_proto_rawDesc = nil
	file_course_proto_goTypes = nil
	file_course_proto_depIdxs = nil
}
