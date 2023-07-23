from __future__ import annotations
from dataclasses import dataclass, field
from typing import (
    Dict,
    Union,
    Callable,
    Optional,
    Coroutine,
    Set,
)
import asyncio

import marshmallow_dataclass

from .tool_base import LayeredTool, ToolType
from .command import Command
from zorro_core.utils.logger import logger

ActionChild = Union["Action", "ActionCommand"]

@dataclass
class ActionCommand(Command):
    """
    Commands that are linked to an action should have an upstream field
    used to declare the dependencies
    """

    upstream: Optional[str] = field(default=None)

    async def traverse(self, task: Callable[[ActionChild], Coroutine]):

        logger.debug("Traversing %s with %s", self, callable)
        await task(self)

@dataclass
class Action(LayeredTool):
    """
    An action holds groups of subactions and commands.
    It allows to chain and organize multiple commands into a dependency graph
    """

    children: Dict[str, ActionChild] = field(default_factory=dict, repr=False)
    upstream: Optional[str] = field(default=None)

    def __post_init__(self):
        super().__post_init__()
        self.type = ToolType.ACTION

    async def _traverse_ready_children(
        self,
        pending: Set[str],
        completed: Set[str],
        task: Callable[[ActionChild], Coroutine],
    ) -> Set[asyncio.Task]:
        """
        Find and traverse children that have all their dependencies (upstream)
        completed.
        """

        asyncio_tasks: Set[asyncio.Task] = set()
        for child_key, child in self.children.items():
            # Skip already process children
            if child_key not in pending:
                continue

            # Process children that don't have dependencies or that have
            # completed dependencies
            if child.upstream is None or child.upstream in completed:
                # We need the task to return it's key so we can mark it
                # as completed once done
                async def wrapped_task(child_key: str, child: Union[Action, ActionCommand]):
                    await child.traverse(task)
                    return child_key

                pending.remove(child_key)
                asyncio_tasks.add(asyncio.create_task(wrapped_task(child_key, child)))

        return asyncio_tasks

    async def traverse(self, task: Callable[[ActionChild], Coroutine]):
        """
        Run the task to all the children, respecting the order of execution
        and dependencies. Multiple might can run concurently.
        """

        logger.debug("Traversing %s with %s", self, callable)
        # We first traverse this action before traversing its children
        await task(self)

        # At first, all the children are pending
        pending = set([child_key for child_key in self.children.keys()])
        completed: Set[str] = set()

        # Loop over task results until there is not running tasks
        running_tasks = await self._traverse_ready_children(
            pending, completed, task
        )
        while running_tasks:
            # We wait for any task to be completed
            tasks_completed, task_left = await asyncio.wait(
                running_tasks, return_when=asyncio.FIRST_COMPLETED
            )
            # Mark completed tasks as completed
            for task_completed in tasks_completed:
                try:
                    completed.add(task_completed.result())
                except asyncio.CancelledError:
                    logger.error("An error occured")

            # Once a task is completed, new children might be ready for execution
            new_tasks = await self._traverse_ready_children(pending, completed, task)
            running_tasks = task_left.union(new_tasks)

    async def _execute_child(self, child: ActionChild):
        if isinstance(child, ActionCommand):
            await child.execute()

    async def execute(self):
        logger.debug("Executing %s with %s", self.name, callable)
        await self.traverse(self._execute_child)

    async def _cancel_child(self, child: ActionChild):
        if isinstance(child, ActionCommand):
            await child.cancel()

    async def cancel(self):
        logger.debug("Canceling %s with %s", self.name, callable)
        await self.traverse(self._cancel_child)

ActionSchema = marshmallow_dataclass.class_schema(Action)()
