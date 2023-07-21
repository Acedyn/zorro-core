from __future__ import annotations
from dataclasses import dataclass, field
from typing import (
    List,
    Dict,
    Union,
    Callable,
    Optional,
    Coroutine,
    AsyncGenerator,
    TypeVar,
    Set,
)
import asyncio

from .base import LayeredTool, ToolType, ToolBase
from .command import Command
from zorro_core.utils.logger import logger


@dataclass
class CommandChild(Command):
    """
    Commands that are linked to an action should have an upstream field
    used to declare the dependencies
    """

    upstream: Optional[str] = field(default=None)


@dataclass
class Action(LayeredTool):
    """
    An action holds groups of subactions and commands.
    It allows to chain and organize multiple commands into a dependency graph
    """

    T = TypeVar("T")

    logs: List[str] = field(default_factory=list)
    children: Dict[str, Union[Action, CommandChild]] = field(default_factory=dict)
    upstream: Optional[str] = field(default=None)

    def __post_init__(self):
        super().__post_init__()
        self.type = ToolType.ACTION

    async def _process_executable_children(
        self,
        pending: Set[str],
        completed: Set[str],
        task: Callable[[ToolBase], Coroutine[None, None, T]],
    ) -> Set[asyncio.Task[T]]:
        """
        Find and process children that have all their dependencies (upstream)
        completed.
        """

        asyncio_tasks: Set[asyncio.Task] = set()
        for child_key, child in self.children.items():
            # Skip already process children
            if child.id not in pending:
                continue
            # Process children that don't have dependencies or that have
            # completed dependencies
            if child.upstream is None or child.upstream in completed:
                pending.remove(child.id)
                asyncio_task = asyncio.create_task(task(child))
                asyncio_task.add_done_callback(lambda _: completed.add(child_key))
                asyncio_tasks.add(asyncio_task)

        return asyncio_tasks

    async def traverse(
        self, task: Callable[[ToolBase], Coroutine[None, None, T]]
    ) -> AsyncGenerator[Optional[T], None]:
        """
        Run the task to all the children, respecting the order of execution
        and dependencies. Multiple might can run concurently.
        """

        logger.debug("Traversing action %s:%s with %s", self.name, self.id, callable)

        # At first, all the children are pending
        pending = set([child_key for child_key in self.children.keys()])
        completed: Set[str] = set()

        # We wait for a task result until there is not running tasks
        running_tasks = await self._process_executable_children(
            pending, completed, task
        )
        while running_tasks:
            # As soon as a task is done, we return the result and look for
            # potentially new executable children
            results, running_tasks = await asyncio.wait(
                running_tasks, return_when=asyncio.FIRST_COMPLETED
            )
            for result in results:
                try:
                    yield result.result()
                except asyncio.CancelledError:
                    yield None

            running_tasks.union(
                await self._process_executable_children(pending, completed, task)
            )
