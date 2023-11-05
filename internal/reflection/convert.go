// This is a hack to access proto package's private methods.
// They are nesessary for marshalling and unmarshalling fields individually
// Feel free to propose any cleaner solution

package reflection

import (
  _ "unsafe"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/encoding/protowire"
)

//go:linkname MarshalField google.golang.org/protobuf/proto.(*MarshalOptions).marshalField
func MarshalField(*proto.MarshalOptions, []byte, protoreflect.FieldDescriptor, protoreflect.Value) ([]byte, error)

//go:linkname UnmarshalList google.golang.org/protobuf/proto.(*UnmarshalOptions).unmarshalList
func UnmarshalList(*proto.UnmarshalOptions, []byte, protowire.Type, protoreflect.List, protoreflect.FieldDescriptor) (int, error)
//go:linkname UnmarshalMap google.golang.org/protobuf/proto.(*UnmarshalOptions).unmarshalMap
func UnmarshalMap(*proto.UnmarshalOptions, []byte, protowire.Type, protoreflect.Map, protoreflect.FieldDescriptor) (int, error)
//go:linkname UnmarshalSingular google.golang.org/protobuf/proto.(*UnmarshalOptions).unmarshalSingular
func UnmarshalSingular(*proto.UnmarshalOptions, []byte, protowire.Type, protoreflect.Message, protoreflect.FieldDescriptor) (int, error)

//go:linkname WireTypes google.golang.org/protobuf/proto.wireTypes
var WireTypes map[protoreflect.Kind]protowire.Type
