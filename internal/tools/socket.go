package tools

import (
	"fmt"

	"github.com/Acedyn/zorro-core/internal/reflection"

	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/life4/genesis/maps"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Wrapped socket with methods attached
type Socket struct {
	*tools_proto.Socket
}

// Get the wrapped socket fields with all their methods
func (socket *Socket) GetFields() map[string]*Socket {
	return maps.Map(socket.Socket.GetFields(), func(k string, v *tools_proto.Socket) (string, *Socket) {
		return k, &Socket{v}
	})
}

// Safe setter for the Field field
func (socket *Socket) SetField(key string, value *Socket) {
	if socket.Socket.GetFields() == nil {
		socket.Socket.Fields = map[string]*tools_proto.Socket{}
	}
	socket.Socket.GetFields()[key] = value.Socket
}

// Update the command with a patch
func (socket *Socket) Update(patch *Socket) bool {
	// Patch the local version of the socket
	isPatched := false

	// Recursively update the socket
	for fieldName, fieldSocket := range socket.GetFields() {
		fieldPatch, ok := patch.GetFields()[fieldName]
		if ok && fieldSocket.Update(fieldPatch) {
			isPatched = true
		}
	}

	return isPatched
}

// Update the socket's raw data with a proto value
func (socket *Socket) UpdateWithMessage(message protoreflect.Message) error {
	rawMessage, err := proto.Marshal(message.Interface())
	if err != nil {
		return fmt.Errorf("could not store raw message %s: %w", message, err)
	}

	// Set the raw value first
	socket.Kind = string(message.Descriptor().FullName())
	socket.Value = &tools_proto.Socket_Raw{
		Raw: rawMessage,
	}

	// Decompose the value into fields so they can be fetched separatly
	messageDescriptor := message.Descriptor()
	for fieldIndex := 0; fieldIndex < messageDescriptor.Fields().Len(); fieldIndex += 1 {
		fieldDescriptor := messageDescriptor.Fields().Get(fieldIndex)
		value := message.Get(fieldDescriptor)

		// Here we set the value or the default value if the value is not there
		fieldSocket := Socket{&tools_proto.Socket{}}
		err = fieldSocket.UpdateWithField(fieldDescriptor, value)
		if err != nil {
			break
		}

		socket.SetField(fieldDescriptor.TextName(), &fieldSocket)
	}

	if err != nil {
		return fmt.Errorf("an error occured when updating a field of socket %s: %w", socket, err)
	}
	return nil
}

// Update the socket's raw data with a message field
func (socket *Socket) UpdateWithField(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value) error {
	// If the value is a message then this function is recursive
	// Maps a considered like messages but we want to treat only messages recursively
	if fieldDescriptor.Message() != nil && !fieldDescriptor.IsMap() {
		return socket.UpdateWithMessage(value.Message())
	}

	rawValue := []byte{}
	rawValue, err := reflection.MarshalField(&proto.MarshalOptions{}, rawValue, fieldDescriptor, value)
	if err != nil {
		return fmt.Errorf("an error occured while marshalling the socket %s: %w", socket, err)
	}

	socket.Kind = formatFieldDescriptorKind(fieldDescriptor)
	socket.Value = &tools_proto.Socket_Raw{
		Raw: rawValue,
	}
	return nil
}

// Apply the socket's values to a message
func (socket *Socket) ApplyValueToMessage(message protoreflect.Message) error {
	// The raw value is applied only if the socket has no fields, otherwise the fields
	// take the priority and their value are applied instead
	if len(socket.GetFields()) == 0 {
		// Try to apply the raw value to the message
		socketRawValue, err := socket.ResolveRawValue()
		if err != nil {
			return fmt.Errorf("could not resolve value of socket %s: %w", socket, err)
		}

		err = proto.Unmarshal(socketRawValue, message.Interface())
		if err != nil {
			return fmt.Errorf("invalid raw message from socket %s: %w", socket, err)
		}
	} else {
		// The socket has fields defined so we apply them instead of applying the raw value
		var err error = nil
		// Override the message's fields with the fields entries
		messageDescriptor := message.Descriptor()
		for fieldIndex := 0; fieldIndex < messageDescriptor.Fields().Len(); fieldIndex += 1 {
			fieldDescriptor := messageDescriptor.Fields().Get(fieldIndex)
			value := message.Get(fieldDescriptor)

			// Some fields might not be set and left with default value
			fieldSocket, ok := socket.GetFields()[fieldDescriptor.TextName()]
			if !ok {
				continue
			}

			// Set the value of that field individually
			err = fieldSocket.ApplyValueToField(fieldDescriptor, value, message)
			if err != nil {
				break
			}
		}

		if err != nil {
			return fmt.Errorf("could not apply field value on socket %s: %w", socket, err)
		}
	}

	return nil
}

// Apply the socket's field value to a message field
func (socket *Socket) ApplyValueToField(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value, message protoreflect.Message) error {
	// Resolve the raw value to apply
	rawValue, err := socket.ResolveRawValue()
	if err != nil {
		return fmt.Errorf("could not resolve value of socket %s: %w", socket, err)
	}

	unmarshalOptions := &proto.UnmarshalOptions{RecursionLimit: protowire.DefaultRecursionLimit}
	_, wtyp, tagLen := protowire.ConsumeTag(rawValue)

	// Apply the value according to the expected data type
	switch {
	case fieldDescriptor.IsMap():
		_, err := reflection.UnmarshalMap(unmarshalOptions, rawValue[tagLen:], wtyp, message.Mutable(fieldDescriptor).Map(), fieldDescriptor)
		if err != nil {
			return fmt.Errorf("error occured while unmarshalling map field %s: %w", fieldDescriptor.FullName(), err)
		}
	case fieldDescriptor.IsList():
		_, err := reflection.UnmarshalList(unmarshalOptions, rawValue[tagLen:], wtyp, message.Mutable(fieldDescriptor).List(), fieldDescriptor)
		if err != nil {
			return fmt.Errorf("error occured while unmarshalling list field %s: %w", fieldDescriptor.FullName(), err)
		}
	case fieldDescriptor.Message() != nil:
		err = socket.ApplyValueToMessage(message.Mutable(fieldDescriptor).Message())
		if err != nil {
			return fmt.Errorf("error occured while unmarshalling message field %s: %w", fieldDescriptor.FullName(), err)
		}
	default:
		_, err := reflection.UnmarshalSingular(unmarshalOptions, rawValue[tagLen:], wtyp, message, fieldDescriptor)
		if err != nil {
			return fmt.Errorf("error occured while unmarshalling singular field %s: %w", fieldDescriptor.FullName(), err)
		}
	}

	return err
}

// Get the raw value after resolving the links
func (socket *Socket) ResolveRawValue() ([]byte, error) {
	switch socket.GetValue().(type) {
	case *tools_proto.Socket_Raw:
		return socket.GetRaw(), nil
	case *tools_proto.Socket_Link:
		return []byte{}, nil
	default:
		return []byte{}, nil
	}
}

// Build a string representing a field's kind
func formatFieldDescriptorKind(fieldDescriptor protoreflect.FieldDescriptor) string {
	kind := fieldDescriptor.Kind().String()
	if fieldDescriptor.IsMap() {
		kind = fmt.Sprintf("map[%s]%s", formatFieldDescriptorKind(fieldDescriptor.MapKey()), formatFieldDescriptorKind(fieldDescriptor.MapValue()))
	} else if fieldDescriptor.IsList() {
		kind = fmt.Sprintf("[]%s", kind)
	} else if fieldDescriptor.Kind() == protoreflect.MessageKind {
		kind = string(fieldDescriptor.Message().FullName())
	}

	return kind
}
