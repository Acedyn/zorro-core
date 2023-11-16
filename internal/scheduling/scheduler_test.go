package scheduling

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	zorro_context "github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/tools"
	"github.com/life4/genesis/maps"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func mockedSocketValueDescriptor(name string) (protoreflect.MessageDescriptor, error) {
	cwdPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get the current working directory: %w", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))
	fileName := "log.proto"
	rootPath := filepath.Join(cwdPath, "testdata", "plugins", "python", "python@3.10", "zorro_python", "commands", "log")
	importPath := filepath.Join(cwdPath, "testdata", "plugins", "python", "python@3.10", "protos")

	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			ImportPaths: []string{rootPath, importPath},
		},
	}
	files, err := compiler.Compile(context.Background(), fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", fileName, err)
	}
	if len(files) != 1 {
		return nil, fmt.Errorf("%d files parsed instead of one", len(files))
	}

	fileDescriptor := files[0]
	return fileDescriptor.Messages().ByName(protoreflect.Name(name)), nil
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

var pythonProcessorQuery = ProcessorQuery{
	ProcessorQuery: &scheduling_proto.ProcessorQuery{
		Name: &[]string{"python"}[0],
	},
}

func TestProcessorRegistration(t *testing.T) {
	host := "127.0.0.1"
	port, err := getFreePort()

	if err != nil {
		t.Errorf("Could not get free port: %s", err.Error())
	}

	// Start the server in its own goroutine
	go func() {
		_, grpcStatus := network.GrpcServer()
		if grpcStatus.IsRunning {
			return
		}

		if err := network.ServeGrpc(host, port); err != nil {
			t.Errorf("An error occured while serving GRPC: %s", err.Error())
		}
	}()

	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))
	fullPath := filepath.Join(cwdPath, "testdata", "plugins")

	resolvedContext, err := zorro_context.NewContext([]string{"python"}, &config_proto.Config{PluginConfig: &config_proto.PluginConfig{
		Repos: []string{fullPath},
	}})

	if err != nil {
		t.Errorf("Could not create context\n\t%s", err)
		return
	}

	_, err = GetOrStartProcessor(resolvedContext, &pythonProcessorQuery)
	if err != nil {
		t.Errorf("An error occured while getting processor from query %s: %s", processorQuery, err.Error())
		return
	}
}

func TestCommandExecution(t *testing.T) {
	logMethodDescriptor, err := mockedSocketValueDescriptor("LogInput")
	if err != nil || logMethodDescriptor == nil {
		t.Errorf("Could not get the log message descriptor: %v", err)
		return
	}

	host := "127.0.0.1"
	port, err := getFreePort()

	if err != nil {
		t.Errorf("Could not get free port: %s", err.Error())
	}

	// Start the server in its own goroutine
	go func() {
		_, grpcStatus := network.GrpcServer()
		if grpcStatus.IsRunning {
			return
		}
		if err := network.ServeGrpc(host, port); err != nil {
			t.Errorf("An error occured while serving GRPC: %s", err.Error())
		}
	}()

	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))
	fullPath := filepath.Join(cwdPath, "testdata", "plugins")

	resolvedContext, err := zorro_context.NewContext([]string{"python"}, &config_proto.Config{PluginConfig: &config_proto.PluginConfig{
		Repos: []string{fullPath},
	}})

	if err != nil {
		t.Errorf("Could not create context\n\t%s", err)
		return
	}

	processorQuery := pythonProcessorQuery
	processorQuery.Context = resolvedContext.Context

	commandMessage := "Hello Zorro"
	commandRawMessage, err := json.Marshal(commandMessage)
	if err != nil {
		t.Errorf("an error occured while marshalling the command message: %v", err)
		return
	}

	command := tools.Command{
		Command: &tools_proto.Command{
			Base: &tools_proto.ToolBase{
				Name: &[]string{"zorro_python.Log"}[0],
				Input: &tools_proto.Socket{
					Fields: map[string]*tools_proto.Socket{
						"message": {
							Value: &tools_proto.Socket_Raw{
								Raw: commandRawMessage,
							},
						},
					},
				},
			},
			ProcessorQuery: processorQuery.ProcessorQuery,
		},
	}

	err = command.Execute(resolvedContext, nil)
	if err != nil {
		t.Errorf("An error occured when executing the command %v: %v", command, err)
		return
	}

	expectedLog := fmt.Sprintf("DEBUG: %s", commandMessage)
	if !maps.HasValue(command.GetBase().GetLogs(), expectedLog) {
		t.Errorf("Expected to find \"%s\" in log message after log command: found %v", expectedLog, command.GetBase().GetLogs())
		return
	}
}
