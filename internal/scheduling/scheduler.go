package scheduling

import (
  "context"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/processor"

	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
)

type schedulingServer struct {
	scheduling_proto.UnimplementedSchedulingServer
}

// As soon as a processor starts, it has to registers itself
func (service *schedulingServer) RegisterProcessor(c context.Context, processorRegistration *scheduling_proto.ProcessorRegistration) (*processor_proto.Processor, error) {
  registeredProcessor := registerProcessor(&processor.PendingProcessor{
    Processor: &processor.Processor{
      Processor: processorRegistration.Processor,
    },
    Registration: make(chan error),
  }, processorRegistration.Host)

  return registeredProcessor.Processor.Processor, nil
}

func init() {
	scheduling_proto.RegisterSchedulingServer(network.GrpcServer(), &schedulingServer{})
}
