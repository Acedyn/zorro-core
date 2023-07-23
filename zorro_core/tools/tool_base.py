from __future__ import annotations
from uuid import uuid4
from dataclasses import dataclass, field
from typing import Dict, Union, List
from enum import IntEnum
from abc import ABC, abstractmethod


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


@dataclass
class Socket:
    raw: Union[str, int, float, list, dict]
    cast: str

    def __post_init__(self):
        self.id = uuid4()


@dataclass
class ToolBase(ABC):
    """
    Every interacion with zorro is done using tools.
    There is multiple types of tools, they are defined using
    a config and used to expose custom functionalities
    """

    name: str
    type: ToolType = field(init=False)
    inputs: Dict[str, Socket] = field(default_factory=dict, repr=False)
    output: Dict[str, Socket] = field(default_factory=dict, repr=False)
    hidden: bool = field(default=False, repr=False)
    tooltip: str = field(default="No tooltip available", repr=False)
    logs: List[str] = field(default_factory=list, repr=False)

    def __post_init__(self):
        self.id = uuid4()

    @abstractmethod
    async def execute(self):
        pass

    @abstractmethod
    async def cancel(self):
        pass


@dataclass
class LayeredTool(ToolBase):
    """
    Some tools can be composed of multiple config that chained
    together will form one final config.
    """

    inherit: List[str] = field(default_factory=list, repr=False)
