from pathlib import Path

import pytest

from zorro_core.context.resolver import resolve_plugins, get_all_plugin_versions, get_matching_plugin_versions
from zorro_core.main.config import PluginConfig


@pytest.mark.asyncio
async def test_plugin_search():
    plugin_paths = [
        Path(__file__).parent / "mock"
    ]
    config = PluginConfig(plugin_paths=plugin_paths)

    assert len(await get_all_plugin_versions("foo", config)) == 3

    plugin_versions = await get_matching_plugin_versions("foo<=3.5", config)

@pytest.mark.asyncio
async def test_plugin_resolution():
    plugin_paths = [
        Path(__file__).parent / "mock"
    ]
    config = PluginConfig(plugin_paths=plugin_paths)

    query = "foo>=3.0.3 bar==4.0 baz<=5.6"
    await resolve_plugins(query, config)
