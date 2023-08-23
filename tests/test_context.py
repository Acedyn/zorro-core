import os
from pathlib import Path
from zorro_core.context.context import Context
from zorro_core.context.plugin import Plugin, PluginEnv


def test_environment_build():
    plugin_a = Plugin(
        name="a",
        version="1",
        path=Path("/a/zorro-plugin.json"),
        env={
            "PATH": PluginEnv(
                append=[Path("./foo/"), Path("./bar/")], prepend=[Path("./baz/")]
            ),
            "PYTHONPATH": PluginEnv(append=[Path("./code")]),
            "DEBUG": PluginEnv(set="info"),
            "TOTO": PluginEnv(set="toto"),
        },
    )
    plugin_b = Plugin(
        name="b",
        version="1",
        path=Path("/b/zorro-plugin.json"),
        env={
            "PATH": PluginEnv(append=[Path("./foo/")], prepend=[Path("./baz/")]),
            "PYTHONPATH": PluginEnv(prepend=[Path("./code")]),
            "DEBUG": PluginEnv(set="debug"),
        },
    )
    plugin_c = Plugin(
        name="a",
        version="1",
        path=Path("/c/zorro-plugin.json"),
        env={
            "PATH": PluginEnv(
                append=[Path("./foo/")], prepend=[Path("./baz/"), Path("./bar/")]
            ),
            "TATA": PluginEnv(set="tata"),
        },
    )

    context = Context(plugins=[plugin_a, plugin_b, plugin_c])
    environment_build = context.build_environment()

    assert len(environment_build.keys()) == 5
    print(environment_build["PATH"])
    assert len(environment_build["PYTHONPATH"].split(os.pathsep)) == 2
    assert environment_build["PATH"].split(os.pathsep)[0] == Path("/c/bar").as_posix()
    assert environment_build["PATH"].split(os.pathsep)[-1] == Path("/c/foo").as_posix()
    assert environment_build["PATH"].split(os.pathsep)[2] == Path("/b/baz").as_posix()
