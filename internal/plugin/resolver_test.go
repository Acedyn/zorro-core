package plugin

import (
	"strings"
	"testing"

	_ "github.com/Acedyn/zorro-core/testdata"

	config_proto "github.com/Acedyn/zorro-proto/gen/protos/config"
	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
)

// Expected version count to be found for each plugins
var getAllPluginVersionsTests = map[string]int{
	"baz": 2,
	"bar": 2,
	"foo": 3,
}

// Test the FindPluginVersions function
func TestGetAllPluginVersion(t *testing.T) {
	for name, expectedPluginCount := range getAllPluginVersionsTests {
		plugins := FindPluginVersions(name, &config_proto.PluginConfig{
			Repositories: []*filesystem_proto.FileSystemSettings{
				{
					FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
				},
			},
		})

		if expectedPluginCount != len(plugins) {
			t.Errorf("incorrect count of plugin (found: %d, expected %d)", len(plugins), expectedPluginCount)
		}
	}
}

// Plugin queries and their expected resolved plugins
var pluginResolutionTests = map[string]map[string]string{
	"foo>=3.0.3 bar==2.3 baz<=5.6": {
		"foo": "3.2",
		"bar": "2.3",
		"baz": "3.1",
	},
	"foo>=3.2 foo<=3.8": {
		"foo": "3.2",
	},
}

// Test the ResolvePlugins function
func TestPluginResolution(t *testing.T) {
	for pluginQuery, expectedVersions := range pluginResolutionTests {
		query := strings.Split(pluginQuery, " ")
		resolvedPlugins, err := ResolvePlugins(query, &config_proto.PluginConfig{
			Repositories: []*filesystem_proto.FileSystemSettings{
				{
					FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
				},
			},
		})
		if err != nil {
			t.Errorf("could not resolve plugin graph: %s", err.Error())
			continue
		}

		for _, resolvedPlugin := range resolvedPlugins {
			expectedVersion, ok := expectedVersions[resolvedPlugin.GetName()]
			if ok && !(expectedVersion == resolvedPlugin.GetVersion()) {
				t.Errorf(
					"incorrect plugin version resolved for %s (resolved %s, expected %s)",
					resolvedPlugin.GetName(),
					resolvedPlugin.GetVersion(),
					expectedVersion,
				)
			}
		}
	}
}
