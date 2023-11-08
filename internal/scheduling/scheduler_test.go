package scheduling

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/tools"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
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

	_, err = GetOrStartProcessor(resolvedContext, &pythonProcessorQuery)
	if err != nil {
		t.Errorf("An error occured while getting processor from query %s: %s", processorQuery, err.Error())
		return
	}
}

func WipTestCommandExecution(t *testing.T) {
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

	// Start the listener
	go ListenCommandQueries()

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

	processorQuery := pythonProcessorQuery
	processorQuery.Context = resolvedContext.Context

	command := tools.Command{
		Command: &tools_proto.Command{
			ProcessorQuery: processorQuery.ProcessorQuery,
		},
	}

	err = command.Execute(resolvedContext)
	if err != nil {
		t.Errorf("An error occured when executing the command %v: %v", command, err)
		return
	}
}
