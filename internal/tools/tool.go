package tools

import (
	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/life4/genesis/maps"
)

// Wrapped tool base with methods attached
type ToolBase struct {
	*tools_proto.ToolBase
}

// Representation of a tool
type Tool interface {
	GetBase() *ToolBase
}

// Traversable tool are nested tools linked via dependencies
type TraversableTool interface {
	Tool
	Traverse(func(TraversableTool) error) error
}

// Get the wrapped output with all its methods
func (tool *ToolBase) GetOutput() *Socket {
	return &Socket{tool.ToolBase.GetOutput()}
}

// Get the wrapped inputt with all its methods
func (tool *ToolBase) GetInput() *Socket {
	return &Socket{tool.ToolBase.GetInput()}
}

func (tool *ToolBase) Update(patch *ToolBase) bool {
	// Patch the local version of the tool
	isPatched := false

	if patch.Label != nil && tool.GetLabel() != patch.GetLabel() {
		tool.Label = patch.Label
		isPatched = true
	}
	if patch.Status != nil && tool.GetStatus() != patch.GetStatus() {
		tool.Status = patch.Status
		isPatched = true
	}
	if patch.Tooltip != nil && tool.GetTooltip() != patch.GetTooltip() {
		tool.Tooltip = patch.Tooltip
		isPatched = true
	}

	// Logs are a special case, they are combined together
	if tool.Logs == nil {
		tool.Logs = map[int64]string{}
	}
	maps.Update(tool.Logs, patch.GetLogs())

	return isPatched
}
