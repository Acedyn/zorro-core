package plugin

import (
	"strconv"
	"strings"

	"github.com/life4/genesis/slices"
)

// List of all possible version operators
type VersionOperator string

const (
	VersionOperator_EQUAL      VersionOperator = "=="
	VersionOperator_LESS_EQUAL VersionOperator = "<="
	VersionOperator_MORE_EQUAL VersionOperator = ">="

	VERSION_ITEM_SEPARATOR = "."
)

// Compaire two versions and define if the second one is less or more than the first
func CompareVersions(versionA string, versionB string) VersionOperator {
	// Handle the edge cases first
	if versionA == versionB {
		return VersionOperator_EQUAL
	} else if versionA == "" {
		return VersionOperator_LESS_EQUAL
	} else if versionB == "" {
		return VersionOperator_MORE_EQUAL
	}

	splittedVersionA := strings.Split(versionA, VERSION_ITEM_SEPARATOR)
	splittedVersionB := strings.Split(versionB, VERSION_ITEM_SEPARATOR)
	minVersionLenght, _ := slices.Min([]int{len(splittedVersionA), len(splittedVersionB)})

	// We compare the version items by items
	for index := 0; index < minVersionLenght; index++ {
		versionItemA := splittedVersionA[index]
		versionItemB := splittedVersionB[index]

		// If the item is a number the comparison is different
		versionNumberA, errA := strconv.Atoi(versionItemA)
		versionNumberB, errB := strconv.Atoi(versionItemB)

		// For number comparison, compare the values
		if errA == nil && errB == nil {
			if versionNumberA > versionNumberB {
				return VersionOperator_MORE_EQUAL
			} else if versionNumberA < versionNumberB {
				return VersionOperator_LESS_EQUAL
			}
			// For string comparison, compare aphabetically
		} else {
			switch strings.Compare(versionItemA, versionItemB) {
			case 1:
				return VersionOperator_MORE_EQUAL
			case -1:
				return VersionOperator_LESS_EQUAL
			}
		}
	}

	return VersionOperator_EQUAL
}

// Parsed version of a version query
type VersionQuery struct {
	Name     string
	Version  string
	Operator VersionOperator
}

func ParseVersionQuery(query string) *VersionQuery {
	versionQuery := VersionQuery{
		Name:     query,
		Version:  "",
		Operator: VersionOperator_EQUAL,
	}

	operators := []VersionOperator{VersionOperator_EQUAL, VersionOperator_LESS_EQUAL, VersionOperator_MORE_EQUAL}
	for _, operator := range operators {
		querySplit := strings.Split(query, string(operator))
		if len(querySplit) == 2 {
			versionQuery.Name = querySplit[0]
			versionQuery.Version = querySplit[1]
			versionQuery.Operator = operator
		}
	}

	return &versionQuery
}

// Test if the given plugin satisfies the query
func (versionQuery *VersionQuery) Match(plugin *Plugin) bool {
	versionComparison := CompareVersions(plugin.GetVersion(), versionQuery.Version)

	if versionComparison == VersionOperator_EQUAL {
		return true
	}
	if versionComparison == versionQuery.Operator {
		return true
	} else {
		return false
	}
}
