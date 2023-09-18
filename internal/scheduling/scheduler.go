package scheduling

import (
  "github.com/Acedyn/zorro-core/internal/network"
)

type schedulingServer struct {
  UnimplementedSchedulingServer
}

func (server *schedulingServer) CommunicateCommands(stream Scheduling_CommunicateCommandsServer) error {
  return nil
}

func init() {
  RegisterSchedulingServer(network.GrpcServer(), &schedulingServer{})
}
