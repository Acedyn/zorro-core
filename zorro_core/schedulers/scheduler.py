from __future__ import annotations
from abc import ABC, abstractmethod
from typing import TYPE_CHECKING
from uuid import UUID

from pydantic import BaseModel, Field

if TYPE_CHECKING:
    from zorro_core.tools.command import Command
    from zorro_core.context.context import Context

scheduled_commands: dict[UUID, Command] = {}


class Scheduler(ABC, BaseModel):
    """
    A scheduler is used to send commands requests.
    Some scheduler can send commands to the local computer, to
    render farm managers, or send command remotely.
    """

    # Static variable: variables starting with a _ are skipped by pydantic
    _scheduled_commands: list[Command] = []

    type: str = Field(default=None)

    @abstractmethod
    async def schedule_command(self, command: Command, context: Context):
        pass

    @property
    def scheduled_commands(cls):
        return scheduled_commands
