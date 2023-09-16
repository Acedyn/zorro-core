package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Acedyn/zorro-core/internal/config"
)

// Test the result of multiple version comparisons
var versionComparisonsTests = map[struct {
	a string
	b string
}]VersionOperator{
	{a: "3.5", b: "3.4"}:            VersionOperator_MORE_EQUAL,
	{a: "2.1.4", b: "2.1.4"}:        VersionOperator_EQUAL,
	{a: "2.1", b: "2.1.4"}:          VersionOperator_EQUAL,
	{a: "3.5.alpha", b: "3.5.beta"}: VersionOperator_LESS_EQUAL,
	{a: "1.5.prod", b: "3.5.alpha"}: VersionOperator_LESS_EQUAL,
}

func TestVersionComparison(t *testing.T) {
	for comparison, expectedResult := range versionComparisonsTests {
		result := CompareVersions(comparison.a, comparison.b)
		if result != expectedResult {
			t.Errorf("invalid version comparison result: %s %s %s (expected: %s)", comparison.a, result, comparison.b, expectedResult)
		}
	}
}

// Test the search of plugins in a directory
var getAllPluginVersionsTests = map[string]int{
	"baz": 2,
	"bar": 2,
	"foo": 3,
}

func TestGetAllPluginVersion(t *testing.T) {
	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))
	fullPath := filepath.Join(cwdPath, "test", "mock")

	for name, expectedPluginCount := range getAllPluginVersionsTests {
		plugins := FindPluginVersions(name, &config.PluginConfig{
			PluginPaths: []string{fullPath},
		})

		if expectedPluginCount != len(plugins) {
			t.Errorf("incorrect count of plugin (found: %d, expected %d)", len(plugins), expectedPluginCount)
		}
	}
}

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

func TestPluginResolution(t *testing.T) {
	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))
	fullPath := filepath.Join(cwdPath, "test", "mock")

	for pluginQuery, expectedVersions := range pluginResolutionTests {
		query := strings.Split(pluginQuery, " ")
		resolvedPlugins, err := ResolvePlugins(query, &config.PluginConfig{
			PluginPaths: []string{fullPath},
		})
		if err != nil {
			t.Errorf("could not resolve plugin graph: %s", err.Error())
			continue
		}

		for _, resolvedPlugin := range resolvedPlugins {
			expectedVersion, ok := expectedVersions[resolvedPlugin.GetName()]
			if ok && !(expectedVersion == resolvedPlugin.GetVersion()) {
				t.Errorf("incorrect plugin version resolved for %s (resolved %s, expected %s)", resolvedPlugin.GetName(), resolvedPlugin.GetVersion(), expectedVersion)
			}
		}
	}
}
