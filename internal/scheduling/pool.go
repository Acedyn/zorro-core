package scheduling

import (
	"fmt"
	"sync"

	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/tools"

	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
)

// Registered processors are ready to receive command requests
type RegisteredProcessor struct {
	*processor.Processor
	// The host to connect to send commands
	Host string
	// Commands waiting to be scheduled
	CommandQueue chan *tools.Command
	// Commands scheduled and still running on the client side
	RunningCommands     map[string]*tools.Command
	RunningCommandsLock *sync.Mutex
}

var (
	processorPoolLock = &sync.Mutex{}
	processorPool     map[string]*RegisteredProcessor
	once              sync.Once
)

// Getter for the clients pool singleton
func ProcessorPool() map[string]*RegisteredProcessor {
	once.Do(func() {
		processorPool = map[string]*RegisteredProcessor{}
	})

	return processorPool
}

// Register the given client to the client pool
func registerProcessor(pendingProcessor *processor.PendingProcessor, host string) *RegisteredProcessor {
	// Check if the client is already registered
	processorPoolLock.Lock()
	defer processorPoolLock.Unlock()
	registeredProcessor, ok := ProcessorPool()[pendingProcessor.GetId()]
	if !ok {
		registeredProcessor = &RegisteredProcessor{
			Processor:           pendingProcessor.Processor,
			Host:                host,
			CommandQueue:        make(chan *tools.Command),
			RunningCommands:     map[string]*tools.Command{},
			RunningCommandsLock: &sync.Mutex{},
		}
		ProcessorPool()[pendingProcessor.GetId()] = registeredProcessor
	}

  processor.UnQueueProcessor(pendingProcessor)
	pendingProcessor.Registration <- nil
	return registeredProcessor
}

// Look among the already registered clients and return the first matching client
func findRegisteredProcessor(query *processor.ProcessorQuery) *RegisteredProcessor {
	processorPoolLock.Lock()
	defer processorPoolLock.Unlock()

	// The look by id is faster since its the primary key
	if query.Id != nil {
		return ProcessorPool()[*query.Id]
	}

	// Test all the registered clients one by one
	for _, registeredClient := range ProcessorPool() {
		if query.MatchProcessor(registeredClient.Processor) {
			return registeredClient
		}
	}

	return nil
}

// Get an already running processor or start a new one from the query
func GetOrStartProcessor(c *context.Context, query *processor.ProcessorQuery) (*RegisteredProcessor, error) {
	// First find a potential running processors that matches the query
	if registeredClient := findRegisteredProcessor(query); registeredClient != nil {
		return registeredClient, nil
	}

	// If no running processors matches the query, try to start a new one
	for _, availableProcessor := range c.AvailableProcessors() {
		if availableProcessor.GetName() == query.GetName() {
			pendingProcessor, err := availableProcessor.Start(query.GetMetadata(), c.Environ(true))
			if err != nil {
				return nil, fmt.Errorf("could not start new client %s: %w", availableProcessor, err)
			}
			// The client should now be registered
			registeredProcessor := findRegisteredProcessor(&processor.ProcessorQuery{
				ProcessorQuery: &processor_proto.ProcessorQuery{
					Id: &pendingProcessor.Id,
				},
			})
			if registeredProcessor == nil {
				return nil, fmt.Errorf("client %s started but did not registered", pendingProcessor.Id)
			}
			return registeredProcessor, nil
		}
	}

	return nil, fmt.Errorf(
		"could not find running client or run new client to satisfy the query %s",
		query,
	)
}
