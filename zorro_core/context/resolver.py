from enum import Enum
from typing import Dict, Set
from collections import defaultdict

from .plugin import Plugin


class VersionOperator(Enum):
    EQUAL = "equal"
    LESS_EQUAL = "less_equal"
    MORE_EQUAL = "more_equal"


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
        raise NotImplemented


def get_all_plugin_versions(name: str) -> Set[Plugin]:
    """Find all available plugin versions in the current context"""
    raise NotImplemented


def get_matching_plugin_versions(query: str) -> Dict[str, Set[Plugin]]:
    """
    Get all available plugins that could be potential candidates to satisfy
    the query
    """

    plugin_versions = defaultdict(set)
    for single_query in query.split(" "):
        version_query = VersionQuery(single_query)

        for plugin in get_all_plugin_versions(version_query.name):
            if version_query.match(plugin):
                plugin_versions[version_query.name].add(plugin)

    return plugin_versions


def get_prefered_plugin_version(plugins: Set[Plugin]) -> Plugin:
    """
    When multiple plugin versions are potential quantidates, we use the
    lasted one of them.
    """

    raise NotImplemented


def resolve_quandidates(quandidates: Dict[str, Set[Plugin]]):
    """
    Find the right quandidates
    """

    for plugin_name, plugins in quandidates.items():
        plugin = get_prefered_plugin_version(plugins)
        quandidates[plugin_name] = set(plugin)


def resolve_plugins(query: str) -> Set[Plugin]:
    """
    Resolve a flat list of plugin that satisfies the given query
    """

    plugin_quandidates = get_matching_plugin_versions(query)

    return set()
