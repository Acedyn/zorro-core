from pathlib import Path

import pytest

from zorro_core.context.plugin import Plugin


@pytest.mark.asyncio
async def test_plugin_load():
    plugin_path = (
        Path(__file__).parent / "mock" / "foo" / "foo@3.2" / "zorro-plugin.json"
    )
    plugin = await Plugin.load(plugin_path)

    assert isinstance(plugin, Plugin)
    assert plugin.name == "foo"
    assert plugin.version == "3.2"
    assert plugin.tools.commands == [plugin_path.parent / Path("./commands")]
    assert plugin.tools.actions == [plugin_path.parent / Path("./actions")]


def test_plugin_comparison():
    assert Plugin(name="foo", version="3.5", path="") > Plugin(
        name="foo", version="3.4", path=""
    )
    assert not Plugin(name="foo", version="2.1.4", path="") < Plugin(
        name="foo", version="2.1.4", path=""
    )
    assert Plugin(name="foo", version="2.1.4", path="") <= Plugin(
        name="foo", version="2.1.4", path=""
    )
    assert not Plugin(name="foo", version="2.1", path="") > Plugin(
        name="foo", version="2.1.4", path=""
    )
    assert Plugin(name="foo", version="3.5.alpha", path="") < Plugin(
        name="foo", version="3.5.beta", path=""
    )
    assert not Plugin(name="foo", version="1.5.prod", path="") >= Plugin(
        name="foo", version="2.5.beta", path=""
    )
