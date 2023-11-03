package reflection

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	protobufRegistryLock = &sync.Mutex{}
	protobufRegistry     map[string]map[string]*protoregistry.Files
	once                 sync.Once
)

// Getter for the clients pool singleton
func ProtobufRegistry() map[string]map[string]*protoregistry.Files {
	once.Do(func() {
		protobufRegistry = map[string]map[string]*protoregistry.Files{}
	})

	return protobufRegistry
}

// Recursive function to gather all the descriptors of a message type
func GatherEmbededMessageDescriptors(messageDescriptors map[string]*descriptorpb.DescriptorProto, messageDescriptor *descriptorpb.DescriptorProto) {
	messageDescriptors[messageDescriptor.GetName()] = messageDescriptor
	for _, embedMessage := range messageDescriptor.GetNestedType() {
		GatherEmbededMessageDescriptors(messageDescriptors, embedMessage)
	}
}

// Build a string representing a field's kind
func FormatFieldDescriptorKind(fieldDescriptor protoreflect.FieldDescriptor) string {
	kind := fieldDescriptor.Kind().String()
	if fieldDescriptor.Kind() == protoreflect.MessageKind {
		kind = string(fieldDescriptor.Message().FullName())
	} else if fieldDescriptor.IsList() {
		kind = fmt.Sprintf("[]%s", kind)
	} else if fieldDescriptor.IsMap() {
		kind = fmt.Sprintf("map[%s]%s", FormatFieldDescriptorKind(fieldDescriptor.MapKey()), FormatFieldDescriptorKind(fieldDescriptor.MapValue()))
	}

	return kind
}
