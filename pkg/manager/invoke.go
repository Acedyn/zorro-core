package manager

import (
	"fmt"

	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/tools"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"

	"github.com/life4/genesis/maps"
)

// Create an action with its associated context
func InvokeAction(name string, pluginQuery []string, customConfig *config_proto.Config) (error, *tools.Action) {
	// The context will determine how the action will be resolved
	actionContext, err := context.NewContext(pluginQuery, customConfig)
	if err != nil {
		return fmt.Errorf("action's context could not be built: %w", err), nil
	}

	// Try to find the requested action among the available ones
	actionPath, actionExists := actionContext.AvailableActions()[name]
	if !actionExists {
		return fmt.Errorf(
			"could not find action named %s in the resolved context from query %s (available: %s)",
			name,
			pluginQuery,
			maps.Keys(actionContext.AvailableActions()),
		), nil
	}

	action, err := tools.LoadAction(actionPath)
	if err != nil {
		return fmt.Errorf("an error occured when loading the action at path %s: %w", actionPath, err), nil
	}

	return nil, action
}
