from pathlib import Path

import pytest

from zorro_core.context.resolver import (
    resolve_plugins,
    get_all_plugin_versions,
    get_matching_plugins_from_query,
)
from zorro_core.main.config import PluginConfig


@pytest.mark.asyncio
async def test_plugin_search():
    plugin_paths = [Path(__file__).parent / "mock"]
    config = PluginConfig(plugin_paths=plugin_paths)

    assert len(await get_all_plugin_versions("foo", config)) == 3
    assert len((await get_matching_plugins_from_query("foo<=3.5", config))["foo"]) == 2


plugin_resolution_queries = [
    (
        "foo>=3.0.3 bar==2.3 baz<=5.6", 
        {
            "foo": "3.2",
            "bar": "2.3",
            "baz": "3.1",
        }
    ),
    (
        "foo>=3.2 foo<=3.8", 
        {
            "foo": "3.2",
        }
    )
]

@pytest.mark.asyncio
@pytest.mark.parametrize("query,expected", plugin_resolution_queries)
async def test_plugin_resolution(query, expected):
    plugin_paths = [Path(__file__).parent / "mock"]
    config = PluginConfig(plugin_paths=plugin_paths)

    plugins = await resolve_plugins(query, config)
    assert len(plugins) == len(expected.keys())
    for plugin in plugins:
        assert plugin.name in expected.keys()
        assert plugin.version == expected[plugin.name]
