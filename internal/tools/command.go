package tools

import (
	"fmt"
  "sync"
)

type CommandHandle struct {
	Command *Command
	Result chan error
}

var (
	commandQueue chan *CommandHandle
	once           sync.Once
)

// Getter for the commands pool singleton
func CommandPool() chan *CommandHandle {
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
func (command *Command) Execute() error {
  executionResult := make(chan error)
  commandQueue <- &CommandHandle{
    Command: command,
    Result: executionResult,
  }

  // Wait for the scheduler to take the command from the queue
  // And let it set the result
  return <- executionResult
}
