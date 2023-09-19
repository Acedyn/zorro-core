package scheduling

import (
	"fmt"
	"io"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/tools"
)

type schedulingServer struct {
	UnimplementedSchedulingServer
}

// Handle update received from the client
func receiveUpdate(registedClient *RegisteredClient, stream Scheduling_CommunicateCommandsServer, stop chan error) error {
	// Loop until we shoud stop listen for updates
	for range stop {
		in, err := stream.Recv()
		// The communication is closed
		if err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("an error occured while receiving message: %w", err)
		}
		response := &ClientCommunication{}

		// Handle client update
		if in.ClientUpdate != nil {
			if registedClient.Client.Patch(in.ClientUpdate) {
				response.ClientUpdate = registedClient.Client
			}
		}

		// Handle command updates
		if in.CommandUpdate != nil {
			registedClient.ScheduledCommandsLock.Lock()
			command, ok := registedClient.ScheduledCommands[in.CommandUpdate.Base.GetId()]
			registedClient.ScheduledCommandsLock.Unlock()

			if !ok {
				return fmt.Errorf(
					"no command with id %s exists under client %s",
					in.CommandUpdate.Base.GetId(),
					registedClient.Client.GetId(),
				)
			}

			if command.Patch(in.CommandUpdate) {
				response.CommandUpdate = command
			}
		}

		err = stream.Send(response)
		if err != nil {
			return fmt.Errorf("an error occured while sending message: %w", err)
		}
	}

	return nil
}

// Send update to the connected client
func sendUpdates(registedClient *RegisteredClient, stream Scheduling_CommunicateCommandsServer, stop chan error) error {
	// Loop until we shoud stop sending updates
	for range stop {
		select {
		// Send the new command queries
		case command := <-registedClient.CommandQueue:
			// Send the command to the client
			err := stream.Send(&ClientCommunication{CommandUpdate: command})
			if err != nil {
				return fmt.Errorf("an error occured while sending message: %w", err)
			}

			// Store the client in the running command list
			registedClient.ScheduledCommandsLock.Lock()
			registedClient.ScheduledCommands[command.Base.GetId()] = command
			registedClient.ScheduledCommandsLock.Unlock()
		case <-stop:
			break
		}
	}

	return nil
}

// Exchange updates about the running commands, context with clients
func (server *schedulingServer) Communicate(stream Scheduling_CommunicateCommandsServer) error {
	// Start with registering the client
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("an error occured while receiving message: %w", err)
	} else if in.ClientUpdate == nil {
		return fmt.Errorf("no client update received: all client communications should start with client registration")
	}
	registeredClient := registerClient(in.ClientUpdate)

	// Send an receive updates
	error := make(chan error)
	go func() {
		error <- receiveUpdate(registeredClient, stream, error)
	}()
	go func() {
		error <- sendUpdates(registeredClient, stream, error)
	}()

	// Wait for any goroutine to exit
	err = <-error
	// Make sure to stop both the coroutines
	close(error)
	return err
}

// Wait for command requests and dispach then to the corresponding registered client
func listenCommandRequests() {
	for commandHandle := range tools.CommandQueue() {
		registeredClient, err := ClientFromQuery(commandHandle.Context, commandHandle.Command.GetClientQuery())
		if err != nil {
			// The error should be handled by the command submitter
			commandHandle.Result <- fmt.Errorf(
				"could not get or create client from query %s in the selected context: %w",
				commandHandle.Command.GetClientQuery(),
				err,
			)
		} else {
			// The command will then be sent to the client
			go func(commandHandle *tools.CommandHandle) {
				registeredClient.CommandQueue <- commandHandle.Command
			}(commandHandle)
		}
	}
}

func init() {
	RegisterSchedulingServer(network.GrpcServer(), &schedulingServer{})
	go listenCommandRequests()
}
