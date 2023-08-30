from grpc.aio import grpc
import pytest

from google.protobuf.empty_pb2 import Empty
from zorro_core.network.grpc_server import serve
from zorro_core.network.protos.scheduler_service_pb2 import ID
from zorro_core.network.protos.scheduler_service_pb2_grpc import CommandSchedulingStub


@pytest.mark.asyncio
async def test_grpc_server():
    server, port = await serve("localhost:0")
    async with grpc.aio.insecure_channel(f"localhost:{port}") as channel:
        stub = CommandSchedulingStub(channel)

        assert (await stub.GetCommand(ID(id=""))).name == "foo"

        commands = [c async for c in stub.GetCommandRequests(Empty())]
        assert len(commands) == 1
        assert commands[0].name == "foo"

    await server.stop(None)
