package manager

import (
	"fmt"

	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/tools"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	"github.com/life4/genesis/maps"
)

// Create an action with its associated context
func InvokeAction(name string, pluginQuery []string, customConfig *config_proto.Config) (*tools.Action, error) {
	// The context will determine how the action will be resolved
	actionContext, err := context.NewContext(pluginQuery, customConfig)
	if err != nil {
		return nil, fmt.Errorf("action's context could not be built: %w", err)
	}

	// Try to find the requested action among the available ones
	actionPath, actionExists := actionContext.AvailableActions()[name]
	if !actionExists {
		return nil, fmt.Errorf(
			"could not find action named %s in the resolved context from query %s (available: %s)",
			name,
			pluginQuery,
			maps.Keys(actionContext.AvailableActions()),
		)
	}

	action, err := tools.LoadAction(actionPath)
	if err != nil {
		return nil, fmt.Errorf("an error occured when loading the action at path %s: %w", actionPath, err)
	}

	// Register the action to the list of invoked tools
	invokedTools := InvokedActions()
	invokedTools = append(invokedTools, action)

	return action, nil
}
