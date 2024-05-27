package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	network_proto "github.com/Acedyn/zorro-proto/gen/protos/network"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Stream struct {
	ctx             context.Context
	cancel          context.CancelCauseFunc
	proto           *network_proto.Stream
	pendingMessages chan []byte
	node            *GrpcOverWrtcNode
	method          string
	header          metadata.MD
	trailer         metadata.MD
	mutex           sync.Mutex
}

// Build and send the request proto header
func (stream *Stream) sendRequestHeaders(ctx context.Context, method string) error {
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
	return stream.node.sendFrame(&network_proto.Request{
		Stream: stream.proto,
		Type: &network_proto.Request_Headers{
			Headers: &headers,
		},
	})
}

// Build and send a request message frame
func (stream *Stream) sendRequestMessage(packet any, endOfStream bool) error {
	var err error
	var data []byte

	// Marshall the input argument of the rpc method
	if packet != nil {
		data, err = proto.Marshal(packet.(proto.Message))
		if err != nil {
			return err
		}
	}

	// Build and send the request proto message
	return stream.node.sendFrame(&network_proto.Request{
		Stream: stream.proto,
		Type: &network_proto.Request_Message{
			Message: &network_proto.RequestMessage{
				Eos:        endOfStream,
				HasMessage: packet != nil,
				PacketMessage: &network_proto.PacketMessage{
					Data: data,
					Eom:  true,
				},
			},
		},
	})
}

// Build and send a response message frame
func (stream *Stream) sendResponseMessage(packet any) error {
	var err error
	var data []byte

	// Marshall the input argument of the rpc method
	if packet != nil {
		data, err = proto.Marshal(packet.(proto.Message))
		if err != nil {
			return err
		}
	}

	// Build and send the response proto message
	return stream.node.sendFrame(&network_proto.Response{
		Stream: stream.proto,
		Type: &network_proto.Response_Message{
			Message: &network_proto.ResponseMessage{
				PacketMessage: &network_proto.PacketMessage{
					Data: data,
					Eom:  true,
				},
			},
		},
	})
}

// Build and send a response trailer that will end the connection
func (stream *Stream) sendResponseTrailer(err error) error {
	// TODO: Find a better way to convert from grpc.status.Status to network_proto.Status
	status := status.FromContextError(err)
	statusRaw, err := proto.Marshal(status.Proto())
	if err != nil {
		return fmt.Errorf("could not build trailer status: %w", err)
	}

	statusProto := &network_proto.Status{}
	err = proto.Unmarshal(statusRaw, statusProto)
	if err != nil {
		return fmt.Errorf("could not build trailer status: %w", err)
	}

	return stream.node.sendFrame(&network_proto.Response{
		Stream: stream.proto,
		Type: &network_proto.Response_Trailers{
			Trailers: &network_proto.ResponseTrailers{
				Status:   statusProto,
				Metadata: metadataToProto(stream.trailer),
			},
		},
	})
}

// Store a message into the pending messages channel so it can be picked up
// later using receivedMessage()
func (stream *Stream) storeIncommingMessage(message []byte) error {
	select {
	case <-stream.ctx.Done():
		return fmt.Errorf("message storing canceled: %w", stream.ctx.Err())
	case stream.pendingMessages <- message:
		return nil
	}
}

// Wait for the next message and apply the result on the given proto
func (stream *Stream) RecvMsg(m any) error {
	select {
	case <-stream.ctx.Done():
		if errors.Is(context.Cause(stream.ctx), io.EOF) {
			return io.EOF
		}
		return fmt.Errorf("message reception canceled: %w", stream.ctx.Err())
	case msgBytes := <-stream.pendingMessages:
		return proto.Unmarshal(msgBytes, m.(proto.Message))
	}
}

// Getter for the context
func (stream *Stream) Context() context.Context {
	return stream.ctx
}

// Getter for the method name
func (stream *Stream) Method() string {
	return stream.method
}

// Getter for the trailer
func (stream *Stream) Trailer() metadata.MD {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	return stream.trailer.Copy()
}

// Getter for the header
func (stream *Stream) Header() (metadata.MD, error) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	return stream.header.Copy(), nil
}

// SetTrailer sets the trailer metadata which will be sent with the RPC status.
// When called more than once, all the provided metadata will be merged.
func (stream *Stream) SetTrailer(trailer metadata.MD) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if stream.trailer == nil {
		stream.trailer = trailer
	} else if trailer != nil {
		stream.trailer = metadata.Join(stream.trailer, trailer)
	}

	return nil
}

// SetHeader sets the header metadata. It may be called multiple times.
// When call multiple times, all the provided metadata will be merged.
func (stream *Stream) SetHeader(header metadata.MD) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if stream.header == nil {
		stream.header = header
	} else if header != nil {
		stream.header = metadata.Join(stream.header, header)
	}

	return nil
}

// SendHeader sends the header metadata. This method is meant to be used by the server side
// stream. This method is named this way to satisfy the ServerTransportStream interface.
func (stream *Stream) SendHeader(header metadata.MD) error {
	if err := stream.SetHeader(header); err != nil {
		return err
	}
	// TODO: Implement the headers sending
	return nil
}

// Create client and server structs to satisfy the grpc.ServerStream and grpc.ClientStream
// interfaces
type ClientStream struct{ *Stream }

func (s *ClientStream) CloseSend() error    { return s.sendRequestMessage(nil, true) }
func (s *ClientStream) SendMsg(m any) error { return s.sendRequestMessage(m, false) }

type ServerStream struct{ *Stream }

func (s *ServerStream) SendMsg(m any) error      { return s.sendResponseMessage(m) }
func (s *ServerStream) SetTrailer(t metadata.MD) { s.Stream.SetTrailer(t) }
