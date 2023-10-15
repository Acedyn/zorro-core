import argparse
import os
import socket
import importlib.util
from typing import List
import uuid
from pathlib import Path
from concurrent import futures

from zorroprotos.processor import processor_pb2, processor_status_pb2
from zorroprotos.scheduling import scheduler_pb2_grpc, scheduler_pb2
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


def discover_commands(server: grpc.Server, commands: List[Path]):
    """
    Fetch all the commands available in the current context
    """
    service_names = [reflection.SERVICE_NAME]

    for command_path in commands:
        logger.info("Registering service at path %s", command_path)
        if not command_path.is_file():
            logger.error(
                "Could not load command at path %s: invalid path",
                command_path.as_posix(),
            )
            continue

        module_spec = importlib.util.spec_from_file_location(
            f"zorro_python.commands.{command_path.stem}", command_path.as_posix()
        )
        if module_spec is None:
            logger.error(
                "Could not load command at path %s: invalid module",
                command_path.as_posix(),
            )
            continue

        module = importlib.util.module_from_spec(module_spec)
        if module_spec.loader is None:
            logger.error(
                "Could not load command at path %s: invalid module loader",
                command_path.as_posix(),
            )
            continue

        module_spec.loader.exec_module(module)
        if not hasattr(module, "register_zorro_commands"):
            logger.error(
                'Could not load command at path %s: missing "register_zorro_commands" function',
                command_path.as_posix(),
            )
            continue

        for service_name in module.register_zorro_commands(server):
            service_names.append(service_name)
            logger.info("Service %s registered", service_name)

        reflection.enable_server_reflection(service_names, server)


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
        "-c",
        "--commands",
        type=Path,
        nargs="*",
        help="List of path to look for commands",
    )
    parser.add_argument(
        "--zorro-core-host",
        type=str,
        default=os.getenv("ZORRO_GRPC_CORE_HOST", "127.0.0.1"),
        help="The host of the zorro core server to connect to",
    )
    parser.add_argument(
        "--zorro-core-port",
        type=int,
        default=os.getenv("ZORRO_GRPC_CORE_PORT", "9865"),
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
    discover_commands(server, arguments.commands)

    server.start()

    # Register the processor in idle mode
    try:
        register_processor(
            arguments.zorro_core_host,
            arguments.zorro_core_port,
            port,
            arguments.id,
            processor_status_pb2.IDLE,
        )
    except grpc.RpcError as e:
        logger.error("Could not register rpc server to a zorro-core instance: %s", e)

    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Stopping gRPC server")


if __name__ == "__main__":
    main()
