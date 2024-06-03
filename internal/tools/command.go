package tools

import (
	tools_proto "github.com/Acedyn/zorro-proto/gen/protos/tools"
)

// Wrapped command with methods attached
type Command struct {
	*tools_proto.Command
}

// Get the wrapped base with all its methods
func (command *Command) GetBase() *ToolBase {
	return &ToolBase{ToolBase: command.Command.GetBase()}
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
