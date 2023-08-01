from enum import Enum
from typing import Dict, Optional, Set
from collections import defaultdict
from copy import deepcopy

from .plugin import Plugin
from zorro_core.main.config import PluginConfig


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

        # The equal operator is the simplest one
        if self.operator == VersionOperator.EQUAL:
            return self.version == plugin.version

        # The less-equal and more-equal operators require to check
        # the versions parts by parts
        for query_version, plugin_version in zip(self.version.split("."), plugin.version.split(".")):
            # The versions can either by strings (like beta, alpha) or numbers
            query_version = int(query_version) if query_version.isdigit() else query_version
            plugin_version = int(plugin_version) if plugin_version.isdigit() else plugin_version

            if self.operator == VersionOperator.LESS_EQUAL:
                if not plugin_version <= query_version:
                    return False
            elif self.operator == VersionOperator.MORE_EQUAL:
                if plugin_version > query_version:
                    return True

        if self.operator == VersionOperator.LESS_EQUAL:
            # The less-equal operator must check all the versions parts
            # and return False if any of the parts is not less-equal
            return True
        elif self.operator == VersionOperator.MORE_EQUAL:
            # The more-equal returns True as soon as a major version
            # is more equal
            return False

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


async def get_matching_plugin_versions(
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


def get_prefered_plugin_version(plugins: Set[Plugin]) -> Plugin:
    """
    When multiple plugin versions are potential quantidates, we use the
    lasted one of them.
    """

    raise NotImplemented


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
    plugin_versions: Set[Plugin],
    quandidates: Dict[str, Set[Plugin]],
    config: PluginConfig,
):
    plugin = get_prefered_plugin_version(plugin_versions)
    quandidates[plugin_name] = set([plugin])

    # Make sure the plugin is fully loaded
    await plugin.load_full()

    # Add all the requirement possible quandidates
    for requirement in plugin.require:
        requirement_quandidates = await get_matching_plugin_versions(
            requirement, config
        )
        # We don't want to keep the quandidates does not match
        # the current requirements
        quandidates = _combine_quandidates(quandidates, requirement_quandidates)

    return plugin


async def resolve_next_plugin(
    quandidates: Dict[str, Set[Plugin]], config: PluginConfig, completed: Optional[Set[str]] = None
):
    """
    Recusive function that will resolve a possible quandidate for a required plugin.
    Tries every possible versions of the plugin in a specific order until a valid choice is found
    """

    completed = completed or set()
    # Used in case we need to try again with a different version choice
    original_quandidates = deepcopy(quandidates)

    plugin_name, plugin_versions = next(
        (
            (plugin_name, versions)
            for plugin_name, versions in quandidates.items()
            if plugin_name not in completed
        ),
        (None, None),
    )

    # There is not plugins to resolve anymore, the resolution is complete
    if plugin_name is None or plugin_versions is None:
        return True

    # A plugin might be required but no quandidates
    # are valid anymore, this means that the graph we
    # are trying to resolve is impossible
    if not len(plugin_versions):
        return False

    resolved_plugin = await _resolve_plugin(plugin_name, plugin_versions, quandidates, config)
    completed.add(plugin_name)

    # Continue to resolve the next plugins,
    # the result will tell us if the resolution path was possible
    while not resolve_next_plugin(quandidates, config, completed):
        # Remove the choice we just made from the quandidates
        # until we make a choice that result in a possible resolution
        original_quandidates[plugin_name].remove(resolved_plugin)
        quandidates = deepcopy(original_quandidates)
        resolved_plugin = await _resolve_plugin(
            plugin_name, plugin_versions, quandidates, config
        )
    else:
        # The path chosen completed the requirements
        return True


async def resolve_plugins(query: str, config: PluginConfig) -> Set[Plugin]:
    """
    Resolve a flat list of plugin that satisfies the given query
    """

    plugin_quandidates = await get_matching_plugin_versions(query, config)
    await resolve_next_plugin(plugin_quandidates, config)

    return set(versions.pop() for versions in plugin_quandidates.values())
