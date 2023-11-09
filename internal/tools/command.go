package tools

import (
	"fmt"
	"sync"

	"github.com/Acedyn/zorro-core/internal/context"

	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	commandQueue chan *CommandQuery
	once         sync.Once
)

type CommandExecutionType string

var EXECUTE_COMMAND CommandExecutionType = "Execute"
var UNDO_COMMAND CommandExecutionType = "Undo"
var TEST_COMMAND CommandExecutionType = "Test"

var COMMAND_PATCH_OUTPUT string = "zorro_command"
var CONTEXT_PATCH_OUTPUT string = "zorro_context"

// Wrapped command with methods attached
type Command struct {
	*tools_proto.Command
}

// Get the wrapped base with all its methods
func (command *Command) GetBase() *ToolBase {
	return &ToolBase{ToolBase: command.Command.GetBase()}
}

// Started command, waiting to be sheduled
type CommandQuery struct {
	Command       *Command
	ExecutionType CommandExecutionType
	Result        chan error
	Context       *context.Context
}

// Getter for the commands queue singleton which holds the queue
// of command waiting to be scheduled
func CommandQueue() chan *CommandQuery {
	once.Do(func() {
		commandQueue = make(chan *CommandQuery)
	})

	return commandQueue
}

// Commands are the smallest tool type, they don't have children
func (command *Command) Traverse(task func(TraversableTool) error) error {
	if err := task(command); err != nil {
		return fmt.Errorf("error occured while traversing command %s: %w", command.GetBase().GetName(), err)
	}

	return nil
}

// The execution of the commands is handled by the scheduler, and processed by the clients
func (command *Command) execute(c *context.Context, executionType CommandExecutionType) chan error {
	result := make(chan error)
	CommandQueue() <- &CommandQuery{
		Command:       command,
		ExecutionType: executionType,
		Result:        result,
		Context:       c,
	}

	// Wait for the scheduler to take the command from the queue
	// And let it set the result
	return result
}

// Start the execution of the command by sending a grpc request to a processor
func (command *Command) Execute(c *context.Context) error {
	// Wait for the scheduler to take the command from the queue
	// And let it set the result
	return <-command.execute(c, EXECUTE_COMMAND)
}

// Start the execution of the command by sending a grpc request to a processor
func (command *Command) Undo(c *context.Context) error {
	// Wait for the scheduler to take the command from the queue
	// And let it set the result
	return <-command.execute(c, UNDO_COMMAND)
}

// Start the execution of the command by sending a grpc request to a processor
func (command *Command) Test(c *context.Context) error {
	// Wait for the scheduler to take the command from the queue
	// And let it set the result
	return <-command.execute(c, TEST_COMMAND)
}

// Used internally to store the result of the command call
func (command *Command) SetOutput(message protoreflect.Message) error {
	if command.Base.Output == nil {
		command.Base.Output = &tools_proto.Socket{}
	}

	// Apply the command update
	commandFieldDescriptor := message.Descriptor().Fields().ByName(protoreflect.Name(COMMAND_PATCH_OUTPUT))
	if commandFieldDescriptor != nil {
		commandField := message.Get(commandFieldDescriptor)
		if commandFieldDescriptor.Message() == nil {
			return fmt.Errorf("invalid datatype for command field: expected message, received %s", commandFieldDescriptor.Kind())
		} else {
			commandPatch := tools_proto.Command{}
			rawMessage, err := proto.Marshal(commandField.Message().Interface())
			if err != nil {
				return fmt.Errorf("could not marshall command field to patch command: %w", err)
			}
			err = proto.Unmarshal(rawMessage, &commandPatch)
			if err != nil {
				return fmt.Errorf("could not parse command field to patch command: %w", err)
			}

			command.Update(&Command{&commandPatch})
		}
		message.Clear(commandFieldDescriptor)
	}

	// Set the output to the message
	err := command.GetBase().GetOutput().UpdateWithMessage(message)
	if err != nil {
		return fmt.Errorf("could not set the output to command %s: %w", command.GetBase().GetId(), err)
	}
	return nil
}

// Update the command with a patch
func (command *Command) Update(patch *Command) bool {
	// Patch the local version of the command
	isPatched := false
	if command.GetBase().Update(patch.GetBase()) {
		isPatched = true
	}

	return isPatched
}
