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

	clientHandle := client.UnQueueClient(clientToRegister)
	clientHandle.Registration <- nil

	return registeredClient
}

// Look among the already registered clients and return the first matching client
func findRegisteredClient(query *client.ClientQuery) *RegisteredClient {
	clientPoolLock.Lock()
	defer clientPoolLock.Unlock()

	// The look by id is faster since its the primary key
	if query.Id != nil {
		return clientPool[*query.Id]
	}

	// Test all the registered clients one by one
	for _, registeredClient := range ClientPool() {
		if query.MatchClient(registeredClient.Client) {
			return registeredClient
		}
	}

	return nil
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
	if registeredClient := findRegisteredClient(query); registeredClient != nil {
		return registeredClient, nil
	}

	// If no running client matches the query, try to start a new one
	for _, availableClient := range c.AvailableClients() {
		if availableClient.GetName() == query.GetName() {
			clientHandle, err := availableClient.Start(c, query.GetMetadata())
			if err != nil {
				return nil, fmt.Errorf("could not start new client %s: %w", availableClient, err)
			}
			// The client should now be registered
			registeredClient := findRegisteredClient(&client.ClientQuery{
				Id: &clientHandle.Client.Id,
			})
			if registeredClient == nil {
				return nil, fmt.Errorf("client %s started but did not registered", clientHandle.Client.Id)
			}
			return registeredClient, nil
		}
	}

	return nil, fmt.Errorf(
		"could not find running client or run new client to satisfy the query %s",
		query,
	)
}
