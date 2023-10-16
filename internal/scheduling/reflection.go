package scheduling

import (
	// "google.golang.org/grpc"
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ReflectionClient struct {
	stub grpc_reflection_v1.ServerReflectionClient
}

func (client *ReflectionClient) CallStream(descriptor protoreflect.MethodDescriptor) {
	// sd := grpc.StreamDesc{
	// 	StreamName:    string(descriptor.Name()),
	// 	ServerStreams: descriptor.IsStreamingServer(),
	// 	ClientStreams: descriptor.IsStreamingClient(),
	// }
}

func (client *ReflectionClient) ListServices() ([]*grpc_reflection_v1.ServiceResponse, error) {
	// The reflection service uses streams
	stream, err := client.stub.ServerReflectionInfo(context.Background())
	if err != nil {
		return nil, fmt.Errorf("an error occured while establishing stream connection with reflection: %w", err)
	}

	waitListResponse := make(chan *grpc_reflection_v1.ListServiceResponse)
	waitError := make(chan error)

	// Gather all the responses from the reflection service
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				waitError <- nil
			} else if err != nil {
				waitError <- err
			}
			// TODO: Handle errors

			switch response := in.MessageResponse.(type) {
			case *grpc_reflection_v1.ServerReflectionResponse_ListServicesResponse:
				waitListResponse <- response.ListServicesResponse
			default:
				waitError <- fmt.Errorf("an unexpected response type was reseived")
			}
		}
	}()

	// Send the request
	err = stream.Send(&grpc_reflection_v1.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1.ServerReflectionRequest_ListServices{},
	})
	if err != nil {
		return nil, fmt.Errorf("an error occured while querying services with reflection: %w", err)
	}
	// Make sure we inform the server that we close the stream right after
	defer stream.CloseSend()

	// Wait for the list services's response
	select {
	case serviceList := <-waitListResponse:
		return serviceList.Service, nil
	case err = <-waitError:
		if err != nil {
			return nil, fmt.Errorf("an error occured on the server side when listing services: %w", err)
		} else {
			return nil, fmt.Errorf("stream closed before reveiving any responses when listing services")
		}
	}
}

func NewReflectedClient(client grpc_reflection_v1.ServerReflectionClient) *ReflectionClient {
	return &ReflectionClient{
		stub: client,
	}
}
