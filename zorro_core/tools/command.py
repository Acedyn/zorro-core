from __future__ import annotations
from typing import Optional, TypeVar, Any

from pydantic import Field

from zorro_core.schedulers.scheduler import Scheduler
from zorro_core.schedulers.local_scheduler import LocalScheduler

from .tool_base import ToolBase
from zorro_core.utils.logger import logger

T = TypeVar("T")


class Command(ToolBase):
    """
    A command is a task that will be sent to a client to be executed.
    """

    implementation_key: str = Field()
    scheduler: Scheduler = Field(default_factory=LocalScheduler)

    def __init__(self, **data: Any):
        data["type"] = "command"
        super().__init__(**data)

    async def execute(self):
        logger.debug("Executing %s with %s", self.name, callable)

    async def cancel(self):
        logger.debug("Canceling %s with %s", self.name, callable)

    @staticmethod
    async def resolve(name: str) -> Optional[Command]:
        return None
