package tools

import (
	"fmt"

	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
)

// Find and traverse children that have all their dependencies (upstream)
// completed.
func (action *Action) getReadyChildren(pending, completed map[string]bool) map[string]TraversableTool {
	readyChildren := map[string]TraversableTool{}
	for childKey, child := range action.GetChildren() {
		// Skip already process children
		if !pending[childKey] {
			continue
		}

		// Process children that don't have dependencies or that have
		// completed dependencies
		if child.Upstream == nil || completed[child.GetUpstream()] {
			switch child.GetChild().(type) {
			case *ActionChild_Action:
				readyChildren[childKey] = child.GetAction()
			}
		}
	}
	return readyChildren

}

// Run the task to all the children, respecting the order of execution
// and dependencies. Multiple might can run concurently.
func (action *Action) Traverse(task func(TraversableTool) error) error {
	// We first traverse this action before traversing its children
	if err := task(action); err != nil {
		return fmt.Errorf("Error occured while traversing action %s: %w", action.GetBase().GetName(), err)
	}

	// At first, all the children are pending
	pending := maps.FromKeys(maps.Keys(action.GetChildren()), true)
	completed := map[string]bool{}
	tasksResults := make(chan error)
	// Hack to make sure the for loop executes at least once
	tasksResults <- nil

	// Wait for the next completed task unil the channel is closed
	for range tasksResults {
		readyChildren := action.getReadyChildren(pending, completed)
		for childKey, child := range readyChildren {
			// All the ready children are executed in their own goroutine
			pending[childKey] = false
			go func(childKey string, child TraversableTool) {
				tasksResults <- task(child)
				completed[childKey] = true
			}(childKey, child)
		}

		// If there is as many completed children as non pending, then the
		// action done
		if len(completed) == len(slices.Filter(maps.Values(pending), func(el bool) bool { return el })) {
			close(tasksResults)
		}
	}

	return nil
}
