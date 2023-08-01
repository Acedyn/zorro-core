from pathlib import Path

import pytest

from zorro_core.context.plugin import Plugin


@pytest.mark.asyncio
async def test_plugin_load():
    plugin_path = Path(__file__).parent / "mock" / "foo" / "foo@3.2" / "zorro-plugin.json"
    plugin = await Plugin.load(plugin_path)

    assert isinstance(plugin, Plugin)
    assert plugin.name == "foo"
    assert plugin.version == "3.2"
