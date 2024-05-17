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
func SetupGrpcOverWrtcServer(wrtcConnection *wrtc.WRTCConnection, options ...GrpcOverWrtcOption) (GrpcOverWrtcServer, error) {
	grpcOverWrtcOption := DefaultGrpcOverWrtcOptions
	for _, option := range options {
		option(&grpcOverWrtcOption)
	}

	grpcOverWrtcServer := GrpcOverWrtcServer{
		handlers: map[string]serverHandlerFunc{},
		services: map[string]*serviceInfo{},
		ctx:      grpcOverWrtcOption.ctx,
	}
	// Wait for the datachannel to be ready before returning
	var wg sync.WaitGroup
	wg.Add(1)

	// Inject the server logic on every messages
	onServerMessage := func(msg webrtc.DataChannelMessage) {
		// Make sure the connection is initialized before handeling any message
		wg.Wait()
		go grpcOverWrtcServer.onMessage(msg)
	}

	grpcOverWrtcBase, err := setupGrpcOverWrtc[*ServerStream](wrtcConnection, onServerMessage, options...)
	grpcOverWrtcServer.GrpcOverWrtcBase = grpcOverWrtcBase
	wg.Done()

	return grpcOverWrtcServer, err
}

// from grpc.
type serviceInfo struct {
	methods  map[string]*grpc.MethodDesc
	streams  map[string]*grpc.StreamDesc
	metadata interface{}
}

type GrpcOverWrtcServer struct {
	*GrpcOverWrtcBase[*ServerStream]
	ctx      context.Context
	handlers map[string]serverHandlerFunc
	services map[string]*serviceInfo
}

// RegisterService registers the given implementation of a service to be handled via
// WebRTC data channels. It extracts the unary and stream methods from a service description
// and calls the methods on the implementation when requested via a data channel.
func (server *GrpcOverWrtcServer) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	// TODO: Make this method threadsafe for handlers and services access
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

	for i := range sd.Methods {
		desc := &sd.Methods[i]
		info.methods[desc.MethodName] = desc
		path := fmt.Sprintf("/%v/%v", sd.ServiceName, desc.MethodName)
		server.handlers[path] = serverUnaryHandler(ss, methodHandler(desc.Handler))
	}
	// TODO: Register streams too

	server.services[sd.ServiceName] = info
}

// Create a new stream in a thread safe manner and generate its ID
func (server *GrpcOverWrtcServer) newStream(id uint64, headers *network_proto.RequestHeaders) *ServerStream {
	// Build the stream's context
	headersMetadata := metadataFromProto(headers.GetMetadata())
	streamContext := metadata.NewIncomingContext(server.ctx, headersMetadata)
	timeout := headers.Timeout.AsDuration()
	var cancelCtx func()
	if timeout == 0 {
		streamContext, cancelCtx = context.WithCancel(streamContext)
	} else {
		streamContext, cancelCtx = context.WithTimeout(streamContext, timeout)
	}

	newStream := &ServerStream{
		Stream: &network_proto.Stream{
			Id: id,
		},
		server:      server,
		mutex:       sync.Mutex{},
		ctx:         streamContext,
		cancel:      cancelCtx,
		method:      headers.GetMethod(),
		header:      headersMetadata,
		trailer:     metadata.MD{},
		messageChan: make(chan []byte),
	}

	server.storeStream(newStream)
	return newStream
}

