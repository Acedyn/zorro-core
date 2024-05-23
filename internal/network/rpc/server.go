package rpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"
	"github.com/Acedyn/zorro-core/internal/utils"

	network_proto "github.com/Acedyn/zorro-proto/gen/protos/network"
	"github.com/pion/webrtc/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Create and setup a datachannel to allow the peer to act as a GRPC server
// with the other peer
func NewServer(wrtcConnection *wrtc.WRTCConnection, options ...GrpcOverWrtcOption) (*GrpcOverWrtcServer, error) {
	grpcOverWrtcServer := GrpcOverWrtcServer{
		handlers: map[string]serverHandlerFunc{},
		services: map[string]*serviceInfo{},
	}

	// Inject the client logic on every messages
	options = append(options, withDatachannelOptions(&wrtc.DatachannelOptions{
		OnMessage: func(msg webrtc.DataChannelMessage) {
			go grpcOverWrtcServer.onMessage(msg)
		},
	}))

	grpcOverWrtcNode, err := setupGrpcOverWrtc(wrtcConnection, options...)
	grpcOverWrtcServer.GrpcOverWrtcNode = grpcOverWrtcNode

	return &grpcOverWrtcServer, err
}

// from grpc.
type serviceInfo struct {
	methods  map[string]*grpc.MethodDesc
	streams  map[string]*grpc.StreamDesc
	metadata interface{}
}

type GrpcOverWrtcServer struct {
	*GrpcOverWrtcNode
	handlers      map[string]serverHandlerFunc
	handlersMutex sync.Mutex
	services      map[string]*serviceInfo
	servicesMutex sync.Mutex
}

// RegisterService registers the given implementation of a service to be handled via
// WebRTC data channels. It extracts the unary and stream methods from a service description
// and calls the methods on the implementation when requested via a data channel.
func (server *GrpcOverWrtcServer) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	info := &serviceInfo{
		methods:  make(map[string]*grpc.MethodDesc, len(sd.Methods)),
		streams:  make(map[string]*grpc.StreamDesc, len(sd.Streams)),
		metadata: sd.Metadata,
	}
	for i := range sd.Methods {
		d := &sd.Methods[i]
		info.methods[d.MethodName] = d
	}
	for i := range sd.Streams {
		d := &sd.Streams[i]
		info.streams[d.StreamName] = d
	}

	server.handlersMutex.Lock()
	defer server.handlersMutex.Unlock()
	server.servicesMutex.Lock()
	defer server.servicesMutex.Unlock()

	for i := range sd.Methods {
		desc := &sd.Methods[i]
		info.methods[desc.MethodName] = desc
		path := fmt.Sprintf("/%v/%v", sd.ServiceName, desc.MethodName)
		server.handlers[path] = unaryRequestHandler(ss, methodHandler(desc.Handler))
	}
	// TODO: Register streams too

	server.services[sd.ServiceName] = info
}

// Create a new stream in a thread safe manner and generate its ID
func (server *GrpcOverWrtcServer) newStream(id uint64, headers *network_proto.RequestHeaders) *Stream {
	// Build the stream's context
	headersMetadata := metadataFromProto(headers.GetMetadata())
	ctx := metadata.NewIncomingContext(server.ctx, headersMetadata)
	timeout := headers.Timeout.AsDuration()
	var cancelCtx func()
	if timeout == 0 {
		ctx, cancelCtx = context.WithCancel(ctx)
	} else {
		ctx, cancelCtx = context.WithTimeout(ctx, timeout)
	}

	newStream := &Stream{
		ctx:             ctx,
		cancel:          cancelCtx,
		pendingMessages: make(chan []byte),
		proto:           &network_proto.Stream{Id: id},
		node:            server.GrpcOverWrtcNode,
		mutex:           sync.Mutex{},
		method:          headers.GetMethod(),
		header:          headersMetadata,
		trailer:         metadata.MD{},
	}

	server.storeStream(newStream)
	return newStream
}

// Handle incomming requests
func (server *GrpcOverWrtcServer) onMessage(msg webrtc.DataChannelMessage) {
	request := &network_proto.Request{}
	err := proto.Unmarshal(msg.Data, request)
	if err != nil {
		server.logError("Invalid message received", err)
		return
	}
	streamInfo := request.GetStream()
	if streamInfo == nil {
		server.logError("Invalid stream data received (%+v)", err, request)
		return
	}

	// Check if the message is for an already existing stream
	stream, err := server.getStream(streamInfo)
	isNewStream := err != nil

	// Create the stream if it didn't already existed
	if isNewStream {
		// When creating a new stream, the first message should be a header
		headers, ok := request.Type.(*network_proto.Request_Headers)
		if !ok || headers.Headers == nil {
			server.logError(
				"Non header message received (%+v) for new stream", err, request,
			)
			return
		}

		id := streamInfo.GetId()
		stream = server.newStream(id, headers.Headers)
	}

	switch r := request.Type.(type) {
	case *network_proto.Request_Headers:
		if !isNewStream {
			server.logError(
				"Header message received (%+v) for existing stream", err, request,
			)
			return
		}

		// Get the handler function to initialize the procedure
		server.handlersMutex.Lock()
		handlerFunc, ok := server.handlers[r.Headers.Method]
		if !ok {
			server.logError("Unknown method %s", err, r.Headers.Method)
			server.handlersMutex.Unlock()
			return
		}
		server.handlersMutex.Unlock()

		// Call the handler
		err := handlerFunc(stream)
		if err != nil {
			server.logError(
				"Handler for method %s failed", err, r.Headers.Method,
			)
			return
		}
	case *network_proto.Request_Message:
		stream.storeIncommingMessage(r.Message.PacketMessage.GetData())
	case *network_proto.Request_RstStream:
		server.logError("Stream (%+v) canceled", err, stream)
		return
	default:
		server.logError("Unkwnown message type (%+v) received", err, request)
		return
	}
}

func (server *GrpcOverWrtcServer) logError(format string, err error, a ...any) {
	a = append(a, server, err)
	utils.Logger().Error(fmt.Sprintf(
		format+" on server (%+v):\n\t+%s",
		a...,
	))
}
