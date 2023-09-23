//go:build exclude
package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

// Test the default initialization of a plugin's field
var pluginInitTests = map[string]*Plugin{
	"/foo/bar/zorro-plugin.json": {
		Name: "foo_bar",
		Tools: &PluginTools{
			Commands: []string{"./commands", "./foo/commands"},
		},
	},
}

func TestPluginInit(t *testing.T) {
	for pluginPath, pluginTest := range pluginInitTests {
		pluginTest.Path = &pluginPath
		pluginTest.InitFields()

		if *pluginTest.Label != "Foo Bar" {
			t.Errorf("invalid plugin label initialized: %s", *pluginTest.Label)
		}

		if pluginTest.Tools.Commands[0] != filepath.Join("/foo/bar/commands") {
			t.Errorf("invalid plugin command paths initialized: %s", pluginTest.Tools.Commands[0])
		}

		if pluginTest.Tools.Commands[1] != filepath.Join("/foo/bar/foo/commands") {
			t.Errorf("invalid plugin command paths initialized: %s", pluginTest.Tools.Commands[1])
		}
	}
}

// Test the loading of a bare plugin
func TestLoadPluginBare(t *testing.T) {
	pluginPath := "/foo/bar@1.2/zorro-plugin.json"
	plugin := LoadPluginBare(pluginPath)

	if plugin.Name != "bar" {
		t.Errorf("invalid bare plugin's name loaded: %s", plugin.Name)
	}

	if plugin.Version != "1.2" {
		t.Errorf("invalid bare plugin's version loaded: %s", plugin.Version)
	}
}

// Test if the loaded plugins correspond to the file's content
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
		ExpectedRequire:  []string{},
	},
	"baz/baz@5.2/zorro-plugin.json": {
		ExpectedName:     "baz",
		ExpectedLabel:    "Baz",
		ExpectedCommands: []string{"./commands"},
		ExpectedActions:  []string{"./actions"},
		ExpectedRequire:  []string{"foo>=5.3"},
	},
}

func TestLoadPluginFromFile(t *testing.T) {
	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))

	for path, expectedPlugin := range loadPluginFromFileTests {
		fullPath := filepath.Join(cwdPath, "test", "mock", path)
		loadedPlugin, err := LoadPluginFromFile(fullPath)
		if err != nil {
			t.Errorf("an error occured while loading the plugin at path %s\n\t%s", path, err)
		}

		if loadedPlugin.GetName() != expectedPlugin.ExpectedName {
			t.Errorf("incorrect name loaded on the plugin at path %s (%s)", path, loadedPlugin.GetName())
		}
		if loadedPlugin.GetLabel() != expectedPlugin.ExpectedLabel {
			t.Errorf("incorrect label loaded on the plugin at path %s (%s)", path, loadedPlugin.GetLabel())
		}
		for index := range loadedPlugin.GetTools().Commands {
			expectedCommand := filepath.Join(filepath.Dir(fullPath), expectedPlugin.ExpectedCommands[index])
			if loadedPlugin.GetTools().Commands[index] != expectedCommand {
				t.Errorf("incorrect command loaded on the plugin at path %s (%s)", path, loadedPlugin.GetTools().Commands[index])
			}
		}
		for index := range loadedPlugin.GetTools().Actions {
			expectedAction := filepath.Join(filepath.Dir(fullPath), expectedPlugin.ExpectedActions[index])
			if loadedPlugin.GetTools().Actions[index] != expectedAction {
				t.Errorf("incorrect action loaded on the plugin at path %s (%s)", path, loadedPlugin.GetTools().Actions[index])
			}
		}
		for index := range loadedPlugin.GetRequire() {
			if loadedPlugin.GetRequire()[index] != expectedPlugin.ExpectedRequire[index] {
				t.Errorf("incorrect require loaded on the plugin at path %s (%s)", path, loadedPlugin.GetRequire()[index])
			}
		}
	}
}
