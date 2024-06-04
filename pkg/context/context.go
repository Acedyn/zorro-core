package context

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Acedyn/zorro-core/internal/plugin"

	config_proto "github.com/Acedyn/zorro-proto/gen/protos/config"
	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
	tools_proto "github.com/Acedyn/zorro-proto/gen/protos/tools"
	"github.com/google/uuid"
	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
)

// Wrapped context with methods attached
type Context struct {
	id      string
	plugins []*plugin.Plugin
}

func (context *Context) GetPlugins() []*plugin.Plugin {
	return context.plugins
}

// Gather the environment variables of all the context's plugins
// in the form "key=value".
func (context *Context) Environ(includeCurrent bool) []string {
	environ := map[string]string{}

	// Add the current environment
	if includeCurrent {
		environ = buildCurrentEnvironment(environ)
	}

	// List of the loaded plugins
	environ["ZORRO_PLUGINS"] = strings.Join(slices.Map(context.GetPlugins(), func(plugin *plugin.Plugin) string {
		return plugin.GetPath()
	}), string(filepath.ListSeparator))

	// Each plugins brings its own set of environment variable modifications
	environ = buildPluginsEnvironment(environ, context.GetPlugins())

	// List the available tools grouped by category
	environ = buildToolsEnvironment(environ, context.GetPlugins())

	// Reformat the environment variables to the "key=value" slice format
	return slices.Map(maps.Keys(environ), func(el string) string {
		return el + "=" + environ[el]
	})
}

// Flatten list of all the tools present in the selected plugins and return their resolved paths
func (context *Context) AvailableToolsPaths(processor *scheduling_proto.Processor) map[tools_proto.ToolType][]string {
	processorSubsets := append(processor.GetSubsets(), processor.GetName())
	availableTools := map[tools_proto.ToolType][]string{
		tools_proto.ToolType_ACTION:  {},
		tools_proto.ToolType_COMMAND: {},
		tools_proto.ToolType_HOOK:    {},
		tools_proto.ToolType_WIDGET:  {},
	}

	for _, plugin := range context.GetPlugins() {
		fileSystemPrefix := ""
		switch fsConfig := plugin.GetRepository().FileSystemSettings.(type) {
		case *filesystem_proto.FileSystemSettings_Os:
			fileSystemPrefix = fsConfig.Os.Directory
		}

		for _, commandDeclaration := range plugin.GetTools().GetCommands() {
			if slices.Contains(processorSubsets, commandDeclaration.GetCategory()) {
				commandPath := strings.ReplaceAll(filepath.Join(fileSystemPrefix, commandDeclaration.GetPath()), string(filepath.Separator), "/")
				availableTools[tools_proto.ToolType_COMMAND] = append(availableTools[tools_proto.ToolType_COMMAND], commandPath)
			}
		}
		for _, actionDeclaration := range plugin.GetTools().GetActions() {
			actionPath := strings.ReplaceAll(filepath.Join(fileSystemPrefix, actionDeclaration.GetPath()), string(filepath.Separator), "/")
			availableTools[tools_proto.ToolType_ACTION] = append(availableTools[tools_proto.ToolType_COMMAND], actionPath)
		}
		for _, widgetDeclaration := range plugin.GetTools().GetWidgets() {
			widgetPath := strings.ReplaceAll(filepath.Join(fileSystemPrefix, widgetDeclaration.GetPath()), string(filepath.Separator), "/")
			availableTools[tools_proto.ToolType_WIDGET] = append(availableTools[tools_proto.ToolType_COMMAND], widgetPath)
		}
		for _, hookDeclaration := range plugin.GetTools().GetHooks() {
			hookPath := strings.ReplaceAll(filepath.Join(fileSystemPrefix, hookDeclaration.GetPath()), string(filepath.Separator), "/")
			availableTools[tools_proto.ToolType_HOOK] = append(availableTools[tools_proto.ToolType_COMMAND], hookPath)
		}
	}

	return availableTools
}

// Flatten list of all the commands present in the selected plugins that can be executed by the given processor
func (context *Context) AvailableCommandPaths(processor *scheduling_proto.Processor) []string {
	return context.AvailableToolsPaths(processor)[tools_proto.ToolType_COMMAND]
}

// Flatten list of all the commands present in the selected plugins that can be executed by the given processor
func (context *Context) AvailableActions() map[string]string {
	availableActions := map[string]string{}

	for _, plugin := range context.GetPlugins() {
		for _, actionDeclaration := range plugin.GetTools().GetActions() {
			actionName := strings.Split(strings.ReplaceAll(strings.ReplaceAll(filepath.Base(actionDeclaration.GetPath()), string(filepath.Separator), "/"), string(filepath.Separator), "/"), ".")[0]
			availableActions[actionName] = actionDeclaration.GetPath()
		}
	}

	return availableActions
}

// Flatten list of all the processors present in the selected plugins
func (context *Context) AvailableProcessors() []*scheduling_proto.Processor {
	availableProcessors := []*scheduling_proto.Processor{}
	for _, plugin := range context.GetPlugins() {
		for _, processor := range plugin.GetProcessors() {
			availableProcessors = append(availableProcessors, processor)
		}
	}

	return availableProcessors
}

// Constructor for a new context
func NewContext(pluginQuery []string, customConfig *config_proto.Config) (*Context, error) {
	var pluginConfig *config_proto.PluginConfig = nil
	if customConfig != nil {
		pluginConfig = customConfig.PluginConfig
	}
	resolvedPlugins, err := plugin.ResolvePlugins(pluginQuery, pluginConfig)
	if err != nil {
		return nil, fmt.Errorf("could not resolve plugins from queries %s: %w", pluginQuery, err)
	}

	return &Context{
		id:      uuid.New().String(),
		plugins: resolvedPlugins,
	}, nil
}
