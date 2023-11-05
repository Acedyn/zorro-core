package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func mockedSocketValueDescriptor() (protoreflect.MessageDescriptor, error) {
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
  return fileDescriptor.Messages().ByName("TestSocket"), nil
}

func initDefaultMessage(message *dynamicpb.Message) {
  messageDescriptor := message.Descriptor()

  for fieldIndex := 0; fieldIndex < messageDescriptor.Fields().Len(); fieldIndex += 1 {
    fieldDescriptor := messageDescriptor.Fields().Get(fieldIndex)
    message.Set(fieldDescriptor, message.NewField(fieldDescriptor))
  }
}

func TestSocketUpdate(t *testing.T) {
  socketValueDescriptor, err := mockedSocketValueDescriptor()
  if err != nil || socketValueDescriptor == nil {
    t.Errorf("Could not get the mocked socket value descriptor: %v", err)
    return
  }

  socket := Socket{&tools_proto.Socket{}}
  message := dynamicpb.NewMessage(socketValueDescriptor)
  initDefaultMessage(message)

  err = socket.UpdateWithMessage(message)
  if err != nil {
    t.Errorf("An error occured whild updating a socket with the message %s: %v", message, err)
    return
  }
  t.Errorf("socket %+v", socket)
}


func TestSocketApplyValue(t *testing.T) {
}