// Handle incomming requests
func (server *GrpcOverWrtcServer) onMessage(msg webrtc.DataChannelMessage) {
	request := &network_proto.Request{}
	err := proto.Unmarshal(msg.Data, request)
	if err != nil {
		utils.Logger().Error(fmt.Sprintf("Invalid message received on server (%+v):\n\t%s",
			server,
			err,
		))
		return
	}
	streamInfo := request.GetStream()
	if streamInfo == nil {
		utils.Logger().Error(fmt.Sprintf(
			"Invalid stream data received (%+v) on server (%+v):\n\t%s",
			request,
			server,
			err,
		))
		return
	}

	// Check if the message is for an already existing stream
	id := streamInfo.GetId()
	server.streamMutex.Lock()
	stream, ok := server.streams[id]
	server.streamMutex.Unlock()
	isNewStream := false

	// Create the stream if it didn't already existed
	if !ok {
		// When creating a new stream, the first message should be a header
		headers, ok := request.Type.(*network_proto.Request_Headers)
		if !ok || headers.Headers == nil {
			utils.Logger().Error(fmt.Sprintf(
				"Non header message received (%+v) for unexisting stream on server (%+v):\n\t%s",
				request,
				server,
				err,
			))
			return
		}

		stream = server.newStream(id, headers.Headers)
		isNewStream = true
	}

	switch r := request.Type.(type) {
	case *network_proto.Request_Headers:
		if !isNewStream {
			utils.Logger().Error(fmt.Sprintf(
				"Header message received (%+v) for already existing stream on server (%+v):\n\t%s",
				request,
				server,
				err,
			))
			return
		}
		// TODO: make the access of the handlers field thread safe
		handlerFunc, ok := server.handlers[r.Headers.Method]
		if !ok {
			utils.Logger().Error("Unknown method %s on server (%s):\n\t%s",
				r.Headers.Method,
				fmt.Sprintf("%+v", server),
				err,
			)
		}

		err := handlerFunc(stream)
		if err != nil {
			utils.Logger().Error(fmt.Sprintf(
				"An error occured when creating handler for method %s on server (%+v):\n\t%s",
				r.Headers.Method,
				server,
				err,
			))
		}
	case *network_proto.Request_Message:
		stream.messageChan <- r.Message.PacketMessage.Data
	case *network_proto.Request_RstStream:
		utils.Logger().Error(fmt.Sprintf(
			"Stream (%+v) canceled on server (%+v):\n\t%s",
			stream,
			server,
			err,
		))
		return
	default:
		utils.Logger().Error(fmt.Sprintf(
			"Unkwnown message type (%+v) received on server (%+v):\n\t+%s",
			request,
			server,
			err,
		))
	}
}

// Logic and data related to a stream on the server side
type ServerStream struct {
	*network_proto.Stream
	server      *GrpcOverWrtcServer
	mutex       sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	method      string
	header      metadata.MD
	trailer     metadata.MD
	messageChan chan []byte
}

func (stream *ServerStream) GetStream() *network_proto.Stream {
	return stream.Stream
}

// Wait for the next message and apply the result on the given proto
func (stream *ServerStream) receiveMessage(m any) error {
	select {
	case <-stream.ctx.Done():
		return fmt.Errorf("message reception canceled: %w", stream.ctx.Err())
	case msgBytes := <-stream.messageChan:
		return proto.Unmarshal(msgBytes, m.(proto.Message))
	}
}

// Send the response for a message, it can be a stream or an unary request
func (stream *ServerStream) sendMessage(m any) error {
	return stream.server.writeMessage(stream, m, false)
}

// Getter for the method name
func (stream *ServerStream) Method() string {
	return stream.method
}

// SetHeader sets the header metadata. It may be called multiple times.
// When call multiple times, all the provided metadata will be merged.
func (stream *ServerStream) SetHeader(header metadata.MD) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if stream.header == nil {
		stream.header = header
	} else if header != nil {
		stream.header = metadata.Join(stream.header, header)
	}

	return nil
}

// SendHeader sends the header metadata.
// The provided md and headers set by SetHeader() will be sent.
func (stream *ServerStream) SendHeader(header metadata.MD) error {
	if err := stream.SetHeader(header); err != nil {
		return err
	}
	return stream.writeHeaders()
}

// SetTrailer sets the trailer metadata which will be sent with the RPC status.
// When called more than once, all the provided metadata will be merged.
func (stream *ServerStream) SetTrailer(trailer metadata.MD) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if stream.trailer == nil {
		stream.trailer = trailer
	} else if trailer != nil {
		stream.trailer = metadata.Join(stream.trailer, trailer)
	}

	return nil
}

func (stream *ServerStream) writeHeaders() error {
	fmt.Println("SENDING HEADERS")
	return nil
	// protoHeaders := metadataToProto(stream.header)
	// return s.ch.writeHeaders(s.stream, &network_proto.RequestHeaders{
	// 	Metadata: protoHeaders,
	// })
}
