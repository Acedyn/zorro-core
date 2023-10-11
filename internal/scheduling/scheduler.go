package scheduling

import (
	"context"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/tools"

	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
)

type schedulingServer struct {
	scheduling_proto.UnimplementedSchedulingServer
}

// As soon as a processor starts, it has to registers itself
func (service *schedulingServer) RegisterProcessor(c context.Context, processorRegistration *scheduling_proto.ProcessorRegistration) (*processor_proto.Processor, error) {
	registeredProcessor := registerProcessor(&processor.Processor{
		Processor: processorRegistration.Processor,
	}, processorRegistration.Host)

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
