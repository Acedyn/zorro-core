package reflection

import (
	"fmt"

	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Build the input protobuf from the socket socket
func BuildCommandMethodMessage(methodDescriptor protoreflect.MethodDescriptor, socket *tools_proto.Socket) (*dynamicpb.Message, error) {
	// First find the expected message descriptor
	inputMessage := dynamicpb.NewMessage(methodDescriptor.Input())

	// Assign the socket values to the message
	err := AssignMessageFromSocket(inputMessage, socket)
	if err != nil {
		return nil, fmt.Errorf("could not build input message for method %s: %w", methodDescriptor.FullName(), err)
	}

	return inputMessage, nil
}

func AssignMessageFromSocket(message *dynamicpb.Message, socket *tools_proto.Socket) error {
	// The socket can either specify the values one by one or contain the entire data on its own
	if len(socket.GetFields()) == 0 {
		socketRawValue, err := GetSocketRawValue(socket)
		if err != nil {
			return fmt.Errorf("invalid socket %s: %w", socket, err)
		}
		// Here the entire message in the raw field
		err = proto.Unmarshal(socketRawValue, message)
		if err != nil {
			return fmt.Errorf("invalid raw message from socket %s: %w", socket, err)
		}

		return nil
	}

	// Holder of any potential errors when settings an child value
	var err error = nil

	// Here the socket contain specific values for each entries
	message.Range(func(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		// Some fields might not be set and left with default value
		child, ok := socket.GetFields()[fieldDescriptor.TextName()]
		childRawValue, err := GetSocketRawValue(child)
		if !ok {
			return true
		}

		// If the field expects a message, make this function recursive for each fields of this message
		if messageDescriptor := fieldDescriptor.Message(); messageDescriptor != nil {
			childMessage := dynamicpb.NewMessage(messageDescriptor)
			err := AssignMessageFromSocket(childMessage, child)
			if err != nil {
				return false
			}
			message.Set(fieldDescriptor, protoreflect.ValueOfMessage(childMessage))
		} else {
			// For non message values, the socket cannot specify more precisely the values
			switch {
			case fieldDescriptor.IsList():
				_, err = unmarshalList(childRawValue, protowire.Type(child.GetWtyp()), message.Mutable(fieldDescriptor).List(), fieldDescriptor)
			case fieldDescriptor.IsMap():
				_, err = unmarshalMap(childRawValue, protowire.Type(child.GetWtyp()), message.Mutable(fieldDescriptor).Map(), fieldDescriptor)
			default:
				value, _, err := unmarshalScalar(childRawValue, protowire.Type(child.GetWtyp()), fieldDescriptor)
				if err == nil {
					message.Set(fieldDescriptor, value)
				}
			}
		}

		// Stop iterating if an error occured
		return err == nil
	})

	if err != nil {
		return fmt.Errorf("an error occured when assigning socket %s to message %s: %w", socket, message, err)
	}

	return nil
}

// Get the value of a socket after resolving the links if any
func GetSocketRawValue(socket *tools_proto.Socket) ([]byte, error) {
	rawValue := []byte{}
	return rawValue, nil
}
