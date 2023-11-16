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

func (socket *Socket) GetSocket() *tools_proto.Socket {
	if socket.Socket == nil {
		socket.Socket = &tools_proto.Socket{}
	}
	return socket.Socket
}

// Get the wrapped socket fields with all their methods
// This method is for accessing the fields, not for editing the map's structure
func (socket *Socket) GetFields() map[string]*Socket {
	if socket.GetSocket().GetFields() == nil {
		socket.GetSocket().Fields = map[string]*tools_proto.Socket{}
	}

	return maps.Map(socket.GetSocket().GetFields(), func(k string, v *tools_proto.Socket) (string, *Socket) {
		return k, &Socket{v}
	})
}

// Safe setter for the Field field
func (socket *Socket) SetField(key string, value *Socket) {
	if socket.GetSocket().GetFields() == nil {
		socket.GetSocket().Fields = map[string]*tools_proto.Socket{}
	}
	socket.GetSocket().GetFields()[key] = value.Socket
}

// Update the socket with a patch
func (socket *Socket) Update(patch *Socket) bool {
	// Patch the local version of the socket
	isPatched := false

	// Update the kind
	if patch.GetKind() != "" && socket.GetKind() != patch.GetKind() {
		socket.Kind = patch.GetKind()
		isPatched = true
	}

	// Update the value
	if patch.Value != nil && socket.GetValue() != patch.GetValue() {
		socket.Value = patch.GetValue()
		isPatched = true
	}

	// Recursively update the socket
	for fieldName, fieldPatch := range patch.GetFields() {
		fieldSocket, ok := socket.GetFields()[fieldName]
		if ok && fieldSocket.Update(fieldPatch) {
			isPatched = true
		} else if !ok {
			socket.SetField(fieldName, fieldPatch)
		}
	}

	return isPatched
}

// Update the socket's raw data with a proto value
func (socket *Socket) UpdateWithMessage(message protoreflect.Message) error {
	messageDescriptor := message.Descriptor()
	socket.Kind = string(messageDescriptor.FullName())

	marshalOptions := protojson.MarshalOptions{EmitUnpopulated: true}
	rawMessage, err := marshalOptions.Marshal(message.Interface())
	if err != nil {
		return fmt.Errorf("could not store raw message %s: %w", message, err)
	}

	// Decompose the value into fields
	rawFields := map[string]json.RawMessage{}
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

		socket.SetField(fieldDescriptor.JSONName(), &fieldSocket)
	}

	if err != nil {
		return fmt.Errorf("an error occured when updating a field of socket %s: %w", socket, err)
	}
	return nil
}

// Apply the socket's values to a message
func (socket *Socket) ApplyFieldsToMessage(message protoreflect.Message) error {
	messageDescriptor := message.Descriptor()
	// Gather the values to apply there is two type of values, the one that are applied directly
	// (jsonPatch) and the ones that will recursively create message fields (nestedSocketPatches)
	jsonPatch := map[string]json.RawMessage{}
	nestedSocketPatch := map[*Socket]protoreflect.FieldDescriptor{}

	// Gather all the values for each fields
	for fieldIndex := 0; fieldIndex < messageDescriptor.Fields().Len(); fieldIndex += 1 {
		fieldDescriptor := messageDescriptor.Fields().Get(fieldIndex)
		socketField, ok := socket.GetFields()[fieldDescriptor.JSONName()]
		if !ok {
			continue
		}

		// We must collect the value to apply first and then apply them because if we apply everything
		// progressively it will override the previously applied values
		if fieldDescriptor.Message() != nil && !fieldDescriptor.IsMap() && !fieldDescriptor.IsList() {
			nestedSocketPatch[socketField] = fieldDescriptor
		} else {
			socketRawValue, err := socketField.ResolveRawValue()
			if err != nil {
				return fmt.Errorf("could not resolve value of socket %s: %w", socket, err)
			}
			jsonPatch[fieldDescriptor.JSONName()] = socketRawValue
		}
	}

	// We apply the values patch first otherwise we can override the nested fields
	encodedJsonPatch, err := json.Marshal(jsonPatch)
	if err != nil {
		return fmt.Errorf("invalid raw message %s: %w", jsonPatch, err)
	}
	err = protojson.Unmarshal(encodedJsonPatch, message.Interface())
	fmt.Println(string(encodedJsonPatch))
	if err != nil {
		return fmt.Errorf("an error occured while applying json patch %s on message %s: %w", encodedJsonPatch, message, err)
	}

	// Apply the nested fields at the end since this won't affect the other fields
	for socketField, messagePatch := range nestedSocketPatch {
		socketField.ApplyFieldsToMessage(message.Mutable(messagePatch).Message())
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
