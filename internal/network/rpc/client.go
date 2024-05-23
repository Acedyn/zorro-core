package rpc

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"
	"github.com/Acedyn/zorro-core/internal/utils"

	network_proto "github.com/Acedyn/zorro-proto/gen/protos/network"
	"github.com/pion/webrtc/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Create and setup a datachannel to allow the peer to act as a GRPC client
// with the other peer
func NewClient(wrtcConnection *wrtc.WRTCConnection, options ...GrpcOverWrtcOption) (*GrpcOverWrtcClient, error) {
	grpcOverWrtcClient := GrpcOverWrtcClient{
		streamIDCounter: 0,
	}

	// Inject the client logic on every messages
	options = append(options, withDatachannelOptions(&wrtc.DatachannelOptions{
		OnMessage: func(msg webrtc.DataChannelMessage) {
			go grpcOverWrtcClient.onMessage(msg)
		},
	}))

	grpcOverWrtcNode, err := setupGrpcOverWrtc(wrtcConnection, options...)
	grpcOverWrtcClient.GrpcOverWrtcNode = grpcOverWrtcNode

	return &grpcOverWrtcClient, err
}

// The client is responsible for creating gRPC requests
type GrpcOverWrtcClient struct {
	*GrpcOverWrtcNode
	streamIDCounter uint64
}

// Invoke will send the request to the server and wait for a response for unary rpc calls
// this is typically called by the generated code
func (client *GrpcOverWrtcClient) Invoke(ctx context.Context, method string, in, out any, options ...grpc.CallOption) error {
	// Register the stream first
	stream := client.newStream(ctx, method)
	defer client.removeStream(stream)

	// TODO: Use the provided options

	// Send the header first, we are counting on the datachannel to make sure the
	// headers is received before the actuall message
	err := stream.sendRequestHeaders(ctx, method)
	if err != nil {
		return fmt.Errorf("an error occured when writing rpc headers: %s", err)
	}

	// Send the message containing the rpc input arg and set the end of stream directly
	// since this is an unary rpc call
	err = stream.sendRequestMessage(in, true)
	if err != nil {
		return fmt.Errorf("an error occured when writing rpc headers: %s", err)
	}

	err = stream.receiveMessage(out)
	if err != nil {
		return fmt.Errorf("an error occured when receiving the server's answer: %s", err)
	}
	return nil
}

// Create a new stream in a thread safe manner and generate its ID
func (client *GrpcOverWrtcClient) newStream(ctx context.Context, method string) *Stream {
	id := atomic.AddUint64(&client.streamIDCounter, 1)
	ctx, cancelCtx := context.WithCancel(ctx)
	headersMetadata, _ := metadata.FromOutgoingContext(ctx)

	newStream := &Stream{
		ctx:             ctx,
		cancel:          cancelCtx,
		pendingMessages: make(chan []byte),
		proto:           &network_proto.Stream{Id: id},
		node:            client.GrpcOverWrtcNode,
		mutex:           sync.Mutex{},
		method:          method,
		header:          headersMetadata,
		trailer:         metadata.MD{},
	}

	client.storeStream(newStream)
	return newStream
}

// Transfer the incomming responses to the appropriate stream
func (client *GrpcOverWrtcClient) onMessage(msg webrtc.DataChannelMessage) {
	response := &network_proto.Response{}
	err := proto.Unmarshal(msg.Data, response)
	if err != nil {
		utils.Logger().Error(fmt.Sprintf("Invalid message received on client (%+v):\n\t%s",
			client,
			err,
		))
		return
	}
	streamInfo := response.GetStream()
	if streamInfo == nil {
		utils.Logger().Error(fmt.Sprintf(
			"Invalid stream data received (%+v) on client (%+v):\n\t%s",
			response,
			client,
			err,
		))
		return
	}

	// Check if the message is for an already existing stream
	stream, err := client.getStream(streamInfo)
	if err != nil {
		utils.Logger().Error(fmt.Sprintf(
			"Message received for unkown stream (%+v) on client (%+v):\n\t%s",
			streamInfo,
			client,
			err,
		))
		return
	}

	// Handle the incomming message according to its type
	switch r := response.Type.(type) {
	case *network_proto.Response_Headers:
		// TODO: Patch the stream headers
		return
	case *network_proto.Response_Trailers:
		// TODO: Patch the stream trailers
		return
	case *network_proto.Response_Message:
		stream.storeIncommingMessage(r.Message.PacketMessage.GetData())
	default:
		utils.Logger().Error(fmt.Sprintf(
			"Unkwnown message type (%+v) received on client (%+v):\n\t+%s",
			response,
			client,
			err,
		))
	}
}
