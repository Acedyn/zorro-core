package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Acedyn/zorro-core/internal/reflection"

	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func mockedSocketValueDescriptor(name string) (protoreflect.MessageDescriptor, error) {
	cwdPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get the current working directory: %w", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))
	fullPath := filepath.Join(cwdPath, "testdata", "mocked_protos", "socket_value.proto")

	reader, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", fullPath, err)
	}
	defer reader.Close()

	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{},
	}
	files, err := compiler.Compile(context.Background(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", fullPath, err)
	}
	if len(files) != 1 {
		return nil, fmt.Errorf("%d files parsed instead of one", len(files))
	}

	fileDescriptor := files[0]
	return fileDescriptor.Messages().ByName(protoreflect.Name(name)), nil
}

func TestSocketUpdate(t *testing.T) {
	socketValueDescriptor, err := mockedSocketValueDescriptor("TestSocket")
	if err != nil || socketValueDescriptor == nil {
		t.Errorf("Could not get the mocked socket value descriptor: %v", err)
		return
	}

	socket := Socket{&tools_proto.Socket{}}
	message := dynamicpb.NewMessage(socketValueDescriptor)

	err = socket.UpdateWithMessage(message)
	if err != nil {
		t.Errorf("An error occured whild updating a socket with the message %s: %v", message, err)
		return
	}

	// Test the foo field
	fooSocket, fooOk := socket.GetFields()["foo"]
	if !fooOk {
		t.Errorf("Foo field expected in the socket %s", socket)
		return
	}
	if fooSocket.Kind != "zorro_testing.MessageField" {
		t.Errorf("Foo of type %s instead of expected type zorro_testing.MessageField", socket.Kind)
		return
	}

	// Test the gault field inside the foo field
	graultSocket, graultOk := fooSocket.GetFields()["grault"]
	if !graultOk {
		t.Errorf("grault field expected in the socket %s", fooSocket)
		return
	}
	if graultSocket.Kind != "float" {
		t.Errorf("grault of type %s instead of expected type float", fooSocket.Kind)
		return
	}

	// Test the baz field
	bazSocket, bazOk := socket.GetFields()["baz"]
	if !bazOk {
		t.Errorf("baz field expected in the socket %s", socket)
		return
	}
	if bazSocket.Kind != "[]string" {
		t.Errorf("baz of type %s instead of expected type []string", socket.Kind)
		return
	}
}

