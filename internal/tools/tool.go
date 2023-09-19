package tools

import (
	"github.com/life4/genesis/maps"
)

// Traversable tool are nested tools linked via dependencies
type TraversableTool interface {
	Traverse(func(TraversableTool) error) error
	GetBase() *ToolBase
}

func (tool *ToolBase) Patch(patch *ToolBase) bool {
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
	maps.Update(tool.Logs, patch.GetLogs())

	return isPatched
}
