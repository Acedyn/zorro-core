package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/plugin"
	"github.com/Acedyn/zorro-core/internal/processor"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	context_proto "github.com/Acedyn/zorro-proto/zorroprotos/context"
	plugin_proto "github.com/Acedyn/zorro-proto/zorroprotos/plugin"
	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
)

// Wrapped context with methods attached
type Context struct {
	*context_proto.Context
}

func (context *Context) GetPlugins() []*plugin.Plugin {
	return slices.Map(context.Context.GetPlugins(), func(p *plugin_proto.Plugin) *plugin.Plugin {
		return &plugin.Plugin{Plugin: p}
	})
}

// Gather the environment variables of all the context's plugins
// in the form "key=value".
func (context *Context) Environ(includeCurrent bool) []string {
	environ := map[string]string{}
	// Convert the current environment to a map rather than
	// a "key=value" slice
	if includeCurrent {
		for _, environVariable := range os.Environ() {
			if splittedEnviron := strings.Split(environVariable, "="); len(splittedEnviron) == 2 {
				environ[splittedEnviron[0]] = splittedEnviron[1]
			}
		}
	}

	// Each plugins brings its own set of environment variable
	// modifications
	for _, plugin := range context.GetPlugins() {
		for key, pluginEnviron := range plugin.GetEnv() {
			// Prepend means insert at the beginning of the current value
			for _, valuePrepend := range pluginEnviron.GetPrepend() {
				if current, ok := environ[key]; ok {
					current = strings.Trim(current, string(filepath.ListSeparator))
					environ[key] = strings.Join([]string{valuePrepend, current}, string(filepath.ListSeparator))
				} else {
					environ[key] = valuePrepend
				}
			}
			// Append means add at the end of the current value
			for _, valueAppend := range pluginEnviron.GetAppend() {
				if current, ok := environ[key]; ok {
					current = strings.Trim(current, string(filepath.ListSeparator))
					environ[key] = strings.Join([]string{current, valueAppend}, string(filepath.ListSeparator))
				} else {
					environ[key] = valueAppend
				}
			}
			// Set will override the current value
			if pluginEnviron.Set != nil {
				environ[key] = pluginEnviron.GetSet()
			}
		}
	}

	// List of the loaded plugins
	environ["ZORRO_PLUGINS"] = strings.Join(slices.Map(context.GetPlugins(), func(plugin *plugin.Plugin) string {
		return plugin.GetPath()
	}), string(filepath.ListSeparator))

	// List of the available actions
	maps.IMerge(environ, concatToolsDeclarations(
		"ZORRO_ACTIONS",
		slices.Reduce(context.GetPlugins(), []*plugin_proto.ToolsDeclaration{}, func(plugin *plugin.Plugin, acc []*plugin_proto.ToolsDeclaration) []*plugin_proto.ToolsDeclaration {
			return append(acc, plugin.Tools.GetActions()...)
		}),
	))

	// List of the available hooks
	maps.IMerge(environ, concatToolsDeclarations(
		"ZORRO_HOOKS",
		slices.Reduce(context.GetPlugins(), []*plugin_proto.ToolsDeclaration{}, func(plugin *plugin.Plugin, acc []*plugin_proto.ToolsDeclaration) []*plugin_proto.ToolsDeclaration {
			return append(acc, plugin.Tools.GetHooks()...)
		}),
	))

	// List of the available widgets
	maps.IMerge(environ, concatToolsDeclarations(
		"ZORRO_WIDGETS",
		slices.Reduce(context.GetPlugins(), []*plugin_proto.ToolsDeclaration{}, func(plugin *plugin.Plugin, acc []*plugin_proto.ToolsDeclaration) []*plugin_proto.ToolsDeclaration {
			return append(acc, plugin.Tools.GetWidgets()...)
		}),
	))

	// List of the available commands
	maps.IMerge(environ, concatToolsDeclarations(
		"ZORRO_COMMANDS",
		slices.Reduce(context.GetPlugins(), []*plugin_proto.ToolsDeclaration{}, func(plugin *plugin.Plugin, acc []*plugin_proto.ToolsDeclaration) []*plugin_proto.ToolsDeclaration {
			return append(acc, plugin.Tools.GetWidgets()...)
		}),
	))

	// Port and host of the grpc server
	_, grpcStatus := network.GrpcServer()
	if grpcStatus.IsRunning {
		environ["ZORRO_GRPC_CORE_PORT"] = strconv.Itoa(grpcStatus.Port)
		environ["ZORRO_GRPC_CORE_HOST"] = grpcStatus.Host
	}

	// Reformat the environment variables to the "key=value" slice format
	return slices.Map(maps.Keys(environ), func(el string) string {
		return el + "=" + environ[el]
	})
}

// List all the tools and groupd them by category
func concatToolsDeclarations(prefix string, toolsDeclarations []*plugin_proto.ToolsDeclaration) map[string]string {
	concatenatedDeclarations := map[string]string{}
	for _, toolsDeclaration := range toolsDeclarations {
		currentTools, ok := concatenatedDeclarations[prefix+":"+toolsDeclaration.GetCategory()]
		if !ok {
			currentTools = toolsDeclaration.GetPath()
		} else {
			currentTools = strings.Join([]string{currentTools, toolsDeclaration.GetPath()}, string(filepath.ListSeparator))
		}
		concatenatedDeclarations[prefix+":"+toolsDeclaration.GetCategory()] = currentTools
	}

	return concatenatedDeclarations
}

// Flatten list of all the processors present in the selected plugins
func (context *Context) AvailableProcessors() []*processor.Processor {
	availableProcessors := []*processor.Processor{}
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
		Context: &context_proto.Context{
			Plugins: slices.Map(resolvedPlugins, func(p *plugin.Plugin) *plugin_proto.Plugin { return p.Plugin }),
		},
	}, nil
}
