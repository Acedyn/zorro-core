package rpc

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"
	"github.com/Acedyn/zorro-core/internal/utils"

	network_proto "github.com/Acedyn/zorro-proto/gen/protos/network"
	"github.com/pion/webrtc/v4"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	stream, err := client.newStream(ctx, &grpc.StreamDesc{}, method, options...)
	defer client.removeStream(stream)
	if err != nil {
		return fmt.Errorf("an error occured when creating stream: %w", err)
	}

	// Send the message containing the rpc input arg and set the end of stream directly
	// since this is an unary rpc call
	err = stream.sendRequestMessage(in, true)
	if err != nil {
		return fmt.Errorf("an error occured when writing rpc headers: %w", err)
	}

	// Wait for the output message
	err = stream.RecvMsg(out)
	if err != nil {
		return fmt.Errorf("an error occured when receiving the server's answer: %w", err)
	}
	return nil
}

// Create a new stream in a thread safe manner and generate its ID
func (client *GrpcOverWrtcClient) newStream(
	ctx context.Context, desc *grpc.StreamDesc, method string, options ...grpc.CallOption,
) (*Stream, error) {
	id := atomic.AddUint64(&client.streamIDCounter, 1)
	ctx, cancelCtx := context.WithCancelCause(ctx)
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

	// Send the header first, we are counting on the datachannel to make sure the
	// headers is received before the actuall message
	err := newStream.sendRequestHeaders(ctx, method)
	if err != nil {
		return nil, fmt.Errorf("an error occured when writing rpc headers: %w", err)
	}

	return newStream, nil
}

// Method to satisfy the grpc.ClientConnInterface interface
func (client *GrpcOverWrtcClient) NewStream(
	ctx context.Context, desc *grpc.StreamDesc, method string, options ...grpc.CallOption,
) (grpc.ClientStream, error) {
	stream, err := client.newStream(ctx, desc, method, options...)
	return &ClientStream{Stream: stream}, err
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
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.trailer = metadataFromProto(r.Trailers.Metadata)

		// TODO: Find a better way to convert from network_proto.Status to grpc.status.Status
		statusRaw, err := proto.Marshal(r.Trailers.Status)
		if err != nil {
			utils.Logger().Error("ERROR")
		}

		statusProto := &spb.Status{}
		err = proto.Unmarshal(statusRaw, statusProto)
		if err != nil {
			utils.Logger().Error("ERROR")
		}
		trailerStatus := status.FromProto(statusProto)

		if trailerStatus.Err() == nil {
			stream.cancel(io.EOF)
		} else {
			stream.cancel(trailerStatus.Err())
		}
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
