from __future__ import annotations
from abc import ABC, abstractmethod
from typing import TYPE_CHECKING

from pydantic import BaseModel, Field

if TYPE_CHECKING:
    from zorro_core.tools.command import Command
    from zorro_core.context.context import Context


class Scheduler(ABC, BaseModel):
    """
    A scheduler is used to send commands requests.
    Some scheduler can send commands to the local computer, to
    render farm managers, or send command remotely.
    """

    type: str = Field(default=None)

    @abstractmethod
    async def schedule_command(self, command: Command, context: Context):
        pass
