package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
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

	if pluginInitTest.Tools.Commands[0].GetPath() != strings.ReplaceAll(filepath.Join("/foo/bar/commands"), string(filepath.Separator), "/") {
		t.Errorf("invalid plugin command paths initialized: %s", pluginInitTest.Tools.Commands[0])
		return
	}

	if pluginInitTest.Tools.Commands[1].GetPath() != strings.ReplaceAll(filepath.Join("/foo/bar/foo/commands"), string(filepath.Separator), "/") {
		t.Errorf("invalid plugin command paths initialized: %s", pluginInitTest.Tools.Commands[1])
		return
	}
}

// Test the loading of a bare plugin
func TestLoadPluginBare(t *testing.T) {
	pluginPath := "/foo/bar@1.2/zorro-plugin.json"
	plugin := GetPluginBare(pluginPath, nil)

	if plugin.Name != "bar" {
		t.Errorf("Invalid bare plugin's name loaded: %s", plugin.Name)
	}

	if plugin.Version != "1.2" {
		t.Errorf("Invalid bare plugin's version loaded: %s", plugin.Version)
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
		t.Errorf("Could not get the current working directory\n\t%s", err)
	}
	cwdPath = strings.ReplaceAll(filepath.Dir(filepath.Dir(filepath.Join(cwdPath))), string(filepath.Separator), "/")

	for path, expectedPlugin := range loadPluginFromFileTests {
		fullPath := strings.ReplaceAll(filepath.Join(cwdPath, "testdata", "plugins"), string(filepath.Separator), "/")
		loadedPlugin, err := GetPluginFromFile(path, &config_proto.RepositoryConfig{
			FileSystemConfig: &config_proto.RepositoryConfig_Os{
				Os: &config_proto.OsFsConfig{
					Directory: fullPath,
				},
			},
		})

		if err != nil {
			t.Errorf("An error occured while loading the plugin at path %s\n\t%s", path, err)
			return
		}

		if loadedPlugin.GetName() != expectedPlugin.ExpectedName {
			t.Errorf("Incorrect name loaded on the plugin at path %s (%s)", path, loadedPlugin.GetName())
			return
		}
		if loadedPlugin.GetLabel() != expectedPlugin.ExpectedLabel {
			t.Errorf("Incorrect label loaded on the plugin at path %s (%s)", path, loadedPlugin.GetLabel())
			return
		}
		for index := range loadedPlugin.GetTools().Commands {
			expectedCommand := strings.ReplaceAll(filepath.Join(filepath.Dir(path), expectedPlugin.ExpectedCommands[index]), string(filepath.Separator), "/")
			if loadedPlugin.GetTools().Commands[index].GetPath() != expectedCommand {
				t.Errorf("Incorrect command loaded on the plugin at path %s (%s)", path, loadedPlugin.GetTools().Commands[index])
				return
			}
		}
		for index := range loadedPlugin.GetTools().Actions {
			expectedAction := strings.ReplaceAll(filepath.Join(filepath.Dir(path), expectedPlugin.ExpectedActions[index]), string(filepath.Separator), "/")
			if loadedPlugin.GetTools().Actions[index].GetPath() != expectedAction {
				t.Errorf("Incorrect action loaded on the plugin at path %s (%s)", path, loadedPlugin.GetTools().Actions[index])
				return
			}
		}
		for index := range loadedPlugin.GetRequire() {
			if loadedPlugin.GetRequire()[index] != expectedPlugin.ExpectedRequire[index] {
				t.Errorf("Incorrect require loaded on the plugin at path %s (%s)", path, loadedPlugin.GetRequire()[index])
				return
			}
		}
	}
}
