package plugin

import (
	"os"
	"path/filepath"
	"testing"

	plugin_proto "github.com/Acedyn/zorro-proto/zorroprotos/plugin"
)

// Mocked plugin to test the InitDefaults methods
var pluginInitTest = Plugin{
	Plugin: &plugin_proto.Plugin{
		Name: "foo_bar",
		Path: "/foo/bar/zorro-plugin.json",
		Tools: &plugin_proto.PluginTools{
			Commands: []*plugin_proto.ToolsDeclaration{
				{
					Path: "./commands",
				},
				{
					Path: "./foo/commands",
				},
			},
		},
	},
}

// Test the InitDefaults methods
func TestPluginInit(t *testing.T) {
	pluginInitTest.InitDefaults()

	if pluginInitTest.Label != "Foo Bar" {
		t.Errorf("invalid plugin label initialized: %s", pluginInitTest.Label)
		return
	}

	if pluginInitTest.Tools.Commands[0].GetPath() != filepath.Join("/foo/bar/commands") {
		t.Errorf("invalid plugin command paths initialized: %s", pluginInitTest.Tools.Commands[0])
		return
	}

	if pluginInitTest.Tools.Commands[1].GetPath() != filepath.Join("/foo/bar/foo/commands") {
		t.Errorf("invalid plugin command paths initialized: %s", pluginInitTest.Tools.Commands[1])
		return
	}
}

// Test the loading of a bare plugin
func TestLoadPluginBare(t *testing.T) {
	pluginPath := "/foo/bar@1.2/zorro-plugin.json"
	plugin := GetPluginBare(pluginPath)

	if plugin.Name != "bar" {
		t.Errorf("invalid bare plugin's name loaded: %s", plugin.Name)
	}

	if plugin.Version != "1.2" {
		t.Errorf("invalid bare plugin's version loaded: %s", plugin.Version)
	}
}

// Expected values for the loaded mocked plugins
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

// Test the loading of the plugins files
func TestLoadPluginFromFile(t *testing.T) {
	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))

	for path, expectedPlugin := range loadPluginFromFileTests {
		fullPath := filepath.Join(cwdPath, "testdata", "mocked_plugins", path)
		loadedPlugin, err := GetPluginFromFile(fullPath)
		if err != nil {
			t.Errorf("an error occured while loading the plugin at path %s\n\t%s", path, err)
			return
		}

		if loadedPlugin.GetName() != expectedPlugin.ExpectedName {
			t.Errorf("incorrect name loaded on the plugin at path %s (%s)", path, loadedPlugin.GetName())
			return
		}
		if loadedPlugin.GetLabel() != expectedPlugin.ExpectedLabel {
			t.Errorf("incorrect label loaded on the plugin at path %s (%s)", path, loadedPlugin.GetLabel())
			return
		}
		for index := range loadedPlugin.GetTools().Commands {
			expectedCommand := filepath.Join(filepath.Dir(fullPath), expectedPlugin.ExpectedCommands[index])
			if loadedPlugin.GetTools().Commands[index].GetPath() != expectedCommand {
				t.Errorf("incorrect command loaded on the plugin at path %s (%s)", path, loadedPlugin.GetTools().Commands[index])
				return
			}
		}
		for index := range loadedPlugin.GetTools().Actions {
			expectedAction := filepath.Join(filepath.Dir(fullPath), expectedPlugin.ExpectedActions[index])
			if loadedPlugin.GetTools().Actions[index].GetPath() != expectedAction {
				t.Errorf("incorrect action loaded on the plugin at path %s (%s)", path, loadedPlugin.GetTools().Actions[index])
				return
			}
		}
		for index := range loadedPlugin.GetRequire() {
			if loadedPlugin.GetRequire()[index] != expectedPlugin.ExpectedRequire[index] {
				t.Errorf("incorrect require loaded on the plugin at path %s (%s)", path, loadedPlugin.GetRequire()[index])
				return
			}
		}
	}
}
