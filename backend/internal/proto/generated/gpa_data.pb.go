// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.21.12
// source: gpa_data.proto

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

type GPAData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Subjects   map[string]float64 `protobuf:"bytes,1,rep,name=subjects,proto3" json:"subjects,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed64,2,opt,name=value,proto3"`
	Professors map[string]float64 `protobuf:"bytes,2,rep,name=professors,proto3" json:"professors,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed64,2,opt,name=value,proto3"`
	CourseGpas map[string]float64 `protobuf:"bytes,3,rep,name=course_gpas,json=courseGpas,proto3" json:"course_gpas,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed64,2,opt,name=value,proto3"`
}

func (x *GPAData) Reset() {
	*x = GPAData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gpa_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GPAData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GPAData) ProtoMessage() {}

func (x *GPAData) ProtoReflect() protoreflect.Message {
	mi := &file_gpa_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GPAData.ProtoReflect.Descriptor instead.
func (*GPAData) Descriptor() ([]byte, []int) {
	return file_gpa_data_proto_rawDescGZIP(), []int{0}
}

func (x *GPAData) GetSubjects() map[string]float64 {
	if x != nil {
		return x.Subjects
	}
	return nil
}

func (x *GPAData) GetProfessors() map[string]float64 {
	if x != nil {
		return x.Professors
	}
	return nil
}

func (x *GPAData) GetCourseGpas() map[string]float64 {
	if x != nil {
		return x.CourseGpas
	}
	return nil
}

var File_gpa_data_proto protoreflect.FileDescriptor

var file_gpa_data_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x67, 0x70, 0x61, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x22, 0x82, 0x03, 0x0a, 0x07, 0x47, 0x50, 0x41,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x39, 0x0a, 0x08, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x47, 0x50, 0x41, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x12,
	0x3f, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x66, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x47, 0x50, 0x41,
	0x44, 0x61, 0x74, 0x61, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x66, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x73,
	0x12, 0x40, 0x0a, 0x0b, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x5f, 0x67, 0x70, 0x61, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x47,
	0x50, 0x41, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x47, 0x70, 0x61,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x47, 0x70,
	0x61, 0x73, 0x1a, 0x3b, 0x0a, 0x0d, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a,
	0x3d, 0x0a, 0x0f, 0x50, 0x72, 0x6f, 0x66, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x3d,
	0x0a, 0x0f, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x47, 0x70, 0x61, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x2d, 0x5a,
	0x2b, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x2d, 0x6f, 0x70, 0x74, 0x69, 0x6d, 0x69,
	0x7a, 0x65, 0x72, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_gpa_data_proto_rawDescOnce sync.Once
	file_gpa_data_proto_rawDescData = file_gpa_data_proto_rawDesc
)

func file_gpa_data_proto_rawDescGZIP() []byte {
	file_gpa_data_proto_rawDescOnce.Do(func() {
		file_gpa_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_gpa_data_proto_rawDescData)
	})
	return file_gpa_data_proto_rawDescData
}

var file_gpa_data_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_gpa_data_proto_goTypes = []any{
	(*GPAData)(nil), // 0: models.GPAData
	nil,             // 1: models.GPAData.SubjectsEntry
	nil,             // 2: models.GPAData.ProfessorsEntry
	nil,             // 3: models.GPAData.CourseGpasEntry
}
var file_gpa_data_proto_depIdxs = []int32{
	1, // 0: models.GPAData.subjects:type_name -> models.GPAData.SubjectsEntry
	2, // 1: models.GPAData.professors:type_name -> models.GPAData.ProfessorsEntry
	3, // 2: models.GPAData.course_gpas:type_name -> models.GPAData.CourseGpasEntry
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_gpa_data_proto_init() }
func file_gpa_data_proto_init() {
	if File_gpa_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_gpa_data_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*GPAData); i {
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
			RawDescriptor: file_gpa_data_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_gpa_data_proto_goTypes,
		DependencyIndexes: file_gpa_data_proto_depIdxs,
		MessageInfos:      file_gpa_data_proto_msgTypes,
	}.Build()
	File_gpa_data_proto = out.File
	file_gpa_data_proto_rawDesc = nil
	file_gpa_data_proto_goTypes = nil
	file_gpa_data_proto_depIdxs = nil
}
