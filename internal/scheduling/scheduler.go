package scheduling

import (
	"context"
	"fmt"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/tools"

	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type schedulingServer struct {
	scheduling_proto.UnimplementedSchedulingServer
}

// As soon as a processor starts, it has to registers itself
func (service *schedulingServer) RegisterProcessor(c context.Context, processorRegistration *scheduling_proto.ProcessorRegistration) (*processor_proto.Processor, error) {
	// Establish the grpc connection with the new processor
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// TODO: Handle closing the connection when the processor deregisters
	conn, err := grpc.Dial(processorRegistration.GetHost(), opts...)
	if err != nil {
		registrationErr := fmt.Errorf("Could not create connection with processor at host %s: %w", processorRegistration.GetHost(), err)
		if pendingProcessor := processor.UnQueueProcessor(processorRegistration.Processor.GetId()); pendingProcessor != nil {
			pendingProcessor.Registration <- registrationErr
		}

		return processorRegistration.Processor, registrationErr
	}

	reflectionClient := grpc_reflection_v1alpha.NewServerReflectionClient(conn)
	registeredProcessor := registerProcessor(&processor.Processor{
		Processor: processorRegistration.Processor,
	}, processorRegistration.GetHost(), NewReflectedClient(reflectionClient))

	return registeredProcessor.Processor.Processor, nil
}

// Listen for the command queue's queries and schedule it to the appropirate processor
func ListenCommandQueries() {
	for commandQuery := range tools.CommandQueue() {
		processorQuery := ProcessorQuery{ProcessorQuery: commandQuery.Command.GetProcessorQuery()}
		_, err := GetOrStartProcessor(&processorQuery)
		if err != nil {
			commandQuery.Result <- err
		}
	}
}

func init() {
	grpcServer, _ := network.GrpcServer()
	scheduling_proto.RegisterSchedulingServer(grpcServer, &schedulingServer{})
}
