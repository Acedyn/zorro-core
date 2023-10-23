package scheduling

import (
	"context"
	"fmt"
	"strings"

	"github.com/life4/genesis/slices"
	"google.golang.org/grpc"
	grpc_reflection "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type FullServiceDescriptor struct {
  ServiceDescriptor *descriptorpb.ServiceDescriptorProto
  Methods []FullMethodDescriptor
}

type FullMethodDescriptor struct {
  MethodDescriptor *descriptorpb.MethodDescriptorProto
  Input *descriptorpb.DescriptorProto
  Output *descriptorpb.DescriptorProto
}

type ReflectionClient struct {
	reflectionStub grpc_reflection.ServerReflectionClient
  services []*FullServiceDescriptor
}

// Recursive function to gather all the descriptors of a message type
func gatherEmbededMessageDescriptors(messageDescriptors map[string]*descriptorpb.DescriptorProto, messageDescriptor *descriptorpb.DescriptorProto) {
  messageDescriptors[messageDescriptor.GetName()] = messageDescriptor
  for _, embedMessage := range messageDescriptor.GetNestedType() {
    gatherEmbededMessageDescriptors(messageDescriptors, embedMessage)
  }
}

func (client *ReflectionClient) CallStream(descriptor protoreflect.MethodDescriptor) {
	// sd := grpc.StreamDesc{
	// 	StreamName:    string(descriptor.Name()),
	// 	ServerStreams: descriptor.IsStreamingServer(),
	// 	ClientStreams: descriptor.IsStreamingClient(),
	// }
}

// Send a request to the reflection service
func (client *ReflectionClient) reflectionRequest(request *grpc_reflection.ServerReflectionRequest) (*grpc_reflection.ServerReflectionResponse, error){
	// The reflection service uses streams, is still don't understand why. For simplicity we will
  // create a stream per request
	stream, err := client.reflectionStub.ServerReflectionInfo(context.Background())
	if err != nil {
		return nil, fmt.Errorf("an error occured while establishing stream connection with reflection: %w", err)
	}

	err = stream.Send(request)
	if err != nil {
		return nil, fmt.Errorf("an error occured while querying services with reflection: %w", err)
	}

	response, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("an error occured when receiving response from reflection service: %w", err)
	}

  return response, nil
}

// Fetch a list of service names that are available
func (client *ReflectionClient) ListServices() ([]*grpc_reflection.ServiceResponse, error) {
  response, err := client.reflectionRequest(&grpc_reflection.ServerReflectionRequest{
		MessageRequest: &grpc_reflection.ServerReflectionRequest_ListServices{},
	})

	if err != nil {
		return nil, fmt.Errorf("could not list services: %w", err)
	}

  listServiceResponse := response.GetListServicesResponse()
  if listServiceResponse == nil {
		return nil, fmt.Errorf("invalid response received to list service request (%s)", response)
  }

  // Remove the reflection service from the list
  services := slices.Filter(listServiceResponse.Service, func(i *grpc_reflection.ServiceResponse) bool {
    return strings.Split(strings.Trim(grpc_reflection.ServerReflection_ServerReflectionInfo_FullMethodName, "/"), "/")[0] != i.GetName()
  })

  return services, nil
}

// Fetch the file descriptors for each service present
func (client *ReflectionClient) ListFileDescriptors() (map[string]*descriptorpb.FileDescriptorProto, error) {
  serviceList, err := client.ListServices()

	if err != nil {
		return nil, fmt.Errorf("could not get service names: %w", err)
	}

  fileDescriptors := map[string]*descriptorpb.FileDescriptorProto{}

  for _, serviceName := range serviceList {
    response, err := client.reflectionRequest(&grpc_reflection.ServerReflectionRequest{
		MessageRequest: &grpc_reflection.ServerReflectionRequest_FileContainingSymbol{
        FileContainingSymbol: serviceName.GetName(),
      },
	  })

    if err != nil {
      return nil, fmt.Errorf("could not get file descriptor for symbol %s: %w", serviceName.GetName(), err)
    }

    fileDescriptorResponse := response.GetFileDescriptorResponse()
    if fileDescriptorResponse == nil {
		  return nil, fmt.Errorf("invalid response received to file descriptor request for symbol %s (%s)", serviceName.GetName(), response)
    }

    for _, rawFileDescriptor := range fileDescriptorResponse.FileDescriptorProto {
      fileDescritor := &descriptorpb.FileDescriptorProto{}
      err := proto.Unmarshal(rawFileDescriptor, fileDescritor)
      if err != nil {
        return nil, fmt.Errorf("invalid proto file format at file descriptor for symbol %s: %w", serviceName.GetName(), err)
      }

      fileDescriptors[fileDescritor.GetName()] = fileDescritor
    }
  }

  return fileDescriptors, nil
}

func (client *ReflectionClient) fetchFullServiceDescriptors() error {
  fileList, err := client.ListFileDescriptors()
	if err != nil {
		return fmt.Errorf("could not get file descriptors: %w", err)
	}

  fullServiceDescriptors := []*FullServiceDescriptor{}

  // First find the service descriptor and method descriptors
  serviceDescriptors := map[string]*descriptorpb.ServiceDescriptorProto{}
  messageDescriptors := map[string]*descriptorpb.DescriptorProto{}
  for _, fileDescriptor := range fileList {
    for _, serviceDescriptor := range fileDescriptor.GetService() {
      serviceDescriptors[serviceDescriptor.GetName()] = serviceDescriptor
    }
    for _, messageDescriptor := range fileDescriptor.GetMessageType() {
      gatherEmbededMessageDescriptors(messageDescriptors, messageDescriptor)
    }
  }

  // Create the full service descriptors
  for _, serviceDescriptor := range serviceDescriptors {
    fullServiceDescriptor := FullServiceDescriptor{
      ServiceDescriptor: serviceDescriptor,
    }
    for _, methodDescriptor := range serviceDescriptor.Method {
      fullMethodDescriptor := FullMethodDescriptor{
        MethodDescriptor: methodDescriptor,
      }
      // Find the concrete input and output
      inputType, inputOk := messageDescriptors[methodDescriptor.GetInputType()]
      outputType, outputOk := messageDescriptors[methodDescriptor.GetOutputType()]

      if !inputOk {
        return fmt.Errorf("missing message descriptors for input with name %s", methodDescriptor.GetInputType())
      }
      if !outputOk {
        return fmt.Errorf("missing message descriptors for output with name %s", methodDescriptor.GetOutputType())
      }
      fullMethodDescriptor.Input = inputType
      fullMethodDescriptor.Output = outputType

      fullServiceDescriptor.Methods = append(fullServiceDescriptor.Methods, fullMethodDescriptor)
    }

    fullServiceDescriptors = append(fullServiceDescriptors, &fullServiceDescriptor)
  }

  client.services = fullServiceDescriptors
  return nil
}

// Fetch the file descriptors for each service present
func (client *ReflectionClient) InvokeRpcServerStream(channel grpc.ClientConnInterface) {

}

func NewReflectedClient(client grpc_reflection.ServerReflectionClient) (*ReflectionClient, error) {
  reflectedClient := &ReflectionClient{
		reflectionStub: client,
	}
  err := reflectedClient.fetchFullServiceDescriptors()
  if err != nil {
    return nil, fmt.Errorf("Could not initialize reflected client: %w", err)
  }

  return reflectedClient, nil
}
