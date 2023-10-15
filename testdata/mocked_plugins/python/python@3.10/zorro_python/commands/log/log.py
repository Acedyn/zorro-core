from typing import Iterable
from datetime import datetime

import grpc

from zorro_python.commands.log import log_pb2, log_pb2_grpc
from zorroprotos.tools import command_pb2, tool_pb2


class LogServicer(log_pb2_grpc.LogServicer):
    def Execute(
        self, request: log_pb2.LogInput, _: grpc.ServicerContext
    ) -> Iterable[log_pb2.LogOutput]:
        message = f"DEBUG: {request.message}"
        if request.level == log_pb2.LogLevels.INFO:
            message = f"INFO: {request.message}"
        elif request.level == log_pb2.LogLevels.WARNING:
            message = f"WARNING: {request.message}"
        elif request.level == log_pb2.LogLevels.ERROR:
            message = f"ERROR: {request.message}"
        elif request.level == log_pb2.LogLevels.CRITICAL:
            message = f"CRITICAL: {request.message}"

        timestamp = int(datetime.now().timestamp() * 1000)
        try:
            yield log_pb2.LogOutput(
                message=message,
                timestamp=timestamp,
                zorro_command=command_pb2.Command(
                    base=tool_pb2.ToolBase(logs={timestamp: message})
                ),
            )
        except Exception as e:
            import traceback

            traceback.print_exception(e)

    def Undo(
        self, request: log_pb2.LogInput, _: grpc.ServicerContext
    ) -> Iterable[log_pb2.LogOutput]:
        message = f"DEBUG: [UNDO] {request.message}"
        if request.level == log_pb2.LogLevels.INFO:
            message = f"INFO: {request.message}"
        elif request.level == log_pb2.LogLevels.WARNING:
            message = f"WARNING: {request.message}"
        elif request.level == log_pb2.LogLevels.ERROR:
            message = f"ERROR: {request.message}"
        elif request.level == log_pb2.LogLevels.CRITICAL:
            message = f"CRITICAL: {request.message}"

        timestamp = int(datetime.now().timestamp() * 1000)
        yield log_pb2.LogOutput(
            message=message,
            timestamp=timestamp,
            zorro_command=command_pb2.Command(
                base=tool_pb2.ToolBase(logs={timestamp: message})
            ),
        )


def register_zorro_commands(server: grpc.Server):
    log_pb2_grpc.add_LogServicer_to_server(LogServicer(), server)
    return [log_pb2.DESCRIPTOR.services_by_name["Log"].full_name]
