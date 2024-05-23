package rpc

import (
	"context"
	"fmt"
	"sync"

	network_proto "github.com/Acedyn/zorro-proto/gen/protos/network"
	"github.com/pion/webrtc/v4"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Base struct for server and clients
type GrpcOverWrtcNode struct {
	datachannel *webrtc.DataChannel
	streams     map[uint64]*Stream
	ctx         context.Context
	streamMutex sync.Mutex
}

// Create a new stream in a thread safe manner
func (node *GrpcOverWrtcNode) storeStream(stream *Stream) {
	node.streamMutex.Lock()
	defer node.streamMutex.Unlock()
	node.streams[stream.proto.GetId()] = stream
}

// Remove a stream in a thread safe manner
func (node *GrpcOverWrtcNode) removeStream(stream *Stream) {
	node.streamMutex.Lock()
	defer node.streamMutex.Unlock()
	delete(node.streams, stream.proto.GetId())
}

// Get a new stream in a thread safe manner
func (node *GrpcOverWrtcNode) getStream(streamProto *network_proto.Stream) (*Stream, error) {
	node.streamMutex.Lock()
	defer node.streamMutex.Unlock()
	stream, ok := node.streams[streamProto.GetId()]
	if !ok {
		return stream, fmt.Errorf("no stream with id %s found in grpc over wrtc node", streamProto.GetId())
	}
	return stream, nil
}

// build a message frame in protobuf format
func (node *GrpcOverWrtcNode) buildMessage(stream Stream, in any, endOfStream bool) ([]byte, error) {
	var err error
	var data []byte

	// Marshall the input argument of the rpc method
	if in != nil {
		data, err = proto.Marshal(in.(proto.Message))
		if err != nil {
			return nil, fmt.Errorf("message %v cannot be marshalled: %w", err)
		}
	}

	return data, nil
}

// Send a message frame to the other peer after marshalling the message and
// also splitting the data into chunks if needed
func (node *GrpcOverWrtcNode) sendFrame(msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not send message, proto marshaling failed: %w", err)
	}

	// TODO: Split the data into chunks if it exeeds the threshold of 1024 * 1024 bytes

	err = node.datachannel.Send(data)
	if err != nil {
		return fmt.Errorf("could not send message, proto marshaling failed: %w", err)
	}
	return nil
}

// Convert grpc metadata to protobuf metadata so it can be serialized and sent
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
