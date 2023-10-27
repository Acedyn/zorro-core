package scheduling

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/tools"

	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

var (
	processorPoolLock = &sync.Mutex{}
	processorPool     map[string]*RegisteredProcessor
	once              sync.Once
)

// Getter for the clients pool singleton
func ProcessorPool() map[string]*RegisteredProcessor {
	once.Do(func() {
		processorPool = map[string]*RegisteredProcessor{}
	})

	return processorPool
}

// Registered processors are ready to receive command requests
type RegisteredProcessor struct {
	*processor.Processor
	// The host to connect to send commands
	Host string
	// Commands waiting to be scheduled
	commandQueue chan *tools.Command
	// Commands scheduled and still running on the client side
	runningCommands     map[string]*tools.Command
	runningCommandsLock *sync.Mutex
	// The client used to send command requests
	Client *ReflectionClient
}

// Send a grpc query to the processor to execute the command request
func (processor *RegisteredProcessor) ProcessCommand(commandQuery tools.CommandQuery) error {
	// Get the method descriptor that correspond to the command request
	methodDescriptor, methodPath, err := processor.Client.GetdMethodDescriptor(commandQuery.Command.Base.GetName(), string(commandQuery.ExecutionType))
	if err != nil {
		return fmt.Errorf("could not find method with processor at host %s: %w", processor.Host, err)
	}

	// Build the input message using the command's input sockets
	missingKeys := []string{}
	inputMessage := dynamicpb.NewMessage(methodDescriptor.Input())
	inputMessage.Range(func(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if socket, ok := commandQuery.Command.Base.Inputs[fieldDescriptor.TextName()]; ok {
			tools_proto.Socket
			switch raw := socket.Raw.(type) {
			case *tools_proto.Socket_RawBinary:
				inputMessage.Set(fieldDescriptor, protoreflect.ValueOfBytes(raw.RawBinary))
			case *tools_proto.Socket_RawString:
				inputMessage.Set(fieldDescriptor, protoreflect.ValueOfString(raw.RawString))
			case *tools_proto.Socket_RawInteger:
				inputMessage.Set(fieldDescriptor, protoreflect.ValueOfInt32(raw.RawInteger))
			case *tools_proto.Socket_RawNumber:
				// TODO: I don't remember why tf I named it RawNumber, rename it to RawFloat
				inputMessage.Set(fieldDescriptor, protoreflect.ValueOfFloat32(raw.RawNumber))
			}
		} else {
			missingKeys = append(missingKeys, fieldDescriptor.TextName())
			return false
		}
		return true
	})

	if len(missingKeys) > 0 {
		return fmt.Errorf("missing input values for method %s: %s", methodDescriptor.FullName(), missingKeys)
	}

	// Start the stream and send the input message
	stream, err := processor.Client.InvokeRpcServerStream(methodDescriptor, methodPath, inputMessage)
	if err != nil {
		return fmt.Errorf("an error occured when invoking method with processor at host %s: %w", processor.Host, err)
	}

	for {
		outputMessage := dynamicpb.NewMessage(methodDescriptor.Output())
		err = stream.RecvMsg(outputMessage)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("an error occured when receiving response by processor at host %s: %w", processor.Host, err)
		}

		// TODO: The processing of the method returned value should be in a different package
		outputMessage.Range(func(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
			kind := FormatFieldDescriptorKind(fieldDescriptor)
			commandQuery.Command.Base.Outputs[fieldDescriptor.TextName()] = &tools_proto.Socket{
				// TODO: Get the raw value
				Cast: kind,
			}
			return true
		})
	}

	return nil
}

// Register the given client to the client pool
func registerProcessor(processorToRegister *processor.Processor, host string, client *ReflectionClient) *RegisteredProcessor {
	// Check if the client is already registered
	processorPoolLock.Lock()
	defer processorPoolLock.Unlock()
	registeredProcessor, ok := ProcessorPool()[processorToRegister.GetId()]
	if !ok {
		registeredProcessor = &RegisteredProcessor{
			Processor:           processorToRegister,
			Host:                host,
			commandQueue:        make(chan *tools.Command),
			runningCommands:     map[string]*tools.Command{},
			runningCommandsLock: &sync.Mutex{},
			Client:              client,
		}
		ProcessorPool()[processorToRegister.GetId()] = registeredProcessor
	}

	// If the processor was in the processor queue, inform that the registration is done
	if pendingProcessor := processor.UnQueueProcessor(processorToRegister.GetId()); pendingProcessor != nil {
		pendingProcessor.Registration <- nil
	}
	return registeredProcessor
}

// Look among the already registered clients and return the first matching client
func findRegisteredProcessor(query *ProcessorQuery) *RegisteredProcessor {
	processorPoolLock.Lock()
	defer processorPoolLock.Unlock()

	// The look by id is faster since its the primary key
	if query.Id != nil {
		return ProcessorPool()[*query.Id]
	}

	// Test all the registered clients one by one
	for _, registeredClient := range ProcessorPool() {
		if query.MatchProcessor(registeredClient.Processor) {
			return registeredClient
		}
	}

	return nil
}

// Get an already running processor or start a new one from the query
func GetOrStartProcessor(query *ProcessorQuery) (*RegisteredProcessor, error) {
	// First find a potential running processors that matches the query
	if registeredClient := findRegisteredProcessor(query); registeredClient != nil {
		return registeredClient, nil
	}

	// If no running processors matches the query, try to start a new one
	for _, availableProcessor := range query.GetContext().AvailableProcessors() {
		if availableProcessor.GetName() == query.GetName() {
			pendingProcessor, err := availableProcessor.Start(query.GetMetadata(), query.GetContext().Environ(true), query.GetContext().AvailableCommandPaths(availableProcessor))
			if err != nil {
				return nil, fmt.Errorf("could not start new processor (%s): %w", availableProcessor, err)
			}
			// The client should now be registered
			registeredProcessor := findRegisteredProcessor(&ProcessorQuery{
				ProcessorQuery: &scheduling_proto.ProcessorQuery{
					Id: &pendingProcessor.Id,
				},
			})
			if registeredProcessor == nil {
				return nil, fmt.Errorf("processor %s started but did not registered", pendingProcessor.Id)
			}
			return registeredProcessor, nil
		}
	}

	return nil, fmt.Errorf(
		"could not find running or run processor to satisfy the query %s",
		query,
	)
}