func TestSocketApplyValue(t *testing.T) {
	socketValueDescriptor, err := mockedSocketValueDescriptor("TestSocket")
	if err != nil || socketValueDescriptor == nil {
		t.Errorf("Could not get the mocked socket value descriptor: %v", err)
		return
	}
	childSocketValueDescriptor, err := mockedSocketValueDescriptor("MessageField")
	if err != nil || childSocketValueDescriptor == nil {
		t.Errorf("Could not get the mocked child socket value descriptor: %v", err)
		return
	}

	socket := Socket{&tools_proto.Socket{}}
	message := dynamicpb.NewMessage(socketValueDescriptor)

	// Set the bar field
	barFieldDescriptor := socketValueDescriptor.Fields().ByName("bar")
	barInitialValue := int32(42)
	barRawValue := []byte{}
	barRawValue, err = reflection.MarshalField(&proto.MarshalOptions{}, barRawValue, barFieldDescriptor, protoreflect.ValueOfInt32(barInitialValue))
	if err != nil {
		t.Errorf("Could not marshall value for field bar: %v", err)
		return
	}

	socket.SetField("bar", &Socket{
		&tools_proto.Socket{
			Value: &tools_proto.Socket_Raw{
				Raw: barRawValue,
			},
		},
	})

	// Set the quux field of the foo field
	fooFieldDescriptor := socketValueDescriptor.Fields().ByName("foo")
	quuxFieldDescriptor := childSocketValueDescriptor.Fields().ByName("quux")
	quuxInitialValue := "hello world"
	quuxRawValue := []byte{}
	quuxRawValue, err = reflection.MarshalField(&proto.MarshalOptions{}, quuxRawValue, quuxFieldDescriptor, protoreflect.ValueOfString(quuxInitialValue))
	if err != nil {
		t.Errorf("Could not marshall value for field quux: %v", err)
		return
	}

	fooSocket := Socket{&tools_proto.Socket{}}
	socket.SetField("foo", &fooSocket)
	fooSocket.SetField("quux", &Socket{
		&tools_proto.Socket{
			Value: &tools_proto.Socket_Raw{
				Raw: quuxRawValue,
			},
		},
	})

	// Set the qux field with the corge fields
	quxFieldDescriptor := message.Descriptor().Fields().ByName("qux")
	quxInitialValue := message.Mutable(quxFieldDescriptor).Map()

	// Create a toto entry
	quxTotoCorgeFirstValue := protoreflect.ValueOfBool(true)
	quxTotoMessage := dynamicpb.NewMessage(childSocketValueDescriptor)
	quxTotoCorgeDescriptor := quxTotoMessage.Descriptor().Fields().ByName("corge")
	quxTotoCorgeValue := quxTotoMessage.Mutable(quxTotoCorgeDescriptor).List()
	quxTotoCorgeValue.Append(quxTotoCorgeFirstValue)
	quxInitialValue.Set(protoreflect.MapKey(protoreflect.ValueOfString("toto")), protoreflect.ValueOfMessage(quxTotoMessage))

	// Create a tata entry
	quxTataCorgeFirstValue := protoreflect.ValueOfBool(true)
	quxTataCorgeSecondValue := protoreflect.ValueOfBool(false)
	quxTataCorgeThirdValue := protoreflect.ValueOfBool(true)
	quxTataMessage := dynamicpb.NewMessage(childSocketValueDescriptor)
	quxTataCorgeDescriptor := quxTataMessage.Descriptor().Fields().ByName("corge")
	quxTataCorgeValue := quxTataMessage.Mutable(quxTataCorgeDescriptor).List()
	quxTataCorgeValue.Append(quxTataCorgeFirstValue)
	quxTataCorgeValue.Append(quxTataCorgeSecondValue)
	quxTataCorgeValue.Append(quxTataCorgeThirdValue)
	quxInitialValue.Set(protoreflect.MapKey(protoreflect.ValueOfString("tata")), protoreflect.ValueOfMessage(quxTataMessage))

	quxRawValue := []byte{}
	quxRawValue, err = reflection.MarshalField(&proto.MarshalOptions{}, quxRawValue, quxFieldDescriptor, protoreflect.ValueOfMap(quxInitialValue))
	if err != nil {
		t.Errorf("Could not marshall value for field quux: %v", err)
		return
	}

	socket.SetField("qux", &Socket{
		&tools_proto.Socket{
			Value: &tools_proto.Socket_Raw{
				Raw: quxRawValue,
			},
		},
	})

	err = socket.ApplyValueToMessage(message)
	if err != nil {
		t.Errorf("An error occured while applying the socket values to a message: %v", err)
		return
	}

	// Check the bar field
	barNewValue := message.Get(barFieldDescriptor)
	if barNewValue.Int() != int64(barInitialValue) {
		t.Errorf("Invalid value applied to the message: reveived %d, expected %d", barNewValue.Int(), barInitialValue)
		return
	}

	// Check the quux field
	fooNewValue := message.Get(fooFieldDescriptor)
	quuxFieldDescriptor = fooNewValue.Message().Descriptor().Fields().ByName("quux")
	quuxNewValue := fooNewValue.Message().Get(quuxFieldDescriptor)
	if quuxNewValue.String() != quuxInitialValue {
		t.Errorf("Invalid value applied to the message: reveived %s, expected %s", quuxNewValue.String(), quuxInitialValue)
		return
	}

	// Check the qux field with the corge fields
	quxNewValue := message.Get(quxFieldDescriptor)

	quxTotoNewValue := quxNewValue.Map().Get(protoreflect.MapKey(protoreflect.ValueOfString("toto")))
	quxTotoNewMessage := quxTotoNewValue.Message()
	quxTotoCorgeNewDescriptor := quxTotoNewMessage.Descriptor().Fields().ByName("corge")
	quxTotoCorgeNewValue := quxTotoNewMessage.Get(quxTotoCorgeNewDescriptor)
	if quxTotoCorgeNewValue.List().Len() != 1 {
		t.Errorf("Invalid value count found in the field corge in the map value of toto: expected 1, recieved %d", quxTotoCorgeNewValue.List().Len())
		return
	}
	if quxTotoCorgeNewValue.List().Get(0).Bool() != quxTotoCorgeFirstValue.Bool() {
		t.Errorf("Invalid value found in the field corge at index 0: expected %v, recieved %v", quxTotoCorgeNewValue.List().Get(0).Bool(), quxTotoCorgeFirstValue.Bool())
		return
	}

	quxTataNewValue := quxNewValue.Map().Get(protoreflect.MapKey(protoreflect.ValueOfString("tata")))
	quxTataNewMessage := quxTataNewValue.Message()
	quxTataCorgeNewDescriptor := quxTataNewMessage.Descriptor().Fields().ByName("corge")
	quxTataCorgeNewValue := quxTataNewMessage.Get(quxTataCorgeNewDescriptor)
	if quxTataCorgeNewValue.List().Len() != 3 {
		t.Errorf("Invalid value count found in the field corge in the map value of tata: expected 3, recieved %d", quxTataCorgeNewValue.List().Len())
		return
	}
	if quxTataCorgeNewValue.List().Get(0).Bool() != quxTataCorgeFirstValue.Bool() {
		t.Errorf("Invalid value found in the field corge at index 0: expected %v, recieved %v", quxTataCorgeNewValue.List().Get(0).Bool(), quxTataCorgeFirstValue.Bool())
		return
	}
	if quxTataCorgeNewValue.List().Get(1).Bool() != quxTataCorgeSecondValue.Bool() {
		t.Errorf("Invalid value found in the field corge at index 1: expected %v, recieved %v", quxTataCorgeNewValue.List().Get(0).Bool(), quxTataCorgeSecondValue.Bool())
		return
	}
	if quxTataCorgeNewValue.List().Get(2).Bool() != quxTataCorgeThirdValue.Bool() {
		t.Errorf("Invalid value found in the field corge at index 2: expected %v, recieved %v", quxTataCorgeNewValue.List().Get(0).Bool(), quxTataCorgeThirdValue.Bool())
		return
	}
}
