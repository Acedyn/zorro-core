package tools

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
	"google.golang.org/protobuf/encoding/protojson"
)

// Wrapped action child with methods attached
type ActionChild struct {
	*tools_proto.ActionChild
}

// Get the wrapped action
func (actionChild *ActionChild) GetAction() *Action {
	return &Action{actionChild.ActionChild.GetAction()}
}

// Get the wrapped command
func (actionChild *ActionChild) GetCommand() *Command {
	return &Command{actionChild.ActionChild.GetCommand()}
}

// Wrapped action with methods attached
type Action struct {
	*tools_proto.Action
}

// Hold the returned value of a child task
type ChildTaskResult struct {
	Err error
	Key string
}

// Get the wrapped base with all its methods
func (action *Action) GetBase() *ToolBase {
	return &ToolBase{ToolBase: action.Action.GetBase()}
}

// Get the wrapped children with all their methods
// This method is for accessing the children, not for editing the map's structure
func (action *Action) GetChildren() map[string]*ActionChild {
	if action.Action.GetChildren() == nil {
		action.Action.Children = map[string]*tools_proto.ActionChild{}
	}

	return maps.Map(action.Action.GetChildren(), func(k string, v *tools_proto.ActionChild) (string, *ActionChild) {
		return k, &ActionChild{v}
	})
}

// Find and traverse children that have all their dependencies (upstream)
// completed.
func (action *Action) getReadyChildren(pending map[string]bool, completed []string) map[string]TraversableTool {
	readyChildren := map[string]TraversableTool{}
	for childKey, child := range action.GetChildren() {
		// Skip already process children
		if !pending[childKey] {
			continue
		}

		// Process children that don't have dependencies or that have
		// completed dependencies
		if slices.All(child.Upstream, func(el string) bool { return slices.Contains(completed, el) }) {
			switch child.GetChild().(type) {
			case *tools_proto.ActionChild_Action:
				readyChildren[childKey] = child.GetAction()
			case *tools_proto.ActionChild_Command:
				readyChildren[childKey] = child.GetCommand()
			}
		}
	}
	return readyChildren
}

// Run the task to all the children, respecting the order of execution
// and dependencies. Multiple might can run concurently (the task MUST be threadsafe !)
func (action *Action) Traverse(task func(TraversableTool) error) error {
	// We first traverse this action before traversing its children
	if err := task(action); err != nil {
		return fmt.Errorf("Error occured while traversing action %s: %w", action.GetBase().GetName(), err)
	}

	// At first, all the children are pending
	pending := maps.FromKeys(maps.Keys(action.GetChildren()), true)
	completed := []string{}
	tasksResults := make(chan *ChildTaskResult, 1)
	errors := []error{}
	// Hack to make sure the for loop executes at least once
	tasksResults <- nil

	// Wait for the next completed task unil the channel is closed
	for taskResult := range tasksResults {
		if taskResult != nil {
			completed = append(completed, taskResult.Key)
			errors = append(errors, taskResult.Err)
		}

		readyChildren := action.getReadyChildren(pending, completed)
		for childKey, child := range readyChildren {
			// All the ready children are executed in their own goroutine
			pending[childKey] = false
			go func(childKey string, child TraversableTool) {
				tasksResults <- &ChildTaskResult{
					Err: child.Traverse(task),
					Key: childKey,
				}
			}(childKey, child)
		}

		// If there is as many completed children as non pending, then the
		// action done
		if len(completed) == len(slices.Filter(maps.Values(pending), func(el bool) bool { return !el })) {
			close(tasksResults)
		}
	}

	// Gather all the potential errors that occured
	if slices.Any(errors, func(el error) bool { return el != nil }) {
		return fmt.Errorf(
			"One or multiple children errored during the execution of the action %s: \n%s",
			action.GetBase().GetName(),
			slices.Join(slices.Filter(errors, func(el error) bool { return el != nil }), "\n"),
		)
	}

	return nil
}

// Update the action with a patch
func (action *Action) Update(patch *Action) bool {
	// Patch the local version of the action
	isPatched := false
	if action.GetBase().Update(patch.GetBase()) {
		isPatched = true
	}

	// Apply the update on the children
	for childKey, patchChild := range patch.GetChildren() {
		actionChild, ok := action.GetChildren()[childKey]
		if !ok {
			action.Children[childKey] = patchChild.ActionChild
			isPatched = true
			continue
		}

		switch patchChild.GetChild().(type) {
		case *tools_proto.ActionChild_Action:
			if !ok || actionChild.GetAction() == nil {
				action.Children[childKey] = patchChild.ActionChild
				isPatched = true
			} else {
				isPatched = patchChild.GetAction().Update(actionChild.GetAction())
			}
		case *tools_proto.ActionChild_Command:
			if actionChild.GetCommand() == nil {
				action.Children[childKey] = patchChild.ActionChild
				isPatched = true
			} else {
				isPatched = patchChild.GetCommand().Update(actionChild.GetCommand())
			}
		}

		// Update the upstream field
		if slices.Equal(actionChild.GetUpstream(), patchChild.GetUpstream()) {
			actionChild.Upstream = patchChild.GetUpstream()
			isPatched = true
		}
	}

	return isPatched
}

// Update the action from json data
func (action *Action) Unmarshall(raw []byte) error {
	actionPatch := Action{&tools_proto.Action{}}
	err := protojson.Unmarshal(raw, &actionPatch)
	if err != nil {
		return fmt.Errorf("an error occured when unmarshalling json to action %s: %w", action, err)
	}

	action.Update(&actionPatch)
	return nil
}

// Initialize the action from json file
func LoadAction(path string) (*Action, error) {
	actionName := strings.Split(filepath.Base(path), ".")[0]
	action := Action{&tools_proto.Action{Base: &tools_proto.ToolBase{
		Name: &actionName,
	}}}

	fileHandle, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file (%s): %w", path, err)
	}
	defer fileHandle.Close()

	// Parse the action data
	fileData, err := io.ReadAll(fileHandle)
	if err != nil {
		return nil, fmt.Errorf("could not read config file (%s): %w", path, err)
	}

	// Apply the json data
	err = action.Unmarshall(fileData)
	if err != nil {
		return nil, err
	}

	return &action, nil
}
