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


@pytest.mark.asyncio
async def test_plugin_resolution():
    plugin_paths = [Path(__file__).parent / "mock"]
    config = PluginConfig(plugin_paths=plugin_paths)

    query = "foo>=3.0.3 bar==2.3 baz<=5.6"
    plugins = await resolve_plugins(query, config)
    assert len(plugins) == 3
    for plugin in plugins:
        if plugin.name == "foo":
            assert plugin.version == "3.2"
        if plugin.name == "bar":
            assert plugin.version == "2.3"
        if plugin.name == "baz":
            assert plugin.version == "3.1"
