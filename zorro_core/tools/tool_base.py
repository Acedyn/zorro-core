from __future__ import annotations
from uuid import uuid4, UUID
from typing import Optional, Any
from enum import IntEnum
from abc import ABC, abstractmethod

from pydantic import BaseModel, Field


class ToolStatus(IntEnum):
    INITIALIZING = 1    # When the class is being instanciated
    INITIALIZED = 2     # The tool is ready to start
    RUNNING = 3         # The tool is running
    PAUSED = 4          # The execution of the tool has been paused
    ERROR = 5           # An error occured during the execution, the tool has stopped
    INVALID = 6         # The tool definition is invalid or similar errors


class Socket(BaseModel):
    raw: str | int | float | list | dict
    cast: str
    id: UUID = Field(default_factory=uuid4)


class ToolBase(BaseModel, ABC):
    """
    Every interacion with zorro is done using tools.
    There is multiple types of tools, they are defined using
    a config and used to expose custom functionalities
    """

    name: str
    type: str = Field(default=None)
    id: UUID = Field(default_factory=uuid4)
    label: str = Field(default="", repr=False)
    status: ToolStatus = Field(default=ToolStatus.INITIALIZING)
    inputs: dict[str, Socket] = Field(default_factory=dict, repr=False)
    output: dict[str, Socket] = Field(default_factory=dict, repr=False)
    hidden: bool = Field(default=False, repr=False)
    tooltip: str = Field(default="No tooltip available", repr=False)
    logs: list[str] = Field(default_factory=list, repr=False)

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

    inherit: list[str] = Field(default_factory=list, repr=False)
