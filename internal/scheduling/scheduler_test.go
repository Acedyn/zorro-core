package scheduling

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/network"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
)

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
		if err := network.ServeGrpc(host, port); err != nil {
			t.Errorf("An error occured while serving GRPC: %s", err.Error())
		}
	}()
	grpcServer, _ := network.GrpcServer()
	defer grpcServer.GracefulStop()

	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))
	fullPath := filepath.Join(cwdPath, "testdata", "mocked_plugins")

	resolvedContext, err := context.NewContext([]string{"python"}, &config_proto.Config{PluginConfig: &config_proto.PluginConfig{
		Repos: []string{fullPath},
	}})

	if err != nil {
		t.Errorf("Could not create context\n\t%s", err)
		return
	}

	pythonProcessorQuery.Context = resolvedContext.Context

	registeredProcessor, err := GetOrStartProcessor(&pythonProcessorQuery)
	if err != nil {
		t.Errorf("An error occured while getting processor from query %s: %s", pythonProcessorQuery, err.Error())
		return
	}

	output, err := registeredProcessor.Client.InvokeRpcServerStream("zorro_python.Log", "Execute")
	if err != nil {
		t.Errorf("An error occured while invoking rpc method %s: %s", "/zorro_python.Log/Execute", err.Error())
		return
	}

	if err != nil {
		t.Errorf("An error occured while invoking rpc method %s: %s", "/zorro_python.Log/Execute", err.Error())
		return
	}

	messageValue, ok := output["message"]
	if !ok {
		t.Errorf("Expected a value in the field 'message': No value found in the returned output")
		return
	}
	if messageValue.String() != "DEBUG: " {
		t.Errorf("Expected value %s in the returned output: Revieved %s", "DEBUG :", messageValue.String())
		return
	}
}
