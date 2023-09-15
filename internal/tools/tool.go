package tools

// Traversable tool are nested tools linked via dependencies
type TraversableTool interface {
	Traverse(func(TraversableTool) error) error
	GetBase() *ToolBase
}
