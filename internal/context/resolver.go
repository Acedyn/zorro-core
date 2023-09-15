package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Acedyn/zorro-core/internal/config"
	"github.com/Acedyn/zorro-core/internal/utils"

	"github.com/life4/genesis/slices"
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
	groupedQueries := map[string][]*VersionQuery{}
	for _, query := range queries {
		versionQuery := ParseVersionQuery(query)
		if queryGroup, ok := groupedQueries[versionQuery.Name]; ok {
			groupedQueries[versionQuery.Name] = append(queryGroup, versionQuery)
		} else {
			groupedQueries[versionQuery.Name] = []*VersionQuery{versionQuery}
		}
	}

	for pluginName, versionQueries := range groupedQueries {
		if _, ok := pluginVersion[pluginName]; !ok {
			pluginVersion[pluginName] = []*Plugin{}
		}

		for _, plugin := range FindPluginVersions(pluginName, pluginConfig) {
			isMatch := true
			for _, versionQuery := range versionQueries {
				if !versionQuery.Match(plugin) {
					isMatch = false
				}
			}
			if isMatch {
				pluginVersion[pluginName] = append(pluginVersion[pluginName], plugin)
			}
		}
	}

	return pluginVersion
}

// When multiple plugin versions are potential quantidates, we use the
// lasted version of them.
func GetPreferedPluginVersion(versions []*Plugin) *Plugin {
	if len(versions) == 0 {
		return nil
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

	return versions[preferedIndex]
}

// Create a new quandidate set with only quandidates that are present in both sets.
func intersectQuandidates(quandidatesA, quandidatesB map[string][]*Plugin) map[string][]*Plugin {
	intersectedCandidates := map[string][]*Plugin{}

	// Get all the keys into a set
	pluginsKeys := make(map[string]bool, len(quandidatesA))
	for key := range quandidatesA {
		pluginsKeys[key] = true
	}
	for key := range quandidatesB {
		pluginsKeys[key] = true
	}

	for key := range pluginsKeys {
		quandidateSetA, okA := quandidatesA[key]
		quandidateSetB, okB := quandidatesB[key]
		// The two first cases are simple: only one of the two has
		// plugins in the key so no intersections to do
		if !okA {
			intersectedCandidates[key] = quandidateSetB
		} else if !okB {
			intersectedCandidates[key] = quandidateSetA
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

// Select a plugin version from quandidates and add its dependencies to the quandidates
func resolvePluginVersion(
	name string,
	quandidates map[string][]*Plugin,
	pluginConfig *config.PluginConfig,
) (*Plugin, map[string][]*Plugin, error) {
	versions, ok := quandidates[name]
	if len(versions) <= 0 || !ok {
		return nil, nil, fmt.Errorf("no quandidates available for plugin %s", name)
	}

	// Reload the plugin to make sure it's not bare
	preferedVersion := GetPreferedPluginVersion(versions)
	plugin, error := LoadPluginFromFile(preferedVersion.GetPath())
	if error != nil {
		return preferedVersion, nil, fmt.Errorf("could not load plugin %s: %w", name, error)
	}

	// We don't want to keep the quandidates that does not match
	// the current requirements
	requirementQuandidates := GetQueryMatchingPlugins(plugin.GetRequire(), pluginConfig)
	newQuandidates := intersectQuandidates(quandidates, requirementQuandidates)

	newQuandidates[name] = []*Plugin{plugin}
	return preferedVersion, newQuandidates, nil
}

// Recusive function that will select a quandidates and resolve its dependencies.
// It will try every possible combinason until a valid one is fund
func resolvePluginGraph(
	quandidates map[string][]*Plugin,
	completed map[string]bool,
	pluginConfig *config.PluginConfig,
) (map[string][]*Plugin, error) {
	if completed == nil {
		completed = map[string]bool{}
	}

	// Select the next plugin that needs to be resolved
	var pluginToResolve *string = nil
	for quandidateName := range quandidates {
		if _, ok := completed[quandidateName]; !ok {
			pluginToResolve = &quandidateName
		}
	}

	// There is not plugins to resolve anymore, the resolution is complete
	if pluginToResolve == nil {
		return quandidates, nil
	}

	completed[*pluginToResolve] = true
	testedVersions := map[string]bool{}

	// Try to resolve a different plugin version until a valid graph is resolved
	for len(quandidates[*pluginToResolve]) > 0 {
		selectedVersion, newQuandidates, iterErr := resolvePluginVersion(*pluginToResolve, quandidates, pluginConfig)

		// Mark the selected version as tested and remove it from the quandidates
		testedVersions[selectedVersion.GetVersion()] = true
		quandidates[*pluginToResolve] = slices.Filter(quandidates[*pluginToResolve], func(plugin *Plugin) bool {
			_, ok := testedVersions[plugin.GetVersion()]
			return !ok
		})

		if iterErr != nil {
			utils.Logger().Debug(fmt.Sprintf("Skipping plugin versions: could not load plugin\n\t" + iterErr.Error()))
			continue
		}

		// First check if the resolved plugin resulted in a valid graph
		isValidGraph := true
		for pluginName, quandidateVersions := range newQuandidates {
			if len(quandidateVersions) == 0 {
				utils.Logger().Debug(fmt.Sprintf("Skipping plugin version %s: invalid resolved graph (no valid quandidates for plugin %s)", selectedVersion.GetVersion(), pluginName))
				isValidGraph = false
				break
			}
		}
		if !isValidGraph {
			continue
		}

		// Continue to resolve the graph
		if newQuandidates, err := resolvePluginGraph(newQuandidates, completed, pluginConfig); err != nil {
			utils.Logger().Debug(fmt.Sprintf("Skipping plugin versions %s: no graph combinason could be resolved", selectedVersion.GetVersion()))
			continue
		} else {
			// The resolved graph is valid
			return newQuandidates, nil
		}
	}

	return nil, fmt.Errorf("invalid graph: no valid combinason could not found")
}

// Resolve a flat list of plugin that satisfies the given query
func ResolvePlugins(query []string, pluginConfig *config.PluginConfig) ([]*Plugin, error) {
	if pluginConfig == nil {
		pluginConfig = config.AppConfig().PluginConfig
	}

	initialQuandidates := GetQueryMatchingPlugins(query, pluginConfig)
	resolvedGraph, err := resolvePluginGraph(initialQuandidates, nil, pluginConfig)
	if err != nil {
		return nil, fmt.Errorf("Plugin graph resolution imposible for query %s: %w", query, err)
	}

	resolvedPlugins := []*Plugin{}
	for _, plugins := range resolvedGraph {
		resolvedPlugins = append(resolvedPlugins, plugins[0])
	}

	return resolvedPlugins, nil
}
