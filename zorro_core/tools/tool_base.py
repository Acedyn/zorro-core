from __future__ import annotations
from uuid import uuid4, UUID
from typing import Dict, Union, List, Optional, Any
from enum import IntEnum
from abc import ABC, abstractmethod

from pydantic import BaseModel, Field


class ToolType(IntEnum):
    COMMAND = 1
    ACTION = 2
    HOOK = 3
    WIDGET = 4


class ToolStatus(IntEnum):
    INITIALIZED = 1
    RUNNING = 2
    PAUSED = 3
    ERROR = 4
    INVALID = 5


class Socket(BaseModel):
    raw: Union[str, int, float, list, dict]
    cast: str
    id: UUID = Field(default_factory=uuid4)


class ToolBase(BaseModel, ABC):
    """
    Every interacion with zorro is done using tools.
    There is multiple types of tools, they are defined using
    a config and used to expose custom functionalities
    """

    name: str
    type: ToolType = Field(default=None)
    id: UUID = Field(default_factory=uuid4)
    label: str = Field(default="", repr=False)
    inputs: Dict[str, Socket] = Field(default_factory=dict, repr=False)
    output: Dict[str, Socket] = Field(default_factory=dict, repr=False)
    hidden: bool = Field(default=False, repr=False)
    tooltip: str = Field(default="No tooltip available", repr=False)
    logs: List[str] = Field(default_factory=list, repr=False)

    def __init__(self, **data: Any):
        super().__init__(**data)
        if not self.label:
            self.label = self.name.replace("_", " ").title()

    @abstractmethod
    async def execute(self):
        pass

    @abstractmethod
    async def cancel(self):
        pass

    @staticmethod
    @abstractmethod
    async def resolve(name: str) -> Optional[ToolBase]:
        pass

class LayeredTool(ToolBase):
    """
    Some tools can be composed of multiple config that chained
    together will form one final config.
    """

    inherit: List[str] = Field(default_factory=list, repr=False)
