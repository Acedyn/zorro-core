from __future__ import annotations
from dataclasses import dataclass, field
from typing import Optional, Coroutine, Callable, TypeVar

from .tool_base import ToolBase, ToolType
from zorro_core.utils.logger import logger


@dataclass
class ClientResolver:
    key: str = field(default="")


@dataclass
class Command(ToolBase):
    """
    A command is a task that will be sent to a client to be executed.
    """
    T = TypeVar("T")

    client: ClientResolver = field(default_factory=ClientResolver)

    def __post_init__(self):
        super().__post_init__()
        self.type = ToolType.COMMAND

    async def execute(self):
        logger.debug("Executing %s with %s", self.name, callable)

    async def cancel(self):
        logger.debug("Canceling %s with %s", self.name, callable)

    @staticmethod
    async def resolve(name: str) -> Optional[Command]:
        return None
