package scheduling

import (
	"fmt"
	"sync"

	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/tools"

	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
)

// Registered processors are ready to receive command requests
type RegisteredProcessor struct {
	*processor.Processor
	// The host to connect to send commands
	Host string
	// Commands waiting to be scheduled
	commandQueue chan *tools.Command
	// Commands scheduled and still running on the client side
	runningCommands     map[string]*tools.Command
	runningCommandsLock *sync.Mutex
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
func registerProcessor(processorToRegister *processor.Processor, host string) *RegisteredProcessor {
	// Check if the client is already registered
	processorPoolLock.Lock()
	defer processorPoolLock.Unlock()
	registeredProcessor, ok := ProcessorPool()[processorToRegister.GetId()]
	if !ok {
		registeredProcessor = &RegisteredProcessor{
			Processor:           processorToRegister,
			Host:                host,
			commandQueue:        make(chan *tools.Command),
			runningCommands:     map[string]*tools.Command{},
			runningCommandsLock: &sync.Mutex{},
		}
		ProcessorPool()[processorToRegister.GetId()] = registeredProcessor
	}

	// If the processor was in the processor queue, inform that the registration is done
	if pendingProcessor := processor.UnQueueProcessor(processorToRegister.GetId()); pendingProcessor != nil {
		pendingProcessor.Registration <- nil
	}
	return registeredProcessor
}

// Look among the already registered clients and return the first matching client
func findRegisteredProcessor(query *ProcessorQuery) *RegisteredProcessor {
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
func GetOrStartProcessor(query *ProcessorQuery) (*RegisteredProcessor, error) {
	// First find a potential running processors that matches the query
	if registeredClient := findRegisteredProcessor(query); registeredClient != nil {
		return registeredClient, nil
	}

	// If no running processors matches the query, try to start a new one
	for _, availableProcessor := range query.GetContext().AvailableProcessors() {
		if availableProcessor.GetName() == query.GetName() {
			pendingProcessor, err := availableProcessor.Start(query.GetMetadata(), query.GetContext().Environ(true), query.GetContext().AvailableCommandPaths(availableProcessor))
			if err != nil {
				return nil, fmt.Errorf("could not start new processor (%s): %w", availableProcessor, err)
			}
			// The client should now be registered
			registeredProcessor := findRegisteredProcessor(&ProcessorQuery{
				ProcessorQuery: &scheduling_proto.ProcessorQuery{
					Id: &pendingProcessor.Id,
				},
			})
			if registeredProcessor == nil {
				return nil, fmt.Errorf("processor %s started but did not registered", pendingProcessor.Id)
			}
			return registeredProcessor, nil
		}
	}

	return nil, fmt.Errorf(
		"could not find running or run processor to satisfy the query %s",
		query,
	)
}
