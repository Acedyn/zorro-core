from typing import Iterable

import grpc

from . import log_pb2_grpc, log_pb2
from protos.zorroprotos.tools import command_pb2


class LogServicer(log_pb2_grpc.LogServicer):
    def Execute(self, request: log_pb2.LogParameters) -> Iterable[command_pb2.Command]:
        yield request.command

    def Undo(self, request: log_pb2.LogParameters) -> Iterable[command_pb2.Command]:
        yield request.command

    def Test(self, request: log_pb2.LogParameters) -> Iterable[command_pb2.Command]:
        yield request.command


def register_zorro_commands(server: grpc.Server):
    log_pb2_grpc.add_LogServicer_to_server(LogServicer(), server)
    return [log_pb2.DESCRIPTOR.services_by_name["Log"].full_name]
