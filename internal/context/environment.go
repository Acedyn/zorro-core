package context

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/plugin"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	plugin_proto "github.com/Acedyn/zorro-proto/zorroprotos/plugin"
	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
)

// Convert the current environment to a map rather than a "key=value" slice
func buildCurrentEnvironment(baseEnvironment map[string]string) map[string]string {
	environment := baseEnvironment

	for _, environVariable := range os.Environ() {
		if splittedEnviron := strings.Split(environVariable, "="); len(splittedEnviron) == 2 {
			environment[splittedEnviron[0]] = splittedEnviron[1]
		}
	}

	return environment
}

// Combine all the plugins environment rules
func buildPluginsEnvironment(baseEnvironment map[string]string, plugins []*plugin.Plugin) map[string]string {
	environment := baseEnvironment

	for _, pluginItem := range plugins {

		fileSystemPrefix := ""
		switch fsConfig := pluginItem.GetRepository().FileSystemConfig.(type) {
		case *config_proto.RepositoryConfig_Os:
			fileSystemPrefix = fsConfig.Os.Directory
		}

		for key, pluginEnviron := range pluginItem.GetEnv() {
			// Prepend means insert at the beginning of the current value
			for _, valuePrepend := range pluginEnviron.GetPrepend() {
				valuePrepend = strings.ReplaceAll(filepath.Join(fileSystemPrefix, valuePrepend), string(filepath.Separator), "/")

				if current, ok := environment[key]; ok {
					current = strings.Trim(current, string(filepath.ListSeparator))
					environment[key] = strings.Join([]string{valuePrepend, current}, string(filepath.ListSeparator))
				} else {
					environment[key] = valuePrepend
				}
			}
			// Append means add at the end of the current value
			for _, valueAppend := range pluginEnviron.GetAppend() {
				valueAppend = strings.ReplaceAll(filepath.Join(fileSystemPrefix, valueAppend), string(filepath.Separator), "/")

				if current, ok := environment[key]; ok {
					current = strings.Trim(current, string(filepath.ListSeparator))
					environment[key] = strings.Join([]string{current, valueAppend}, string(filepath.ListSeparator))
				} else {
					environment[key] = valueAppend
				}
			}
			// Set will override the current value
			if pluginEnviron.Set != nil {
				environment[key] = pluginEnviron.GetSet()
			}
		}
	}

	return environment
}

// Create variables to indicate infos about the server
func buildGrpcEnvironment(baseEnvironment map[string]string) map[string]string {
	environment := baseEnvironment

	_, grpcStatus := network.GrpcServer()
	if grpcStatus.IsRunning {
		environment["ZORRO_GRPC_CORE_PORT"] = strconv.Itoa(grpcStatus.Port)
		environment["ZORRO_GRPC_CORE_HOST"] = grpcStatus.Host
	}

	return environment
}

// Build the list of paths to each tools and group them by category
func buildToolsEnvironment(baseEnvironment map[string]string, plugins []*plugin.Plugin) map[string]string {
	environment := baseEnvironment

	// List of the available actions
	maps.IMerge(environment, concatToolsDeclarations(
		"ZORRO_ACTIONS",
		slices.Reduce(plugins, []*plugin_proto.ToolsDeclaration{}, func(plugin *plugin.Plugin, acc []*plugin_proto.ToolsDeclaration) []*plugin_proto.ToolsDeclaration {
			return append(acc, plugin.Tools.GetActions()...)
		}),
	))

	// List of the available hooks
	maps.IMerge(environment, concatToolsDeclarations(
		"ZORRO_HOOKS",
		slices.Reduce(plugins, []*plugin_proto.ToolsDeclaration{}, func(plugin *plugin.Plugin, acc []*plugin_proto.ToolsDeclaration) []*plugin_proto.ToolsDeclaration {
			return append(acc, plugin.Tools.GetHooks()...)
		}),
	))

	// List of the available widgets
	maps.IMerge(environment, concatToolsDeclarations(
		"ZORRO_WIDGETS",
		slices.Reduce(plugins, []*plugin_proto.ToolsDeclaration{}, func(plugin *plugin.Plugin, acc []*plugin_proto.ToolsDeclaration) []*plugin_proto.ToolsDeclaration {
			return append(acc, plugin.Tools.GetWidgets()...)
		}),
	))

	// List of the available commands
	maps.IMerge(environment, concatToolsDeclarations(
		"ZORRO_COMMANDS",
		slices.Reduce(plugins, []*plugin_proto.ToolsDeclaration{}, func(plugin *plugin.Plugin, acc []*plugin_proto.ToolsDeclaration) []*plugin_proto.ToolsDeclaration {
			return append(acc, plugin.Tools.GetCommands()...)
		}),
	))

	return environment
}

// List all the tools present in a context and groupd them by category
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
