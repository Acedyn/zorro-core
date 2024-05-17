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

type StreamInterface interface {
	GetId() uint64
	GetStream() *network_proto.Stream
}

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

// Send a message to the other peer after marshalling the message and
// also splitting the data into chunks if needed
func (base *GrpcOverWrtcBase[T]) send(msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not send message, proto marshaling failed: %w", err)
	}

	// TODO: Split the data into chunks if it exeeds the threshold of 1024 * 1024 bytes

	err = base.datachannel.Send(data)
	if err != nil {
		return fmt.Errorf("could not send message, proto marshaling failed: %w", err)
	}
	return nil
}

// Build and send the header frame to the server to initialize an rpc invokation
func (base *GrpcOverWrtcBase[T]) writeHeaders(ctx context.Context, stream *ClientStream, method string) error {
	// Get the metadata from the context's metadata
	headersMD, _ := metadata.FromOutgoingContext(ctx)
	// Get the timeout from the context's deadline
	var timeout time.Duration
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	// Build the headers as a proto message
	headers := network_proto.RequestHeaders{
		Method:   method,
		Timeout:  durationpb.New(timeout),
		Metadata: metadataToProto(headersMD),
	}

	// Build and send the request proto header
	return base.send(&network_proto.Request{
		Stream: stream.Stream,
		Type: &network_proto.Request_Headers{
			Headers: &headers,
		},
	})
}

// Build and send the message frame
func (base *GrpcOverWrtcBase[T]) writeMessage(stream T, in any, endOfStream bool) error {
	var err error
	var data []byte

	// Marshall the input argument of the rpc method
	if in != nil {
		data, err = proto.Marshal(in.(proto.Message))
		if err != nil {
			return err
		}
	}

	// Build and send the request proto message
	return base.send(&network_proto.Request{
		Stream: stream.GetStream(),
		Type: &network_proto.Request_Message{
			Message: &network_proto.RequestMessage{
				Eos:        endOfStream,
				HasMessage: in != nil,
				PacketMessage: &network_proto.PacketMessage{
					Data: data,
				},
			},
		},
	})
}

// Convert grpc metadata to protobuf metadata so it can be serialized
func metadataToProto(md metadata.MD) *network_proto.Metadata {
	if md == nil || md.Len() == 0 {
		return nil
	}
	result := make(map[string]*network_proto.Strings, md.Len())
	for key, values := range md {
		result[key] = &network_proto.Strings{
			Values: values,
		}
	}
	return &network_proto.Metadata{
		Md: result,
	}
}

// Convert protobuf metadata to grpc metadata so it can be use with the grpc api
func metadataFromProto(md *network_proto.Metadata) metadata.MD {
	if md == nil || md.Md == nil || len(md.Md) == 0 {
		return nil
	}
	result := make(metadata.MD, len(md.Md))
	for key, values := range md.Md {
		valuesCopy := make([]string, len(values.Values))
		copy(valuesCopy, values.Values)
		result[key] = valuesCopy
	}

	return result
}

// Create and setup a datachannel with an injected onMessage handler
func setupGrpcOverWrtc[T StreamInterface](
	wrtcConnection *wrtc.WRTCConnection,
	onMessage func(msg webrtc.DataChannelMessage),
	options ...GrpcOverWrtcOption,
) (*GrpcOverWrtcBase[T], error) {
	// Patch the onMessage datachannel option to inject the server logic
	grpcOverWrtcOption := DefaultGrpcOverWrtcOptions
	for _, option := range options {
		option(&grpcOverWrtcOption)
	}

	// If the onMessage option is already set, create an handler that will call
	// both the handler and the server logic's handler
	if grpcOverWrtcOption.datachannelOptions.OnMessage == nil {
		grpcOverWrtcOption.datachannelOptions.OnMessage = onMessage
	} else {
		patchedOnMessage := func(msg webrtc.DataChannelMessage) {
			onMessage(msg)
			grpcOverWrtcOption.datachannelOptions.OnMessage(msg)
		}
		grpcOverWrtcOption.datachannelOptions.OnMessage = patchedOnMessage
	}

	// Setup the connection with the patched options
	datachannel, err := SetupGrpcOverWrtc(wrtcConnection,
		WithContext(grpcOverWrtcOption.ctx),
		WithLabel(grpcOverWrtcOption.label),
		WithAsOfferer(grpcOverWrtcOption.asOfferer),
		WithDatachannelOptions(&grpcOverWrtcOption.datachannelOptions),
	)

	return &GrpcOverWrtcBase[T]{
		datachannel: datachannel,
		streams:     map[uint64]T{},
		streamMutex: sync.Mutex{},
	}, err
}
