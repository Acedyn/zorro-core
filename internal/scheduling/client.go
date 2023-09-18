package scheduling

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"

	"github.com/Acedyn/zorro-core/internal/tools"
)

type Context interface {
  AvailableClients() []*Client
  Environ(bool) []string
}

// Internal struct to keep track of the running clients
type ClientHandle struct {
	Client *Client
	Process       *os.Process
}

var (
  runningClientsLock = &sync.Mutex{}
	clientPool map[int]*ClientHandle
	once           sync.Once
)

// Getter for the clients singleton
func ClientPool() map[int]*ClientHandle {
	once.Do(func() {
		clientPool = map[int]*ClientHandle{}
	})

	return clientPool
}

// Test if running client matches the query's requirements
func MatchClientQuery(query *tools.ClientQuery, client *Client) bool {
	// Test the name
	if query.Name != nil {
		// Some clients are supersets of other clients
		// If so they should match also their subsets
		subsets := append(client.GetSubsets(), client.GetName())
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
func (client *Client) Start(
	context Context,
	metadata map[string]string,
) (*ClientHandle, error) {
  clientHandle := &ClientHandle{
		Client: &(*client),
	}
  clientHandle.Client.Status = ClientStatus_STARTING
  clientHandle.Client.Metadata = maps.Merge(clientHandle.Client.GetMetadata(), metadata)

	// Build the command template
	template, err := template.New(client.GetName()).Parse(clientHandle.Client.GetStartClientTemplate())
	if err != nil {
		return nil, fmt.Errorf(
      "could not run client %s: Invalid launch template %w", 
      clientHandle.Client.GetName(), 
      err,
    )
	}

	// Apply the metadata and the name on the template
	runCommand := &bytes.Buffer{}
	err = template.Execute(runCommand, struct {
		Name     string
		Label    string
		Version  string
		Metadata map[string]string
	}{
		Name:     clientHandle.Client.GetName(),
		Label:    clientHandle.Client.GetLabel(),
		Version:  clientHandle.Client.GetVersion(),
		Metadata: clientHandle.Client.GetMetadata(),
	})
	if err != nil {
		return nil, fmt.Errorf("could not run client %s: Templating error %w", clientHandle.Client.GetName(), err)
	}

	// Build the subprocess's env with the context's environment variables
	splittedCommand := strings.Split(runCommand.String(), " ")
	clientCommand := exec.Command(splittedCommand[0], splittedCommand[1:]...)
	clientCommand.Env = context.Environ(true)

  // Start the subprocess
  err = clientCommand.Start()
  if err != nil {
    return nil, fmt.Errorf("an error occured while starting process for client %s: %w", client, err)
  }

  // Register the new client into the client pool
  clientHandle.Client.Pid = int32(clientCommand.Process.Pid)
  clientHandle.Process = clientCommand.Process
  runningClientsLock.Lock()
  defer runningClientsLock.Unlock()
  ClientPool()[clientCommand.Process.Pid] = clientHandle

	return clientHandle, nil
}

// Get an already running client or start a new one from the query
func ClientFromQuery(context Context, query *tools.ClientQuery) (*ClientHandle, error) {
	// First find a potential running client that matches the query
  runningClientsLock.Lock()
	for _, clientHandle := range ClientPool() {
		if MatchClientQuery(query, clientHandle.Client) {
      runningClientsLock.Unlock()
			return clientHandle, nil
		}
	}
  // We must unlock the mutex here because client.Start will need it
  runningClientsLock.Unlock()

	// If no running client matches the query, try to run a new one
	for _, client := range context.AvailableClients() {
		if client.GetName() == query.GetName() {
			return client.Start(context, query.GetMetadata())
		}
	}

	return nil, fmt.Errorf(
		"could not find running client or run new client to satisfy the query %s",
		query,
	)
}
