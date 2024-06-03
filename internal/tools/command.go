package tools

import (
	"fmt"

	tools_proto "github.com/Acedyn/zorro-proto/gen/protos/tools"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type CommandExecutionType string

var EXECUTE_COMMAND CommandExecutionType = "Execute"
var UNDO_COMMAND CommandExecutionType = "Undo"
var TEST_COMMAND CommandExecutionType = "Test"

// Wrapped command with methods attached
type Command struct {
	*tools_proto.Command
}

// Get the wrapped base with all its methods
func (command *Command) GetBase() *ToolBase {
	return &ToolBase{ToolBase: command.Command.GetBase()}
}

// Used internally to store the result of the command call
func (command *Command) SetOutput(message protoreflect.Message) error {
	if command.Base.Output == nil {
		command.Base.Output = &tools_proto.Socket{}
	}

	// Set the output to the message
	err := command.GetBase().GetOutput().UpdateValueWithMessage(message)
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
