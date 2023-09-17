package scheduling

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	"github.com/life4/genesis/slices"

	"github.com/Acedyn/zorro-core/internal/context"
)

// Internal struct to keep track of the running clients
type ClientHandle struct {
	RunningClient *RunningClient
	Process       *os.Process
}

var (
  runningClientsLock = &sync.Mutex{}
	runningClients map[int]*ClientHandle
	once           sync.Once
)

// Getter for the clients singleton
func RunningClients() map[int]*ClientHandle {
	once.Do(func() {
		runningClients = map[int]*ClientHandle{}
	})

	return runningClients
}

// Test if running client matches the query's requirements
func (query *ClientQuery) Match(client *RunningClient) bool {
	// Test the name
	if query.Name != nil {
		// Some clients are supersets of other clients
		// If so they should match also their subsets
		subsets := append(client.GetClient().GetSubsets(), client.GetClient().GetName())
		if !slices.Contains(subsets, query.GetName()) {
			return false
		}
	}
	// Test the PID
	if query.Pid != nil {
		if query.GetPid() != client.GetPid() {
			return false
		}
	}
	// Test the Metadata
	for key, value := range query.GetMetadata() {
		metadata, ok := client.GetMetadata()[key]
		if !ok || metadata != value {
			return false
		}
	}
	return true
}

// Start the client into a running client
func RunClient(
	client *context.Client,
	context *context.Context,
	metadata map[string]string,
) (*ClientHandle, error) {
	runningClient := &RunningClient{
		Client:   client,
		Status:   ClientStatus_STARTING,
		Metadata: metadata,
	}

	// Build the command template
	template, err := template.New(client.GetName()).Parse(client.GetRunClientTemplate())
	if err != nil {
		return nil, fmt.Errorf("could not run client %s: Invalid launch template %w", client.GetName(), err)
	}

	// Apply the metadata and the name on the template
	runCommand := &bytes.Buffer{}
	err = template.Execute(runCommand, struct {
		Name     string
		Label    string
		Version  string
		Metadata map[string]string
	}{
		Name:     client.GetName(),
		Label:    client.GetLabel(),
		Version:  client.GetVersion(),
		Metadata: metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("could not run client %s: Templating error %w", client.GetName(), err)
	}

	// Build the subprocess's env with the context's environment variables
	splittedCommand := strings.Split(runCommand.String(), " ")
	clientCommand := exec.Command(splittedCommand[0], splittedCommand[1:]...)
	clientCommand.Env = context.Environ(true)

  // Run the subprocess
  err = clientCommand.Start()
  if err != nil {
    return nil, fmt.Errorf("an error occured while starting process for client %s: %w", client, err)
  }

  // Register the new client into the client pool
  runningClient.Pid = int32(clientCommand.Process.Pid)
  clientHandle := &ClientHandle{
		RunningClient: runningClient,
		Process:       clientCommand.Process,
	}
  runningClientsLock.Lock()
  defer runningClientsLock.Unlock()
  RunningClients()[clientCommand.Process.Pid] = clientHandle

	return clientHandle, nil
}

// Get an already running client or start a new one from the query
func ClientFromQuery(context *context.Context, query *ClientQuery) (*ClientHandle, error) {
	// First find a potential running client that matches the query
  runningClientsLock.Lock()
	for _, client := range RunningClients() {
		if query.Match(client.RunningClient) {
      runningClientsLock.Unlock()
			return client, nil
		}
	}
  // We must unlock the mutex here because RunClient will need it
  runningClientsLock.Unlock()

	// If no running client matches the query, try to run a new one
	for _, client := range context.AvailableClients() {
		if client.GetName() == query.GetName() {
			return RunClient(client, context, query.GetMetadata())
		}
	}

	return nil, fmt.Errorf(
		"could not find running client or run new client to satisfy the query %s",
		query,
	)
}
