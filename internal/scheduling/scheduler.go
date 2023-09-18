package scheduling

import (
	"context"
	"fmt"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/tools"
	"github.com/life4/genesis/maps"
)

type schedulingServer struct {
  UnimplementedSchedulingServer
}

func (server *schedulingServer) RegisterClient(c context.Context, client *Client) (*Client, error) {
  clientPoolLock.Lock()
  defer clientPoolLock.Unlock()
  clientHandle, ok := ClientPool()[int(client.GetPid())]
  if !ok {
    return nil, fmt.Errorf("could not register client %s: client not found in client pool", client)
  }

  // Patch the local version of the client
  maps.Update(clientHandle.Client.Metadata, client.GetMetadata())
  if client.Label != nil {
    clientHandle.Client.Label = client.Label
  }
  if client.Status != nil {
    clientHandle.Client.Status = client.Status
  }

  // Send the patched version
  return clientHandle.Client, nil
}

func (server *schedulingServer) CommunicateCommands(stream Scheduling_CommunicateCommandsServer) error {
  return nil
}

func listenCommandRequests() {
  // for command := range tools.CommandPool() {
    // client := ClientFromQuery(command.Command.GetClientQuery())
  // }
}

func init() {
  RegisterSchedulingServer(network.GrpcServer(), &schedulingServer{})
}
