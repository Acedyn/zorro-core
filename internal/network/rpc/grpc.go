package rpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"

	network_proto "github.com/Acedyn/zorro-proto/gen/protos/network"
	"github.com/pion/webrtc/v4"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type GrpcOverWrtcBase[T StreamInterface] struct {
	datachannel *webrtc.DataChannel
	streams     map[uint64]T
	streamMutex sync.Mutex
}

// Create a new stream in a thread safe manner
func (base *GrpcOverWrtcBase[T]) storeStream(stream T) {
	base.streamMutex.Lock()
	base.streams[stream.GetId()] = stream
	base.streamMutex.Unlock()
}
