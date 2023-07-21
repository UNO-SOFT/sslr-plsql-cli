// Copyright 2021 Tamás Gulácsi. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.4.0
// source: functions.proto

package pb

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

type Object struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string      `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Functions []*Function `protobuf:"bytes,2,rep,name=Functions,proto3" json:"Functions,omitempty"`
}

func (x *Object) Reset() {
	*x = Object{}
	if protoimpl.UnsafeEnabled {
		mi := &file_functions_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Object) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Object) ProtoMessage() {}

func (x *Object) ProtoReflect() protoreflect.Message {
	mi := &file_functions_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Object.ProtoReflect.Descriptor instead.
func (*Object) Descriptor() ([]byte, []int) {
	return file_functions_proto_rawDescGZIP(), []int{0}
}

func (x *Object) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Object) GetFunctions() []*Function {
	if x != nil {
		return x.Functions
	}
	return nil
}

type Function struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string  `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Parent string  `protobuf:"bytes,2,opt,name=Parent,proto3" json:"Parent,omitempty"`
	Begin  uint32  `protobuf:"varint,3,opt,name=Begin,proto3" json:"Begin,omitempty"`
	End    uint32  `protobuf:"varint,4,opt,name=End,proto3" json:"End,omitempty"`
	Level  uint32  `protobuf:"varint,5,opt,name=Level,proto3" json:"Level,omitempty"`
	Calls  []*Call `protobuf:"bytes,6,rep,name=Calls,proto3" json:"Calls,omitempty"`
}

func (x *Function) Reset() {
	*x = Function{}
	if protoimpl.UnsafeEnabled {
		mi := &file_functions_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Function) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Function) ProtoMessage() {}

func (x *Function) ProtoReflect() protoreflect.Message {
	mi := &file_functions_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Function.ProtoReflect.Descriptor instead.
func (*Function) Descriptor() ([]byte, []int) {
	return file_functions_proto_rawDescGZIP(), []int{1}
}

func (x *Function) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Function) GetParent() string {
	if x != nil {
		return x.Parent
	}
	return ""
}

func (x *Function) GetBegin() uint32 {
	if x != nil {
		return x.Begin
	}
	return 0
}

func (x *Function) GetEnd() uint32 {
	if x != nil {
		return x.End
	}
	return 0
}

func (x *Function) GetLevel() uint32 {
	if x != nil {
		return x.Level
	}
	return 0
}

func (x *Function) GetCalls() []*Call {
	if x != nil {
		return x.Calls
	}
	return nil
}

type Call struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Other     string `protobuf:"bytes,1,opt,name=Other,proto3" json:"Other,omitempty"`
	Line      uint32 `protobuf:"varint,2,opt,name=Line,proto3" json:"Line,omitempty"`
	Procedure bool   `protobuf:"varint,3,opt,name=Procedure,proto3" json:"Procedure,omitempty"`
}

func (x *Call) Reset() {
	*x = Call{}
	if protoimpl.UnsafeEnabled {
		mi := &file_functions_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Call) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Call) ProtoMessage() {}

func (x *Call) ProtoReflect() protoreflect.Message {
	mi := &file_functions_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Call.ProtoReflect.Descriptor instead.
func (*Call) Descriptor() ([]byte, []int) {
	return file_functions_proto_rawDescGZIP(), []int{2}
}

func (x *Call) GetOther() string {
	if x != nil {
		return x.Other
	}
	return ""
}

func (x *Call) GetLine() uint32 {
	if x != nil {
		return x.Line
	}
	return 0
}

func (x *Call) GetProcedure() bool {
	if x != nil {
		return x.Procedure
	}
	return false
}

var File_functions_proto protoreflect.FileDescriptor

var file_functions_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x45, 0x0a, 0x06, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x27, 0x0a, 0x09, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x09, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x46,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x91, 0x01, 0x0a, 0x08, 0x46, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x61, 0x72,
	0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x50, 0x61, 0x72, 0x65, 0x6e,
	0x74, 0x12, 0x14, 0x0a, 0x05, 0x42, 0x65, 0x67, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x05, 0x42, 0x65, 0x67, 0x69, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x45, 0x6e, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x45, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x4c, 0x65, 0x76,
	0x65, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12,
	0x1b, 0x0a, 0x05, 0x43, 0x61, 0x6c, 0x6c, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x05,
	0x2e, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x05, 0x43, 0x61, 0x6c, 0x6c, 0x73, 0x22, 0x4e, 0x0a, 0x04,
	0x43, 0x61, 0x6c, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x4f, 0x74, 0x68, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x4f, 0x74, 0x68, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x4c, 0x69,
	0x6e, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x4c, 0x69, 0x6e, 0x65, 0x12, 0x1c,
	0x0a, 0x09, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x64, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x09, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x64, 0x75, 0x72, 0x65, 0x42, 0x27, 0x5a, 0x25,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x55, 0x4e, 0x4f, 0x2d, 0x53,
	0x4f, 0x46, 0x54, 0x2f, 0x73, 0x73, 0x6c, 0x72, 0x2d, 0x70, 0x6c, 0x73, 0x71, 0x6c, 0x2d, 0x63,
	0x6c, 0x69, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_functions_proto_rawDescOnce sync.Once
	file_functions_proto_rawDescData = file_functions_proto_rawDesc
)

func file_functions_proto_rawDescGZIP() []byte {
	file_functions_proto_rawDescOnce.Do(func() {
		file_functions_proto_rawDescData = protoimpl.X.CompressGZIP(file_functions_proto_rawDescData)
	})
	return file_functions_proto_rawDescData
}

var file_functions_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_functions_proto_goTypes = []interface{}{
	(*Object)(nil),   // 0: Object
	(*Function)(nil), // 1: Function
	(*Call)(nil),     // 2: Call
}
var file_functions_proto_depIdxs = []int32{
	1, // 0: Object.Functions:type_name -> Function
	2, // 1: Function.Calls:type_name -> Call
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_functions_proto_init() }
func file_functions_proto_init() {
	if File_functions_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_functions_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Object); i {
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
		file_functions_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Function); i {
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
		file_functions_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Call); i {
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
			RawDescriptor: file_functions_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_functions_proto_goTypes,
		DependencyIndexes: file_functions_proto_depIdxs,
		MessageInfos:      file_functions_proto_msgTypes,
	}.Build()
	File_functions_proto = out.File
	file_functions_proto_rawDesc = nil
	file_functions_proto_goTypes = nil
	file_functions_proto_depIdxs = nil
}
