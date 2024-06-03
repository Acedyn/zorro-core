package processor

import (
	"sync"

	"github.com/Acedyn/zorro-core/internal/network/rpc"
	"github.com/Acedyn/zorro-core/internal/tools"

	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
)

// Registered processors are ready to receive command requests
type RegisteredProcessor struct {
	*scheduling_proto.Processor
	// The host to connect to send commands
	Host string
	// Commands waiting to be scheduled
	commandQueue chan *tools.Command
	// Commands scheduled and still running on the client side
	runningCommands     map[string]*tools.Command
	runningCommandsLock *sync.Mutex
	// The client used to send command requests
	Client *rpc.GrpcOverWrtcClient
}
