import argparse
import os
import socket
import uuid
from concurrent import futures

from protos.zorroprotos.processor import processor_pb2, processor_status_pb2
from protos.zorroprotos.scheduling import scheduler_pb2_grpc, scheduler_pb2
from zorro_python.logger import logger

import grpc
from grpc_reflection.v1alpha import reflection

python_processor = processor_pb2.Processor()


def create_server() -> tuple[grpc.Server, int]:
    """
    Create an initialize the gRPC server
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=100))
    port = server.add_insecure_port("[::]:")
    logger.info("Server created on port %d", port)
    return server, port


def register_commands(server: grpc.Server, processor: processor_pb2.Processor):
    """
    Fetch all the commands available in the current context
    """
    reflection.enable_server_reflection((reflection.SERVICE_NAME), server)


def register_processor(
    core_host: str,
    core_port: int,
    processor_port: int,
    processor_id: str,
    processor_status: processor_status_pb2.ProcessorStatus,
) -> processor_pb2.Processor:
    """
    Register the processor to the scheduler so it knows this processor its
    ready to accept commands
    """

    # Find the zorro core's url
    zorro_core_url = f"{core_host}:{core_port}"

    # If the zorro-core is running on a different machine we must passe our ip
    # otherwise use localhost
    processor_host = "127.0.0.1"
    current_host = socket.gethostbyname(socket.gethostname())
    if core_host != current_host or core_host in ["127.0.0.1", "localhost"]:
        processor_host = current_host
    processor_url = f"{processor_host}:{processor_port}"

    # The registration is done via a gRPC endpoint
    logger.info("Connecting to zorro core on %s", zorro_core_url)
    with grpc.insecure_channel(zorro_core_url) as channel:
        global python_processor
        logger.info("Registering processor with url %s", processor_url)
        stub = scheduler_pb2_grpc.SchedulingStub(channel)

        # Set the values of the processor to patch
        python_processor.id = processor_id
        python_processor.status = processor_status
        python_processor = stub.RegisterProcessor(
            scheduler_pb2.ProcessorRegistration(
                processor=python_processor,
                host=processor_url,
            )
        )
        logger.info("Processor %s registered", python_processor)
        return python_processor


def parse_cli():
    parser = argparse.ArgumentParser(
        prog="Python zorro processor",
        description="Start a zorro processor ready to execute commands",
    )

    parser.add_argument(
        "-i",
        "--id",
        type=str,
        default=str(uuid.uuid4()),
        help="The id is used when we are waiting for a processor to start, to recognize it from the others",
    )
    parser.add_argument(
        "--zorro-core-host",
        type=str,
        default=os.getenv("ZORRO_CORE_HOST", "127.0.0.1"),
        help="The host of the zorro core server to connect to",
    )
    parser.add_argument(
        "--zorro-core-port",
        type=int,
        default=os.getenv("ZORRO_CORE_PORT")
        if os.getenv("ZORRO_CORE_PORT", "").isdigit()
        else 9865,
        help="The port of the zorro core server to connect to",
    )
    parser.add_argument(
        "-a",
        "--asyncio",
        action="store_true",
        help="Start the asyncio processor",
    )

    return parser.parse_args()


def main():
    arguments = parse_cli()
    server, port = create_server()

    # Register the processor one first time
    processor = register_processor(
        arguments.zorro_core_host,
        arguments.zorro_core_port,
        port,
        arguments.id,
        processor_status_pb2.STARTING,
    )
    register_commands(server, processor)
    server.start()

    # Register the processor again to update its status
    processor = register_processor(
        arguments.zorro_core_host,
        arguments.zorro_core_port,
        port,
        arguments.id,
        processor_status_pb2.IDLE,
    )

    server.wait_for_termination()


if __name__ == "__main__":
    main()
