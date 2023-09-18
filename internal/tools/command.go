package tools

import (
	"fmt"
)

// Commands are the smallest tool type, they don't have children
func (action *Command) Traverse(task func(TraversableTool) error) error {
	if err := task(action); err != nil {
		return fmt.Errorf("error occured while traversing command %s: %w", action.GetBase().GetName(), err)
	}

	return nil
}
