from __future__ import annotations
from uuid import uuid4
from dataclasses import dataclass, field
from typing import Dict, Union, Callable, Awaitable, TypeVar, List, AsyncGenerator
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

    T = TypeVar("T")

    name: str
    type: ToolType
    inputs: Dict[str, Socket] = field(default_factory=dict)
    output: Dict[str, Socket] = field(default_factory=dict)
    hidden: bool = field(default=False)
    tooltip: str = field(default="No tooltip available")

    def __post_init__(self):
        self.id = uuid4()

    @abstractmethod
    async def traverse(
        self, task: Callable[[ToolBase], Awaitable[T]]
    ) -> AsyncGenerator[T, None]:
        pass


@dataclass
class LayeredTool(ToolBase):
    """
    Some tools can be composed of multiple config that chained
    together will form one final config.
    """

    inherit: List[str] = field(default_factory=list)
