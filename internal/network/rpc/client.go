package rpc

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"

	network_proto "github.com/Acedyn/zorro-proto/gen/protos/network"
	"github.com/pion/webrtc/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Create and setup a datachannel to allow the peer to act as a GRPC client
// with the other peer
func SetupGrpcOverWrtcClient(wrtcConnection *wrtc.WRTCConnection, options ...GrpcOverWrtcOption) (GrpcOverWrtcClient, error) {
	grpcOverWrtcBase, err := setupGrpcOverWrtc[*ClientStream](wrtcConnection, onClientMessage, options...)

	return GrpcOverWrtcClient{
		GrpcOverWrtcBase: grpcOverWrtcBase,
	}, err
}

type GrpcOverWrtcClient struct {
	*GrpcOverWrtcBase[*ClientStream]
	streamIDCounter uint64
}

// Invoke will send the request to the server and wait for a response for unary rpc calls
// this is typically called by the generated code
func (client *GrpcOverWrtcClient) Invoke(ctx context.Context, method string, in, out any, options ...grpc.CallOption) error {
	// Register the stream first
	stream := client.newStream()

	// TODO: Use the provided options

	// Send the header first, we are counting on the datachannel to make sure the
	// headers is received before the actuall message
	err := client.writeHeaders(ctx, stream, method)
	if err != nil {
		return fmt.Errorf("an error occured when writing rpc headers: %s", err)
	}

	// Send the message containing the rpc input arg and set the end of stream directly
	// since this is an unary rpc call
	err = client.writeMessage(stream, in, true)
	if err != nil {
		return fmt.Errorf("an error occured when writing rpc headers: %s", err)
	}

	return nil
}

// NewStream creates a new Stream for the client side. This is typically
// called by generated code. ctx is used for the lifetime of the stream.
func (client *GrpcOverWrtcClient) NewStream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	method string,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {

	return nil, nil
}

// Build and send the header frame to the server to initialize an rpc invokation
func (client *GrpcOverWrtcClient) writeHeaders(ctx context.Context, stream *ClientStream, method string) error {
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
	return client.send(&network_proto.Request{
		Stream: stream.Stream,
		Type: &network_proto.Request_Headers{
			Headers: &headers,
		},
	})
}

// Build and send the message frame
func (client *GrpcOverWrtcClient) writeMessage(stream *ClientStream, in any, endOfStream bool) error {
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
	return client.send(&network_proto.Request{
		Stream: stream.Stream,
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

// Create a new stream in a thread safe manner and generate its ID
func (client *GrpcOverWrtcClient) newStream() *ClientStream {
	id := atomic.AddUint64(&client.streamIDCounter, 1)
	newStream := &ClientStream{
		Stream: &network_proto.Stream{Id: id},
	}

	client.storeStream(newStream)
	return newStream
}

// Inject the client logic on every messages
func onClientMessage(msg webrtc.DataChannelMessage) {

}

// Logic and data related to a stream on the client side
type ClientStream struct {
	*network_proto.Stream
}
