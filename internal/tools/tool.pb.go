// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.24.3
// source: tools/tool.proto

package tools

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

// List of all the possible tools
type ToolType int32

const (
	ToolType_COMMAND ToolType = 0
	ToolType_ACTION  ToolType = 1
	ToolType_WIDGET  ToolType = 2
	ToolType_HOOK    ToolType = 3
)

// Enum value maps for ToolType.
var (
	ToolType_name = map[int32]string{
		0: "COMMAND",
		1: "ACTION",
		2: "WIDGET",
		3: "HOOK",
	}
	ToolType_value = map[string]int32{
		"COMMAND": 0,
		"ACTION":  1,
		"WIDGET":  2,
		"HOOK":    3,
	}
)

func (x ToolType) Enum() *ToolType {
	p := new(ToolType)
	*p = x
	return p
}

func (x ToolType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ToolType) Descriptor() protoreflect.EnumDescriptor {
	return file_tools_tool_proto_enumTypes[0].Descriptor()
}

func (ToolType) Type() protoreflect.EnumType {
	return &file_tools_tool_proto_enumTypes[0]
}

func (x ToolType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ToolType.Descriptor instead.
func (ToolType) EnumDescriptor() ([]byte, []int) {
	return file_tools_tool_proto_rawDescGZIP(), []int{0}
}

// Various states a tool can have, they are ordered by importance
type ToolStatus int32

const (
	ToolStatus_INITIALIZING ToolStatus = 0
	ToolStatus_INITIALIZED  ToolStatus = 1
	ToolStatus_RUNNING      ToolStatus = 2
	ToolStatus_PAUSED       ToolStatus = 3
	ToolStatus_ERROR        ToolStatus = 4
	ToolStatus_INVALID      ToolStatus = 5
)

// Enum value maps for ToolStatus.
var (
	ToolStatus_name = map[int32]string{
		0: "INITIALIZING",
		1: "INITIALIZED",
		2: "RUNNING",
		3: "PAUSED",
		4: "ERROR",
		5: "INVALID",
	}
	ToolStatus_value = map[string]int32{
		"INITIALIZING": 0,
		"INITIALIZED":  1,
		"RUNNING":      2,
		"PAUSED":       3,
		"ERROR":        4,
		"INVALID":      5,
	}
)

func (x ToolStatus) Enum() *ToolStatus {
	p := new(ToolStatus)
	*p = x
	return p
}

func (x ToolStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ToolStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_tools_tool_proto_enumTypes[1].Descriptor()
}

func (ToolStatus) Type() protoreflect.EnumType {
	return &file_tools_tool_proto_enumTypes[1]
}

func (x ToolStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ToolStatus.Descriptor instead.
func (ToolStatus) EnumDescriptor() ([]byte, []int) {
	return file_tools_tool_proto_rawDescGZIP(), []int{1}
}

// A socket acts as a payload for an input/output between two tools
type Socket struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The raw data can then be casted into a different datatype
	//
	// Types that are assignable to Raw:
	//
	//	*Socket_RawString
	//	*Socket_RawInteger
	//	*Socket_RawNumber
	//	*Socket_RawBinary
	Raw isSocket_Raw `protobuf_oneof:"raw"`
	// Indicator for the client to cast the raw data into any datatype
	Cast string `protobuf:"bytes,5,opt,name=cast,proto3" json:"cast,omitempty"`
}

func (x *Socket) Reset() {
	*x = Socket{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tools_tool_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Socket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Socket) ProtoMessage() {}

func (x *Socket) ProtoReflect() protoreflect.Message {
	mi := &file_tools_tool_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Socket.ProtoReflect.Descriptor instead.
func (*Socket) Descriptor() ([]byte, []int) {
	return file_tools_tool_proto_rawDescGZIP(), []int{0}
}

func (m *Socket) GetRaw() isSocket_Raw {
	if m != nil {
		return m.Raw
	}
	return nil
}

func (x *Socket) GetRawString() string {
	if x, ok := x.GetRaw().(*Socket_RawString); ok {
		return x.RawString
	}
	return ""
}

func (x *Socket) GetRawInteger() int32 {
	if x, ok := x.GetRaw().(*Socket_RawInteger); ok {
		return x.RawInteger
	}
	return 0
}

func (x *Socket) GetRawNumber() float32 {
	if x, ok := x.GetRaw().(*Socket_RawNumber); ok {
		return x.RawNumber
	}
	return 0
}

func (x *Socket) GetRawBinary() []byte {
	if x, ok := x.GetRaw().(*Socket_RawBinary); ok {
		return x.RawBinary
	}
	return nil
}

func (x *Socket) GetCast() string {
	if x != nil {
		return x.Cast
	}
	return ""
}

type isSocket_Raw interface {
	isSocket_Raw()
}

type Socket_RawString struct {
	RawString string `protobuf:"bytes,1,opt,name=raw_string,json=rawString,proto3,oneof"`
}

type Socket_RawInteger struct {
	RawInteger int32 `protobuf:"varint,2,opt,name=raw_integer,json=rawInteger,proto3,oneof"`
}

type Socket_RawNumber struct {
	RawNumber float32 `protobuf:"fixed32,3,opt,name=raw_number,json=rawNumber,proto3,oneof"`
}

type Socket_RawBinary struct {
	RawBinary []byte `protobuf:"bytes,4,opt,name=raw_binary,json=rawBinary,proto3,oneof"`
}

func (*Socket_RawString) isSocket_Raw() {}

func (*Socket_RawInteger) isSocket_Raw() {}

func (*Socket_RawNumber) isSocket_Raw() {}

func (*Socket_RawBinary) isSocket_Raw() {}

// A tool expose functionalities to the user. There is multiple types of tools,
// like actions, or commands. Almost all the fields are optional because we
// might receive tool patches
type ToolBase struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The Id is the only required field since it is used to differentiate tools
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The name should be unique and be as simple as possible
	Name *string `protobuf:"bytes,2,opt,name=name,proto3,oneof" json:"name,omitempty"`
	// This helps defining the type of tool definition to use when
	// deserializing the protobuf
	Type *ToolType `protobuf:"varint,3,opt,name=type,proto3,enum=zorro.ToolType,oneof" json:"type,omitempty"`
	// User friently version of the name without all its contraints
	Label *string `protobuf:"bytes,4,opt,name=label,proto3,oneof" json:"label,omitempty"`
	// The status is only used for user feedback
	Status ToolStatus `protobuf:"varint,5,opt,name=status,proto3,enum=zorro.ToolStatus" json:"status,omitempty"`
	// Inputs and outputs
	Inputs  map[string]*Socket `protobuf:"bytes,6,rep,name=inputs,proto3" json:"inputs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Outputs map[string]*Socket `protobuf:"bytes,7,rep,name=outputs,proto3" json:"outputs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The tooltip is only for the user
	Tooltip *string `protobuf:"bytes,8,opt,name=tooltip,proto3,oneof" json:"tooltip,omitempty"`
	// Logs can be heavy, only the last ones are usually sent
	Logs map[int32]string `protobuf:"bytes,9,rep,name=logs,proto3" json:"logs,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ToolBase) Reset() {
	*x = ToolBase{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tools_tool_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ToolBase) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToolBase) ProtoMessage() {}

func (x *ToolBase) ProtoReflect() protoreflect.Message {
	mi := &file_tools_tool_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToolBase.ProtoReflect.Descriptor instead.
func (*ToolBase) Descriptor() ([]byte, []int) {
	return file_tools_tool_proto_rawDescGZIP(), []int{1}
}

func (x *ToolBase) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ToolBase) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

func (x *ToolBase) GetType() ToolType {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return ToolType_COMMAND
}

func (x *ToolBase) GetLabel() string {
	if x != nil && x.Label != nil {
		return *x.Label
	}
	return ""
}

func (x *ToolBase) GetStatus() ToolStatus {
	if x != nil {
		return x.Status
	}
	return ToolStatus_INITIALIZING
}

func (x *ToolBase) GetInputs() map[string]*Socket {
	if x != nil {
		return x.Inputs
	}
	return nil
}

func (x *ToolBase) GetOutputs() map[string]*Socket {
	if x != nil {
		return x.Outputs
	}
	return nil
}

func (x *ToolBase) GetTooltip() string {
	if x != nil && x.Tooltip != nil {
		return *x.Tooltip
	}
	return ""
}

func (x *ToolBase) GetLogs() map[int32]string {
	if x != nil {
		return x.Logs
	}
	return nil
}

// An abstract tool representation
type Tool struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// All tools are composed of this field that contains required infos
	Base *ToolBase `protobuf:"bytes,1,opt,name=base,proto3" json:"base,omitempty"`
}

func (x *Tool) Reset() {
	*x = Tool{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tools_tool_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tool) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tool) ProtoMessage() {}

func (x *Tool) ProtoReflect() protoreflect.Message {
	mi := &file_tools_tool_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tool.ProtoReflect.Descriptor instead.
func (*Tool) Descriptor() ([]byte, []int) {
	return file_tools_tool_proto_rawDescGZIP(), []int{2}
}

func (x *Tool) GetBase() *ToolBase {
	if x != nil {
		return x.Base
	}
	return nil
}

var File_tools_tool_proto protoreflect.FileDescriptor

var file_tools_tool_proto_rawDesc = []byte{
	0x0a, 0x10, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x05, 0x7a, 0x6f, 0x72, 0x72, 0x6f, 0x22, 0xa9, 0x01, 0x0a, 0x06, 0x53, 0x6f,
	0x63, 0x6b, 0x65, 0x74, 0x12, 0x1f, 0x0a, 0x0a, 0x72, 0x61, 0x77, 0x5f, 0x73, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x09, 0x72, 0x61, 0x77, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x21, 0x0a, 0x0b, 0x72, 0x61, 0x77, 0x5f, 0x69, 0x6e, 0x74,
	0x65, 0x67, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x0a, 0x72, 0x61,
	0x77, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x65, 0x72, 0x12, 0x1f, 0x0a, 0x0a, 0x72, 0x61, 0x77, 0x5f,
	0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x02, 0x48, 0x00, 0x52, 0x09,
	0x72, 0x61, 0x77, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1f, 0x0a, 0x0a, 0x72, 0x61, 0x77,
	0x5f, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52,
	0x09, 0x72, 0x61, 0x77, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x61,
	0x73, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x61, 0x73, 0x74, 0x42, 0x05,
	0x0a, 0x03, 0x72, 0x61, 0x77, 0x22, 0xd4, 0x04, 0x0a, 0x08, 0x54, 0x6f, 0x6f, 0x6c, 0x42, 0x61,
	0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x17, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x12, 0x28, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0f, 0x2e, 0x7a, 0x6f, 0x72, 0x72,
	0x6f, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x48, 0x01, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x88, 0x01, 0x01, 0x12, 0x19, 0x0a, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x02, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x88, 0x01, 0x01,
	0x12, 0x29, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x11, 0x2e, 0x7a, 0x6f, 0x72, 0x72, 0x6f, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x33, 0x0a, 0x06, 0x69,
	0x6e, 0x70, 0x75, 0x74, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x7a, 0x6f,
	0x72, 0x72, 0x6f, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x42, 0x61, 0x73, 0x65, 0x2e, 0x49, 0x6e, 0x70,
	0x75, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73,
	0x12, 0x36, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x1c, 0x2e, 0x7a, 0x6f, 0x72, 0x72, 0x6f, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x42, 0x61,
	0x73, 0x65, 0x2e, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x12, 0x1d, 0x0a, 0x07, 0x74, 0x6f, 0x6f, 0x6c,
	0x74, 0x69, 0x70, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x48, 0x03, 0x52, 0x07, 0x74, 0x6f, 0x6f,
	0x6c, 0x74, 0x69, 0x70, 0x88, 0x01, 0x01, 0x12, 0x2d, 0x0a, 0x04, 0x6c, 0x6f, 0x67, 0x73, 0x18,
	0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x7a, 0x6f, 0x72, 0x72, 0x6f, 0x2e, 0x54, 0x6f,
	0x6f, 0x6c, 0x42, 0x61, 0x73, 0x65, 0x2e, 0x4c, 0x6f, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x04, 0x6c, 0x6f, 0x67, 0x73, 0x1a, 0x48, 0x0a, 0x0b, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x23, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x7a, 0x6f, 0x72, 0x72, 0x6f, 0x2e, 0x53,
	0x6f, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x1a, 0x49, 0x0a, 0x0c, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x23, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0d, 0x2e, 0x7a, 0x6f, 0x72, 0x72, 0x6f, 0x2e, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x37, 0x0a, 0x09, 0x4c,
	0x6f, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x07, 0x0a,
	0x05, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x6c, 0x61, 0x62, 0x65, 0x6c,
	0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x74, 0x6f, 0x6f, 0x6c, 0x74, 0x69, 0x70, 0x22, 0x2b, 0x0a, 0x04,
	0x54, 0x6f, 0x6f, 0x6c, 0x12, 0x23, 0x0a, 0x04, 0x62, 0x61, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x7a, 0x6f, 0x72, 0x72, 0x6f, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x42,
	0x61, 0x73, 0x65, 0x52, 0x04, 0x62, 0x61, 0x73, 0x65, 0x2a, 0x39, 0x0a, 0x08, 0x54, 0x6f, 0x6f,
	0x6c, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x4d, 0x4d, 0x41, 0x4e, 0x44,
	0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x01, 0x12, 0x0a,
	0x0a, 0x06, 0x57, 0x49, 0x44, 0x47, 0x45, 0x54, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x48, 0x4f,
	0x4f, 0x4b, 0x10, 0x03, 0x2a, 0x60, 0x0a, 0x0a, 0x54, 0x6f, 0x6f, 0x6c, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x10, 0x0a, 0x0c, 0x49, 0x4e, 0x49, 0x54, 0x49, 0x41, 0x4c, 0x49, 0x5a, 0x49,
	0x4e, 0x47, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x49, 0x4e, 0x49, 0x54, 0x49, 0x41, 0x4c, 0x49,
	0x5a, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x55, 0x4e, 0x4e, 0x49, 0x4e, 0x47,
	0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x50, 0x41, 0x55, 0x53, 0x45, 0x44, 0x10, 0x03, 0x12, 0x09,
	0x0a, 0x05, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x04, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x4e, 0x56,
	0x41, 0x4c, 0x49, 0x44, 0x10, 0x05, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x41, 0x63, 0x65, 0x64, 0x79, 0x6e, 0x2f, 0x7a, 0x6f, 0x72, 0x72,
	0x6f, 0x2d, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f,
	0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_tools_tool_proto_rawDescOnce sync.Once
	file_tools_tool_proto_rawDescData = file_tools_tool_proto_rawDesc
)

func file_tools_tool_proto_rawDescGZIP() []byte {
	file_tools_tool_proto_rawDescOnce.Do(func() {
		file_tools_tool_proto_rawDescData = protoimpl.X.CompressGZIP(file_tools_tool_proto_rawDescData)
	})
	return file_tools_tool_proto_rawDescData
}

var file_tools_tool_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_tools_tool_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_tools_tool_proto_goTypes = []interface{}{
	(ToolType)(0),    // 0: zorro.ToolType
	(ToolStatus)(0),  // 1: zorro.ToolStatus
	(*Socket)(nil),   // 2: zorro.Socket
	(*ToolBase)(nil), // 3: zorro.ToolBase
	(*Tool)(nil),     // 4: zorro.Tool
	nil,              // 5: zorro.ToolBase.InputsEntry
	nil,              // 6: zorro.ToolBase.OutputsEntry
	nil,              // 7: zorro.ToolBase.LogsEntry
}
var file_tools_tool_proto_depIdxs = []int32{
	0, // 0: zorro.ToolBase.type:type_name -> zorro.ToolType
	1, // 1: zorro.ToolBase.status:type_name -> zorro.ToolStatus
	5, // 2: zorro.ToolBase.inputs:type_name -> zorro.ToolBase.InputsEntry
	6, // 3: zorro.ToolBase.outputs:type_name -> zorro.ToolBase.OutputsEntry
	7, // 4: zorro.ToolBase.logs:type_name -> zorro.ToolBase.LogsEntry
	3, // 5: zorro.Tool.base:type_name -> zorro.ToolBase
	2, // 6: zorro.ToolBase.InputsEntry.value:type_name -> zorro.Socket
	2, // 7: zorro.ToolBase.OutputsEntry.value:type_name -> zorro.Socket
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_tools_tool_proto_init() }
func file_tools_tool_proto_init() {
	if File_tools_tool_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_tools_tool_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Socket); i {
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
		file_tools_tool_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ToolBase); i {
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
		file_tools_tool_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tool); i {
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
	file_tools_tool_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Socket_RawString)(nil),
		(*Socket_RawInteger)(nil),
		(*Socket_RawNumber)(nil),
		(*Socket_RawBinary)(nil),
	}
	file_tools_tool_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tools_tool_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_tools_tool_proto_goTypes,
		DependencyIndexes: file_tools_tool_proto_depIdxs,
		EnumInfos:         file_tools_tool_proto_enumTypes,
		MessageInfos:      file_tools_tool_proto_msgTypes,
	}.Build()
	File_tools_tool_proto = out.File
	file_tools_tool_proto_rawDesc = nil
	file_tools_tool_proto_goTypes = nil
	file_tools_tool_proto_depIdxs = nil
}
