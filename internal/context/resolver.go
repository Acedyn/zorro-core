package context

import (
	"fmt"
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
func FindPluginVersions(name string, pluginConfig *config.PluginConfig) []*Plugin {
	if pluginConfig == nil {
		pluginConfig = config.AppConfig().PluginConfig
	}
	versions := []*Plugin{}

	for _, pluginSearchPath := range pluginConfig.PluginPaths {
		err := filepath.WalkDir(pluginSearchPath, func(path string, f os.DirEntry, _ error) error {
			pathStem := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			if pathStem == PLUGIN_DEFINITION_NAME {
				plugin := LoadPluginBare(path)
				if plugin.GetName() == name {
					versions = append(versions, LoadPluginBare(path))
				}
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			utils.Logger().Warn("An error occured while looking for plugins in path %s:\n\t%s", pluginSearchPath, err)
		}
	}

	return versions
}

// Get all available plugins that could be potential candidates to satisfy the query
func GetQueryMatchingPlugins(queries []string, pluginConfig *config.PluginConfig) map[string][]*Plugin {
	pluginVersion := map[string][]*Plugin{}

	for _, query := range queries {
		versionQuery := ParseVersionQuery(query)
		if _, ok := pluginVersion[versionQuery.Name]; !ok {
			pluginVersion[versionQuery.Name] = []*Plugin{}
		}

		for _, plugin := range FindPluginVersions(versionQuery.Name, pluginConfig) {
			if versionQuery.Match(plugin) {
				pluginVersion[versionQuery.Name] = append(pluginVersion[versionQuery.Name], plugin)
			}
		}
	}

	return pluginVersion
}

// When multiple plugin versions are potential quantidates, we use the
// lasted version of them.
func GetPreferedPluginVersion(versions []*Plugin) (*Plugin, int) {
	if len(versions) == 0 {
		return nil, 0
	}

	preferedIndex := 0
	for pluginIndex, plugin := range versions {
	  preferedPlugin := versions[preferedIndex]
		versionComparison := CompareVersions(plugin.GetVersion(), preferedPlugin.GetVersion())

		switch versionComparison {
		// If the version is higher the plugin take the spot of the prefered plugin
		case VersionOperator_MORE_EQUAL:
	    preferedIndex = pluginIndex
		// For equal versions the more precise one win
		case VersionOperator_EQUAL:
			splittedPreferedVersion := strings.Split(preferedPlugin.GetVersion(), VERSION_ITEM_SEPARATOR)
			splittedVersion := strings.Split(plugin.GetVersion(), VERSION_ITEM_SEPARATOR)
			if len(splittedVersion) > len(splittedPreferedVersion) {
	      preferedIndex = pluginIndex
			}
		}
	}

	return versions[preferedIndex], preferedIndex
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

// Recursive function that will select a plugin version from quandidates
// and add its dependencies to the quandidates until a valid version is found
func resolvePluginVersion(
  name string, 
  quandidates map[string][]*Plugin, 
  pluginConfig *config.PluginConfig,
) (*Plugin, map[string][]*Plugin, error) {
	versions, ok := quandidates[name]
	if len(versions) == 0 || !ok {
		return nil, nil, fmt.Errorf("no quandidates available for plugin %s", name)
	}

	// Reload the plugin to make sure it's not bare
  preferedVersion, preferedVersionIndex := GetPreferedPluginVersion(versions)
	plugin, error := LoadPluginFromFile(preferedVersion.GetPath())
	if error != nil {
		return nil, nil, fmt.Errorf("could not load plugin %s: %w", name, error)
	}

	// We don't want to keep the quandidates that does not match
	// the current requirements
	requirementQuandidates := GetQueryMatchingPlugins(plugin.GetRequire(), pluginConfig)
	newQuandidates := intersectQuandidates(quandidates, requirementQuandidates)

	// The prefered plugin's requirement might not be compatible with
	// the currently selected quandidate
	for _, quandidateVersions := range newQuandidates {
		// If a plugin does not have any quandidate versions anymore
		// we must look for a different combinason
		if len(quandidateVersions) == 0 {
      // Remove the selected version so it is not selected next time
      versions[preferedVersionIndex] = versions[len(versions) - 1]
      quandidates[name] = versions[:len(versions) - 1]
			return resolvePluginVersion(name, quandidates, pluginConfig)
		}
	}

  newQuandidates[name] = []*Plugin{plugin}
  return plugin, newQuandidates, nil
}

// Recusive function that will select a quandidates and resolve its dependencies.
// It will try every possible combinason until a valid one is fund
func resolvePluginGraph(
  quandidates map[string][]*Plugin, 
  pluginConfig *config.PluginConfig,
) (map[string][]*Plugin, error) {

  // Select the next plugin that needs to be resolved
  var pluginToResolve *string = nil
  for quandidateName, quandidateVersions := range quandidates {
    if len(quandidateVersions) > 1 {
      pluginToResolve = &quandidateName
    }
  }

  // There is not plugins to resolve anymore, the resolution is complete
  if pluginToResolve == nil {
    return nil, nil
  }

  _, newQuandidates, err := resolvePluginVersion(*pluginToResolve, quandidates, pluginConfig)

  // If no valid plugins where found, it means this path is impossible to
  // resolve and we should try an other combinason
  if err != nil {
    return nil, fmt.Errorf("no valid version found for plugin %s: invalid combinason (%w)", *pluginToResolve, err)
  }

  return newQuandidates, nil
}

func ResolvePlugins(query string, pluginConfig *config.PluginConfig) []*Plugin {
	if pluginConfig == nil {
		pluginConfig = config.AppConfig().PluginConfig
	}

  return nil
}

