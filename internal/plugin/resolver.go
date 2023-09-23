package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Acedyn/zorro-core/internal/config"
	"github.com/Acedyn/zorro-core/internal/utils"

	"github.com/life4/genesis/slices"
  config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
)

// Find all available plugin versions with the given name
func FindPluginVersions(name string, pluginConfig *config_proto.PluginConfig) []*Plugin {
	if pluginConfig == nil {
		pluginConfig = config.AppConfig().PluginConfig
	}
	versions := []*Plugin{}

	for _, pluginSearchPath := range pluginConfig.GetRepos() {
		err := filepath.WalkDir(pluginSearchPath, func(path string, f os.DirEntry, _ error) error {
			pathStem := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			if pathStem == PLUGIN_DEFINITION_NAME {
				plugin := GetPluginBare(path)
				if plugin.GetName() == name {
					versions = append(versions, GetPluginBare(path))
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
func GetQueryMatchingPlugins(queries []string, pluginConfig *config_proto.PluginConfig) map[string][]*Plugin {
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
	pluginConfig *config_proto.PluginConfig,
) (*Plugin, map[string][]*Plugin, error) {
	versions, ok := quandidates[name]
	if len(versions) <= 0 || !ok {
		return nil, nil, fmt.Errorf("no quandidates available for plugin %s", name)
	}

	// Reload the plugin to make sure it's not bare
	preferedVersion := GetPreferedPluginVersion(versions)
	plugin, error := GetPluginFromFile(preferedVersion.GetPath())
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
	pluginConfig *config_proto.PluginConfig,
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
func ResolvePlugins(query []string, pluginConfig *config_proto.PluginConfig) ([]*Plugin, error) {
	if pluginConfig == nil {
		pluginConfig = config.AppConfig().PluginConfig
	}

	initialQuandidates := GetQueryMatchingPlugins(query, pluginConfig)
	resolvedGraph, err := resolvePluginGraph(initialQuandidates, nil, pluginConfig)
	if err != nil {
		return nil, fmt.Errorf("plugin graph resolution imposible for query %s: %w", query, err)
	}

	resolvedPlugins := []*Plugin{}
	for _, plugins := range resolvedGraph {
		resolvedPlugins = append(resolvedPlugins, plugins[0])
	}

	return resolvedPlugins, nil
}
