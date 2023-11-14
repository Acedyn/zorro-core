package tools

import (
	"encoding/json"
	"fmt"

	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/life4/genesis/maps"
	"google.golang.org/protobuf/encoding/protojson"
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
	messageDescriptor := message.Descriptor()
	socket.Kind = string(messageDescriptor.FullName())

  // First make sure all the fields are set
	for fieldIndex := 0; fieldIndex < messageDescriptor.Fields().Len(); fieldIndex += 1 {
		fieldDescriptor := messageDescriptor.Fields().Get(fieldIndex)
		value := message.Get(fieldDescriptor)
    if fieldDescriptor.IsList() {
      value = protoreflect.ValueOfList(message.Mutable(fieldDescriptor).List())
    } else if fieldDescriptor.IsMap() {
      value = protoreflect.ValueOfMap(message.Mutable(fieldDescriptor).Map())
    } else if fieldDescriptor.Message() != nil {
		  value = message.Mutable(fieldDescriptor)
    }
    message.Set(fieldDescriptor, value)
  }

  rawMessage, err := protojson.Marshal(message.Interface())
	if err != nil {
		return fmt.Errorf("could not store raw message %s: %w", message, err)
	}

  // Decompose the value into fields
  rawFields := map[string][]byte{}
  err = json.Unmarshal(rawMessage, &rawFields)
  if err != nil {
    return fmt.Errorf("could not decompose message fields: %w", err)
  }

  // For each field of the message store the value in a new socket
	for fieldIndex := 0; fieldIndex < messageDescriptor.Fields().Len(); fieldIndex += 1 {
		fieldDescriptor := messageDescriptor.Fields().Get(fieldIndex)

		// Here we set the value or the default value if the value is not there
		fieldSocket := Socket{&tools_proto.Socket{}}
    if fieldDescriptor.Message() != nil && !fieldDescriptor.IsMap() && !fieldDescriptor.IsList() {
      // For nested messages make this method recursive
		  err = fieldSocket.UpdateWithMessage(message.Mutable(fieldDescriptor).Message())
      if err != nil {
        return fmt.Errorf("could not update socket field with message: %w", err)
      }
    } else {
      // Store the raw value individualy for value fields
      rawField, ok := rawFields[fieldDescriptor.JSONName()]
      if !ok {
        continue
      }

      fieldSocket.Update(&Socket{
        &tools_proto.Socket{
          Kind: formatFieldDescriptorKind(fieldDescriptor),
          Value: &tools_proto.Socket_Raw{
            Raw: rawField,
          },
        },
      })
    }

		socket.SetField(fieldDescriptor.TextName(), &fieldSocket)
	}

	if err != nil {
		return fmt.Errorf("an error occured when updating a field of socket %s: %w", socket, err)
	}
	return nil
}

// Apply the socket's values to a message
func (socket *Socket) ApplyFieldsToMessage(message protoreflect.Message) error {

	messageDescriptor := message.Descriptor()
	for fieldIndex := 0; fieldIndex < messageDescriptor.Fields().Len(); fieldIndex += 1 {
		fieldDescriptor := messageDescriptor.Fields().Get(fieldIndex)
    socketField, ok := socket.GetFields()[fieldDescriptor.JSONName()]
    if !ok {
      continue
    }

    if fieldDescriptor.Message() != nil && !fieldDescriptor.IsMap() && !fieldDescriptor.IsList() {
      socketField.ApplyFieldsToMessage(message.Mutable(fieldDescriptor).Message())
    } else {
		  socketRawValue, err := socketField.ResolveRawValue()
      if err != nil {
        return fmt.Errorf("could not resolve value of socket %s: %w", socket, err)
      }

      jsonPatch := map[string][]byte{fieldDescriptor.JSONName(): socketRawValue}
      encodedJsonPatch, err := json.Marshal(jsonPatch)
      if err != nil {
        return fmt.Errorf("invalid raw message %s from socket field %s: %w", jsonPatch, socketField, err)
      }

      err = protojson.Unmarshal(encodedJsonPatch, message.Interface())
      if err != nil {
        return fmt.Errorf("an error occured while applying json patch %s on message %s: %w", jsonPatch, message, err)
      }
    }
  }

	return nil
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
