import asyncio
from asyncio.subprocess import Process
from pathlib import Path
from enum import IntEnum
from typing import Optional

from pydantic import BaseModel, ConfigDict, Field

from zorro_core.context.context import Context


class Program(BaseModel):
    """
    A program defines an installed program an how to
    launch it
    """

    name: str
    path: Path = Field(default=Path())
    launch_client_template: list[str] = Field(default_factory=lambda: ["{name}"])

    async def start_as_client(self, context: Context, metadata: Optional[dict] = None):
        """
        Start the program and return a client as a handle for it
        """

        metadata = metadata or {}
        client = Client(program=self, context=context, metadata=metadata)

        command = [
            key.format({**metadata, "name": self.name, "path": self.path})
            for key in self.launch_client_template
        ]
        process = await asyncio.create_subprocess_exec(*command, env=context.build_environment())

        client.process = process
        client.pid = process.pid
        return client


class ClientStatus(IntEnum):
    INITIALIZED = 1  # When the class is instanciated but the program is not running
    STARTING = 2  # The program is starting
    IDLE = 3  # The program is running but not executing any commands
    PROCESSING = 4  # The program is running and executing commands
    SHUTTING_DOWN = 5  # The program received a shutting down command
    SHUT_DOWN = 6  # The program is now off
    NOT_RESPONDING = 7  # No ping received from the client for a certain amound of time


class Client(BaseModel):
    """
    A client is bound to a running program, with a client running to
    receive and process commands
    """
    model_config = ConfigDict(arbitrary_types_allowed=True)

    program: Program
    context: Context
    pid: Optional[int] = Field(default=None)
    process: Optional[Process] = Field(default=None, exclude=True)
    metadata: dict = Field(default_factory=dict)
    status: ClientStatus = Field(default=ClientStatus.INITIALIZED)
