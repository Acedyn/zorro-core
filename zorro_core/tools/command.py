from dataclasses import dataclass, field
from typing import List, Callable

from .base import ToolBase, ToolType
from zorro_core.utils.logger import logger


@dataclass
class ClientResolver:
    key: str = field(default="")


@dataclass
class Command(ToolBase):
    """
    A command is a task that will be sent to a client to be executed.
    """

    logs: List[str] = field(default_factory=list)
    client: ClientResolver = field(default_factory=ClientResolver)

    def __post_init__(self):
        super().__post_init__()
        self.type = ToolType.COMMAND

    async def traverse(self, callable: Callable[[ToolBase], None]):
        logger.debug("Traversing command %s:%s with %s", self.name, self.id, callable)
