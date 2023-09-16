package scheduling

import (
	"sync"

	"github.com/life4/genesis/slices"
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
