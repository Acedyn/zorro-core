package scheduling

import (
	"context"
	"fmt"
	"strings"

	"github.com/life4/genesis/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	grpc_reflection "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type ReflectionClient struct {
	reflectionStub grpc_reflection.ServerReflectionClient
	registry       *protoregistry.Files
	connection     grpc.ClientConnInterface
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

	fileDescriptors := []*descriptorpb.FileDescriptorProto{}

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

			fileDescriptors = append(fileDescriptors, fileDescriptor)
		}
	}

	return fileDescriptors, nil
}

func (client *ReflectionClient) fetchFullServiceDescriptors() error {
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

// Fetch the file descriptors for each service present
func (client *ReflectionClient) InvokeRpcServerStream(serviceName, methodName string) (map[string]protoreflect.Value, error) {
	// Get the service
	var serviceDescriptor protoreflect.ServiceDescriptor
	for _, service := range client.GetServiceDescriptors() {
		if string(service.FullName()) == serviceName {
			serviceDescriptor = service
		}
	}

	if serviceDescriptor == nil {
		return nil, fmt.Errorf("no service with name %s where found", serviceName)
	}

	// Get the method
	methodDescriptor := serviceDescriptor.Methods().ByName(protoreflect.Name(methodName))
	if methodDescriptor == nil {
		return nil, fmt.Errorf("no method with name %s where found in the service %s", methodName, serviceName)
	}

	streamDescriptor := grpc.StreamDesc{
		StreamName:    string(methodDescriptor.Name()),
		ServerStreams: methodDescriptor.IsStreamingServer(),
		ClientStreams: methodDescriptor.IsStreamingClient(),
	}

	// Prepare the stream
	ctx, cancel := context.WithCancel(context.Background())
	requestName := fmt.Sprintf("/%s/%s", serviceDescriptor.FullName(), methodDescriptor.Name())
	stream, err := client.connection.NewStream(ctx, &streamDescriptor, requestName)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not create stream with request %s: %w", requestName, err)
	}

	// When the new stream is finished, also cleanup the parent context
	go func() {
		<-stream.Context().Done()
		cancel()
	}()

	// Build the dynamic message and send the first message
	inputMessage := dynamicpb.NewMessage(methodDescriptor.Input())
	outputMessage := dynamicpb.NewMessage(methodDescriptor.Output())
	err = stream.SendMsg(inputMessage)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not send message %s: %w", inputMessage, err)
	}
	err = stream.CloseSend()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not close send on stream %s: %w", stream, err)
	}

	err = stream.RecvMsg(outputMessage)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not receive message %s: %w", outputMessage, err)
	}

	formattedOutput := map[string]protoreflect.Value{}
	outputMessage.Range(func(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		formattedOutput[fieldDescriptor.TextName()] = value
		return true
	})
	return formattedOutput, nil
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
	client := grpc_reflection_v1alpha.NewServerReflectionClient(connection)

	// Create the reflected client
	reflectedClient := &ReflectionClient{
		reflectionStub: client,
		connection:     connection,
	}
	// Fetch all the available methods
	err = reflectedClient.fetchFullServiceDescriptors()
	if err != nil {
		return nil, fmt.Errorf("Could not initialize reflected client: %w", err)
	}

	return reflectedClient, nil
}
