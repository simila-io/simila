// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: format.proto

package format

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Id allows to provide pure id for an entity
type Id struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Id) Reset() {
	*x = Id{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Id) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Id) ProtoMessage() {}

func (x *Id) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Id.ProtoReflect.Descriptor instead.
func (*Id) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{0}
}

func (x *Id) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

// Format describes a document format
type Format struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// name uniquely identifies a format
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// basis specifies format dimensions
	Basis []byte `protobuf:"bytes,2,opt,name=basis,proto3" json:"basis,omitempty"`
}

func (x *Format) Reset() {
	*x = Format{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Format) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Format) ProtoMessage() {}

func (x *Format) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Format.ProtoReflect.Descriptor instead.
func (*Format) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{1}
}

func (x *Format) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Format) GetBasis() []byte {
	if x != nil {
		return x.Basis
	}
	return nil
}

// Formats uses as a result of List() function
type Formats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Formats []*Format `protobuf:"bytes,1,rep,name=formats,proto3" json:"formats,omitempty"`
}

func (x *Formats) Reset() {
	*x = Formats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Formats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Formats) ProtoMessage() {}

func (x *Formats) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Formats.ProtoReflect.Descriptor instead.
func (*Formats) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{2}
}

func (x *Formats) GetFormats() []*Format {
	if x != nil {
		return x.Formats
	}
	return nil
}

var File_format_proto protoreflect.FileDescriptor

var file_format_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x14, 0x0a, 0x02, 0x49, 0x64, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x32, 0x0a, 0x06,
	0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x61,
	0x73, 0x69, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x62, 0x61, 0x73, 0x69, 0x73,
	0x22, 0x36, 0x0a, 0x07, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x73, 0x12, 0x2b, 0x0a, 0x07, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x52,
	0x07, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x73, 0x32, 0xc7, 0x01, 0x0a, 0x07, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x2e, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x11,
	0x2e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61,
	0x74, 0x1a, 0x11, 0x2e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6f,
	0x72, 0x6d, 0x61, 0x74, 0x12, 0x27, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x0d, 0x2e, 0x66, 0x6f,
	0x72, 0x6d, 0x61, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x64, 0x1a, 0x11, 0x2e, 0x66, 0x6f, 0x72,
	0x6d, 0x61, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x2f, 0x0a,
	0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x0d, 0x2e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74,
	0x2e, 0x76, 0x31, 0x2e, 0x49, 0x64, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x32,
	0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x12,
	0x2e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61,
	0x74, 0x73, 0x42, 0x14, 0x5a, 0x12, 0x2e, 0x2f, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2f, 0x76,
	0x31, 0x3b, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_format_proto_rawDescOnce sync.Once
	file_format_proto_rawDescData = file_format_proto_rawDesc
)

func file_format_proto_rawDescGZIP() []byte {
	file_format_proto_rawDescOnce.Do(func() {
		file_format_proto_rawDescData = protoimpl.X.CompressGZIP(file_format_proto_rawDescData)
	})
	return file_format_proto_rawDescData
}

var file_format_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_format_proto_goTypes = []interface{}{
	(*Id)(nil),            // 0: format.v1.Id
	(*Format)(nil),        // 1: format.v1.Format
	(*Formats)(nil),       // 2: format.v1.Formats
	(*emptypb.Empty)(nil), // 3: google.protobuf.Empty
}
var file_format_proto_depIdxs = []int32{
	1, // 0: format.v1.Formats.formats:type_name -> format.v1.Format
	1, // 1: format.v1.Service.Create:input_type -> format.v1.Format
	0, // 2: format.v1.Service.Get:input_type -> format.v1.Id
	0, // 3: format.v1.Service.Delete:input_type -> format.v1.Id
	3, // 4: format.v1.Service.List:input_type -> google.protobuf.Empty
	1, // 5: format.v1.Service.Create:output_type -> format.v1.Format
	1, // 6: format.v1.Service.Get:output_type -> format.v1.Format
	3, // 7: format.v1.Service.Delete:output_type -> google.protobuf.Empty
	2, // 8: format.v1.Service.List:output_type -> format.v1.Formats
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_format_proto_init() }
func file_format_proto_init() {
	if File_format_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_format_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Id); i {
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
		file_format_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Format); i {
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
		file_format_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Formats); i {
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
			RawDescriptor: file_format_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_format_proto_goTypes,
		DependencyIndexes: file_format_proto_depIdxs,
		MessageInfos:      file_format_proto_msgTypes,
	}.Build()
	File_format_proto = out.File
	file_format_proto_rawDesc = nil
	file_format_proto_goTypes = nil
	file_format_proto_depIdxs = nil
}
