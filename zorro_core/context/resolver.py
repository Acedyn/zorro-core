from enum import Enum
from typing import Dict, Optional, Set, List
from collections import defaultdict
from copy import deepcopy

from .plugin import Plugin
from zorro_core.main.config import PluginConfig
from zorro_core.utils.logger import logger


class VersionOperator(Enum):
    EQUAL = "=="
    LESS_EQUAL = ">="
    MORE_EQUAL = "<="


class VersionQuery:
    """
    Interpretation of a version query used in cases like plugin requirement.
    The query strings can look like myplugin>=2.0.3
    """

    OPERATORS_MAPPING: Dict[VersionOperator, str] = {
        VersionOperator.EQUAL: "==",
        VersionOperator.MORE_EQUAL: ">=",
        VersionOperator.LESS_EQUAL: "<=",
    }

    def __init__(self, query: str):
        self.name = query
        self.version = ""
        self.operator = self.OPERATORS_MAPPING[VersionOperator.EQUAL]

        for operator, token in self.OPERATORS_MAPPING.items():
            if len(query_split := query.split(token)) == 2:
                self.operator = operator
                self.name, self.version = query_split

    def match(self, plugin: Plugin) -> bool:
        """Test if the given plugin satisfies the query"""

        if self.operator == VersionOperator.EQUAL:
            return self.version == plugin.version
        if self.operator == VersionOperator.LESS_EQUAL:
            return plugin <= self
        elif self.operator == VersionOperator.MORE_EQUAL:
            return plugin >= self

        logger.warning("Could not match with version query %s: Invalid operator", self)
        return False


async def get_all_plugin_versions(name: str, config: PluginConfig) -> Set[Plugin]:
    """Find all available plugin versions in the current context"""

    plugins = set()
    for plugin_search_path in config.plugin_paths:
        # TODO: Customise the glob to stop the recursion when finding a zorro-plugin.json
        for plugin_path in plugin_search_path.glob("**/zorro-plugin.json"):
            plugin = await Plugin.load_bare(plugin_path)
            if plugin.name == name:
                plugins.add(plugin)

    return plugins


async def get_matching_plugins_from_query(
    query: str, config: PluginConfig
) -> Dict[str, Set[Plugin]]:
    """
    Get all available plugins that could be potential candidates to satisfy
    the query
    """

    plugin_versions = defaultdict(set)
    for single_query in query.split(" "):
        version_query = VersionQuery(single_query)

        for plugin in await get_all_plugin_versions(version_query.name, config):
            if version_query.match(plugin):
                plugin_versions[version_query.name].add(plugin)

    return plugin_versions


def get_prefered_plugin_version(plugins: Set[Plugin]) -> Optional[Plugin]:
    """
    When multiple plugin versions are potential quantidates, we use the
    lasted version of them.
    """

    if len(plugins) == 0:
        return None

    prefered_plugin = list(plugins)[0]
    for plugin in plugins:
        if plugin > prefered_plugin:
            prefered_plugin = plugin

    return prefered_plugin


def _combine_quandidates(
    quandidates_a: Dict[str, Set[Plugin]], quandidates_b: Dict[str, Set[Plugin]]
):
    """
    Create a new quandidate set with only quandidates that are present in both sets.
    """

    combined_quandidates: Dict[str, Set[Plugin]] = {}
    for key in set([*quandidates_a.keys(), *quandidates_b.keys()]):
        if key not in quandidates_a.keys():
            combined_quandidates[key] = quandidates_b.get(key, set())
        elif key not in quandidates_b.keys():
            combined_quandidates[key] = quandidates_a.get(key, set())
        else:
            combined_quandidates[key] = quandidates_a.get(key, set()).intersection(
                quandidates_b.get(key, set())
            )

    return combined_quandidates


async def _resolve_plugin(
    plugin_name: str,
    quandidates: Dict[str, Set[Plugin]],
    config: PluginConfig,
):
    print("RESOLVE")
    print(plugin_name)
    print([i.version for i in quandidates[plugin_name]])
    plugin = get_prefered_plugin_version(quandidates[plugin_name])
    if plugin is None:
        logger.error(
            "Could not resolve plugin %s: No valid versions available", plugin_name
        )
        return None
    print(plugin)

    # Make sure the plugin is fully loaded
    loaded_plugin = await plugin.reload()

    # Add all the requirement possible quandidates
    new_quandidates = quandidates
    for requirement in loaded_plugin.require:
        requirement_quandidates = await get_matching_plugins_from_query(
            requirement, config
        )
        print("REQUIREMENT")
        print(
            [
                {key: [i.version for i in value]}
                for key, value in requirement_quandidates.items()
            ]
        )
        print("ORIGINAL")
        print(
            [
                {key: [i.version for i in value]}
                for key, value in new_quandidates.items()
            ]
        )
        # We don't want to keep the quandidates does not match
        # the current requirements
        new_quandidates = _combine_quandidates(new_quandidates, requirement_quandidates)
        print("COMBINED")
        print(
            [
                {key: [i.version for i in value]}
                for key, value in new_quandidates.items()
            ]
        )

    # The prefered plugin's requirement might not be compatible with
    # the currently selected quandidates
    if any(len(quandidates) == 0 for quandidates in new_quandidates.values()):
        return None

    quandidates.update(new_quandidates)
    quandidates[plugin_name] = set([plugin])
    return plugin


async def _resolve_next_plugin_graph_iteration(
    quandidates: Dict[str, Set[Plugin]],
    config: PluginConfig,
    completed: Optional[Set[str]] = None,
):
    """
    Recusive function that will select a quandidates and resolve its dependencies.
    It will try every possible versions of the plugin until a valid combinason is found
    """

    completed = completed or set()
    # Used in case we need to try again with a different version choice
    original_quandidates = deepcopy(quandidates)

    # Select the next plugin that needs to be resolved
    plugin_name = next(
        (
            plugin_name
            for plugin_name in quandidates.keys()
            if plugin_name not in completed
        ),
        None,
    )

    # There is not plugins to resolve anymore, the resolution is complete
    if plugin_name is None:
        return True

    resolved_plugin = await _resolve_plugin(plugin_name, quandidates, config)
    # A plugin might be required but no quandidates
    # are valid anymore, this means that the graph we
    # are trying to resolve is impossible
    if resolved_plugin is None:
        return False

    completed.add(plugin_name)

    # Continue to resolve the next plugins,
    # the result will tell us if the resolution path was possible
    while not await _resolve_next_plugin_graph_iteration(
        quandidates, config, completed
    ):
        # Remove the choice we just made from the quandidates
        # until we make a choice that result in a possible resolution
        original_quandidates[plugin_name].remove(resolved_plugin)
        quandidates.update(deepcopy(original_quandidates))
        resolved_plugin = await _resolve_plugin(plugin_name, quandidates, config)
        if resolved_plugin is None:
            return False
    else:
        # The path chosen completed the requirements
        return True


async def resolve_plugins(query: str, config: PluginConfig) -> List[Plugin]:
    """
    Resolve a flat list of plugin that satisfies the given query
    """

    plugin_quandidates = await get_matching_plugins_from_query(query, config)
    if not await _resolve_next_plugin_graph_iteration(plugin_quandidates, config):
        logger.error("Could not resolve the depencency graph for the query %s", query)
        return set()

    return [versions.pop() for versions in plugin_quandidates.values()]
