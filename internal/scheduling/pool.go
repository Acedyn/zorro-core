package scheduling

import (
	"fmt"
	"sync"

	"github.com/Acedyn/zorro-core/internal/client"
	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/tools"
)

// Internal struct to keep track of the running clients
type RegisteredClient struct {
	Client                *client.Client
	CommandQueue          chan *tools.Command
	ScheduledCommands     map[string]*tools.Command
	ScheduledCommandsLock *sync.Mutex
}

// Register the given client to the client pool
func registerClient(clientToRegister *client.Client) *RegisteredClient {
	// Check if the client is already registered
	clientPoolLock.Lock()
	defer clientPoolLock.Unlock()
	registeredClient, ok := ClientPool()[clientToRegister.GetId()]
	if !ok {
		registeredClient = &RegisteredClient{
			Client:                clientToRegister,
			CommandQueue:          make(chan *tools.Command),
			ScheduledCommands:     map[string]*tools.Command{},
			ScheduledCommandsLock: &sync.Mutex{},
		}
		ClientPool()[clientToRegister.GetId()] = registeredClient
	}

	// Check if the client was queued, we must inform the submitter that the registering has happened
	client.ClientQueueLock.Lock()
	defer client.ClientQueueLock.Unlock()
	if clientHandle, ok := client.ClientQueue()[clientToRegister.GetId()]; ok {
		clientHandle.Registration <- nil
	}

	return registeredClient

}

var (
	clientPoolLock = &sync.Mutex{}
	clientPool     map[string]*RegisteredClient
	once           sync.Once
)

// Getter for the clients pool singleton
func ClientPool() map[string]*RegisteredClient {
	once.Do(func() {
		clientPool = map[string]*RegisteredClient{}
	})

	return clientPool
}

// Get an already running client or start a new one from the query
func ClientFromQuery(c *context.Context, query *client.ClientQuery) (*RegisteredClient, error) {
	// First find a potential running client that matches the query
	clientPoolLock.Lock()
	for _, registeredClient := range ClientPool() {
		if query.MatchClient(registeredClient.Client) {
			clientPoolLock.Unlock()
			return registeredClient, nil
		}
	}
	// We must unlock the mutex here because client.Start will need it
	clientPoolLock.Unlock()

	// If no running client matches the query, try to start a new one
	for _, client := range c.AvailableClients() {
		if client.GetName() == query.GetName() {
			clientHandle, err := client.Start(c, query.GetMetadata())
			if err != nil {
				return nil, fmt.Errorf("could not start new client %s: %w", client, err)
			}
			return registerClient(clientHandle.Client), nil
		}
	}

	return nil, fmt.Errorf(
		"could not find running client or run new client to satisfy the query %s",
		query,
	)
}
