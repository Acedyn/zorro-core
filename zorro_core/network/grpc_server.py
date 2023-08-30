from typing import AsyncIterable
from uuid import UUID

from google.protobuf.empty_pb2 import Empty
import grpc

from .protos import scheduler_service_pb2_grpc, scheduler_service_pb2, command_pb2


class CommandSchedulingServicer(scheduler_service_pb2_grpc.CommandSchedulingServicer):
    async def GetCommand(
        self, request: scheduler_service_pb2.ID, context
    ) -> command_pb2.CommandRequest:
        return command_pb2.CommandRequest(name="foo")

    async def GetCommandRequests(
        self, request: Empty, context
    ) -> AsyncIterable[command_pb2.CommandRequest]:
        yield command_pb2.CommandRequest(name="foo")

    async def GetAndSendCommandUpdates(
        self, request: AsyncIterable[command_pb2.CommandUpdate], context
    ) -> AsyncIterable[command_pb2.CommandUpdate]:
        yield command_pb2.CommandUpdate(name="foo")


async def serve(address: str = "127.0.0.1:0"):
    server = grpc.aio.server()
    scheduler_service_pb2_grpc.add_CommandSchedulingServicer_to_server(
        CommandSchedulingServicer(), server
    )
    port = server.add_insecure_port(address)

    await server.start()
    return server, port
