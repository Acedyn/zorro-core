package processor

import (
	"sync"

	"github.com/Acedyn/zorro-core/internal/network/rpc"
	"github.com/Acedyn/zorro-core/internal/tools"

	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
)

// Pool of available processors
type ProcessorPool struct {
	registeredMutex sync.Mutex
	registered      map[string]*RegisteredProcessor
}

// Register the given client to the client pool
func (pool *ProcessorPool) RegisterProcessor(
	processorToRegister *scheduling_proto.Processor,
	client *rpc.GrpcOverWrtcClient,
) *RegisteredProcessor {
	// Check if the client is already registered
	pool.registeredMutex.Lock()
	defer pool.registeredMutex.Unlock()
	registeredProcessor, ok := pool.registered[processorToRegister.GetId()]
	if !ok {
		registeredProcessor = &RegisteredProcessor{
			Processor:           processorToRegister,
			commandQueue:        make(chan *tools.Command),
			runningCommands:     map[string]*tools.Command{},
			runningCommandsLock: &sync.Mutex{},
			Client:              client,
		}
		pool.registered[processorToRegister.GetId()] = registeredProcessor
	}

	return registeredProcessor
}

// Look among the already registered clients and return the first matching client
func (pool *ProcessorPool) FindRegisteredProcessor(query *ProcessorQuery) *RegisteredProcessor {
	pool.registeredMutex.Lock()
	defer pool.registeredMutex.Unlock()

	// The look by id is faster since its the primary key
	if query.Id != nil {
		return pool.registered[*query.Id]
	}

	// Test all the registered clients one by one
	for _, registeredClient := range pool.registered {
		if query.MatchProcessor(registeredClient.Processor) {
			return registeredClient
		}
	}

	return nil
}
