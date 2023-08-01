from __future__ import annotations
from typing import Optional, TypeVar, Any

from pydantic import BaseModel, Field

from .tool_base import ToolBase, ToolType
from zorro_core.utils.logger import logger

T = TypeVar("T")


class ClientResolver(BaseModel):
    key: str = Field(default="")


class Command(ToolBase):
    """
    A command is a task that will be sent to a client to be executed.
    """

    client: ClientResolver = Field(default_factory=ClientResolver)

    def __init__(self, **data: Any):
        data["type"] = ToolType.COMMAND
        super().__init__(**data)

    async def execute(self):
        logger.debug("Executing %s with %s", self.name, callable)

    async def cancel(self):
        logger.debug("Canceling %s with %s", self.name, callable)

    @staticmethod
    async def resolve(name: str) -> Optional[Command]:
        return None
