package tools

import (
	"encoding/json"
	"testing"

	test_proto "github.com/Acedyn/zorro-core/testdata/protos"
	tools_proto "github.com/Acedyn/zorro-proto/gen/protos/tools"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestSocketUpdate(t *testing.T) {
	socket := Socket{&tools_proto.Socket{}}
	testSocketValue := test_proto.TestSocket{}
	socketValueDescriptor := testSocketValue.ProtoReflect().Descriptor()
	message := dynamicpb.NewMessage(socketValueDescriptor)

	err := socket.UpdateValueWithMessage(message)
	if err != nil {
		t.Errorf("An error occured while updating a socket with the message %s: %v", message, err)
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

func TestSocketApplyFields(t *testing.T) {
	testSocketValue := test_proto.TestSocket{}
	socketValueDescriptor := testSocketValue.ProtoReflect().Descriptor()

	childTestSocketValue := test_proto.MessageField{}
	childSocketValueDescriptor := childTestSocketValue.ProtoReflect().Descriptor()

	socket := Socket{&tools_proto.Socket{}}
	message := dynamicpb.NewMessage(socketValueDescriptor)

	// Set the bar field
	barFieldDescriptor := socketValueDescriptor.Fields().ByName("bar")
	barInitialValue := int32(42)
	barRawValue, err := json.Marshal(barInitialValue)
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
	quuxRawValue, err := json.Marshal(quuxInitialValue)
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
	if err != nil {
		t.Errorf("Could not marshall value for field quux: %v", err)
		return
	}

	socket.SetField("qux", &Socket{
		&tools_proto.Socket{
			Value: &tools_proto.Socket_Raw{
				Raw: []byte("{\"toto\": {\"corge\": [true]}, \"tata\": {\"corge\": [true, false, true]}}"),
			},
		},
	})

	message = dynamicpb.NewMessage(socketValueDescriptor)
	err = socket.ApplyFieldsToMessage(message, nil)
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
	if quxTotoCorgeNewValue.List().Get(0).Bool() != true {
		t.Errorf("Invalid value found in the field corge at index 0: expected %v, recieved %v", quxTotoCorgeNewValue.List().Get(0).Bool(), true)
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
	if quxTataCorgeNewValue.List().Get(0).Bool() != true {
		t.Errorf("Invalid value found in the field corge at index 0: expected %v, recieved %v", quxTataCorgeNewValue.List().Get(0).Bool(), true)
		return
	}
	if quxTataCorgeNewValue.List().Get(1).Bool() != false {
		t.Errorf("Invalid value found in the field corge at index 1: expected %v, recieved %v", quxTataCorgeNewValue.List().Get(0).Bool(), false)
		return
	}
	if quxTataCorgeNewValue.List().Get(2).Bool() != true {
		t.Errorf("Invalid value found in the field corge at index 2: expected %v, recieved %v", quxTataCorgeNewValue.List().Get(0).Bool(), true)
		return
	}
}
