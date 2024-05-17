package rpc

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"

	network_proto "github.com/Acedyn/zorro-proto/gen/protos/network"
	"github.com/pion/webrtc/v4"
	"google.golang.org/grpc"
)

// Create and setup a datachannel to allow the peer to act as a GRPC client
// with the other peer
func SetupGrpcOverWrtcClient(wrtcConnection *wrtc.WRTCConnection, options ...GrpcOverWrtcOption) (GrpcOverWrtcClient, error) {
	grpcOverWrtcClient := GrpcOverWrtcClient{
		streamIDCounter: 0,
	}

	// Inject the server logic on every messages
	onClientMessage := func(msg webrtc.DataChannelMessage) {
		go grpcOverWrtcClient.onMessage(msg)
	}

	grpcOverWrtcBase, err := setupGrpcOverWrtc[*ClientStream](wrtcConnection, onClientMessage, options...)
	grpcOverWrtcClient.GrpcOverWrtcBase = grpcOverWrtcBase

	return grpcOverWrtcClient, err
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
func (client *GrpcOverWrtcClient) onMessage(msg webrtc.DataChannelMessage) {
	fmt.Println("MESSAGE RECEIVED")
}

// Logic and data related to a stream on the client side
type ClientStream struct {
	*network_proto.Stream
}

func (stream *ClientStream) GetStream() *network_proto.Stream {
	return stream.Stream
}
