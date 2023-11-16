package tools

import (
	"strings"

	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/life4/genesis/maps"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var TOOL_SEPARATOR string = "/"

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
	Traverse(func(Tool) error) error
	GetChild(path string) (Tool, TraversableTool)
}

// Get the wrapped output with all its methods
func (tool *ToolBase) GetOutput() *Socket {
	if tool.ToolBase.GetOutput() == nil {
		tool.ToolBase.Output = &tools_proto.Socket{}
	}
	return &Socket{tool.ToolBase.GetOutput()}
}

// Get the wrapped inputt with all its methods
func (tool *ToolBase) GetInput() *Socket {
	if tool.ToolBase.GetInput() == nil {
		tool.ToolBase.Input = &tools_proto.Socket{}
	}
	return &Socket{tool.ToolBase.GetInput()}
}

func (tool *ToolBase) Update(patch *ToolBase) bool {
	// Patch the local version of the tool
	isPatched := false

	// Update the fields
	if patch.Name != nil && tool.GetName() != patch.GetName() {
		tool.Name = patch.Name
		isPatched = true
	}
	if patch.Label != nil && tool.GetLabel() != patch.GetLabel() {
		tool.Label = patch.Label
		isPatched = true
	} else if tool.Label == nil {
		generatedLabel := strings.Replace(
			cases.Title(language.Und, cases.NoLower).String(tool.GetName()),
			"_", " ", 0,
		)
		tool.Label = &generatedLabel
	}
	if patch.Status != nil && tool.GetStatus() != patch.GetStatus() {
		tool.Status = patch.Status
		isPatched = true
	}
	if patch.Tooltip != nil && tool.GetTooltip() != patch.GetTooltip() {
		tool.Tooltip = patch.Tooltip
		isPatched = true
	}

	// Update the inputs and outputs
	if patch.GetInput() != nil {
		tool.GetInput().Update(patch.GetInput())
	}
	if patch.GetOutput() != nil {
		tool.GetOutput().Update(patch.GetOutput())
	}

	// Logs are a special case, they are combined together
	if tool.Logs == nil {
		tool.Logs = map[int64]string{}
	}
	maps.Update(tool.Logs, patch.GetLogs())

	return isPatched
}
