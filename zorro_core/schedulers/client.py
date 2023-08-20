from pathlib import Path
from enum import IntEnum
from typing import Optional

from pydantic import BaseModel, Field

class Program(BaseModel):
    """
    A program defines an installed program an how to
    launch it
    """
    name: str
    path: Path
    launch_client_template: str

class ClientStatus(IntEnum):
    INITIALIZED = 1     # When the class is instanciated but the program is not running
    STARTING = 2        # The program is starting
    IDLE = 3            # The program is running but not executing any commands
    PROCESSING = 4      # The program is running and executing commands
    SHUTTING_DOWN = 5   # The program received a shutting down command
    SHUT_DOWN = 6       # The program is now off
    NOT_RESPONDING = 7  # No ping received from the client for a certain amound of time

class Client(BaseModel):
    """
    A client is bound to a running program, with a client running to
    receive and process commands
    """
    program: Program
    pid: Optional[int] = Field(default=None)
    metadata: dict = Field(default_factory=dict)
    status: ClientStatus = Field(default=ClientStatus.INITIALIZED)
