package context

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Acedyn/zorro-core/internal/config"
	"github.com/Acedyn/zorro-core/internal/utils"
)

// List of all possible version operators
type VersionOperator string

const (
	VersionOperator_EQUAL      VersionOperator = "=="
	VersionOperator_LESS_EQUAL VersionOperator = "<="
	VersionOperator_MORE_EQUAL VersionOperator = ">="

	VERSION_ITEM_SEPARATOR = "."
	PLUGIN_DEFINITION_NAME = "zorro-plugin"
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
	minVersionLenght := int(math.Min(float64(len(splittedVersionA)), float64(len(splittedVersionB))))

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

// Test if the given plugin satisfies the query
func (versionQuery *VersionQuery) Match(plugin *Plugin) bool {
	versionComparison := CompareVersions(versionQuery.Version, plugin.GetVersion())

	if versionComparison == versionQuery.Operator {
		return true
	} else {
		return false
	}
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

// Find all available plugin versions with the given name
func GetAllPluginVersion(name string, pluginConfig *config.PluginConfig) []*Plugin {
	if pluginConfig == nil {
		pluginConfig = config.AppConfig().PluginConfig
	}
	plugins := []*Plugin{}

	for _, pluginSearchPath := range pluginConfig.PluginPaths {
		err := filepath.WalkDir(pluginSearchPath, func(path string, f os.DirEntry, _ error) error {
			pathStem := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			if pathStem == PLUGIN_DEFINITION_NAME {
				plugin := LoadPluginBare(path)
				if plugin.GetName() == name {
					plugins = append(plugins, LoadPluginBare(path))
				}
				return filepath.SkipDir
			}
			return nil
		})

		if err != nil {
			utils.Logger().Warn("An error occured while looking for plugins in path %s:\n\t%s", pluginSearchPath, err)
		}
	}

	return plugins
}

// Get all available plugins that could be potential candidates to satisfy the query
func GetQueryMatchingQuery(queries []string, pluginConfig *config.PluginConfig) map[string][]*Plugin {
	pluginVersion := map[string][]*Plugin{}

	for _, query := range queries {
		versionQuery := ParseVersionQuery(query)
		if _, ok := pluginVersion[versionQuery.Name]; !ok {
			pluginVersion[versionQuery.Name] = []*Plugin{}
		}

		for _, plugin := range GetAllPluginVersion(versionQuery.Name, pluginConfig) {
			if versionQuery.Match(plugin) {
				pluginVersion[versionQuery.Name] = append(pluginVersion[versionQuery.Name], plugin)
			}
		}
	}

	return pluginVersion
}

// When multiple plugin versions are potential quantidates, we use the
// lasted version of them.
func GetPreferedPluginVersion(plugins []*Plugin) *Plugin {
	if len(plugins) == 0 {
		return nil
	}

	preferedPlugin := plugins[0]
	for _, plugin := range plugins {
		versionComparison := CompareVersions(plugin.GetVersion(), preferedPlugin.GetVersion())

		switch versionComparison {
		// If the version is higher the plugin take the spot of the prefered plugin
		case VersionOperator_MORE_EQUAL:
			preferedPlugin = plugin
		// For equal plugins the more precise one win
		case VersionOperator_EQUAL:
			splittedPreferedVersion := strings.Split(preferedPlugin.GetVersion(), VERSION_ITEM_SEPARATOR)
			splittedVersion := strings.Split(plugin.GetVersion(), VERSION_ITEM_SEPARATOR)
			if len(splittedVersion) > len(splittedPreferedVersion) {
				preferedPlugin = plugin
			}
		}
	}

	return preferedPlugin
}

// Create a new quandidate set with only quandidates that are present in both sets.
func intersectQuandidates(quandidatesA, quandidatesB map[string][]*Plugin) map[string][]*Plugin {
	intersectedCandidates := map[string][]*Plugin{}

	// Get all the keys into a set (there is no set in go so maps are used insead)
	pluginsKeys := make(map[string]any, len(quandidatesA))
	for key := range quandidatesA {
		pluginsKeys[key] = nil
	}
	for key := range quandidatesB {
		pluginsKeys[key] = nil
	}

	for key := range pluginsKeys {
		quandidateSetA, okA := quandidatesA[key]
		quandidateSetB, okB := quandidatesB[key]
		// The two first cases are simple: only one of the two has
		// plugins in the key so no intersections to do
		if !okA {
			intersectedCandidates[key] = quandidateSetA
		} else if !okB {
			intersectedCandidates[key] = quandidateSetB
		} else {
			// Both quandidates sets have plugins, we must only keep the ones
			// That are present in both
			intersectedCandidates[key] = []*Plugin{}

			for _, quandidateA := range quandidateSetA {
				for _, quandidateB := range quandidateSetB {

					nameEqual := quandidateA.GetName() == quandidateB.GetName()
					versionEqual := quandidateA.GetVersion() == quandidateB.GetVersion()
					pathEqual := quandidateA.GetPath() == quandidateB.GetPath()
					labelEqual := quandidateA.GetLabel() == quandidateB.GetLabel()

					if nameEqual && versionEqual && pathEqual && labelEqual {
						// Since they are equal, we don't care if we put the quandidate A or B
						intersectedCandidates[key] = append(intersectedCandidates[key], quandidateA)
						continue
					}
				}
			}
		}
	}

	return intersectedCandidates
}
