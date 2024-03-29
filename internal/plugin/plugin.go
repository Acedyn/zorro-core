package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/pkg/config"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	plugin_proto "github.com/Acedyn/zorro-proto/zorroprotos/plugin"
	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	"github.com/life4/genesis/slices"
	"golang.org/x/text/cases"
)

const (
	PLUGIN_DEFINITION_NAME = "zorro-plugin"
	VERSION_SPERARATOR     = "@"
	DEFAULT_VERSION        = "v0.0.0"
)

// Wrapped plugin with methods attached
type Plugin struct {
	*plugin_proto.Plugin
}

func (plugin *Plugin) GetProcessors() []*processor.Processor {
	return slices.Map(plugin.Plugin.GetProcessors(), func(p *processor_proto.Processor) *processor.Processor {
		return &processor.Processor{Processor: p}
	})
}

func (plugin *Plugin) GetRepository() *config_proto.RepositoryConfig {
	if plugin.Plugin.GetRepository() != nil {
		return plugin.Plugin.GetRepository()
	}

	return &config_proto.RepositoryConfig{
		FileSystemConfig: &config_proto.RepositoryConfig_Memory{},
	}
}

// Initialize the plugin's fields by expanding paths and initializing
// default values
func (plugin *Plugin) InitDefaults() {
	// Build the default label if none is set
	if plugin.GetLabel() == "" {
		caser := cases.Title(config.GetLanguage())
		plugin.Label = caser.String(strings.ReplaceAll(plugin.GetName(), "_", " "))
	}

	// Make sure structs are not nil
	if plugin.Tools == nil {
		plugin.Tools = &plugin_proto.PluginTools{}
	}
	if plugin.Env == nil {
		plugin.Env = map[string]*plugin_proto.PluginEnv{}
	}

	// Expand the relative paths of the env keys
	fieldsToExpand := [][]string{}
	for _, env := range plugin.GetEnv() {
		fieldsToExpand = append(fieldsToExpand, env.Prepend)
		fieldsToExpand = append(fieldsToExpand, env.Append)
	}

	for _, pathsToExpand := range fieldsToExpand {
		for index, pathToExpand := range pathsToExpand {
			pathsToExpand[index] = plugin.resolveRelativePath(pathToExpand)
		}
	}

	// Expand the relative paths of the declared tools
	pluginToolsDeclarations := [][]*plugin_proto.ToolsDeclaration{
		plugin.GetTools().GetActions(),
		plugin.GetTools().GetCommands(),
		plugin.GetTools().GetHooks(),
		plugin.GetTools().GetWidgets(),
	}
	for _, toolsDeclarations := range pluginToolsDeclarations {
		for _, toolsDeclaration := range toolsDeclarations {
			toolsDeclaration.Path = plugin.resolveRelativePath(toolsDeclaration.GetPath())
		}
	}

	// Normalize the path to be specific to the current os
	plugin.Path = strings.ReplaceAll(filepath.Join(plugin.GetPath()), string(filepath.Separator), "/")

	// Make sure the plugin doesn't require itself
	filteredRequires := make([]string, 0, len(plugin.GetRequire()))
	for _, requirement := range plugin.GetRequire() {
		if ParseVersionQuery(requirement).Name != plugin.GetName() {
			filteredRequires = append(filteredRequires, requirement)
		}
	}
	plugin.Require = filteredRequires
}

// Expand the path, relative to the plugin
func (plugin *Plugin) resolveRelativePath(path string) string {
	if filepath.IsAbs(path) {
		// Normalize the path
		return strings.ReplaceAll(filepath.Join(path), string(filepath.Separator), "/")
	}

	return strings.ReplaceAll(filepath.Join(filepath.Dir(plugin.GetPath()), path), string(filepath.Separator), "/")
}

func (plugin *Plugin) Load() error {
	fileSystem, err := GetFileSystem(plugin.GetRepository())
	if err != nil {
		return fmt.Errorf("invalid file system (%s): %w", plugin.GetRepository(), err)
	}

	fmt.Println(fileSystem)
	fmt.Println(plugin.GetPath())
	fileHandle, err := fileSystem.Open(plugin.GetPath())
	if err != nil {
		return fmt.Errorf("could not open file (%s): %w", plugin.GetPath(), err)
	}
	defer fileHandle.Close()

	// Parse the plugin data
	fileData, err := io.ReadAll(fileHandle)
	if err != nil {
		return fmt.Errorf("could not read config file (%s): %w", plugin.GetPath(), err)
	}

	switch filepath.Ext(plugin.GetPath()) {
	case ".json":
		return plugin.LoadJson(fileData)
	default:
		return fmt.Errorf("unhandled filetype for plugin file (%s)", filepath.Ext(plugin.GetPath()))
	}
}

// Initialize the plugin after parsing json config
func (plugin *Plugin) LoadJson(config []byte) error {
	err := json.Unmarshal(config, plugin)
	if err != nil {
		return fmt.Errorf("invalid plugin json config (%s): %w", plugin.GetPath(), err)
	}

	plugin.InitDefaults()
	return nil
}

// Get a minial version of a plugin without openning any files
func GetPluginBare(path string, repository *config_proto.RepositoryConfig) *Plugin {
	// Guess the version and the name from the path
	splittedName := strings.Split(strings.ReplaceAll(filepath.Base(filepath.Dir(path)), string(filepath.Separator), "/"), VERSION_SPERARATOR)
	name := strings.ReplaceAll(filepath.Base(filepath.Dir(path)), string(filepath.Separator), "/")
	version := DEFAULT_VERSION
	if len(splittedName) == 2 {
		name, version = splittedName[0], splittedName[1]
	}

	// The file system is optional
	if repository == nil {
		repository = &config_proto.RepositoryConfig{
			FileSystemConfig: &config_proto.RepositoryConfig_Memory{},
		}
	}

	return &Plugin{
		Plugin: &plugin_proto.Plugin{
			Name:       name,
			Version:    version,
			Path:       path,
			Repository: repository,
		},
	}
}

// Get a plugin from a file
func GetPluginFromFile(path string, repository *config_proto.RepositoryConfig) (*Plugin, error) {
	plugin := GetPluginBare(path, repository)
	err := plugin.Load()
	return plugin, err
}
