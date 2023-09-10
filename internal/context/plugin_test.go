package context

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type loadPluginFromFileTest struct {
	ExpectedName     string
	ExpectedVersion  string
	ExpectedLabel    string
	ExpectedCommands []string
	ExpectedActions  []string
	ExpectedRequire  []string
}

var loadPluginFromFileTests = map[string]*loadPluginFromFileTest{
	"foo/foo@3.1/zorro-plugin.json": {
		ExpectedName:     "foo",
		ExpectedLabel:    "The foo plugin",
		ExpectedCommands: []string{"./commands"},
		ExpectedActions:  []string{"./actions"},
	},
	"baz/baz@5.2/zorro-plugin.json": {
		ExpectedName:     "baz",
		ExpectedCommands: []string{"./commands"},
		ExpectedActions:  []string{"./actions"},
		ExpectedRequire:  []string{"foo>=5.3"},
	},
}

// Test if the loaded plugins correspond to the file's content
func TestLoadPluginFromFile(t *testing.T) {
	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(strings.ReplaceAll(cwdPath, string(os.PathSeparator), "/")))

	for path, expectedPlugin := range loadPluginFromFileTests {
		fullPath := filepath.Join(cwdPath, "test", "mock", path)
		loadedPlugin, err := LoadPluginFromFile(fullPath)
		if err != nil {
			t.Errorf("An error occured while loading the plugin at path %s\n\t%s", path, err)
		}

		if loadedPlugin.Name != expectedPlugin.ExpectedName {
			t.Errorf("Incorrect name loaded on the plugin at path %s", path)
		}
		if loadedPlugin.GetLabel() != expectedPlugin.ExpectedLabel {
			t.Errorf("Incorrect label loaded on the plugin at path %s", path)
		}
		if !reflect.DeepEqual(loadedPlugin.Tools.Commands, expectedPlugin.ExpectedCommands) {
			t.Errorf("Incorrect commands loaded on the plugin at path %s", path)
		}
		if !reflect.DeepEqual(loadedPlugin.Tools.Actions, expectedPlugin.ExpectedActions) {
			t.Errorf("Incorrect actions loaded on the plugin at path %s", path)
		}
		if !reflect.DeepEqual(loadedPlugin.Require, expectedPlugin.ExpectedRequire) {
			t.Errorf("Incorrect require loaded on the plugin at path %s", path)
		}
	}
}

// Test the loading of a bare plugin
func TestLoadPluginBare(t *testing.T) {
	pluginPath := "/foo/bar@1.2/zorro-plugin.json"
	plugin := LoadPluginBare(pluginPath)

	if plugin.Name != "bar" {
		t.Errorf("Invalid bare plugin's name loaded: %s", plugin.Name)
	}

	if plugin.Version != "1.2" {
		t.Errorf("Invalid bare plugin's version loaded: %s", plugin.Version)
	}
}

// Test the default initialization of a plugin's field
func TestPluginInit(t *testing.T) {
	pluginPath := "/foo/bar/zorro-plugin.json"
	plugin := &Plugin{
		Name: "foo_bar",
		Path: &pluginPath,
		Tools: &PluginTools{
			Commands: []string{"./commands", "/foo/commands"},
		},
	}

	plugin.InitFields()

	if *plugin.Label != "Foo Bar" {
		t.Errorf("Invalid plugin label initialized: %s", *plugin.Label)
	}

	if plugin.Tools.Commands[0] != "/foo/bar/commands" {
		t.Errorf("Invalid plugin command paths initialized: %s", plugin.Tools.Commands)
	}

	if plugin.Tools.Commands[1] != "/foo/commands" {
		t.Errorf("Invalid plugin command paths initialized: %s", plugin.Tools.Commands)
	}
}
