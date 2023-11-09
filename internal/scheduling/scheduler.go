package scheduling

import (
	"context"
	"fmt"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/reflection"
	"github.com/Acedyn/zorro-core/internal/tools"

	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
)

type schedulingServer struct {
	scheduling_proto.UnimplementedSchedulingServer
}

// As soon as a processor starts, it has to registers itself
func (service *schedulingServer) RegisterProcessor(c context.Context, processorRegistration *scheduling_proto.ProcessorRegistration) (*processor_proto.Processor, error) {
	// TODO: Handle closing the connection when the processor deregisters
	reflectedClient, err := reflection.NewReflectedClient(processorRegistration.GetHost())
	if err != nil {
		registrationErr := fmt.Errorf("could not create reflection client with processor at host %s: %w", processorRegistration.GetHost(), err)
		if pendingProcessor := processor.UnQueueProcessor(processorRegistration.Processor.GetId()); pendingProcessor != nil {
			pendingProcessor.Registration <- registrationErr
		}

		return processorRegistration.Processor, registrationErr
	}

	// Inform the processor queue that the processor was registered
	registeredProcessor := registerProcessor(&processor.Processor{
		Processor: processorRegistration.Processor,
	}, processorRegistration.GetHost(), reflectedClient)

	return registeredProcessor.Processor.Processor, nil
}

// Listen for the command queue's queries and schedule it to the appropirate processor
func ListenCommandQueries() {
	for commandQuery := range tools.CommandQueue() {
		processorQuery := ProcessorQuery{ProcessorQuery: commandQuery.Command.GetProcessorQuery()}
		// Get the processor that will execute the command query
		registeredProcessor, err := GetOrStartProcessor(commandQuery.Context, &processorQuery)
		if err != nil {
			commandQuery.Result <- err
		}

		// Execute the command query
		commandQuery.Result <- registeredProcessor.ProcessCommand(commandQuery)
	}
}

func init() {
	grpcServer, _ := network.GrpcServer()
	scheduling_proto.RegisterSchedulingServer(grpcServer, &schedulingServer{})
	go ListenCommandQueries()
}
