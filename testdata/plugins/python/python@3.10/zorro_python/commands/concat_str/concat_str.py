from typing import Iterable

import grpc

from zorro_python.logger import logger
from zorro_python.commands.concat_str import concat_str_pb2, concat_str_pb2_grpc


class ConcatStrServicer(concat_str_pb2_grpc.ConcatStrServicer):
    def Execute(
        self, request: concat_str_pb2.ConcatStrInput, _: grpc.ServicerContext
    ) -> Iterable[concat_str_pb2.ConcatStrOutput]:
        logger.info("Executing the concat_str command")

        yield concat_str_pb2.ConcatStrOutput(
            string=request.stringA + request.stringB
        )

    def Undo(
        self, request: concat_str_pb2.ConcatStrInput, _: grpc.ServicerContext
    ) -> Iterable[concat_str_pb2.ConcatStrOutput]:
        logger.info("Undoing the concat_str command")

        yield concat_str_pb2.ConcatStrOutput()


def register_zorro_commands(server: grpc.Server):
    concat_str_pb2_grpc.add_ConcatStrServicer_to_server(ConcatStrServicer(), server)
    return [concat_str_pb2.DESCRIPTOR.services_by_name["ConcatStr"].full_name]

