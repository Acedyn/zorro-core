import sys

from pathlib import Path
import pytest

from zorro_core.context.context import Context
from zorro_core.context.plugin import Plugin, PluginEnv
from zorro_core.schedulers.client import Client, Program


@pytest.mark.asyncio
async def test_start_client():
    python_program = Program(name="python")
    python_plugin = Plugin(
        name="python",
        path=Path(),
        version="3.10",
        programs=[python_program],
        env={"PATH": PluginEnv(append=[Path(sys.executable).parent])},
    )

    context = Context(plugins=[python_plugin])
    client = await Client.start_program(python_program, context)
    assert client.process is not None


# @pytest.mark.asyncio
# async def test_local_scheduler():
#     client_query = ClientQuery(program_name="python")
#     scheduler = LocalScheduler(query=client_query)
#     command = Command()
#     print(scheduler, command)
