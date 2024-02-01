package subprocess

import (
	"context"
	"fmt"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/reflection"
	"github.com/Acedyn/zorro-core/internal/tools"
	"github.com/Acedyn/zorro-core/pkg/scheduling"

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

// The suprocess scheduler will send command queries to suprocesses via gRPC
type SubprocessScheduler struct {
	grpcStatus       *network.GrpcServerStatus
	schedulingServer *schedulingServer
}

// Identifiers used to match againts the scheduler query
func (*SubprocessScheduler) GetInfo() scheduling.SchedulerInfo {
	return scheduling.SchedulerInfo{
		Name: "subprocess",
	}
}

// Send the command query to the appropriate processor
func (*SubprocessScheduler) ScheduleCommand(commandQuery *tools.CommandQuery) {
	processorQuery := ProcessorQuery{ProcessorQuery: commandQuery.Command.GetProcessorQuery()}
	// Get the processor that will execute the command query
	registeredProcessor, err := GetOrStartProcessor(commandQuery.Context, &processorQuery)
	if err != nil {
		commandQuery.Result <- err
	}

	// Execute the command query
	commandQuery.Result <- registeredProcessor.ProcessCommand(commandQuery)
}

// Start the suprocess scheduling server
func (subprocessScheduler *SubprocessScheduler) Initialize() {
	grpcServer, grpcStatus := network.GrpcServer()
	subprocessScheduler.grpcStatus = grpcStatus
	subprocessScheduler.schedulingServer = &schedulingServer{}
	scheduling_proto.RegisterSchedulingServer(grpcServer, subprocessScheduler.schedulingServer)
}

// Register the subprocess scheduler to the list of available schedulers
func init() {
	subprocessScheduler := &SubprocessScheduler{}
	scheduling.AvailableSchedulers()[subprocessScheduler.GetInfo().Name] = subprocessScheduler
}
