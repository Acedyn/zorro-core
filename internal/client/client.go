package client

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
)

type StartClientContext interface {
	AvailableClients() []*Client
	Environ(bool) []string
}

type ClientHandle struct {
	Client       *Client
	Process      *os.Process
	Registration chan error
}

var (
	ClientQueueLock = &sync.Mutex{}
	clientQueue     map[string]*ClientHandle
	once            sync.Once
)

// Getter for the clients queue singleton which holds the queue
// of client waiting to be registered
func ClientQueue() map[string]*ClientHandle {
	once.Do(func() {
		clientQueue = map[string]*ClientHandle{}
	})

	return clientQueue
}

// Test if a client matches the query's requirements
func (query *ClientQuery) MatchClient(client *Client) bool {
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
	context StartClientContext,
	metadata map[string]string,
) (*ClientHandle, error) {
	registration := make(chan error)
	clientHandle := &ClientHandle{
		Client:       &(*client),
		Registration: registration,
	}
	startingStatus := ClientStatus_STARTING
	clientHandle.Client.Status = &startingStatus
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

	// Register the new client into the client queue and wait for it to be registered
	clientHandle.Client.Pid = int32(clientCommand.Process.Pid)
	clientHandle.Process = clientCommand.Process
	ClientQueueLock.Lock()
	defer ClientQueueLock.Unlock()
	ClientQueue()[clientHandle.Client.GetId()] = clientHandle

	return clientHandle, <-registration
}

// Update the client with a patch
func (client *Client) Patch(patch *Client) bool {
	// Patch the local version of the client
	isPatched := false
	if maps.Equal(client.Metadata, patch.GetMetadata()) {
		maps.Update(client.Metadata, patch.GetMetadata())
		isPatched = true
	}
	if patch.Label != nil && client.Label != patch.Label {
		client.Label = patch.Label
		isPatched = true
	}
	if patch.Status != nil && client.Status != patch.Status {
		client.Status = patch.Status
		isPatched = true
	}

	return isPatched
}
