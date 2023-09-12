package context

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Acedyn/zorro-core/internal/config"

	"golang.org/x/text/cases"
)

const VERSION_SPERARATOR string = "@"

// Initialize the plugin's fields by expanding paths and initializing
// default values
func (plugin *Plugin) InitFields() {
	// Build the default label if none is set
	if plugin.Label == nil {
		caser := cases.Title(config.GetLanguage())
		generatedLabel := caser.String(strings.ReplaceAll(plugin.GetName(), "_", " "))
		plugin.Label = &generatedLabel
	}

	// Expand the relative paths
	pluginTools := [][]string{
		plugin.Tools.Commands,
		plugin.Tools.Actions,
		plugin.Tools.Hooks,
		plugin.Tools.Widgets,
	}
	for _, tool := range pluginTools {
		for index, tool_path := range tool {
			tool[index] = plugin.resolveRelativePath(tool_path)
		}
	}

	// Normalize the path to be specific to the current os
	normalizedPath := filepath.Join(plugin.GetPath())
	plugin.Path = &normalizedPath
}

// Expand the path, relative to the plugin
func (plugin *Plugin) resolveRelativePath(path string) string {
	if filepath.IsAbs(path) {
		// Normalize the path
		return filepath.Join(path)
	}

	return filepath.Join(filepath.Dir(plugin.GetPath()), path)
}

// Load a minial version of a plugin without looking for any files
func LoadPluginBare(path string) *Plugin {
	splittedName := strings.Split(filepath.Base(filepath.Dir(path)), VERSION_SPERARATOR)
	name := filepath.Base(filepath.Dir(path))
	version := "0.0.0"
	if len(splittedName) == 2 {
		name, version = splittedName[0], splittedName[1]
	}
	return &Plugin{
		Name:    name,
		Version: version,
		Path:    &path,
	}
}

// Initialize a plugin from a file
func LoadPluginFromFile(path string) (*Plugin, error) {
	fileHandle, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Could not open file (%s): %w", path, err)
	}
	defer fileHandle.Close()

	// Parse the plugin data
	fileData, err := io.ReadAll(fileHandle)
	if err != nil {
		return nil, fmt.Errorf("Could not read config file (%s): %w", path, err)
	}

	// Handle multiple file types
	plugin := LoadPluginBare(path)
	switch filepath.Ext(path) {
	case ".json":
		return LoadPluginFromJson(fileData, plugin)
	default:
		return nil, fmt.Errorf("Unhandled filetype for plugin file (%s)", filepath.Ext(path))
	}
}

// Initialize the plugin after parsing json config
func LoadPluginFromJson(config []byte, plugin *Plugin) (*Plugin, error) {
	if plugin == nil {
		plugin = &Plugin{}
	}
	err := json.Unmarshal(config, plugin)
	if err != nil {
		return nil, fmt.Errorf("Invalid plugin json config (%s): %w", plugin.Name, err)
	}

	return plugin, nil
}
