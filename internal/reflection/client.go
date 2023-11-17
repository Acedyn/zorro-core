package reflection

import (
	"context"
	"fmt"
	"strings"

	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_reflection "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Client that discovers the available gRPC method and invoke them
type ReflectionClient struct {
	reflectionStub grpc_reflection.ServerReflectionClient
	registry       *protoregistry.Files
	connection     grpc.ClientConnInterface
}

// Send a request to the reflection service
func (client *ReflectionClient) reflectionRequest(request *grpc_reflection.ServerReflectionRequest) (*grpc_reflection.ServerReflectionResponse, error) {
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
func (client *ReflectionClient) ListFileDescriptors() ([]*descriptorpb.FileDescriptorProto, error) {
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
			fileDescriptor := &descriptorpb.FileDescriptorProto{}
			err := proto.Unmarshal(rawFileDescriptor, fileDescriptor)
			if err != nil {
				return nil, fmt.Errorf("invalid proto file format at file descriptor for symbol %s: %w", serviceName.GetName(), err)
			}

			fileDescriptors[fileDescriptor.GetName()] = fileDescriptor
		}
	}

	return maps.Values(fileDescriptors), nil
}

func (client *ReflectionClient) registerFileDescriptors() error {
	fileList, err := client.ListFileDescriptors()
	if err != nil {
		return fmt.Errorf("could not get file descriptors: %w", err)
	}

	files, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{
		File: fileList,
	})

	if err != nil {
		return fmt.Errorf("coult not interpret file descriptors: %w", err)
	}

	client.registry = files
	return nil
}

// List the registered service descriptors
func (client *ReflectionClient) GetServiceDescriptors() []protoreflect.ServiceDescriptor {
	serviceDescriptors := []protoreflect.ServiceDescriptor{}
	client.registry.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
		for serviceIndex := 0; serviceIndex < fileDescriptor.Services().Len(); serviceIndex += 1 {
			serviceDescriptors = append(serviceDescriptors, fileDescriptor.Services().Get(serviceIndex))
		}

		return true
	})

	return serviceDescriptors
}

// Find a registered service via its full name (generally [package_name].[service_name] unless its nested)
func (client *ReflectionClient) GetMethodDescriptor(serviceName, methodName string) (protoreflect.MethodDescriptor, string, error) {
	// Get the service
	var serviceDescriptor protoreflect.ServiceDescriptor
	client.registry.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
		for serviceIndex := 0; serviceIndex < fileDescriptor.Services().Len(); serviceIndex += 1 {
			service := fileDescriptor.Services().Get(serviceIndex)
			if string(service.FullName()) == serviceName {
				serviceDescriptor = service
				return false
			}
		}
		return true
	})

	if serviceDescriptor == nil {
		return nil, "", fmt.Errorf("no service with name %s where found", serviceName)
	}

	// Get the method
	methodDescriptor := serviceDescriptor.Methods().ByName(protoreflect.Name(methodName))
	if methodDescriptor == nil {
		return nil, "", fmt.Errorf("no method with name %s where found in the service %s", methodName, serviceName)
	}

	// Build the full method path
	methodPath := fmt.Sprintf("/%s/%s", serviceDescriptor.FullName(), methodDescriptor.Name())
	return methodDescriptor, methodPath, nil
}

// Fetch the file descriptors for each service present
func (client *ReflectionClient) InvokeRpcServerStream(method protoreflect.MethodDescriptor, methodPath string, input any) (grpc.ClientStream, error) {
	streamDescriptor := grpc.StreamDesc{
		StreamName:    string(method.Name()),
		ServerStreams: method.IsStreamingServer(),
		ClientStreams: method.IsStreamingClient(),
	}

	// Prepare the stream
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.connection.NewStream(ctx, &streamDescriptor, methodPath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not create stream with request %s: %w", methodPath, err)
	}

	// When the new stream is finished, also cleanup the parent context
	go func() {
		<-stream.Context().Done()
		cancel()
	}()

	// Send the first message and close since this method is for server side streaming
	err = stream.SendMsg(input)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not send message %s: %w", input, err)
	}
	err = stream.CloseSend()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not close send on stream %s: %w", stream, err)
	}

	return stream, nil
}

// Create a client that wil fetch all the available methods and offer and interface to call them
func NewReflectedClient(host string) (*ReflectionClient, error) {
	// Establish the grpc connection with the new processor
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	connection, err := grpc.Dial(host, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create connection with processor at host %s: %w", host, err)
	}
	client := grpc_reflection.NewServerReflectionClient(connection)

	// Create the reflected client
	reflectedClient := &ReflectionClient{
		reflectionStub: client,
		connection:     connection,
	}
	// Fetch all the available methods
	err = reflectedClient.registerFileDescriptors()
	if err != nil {
		return nil, fmt.Errorf("could not initialize reflected client: %w", err)
	}

	return reflectedClient, nil
}
