package context

import (
	"os"
	"path/filepath"
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
			t.Errorf("Invalid version comparison result: %s %s %s (expected: %s)", comparison.a, result, comparison.b, expectedResult)
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
		t.Errorf("Could not get the current working directory\n\t%s", err)
	}
	cwdPath = filepath.Dir(filepath.Dir(filepath.Join(cwdPath)))

	fullPath := filepath.Join(cwdPath, "test", "mock")

	for name, expectedPluginCount := range getAllPluginVersionsTests {
		plugins := FindPluginVersions(name, &config.PluginConfig{
			PluginPaths: []string{fullPath},
		})

		if expectedPluginCount != len(plugins) {
			t.Errorf("Incorrect count of plugin found: %d (expected %d)", len(plugins), expectedPluginCount)
		}
	}
}
