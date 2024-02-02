package plugin

import (
	"testing"
)

// Mocked version comparisons and their expected values
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

// Test the result of multiple version comparisons
func TestVersionComparison(t *testing.T) {
	for comparison, expectedResult := range versionComparisonsTests {
		result := CompareVersions(comparison.a, comparison.b)
		if result != expectedResult {
			t.Errorf("Invalid version comparison result: %s %s %s (expected: %s)", comparison.a, result, comparison.b, expectedResult)
		}
	}
}
