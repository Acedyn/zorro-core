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

var (
	runningClients []*RunningClient
	once   sync.Once
)

// Getter for the clients singleton
func RunningClients() []*RunningClient {
	once.Do(func() {
		runningClients = []*RunningClient{}
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
) (*RunningClient, *os.Process, error) {
  runningClient := &RunningClient{
    Client: client,
    Status: ClientStatus_STARTING,
    Metadata: metadata,
  }

  // Build the command template
  template, err := template.New(client.GetName()).Parse(client.GetRunClientTemplate())
  if err != nil {
    return nil, nil, fmt.Errorf("Could not run client %s: Invalid launch template %w", client.GetName(), err)
  }
  
  // Apply the metadata and the name on the template
  runCommand := &bytes.Buffer{}
  err = template.Execute(runCommand, struct {
    Name string
    Metadata map[string]string
  }{
    Name: client.GetName(),
    Metadata: metadata,
  })
  if err != nil {
    return nil, nil, fmt.Errorf("Could not run client %s: Templating error %w", client.GetName(), err)
  }

  // Run the subprocess with the context's environment variables
  splittedCommand := strings.Split(runCommand.String(), " ")
  clientCommand := exec.Command(splittedCommand[0], splittedCommand[1:]...)
  clientCommand.Env = context.Environ(true)

  return runningClient, clientCommand.Process, clientCommand.Start()
}

// Get an already running client or start a new one from the query
func RunningClientFromQuery(query *ClientQuery) *RunningClient {
  // First find a potential running client that matches the query
  for _, client := range RunningClients() {
    if query.Match(client) {
      return client
    }
  }
  return nil
}
