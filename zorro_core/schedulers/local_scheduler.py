from __future__ import annotations
from typing import TYPE_CHECKING, Optional, Any

from pydantic import BaseModel, Field

from .client import Client
from .scheduler import Scheduler

if TYPE_CHECKING:
    from zorro_core.tools.command import Command
    from zorro_core.context.context import Context


class LocalClientQuery(BaseModel):
    """
    Filter set used to select one client among the available ones
    """

    program_name: Optional[str] = Field(default=None)
    client_pid: Optional[str] = Field(default=None)
    client_metadata: Optional[dict[str, str]] = Field(default=None)


class LocalScheduler(Scheduler):
    """
    The local scheduler will send the commands to clients running
    on the current computer. It can start a new client if no running
    clients matches the query
    """

    # Static variable: variables starting with a _ are skipped by pydantic
    _clients: list[Client]

    query: LocalClientQuery = Field(default_factory=LocalClientQuery)

    def __init__(self, **data: Any):
        data["type"] = "local"
        super().__init__(**data)

    async def schedule_command(self, command: Command, context: Context):
        self.scheduled_commands[command.id] = command
