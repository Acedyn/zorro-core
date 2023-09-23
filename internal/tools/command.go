package tools

import (
	"fmt"
	"sync"

	"github.com/Acedyn/zorro-core/internal/context"

  tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
)

var (
	commandQueue chan *CommandHandle
	once         sync.Once
)

// Wrapped command with methods attached
type Command struct {
  *tools_proto.Command
}

// Started command, waiting to be sheduled
type CommandHandle struct {
	Command *Command
	Result  chan error
	Context *context.Context
}

// Get the wrapped base with all its methods
func (command *Command) GetBase() *ToolBase {
  return &ToolBase{ToolBase: command.Command.GetBase()}
}

// Getter for the commands queue singleton which holds the queue
// of command waiting to be scheduled
func CommandQueue() chan *CommandHandle {
	once.Do(func() {
		commandQueue = make(chan *CommandHandle)
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
func (command *Command) Execute(c *context.Context) error {
	result := make(chan error)
	CommandQueue() <- &CommandHandle{
		Command: command,
		Result:  result,
		Context: c,
	}

	// Wait for the scheduler to take the command from the queue
	// And let it set the result
	return <-result
}

// Update the command with a patch
func (command *Command) Patch(patch *Command) bool {
	// Patch the local version of the command
	isPatched := false
	if command.GetBase().Patch(patch.GetBase()) {
		isPatched = true
	}

	return isPatched
}
