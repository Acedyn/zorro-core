from __future__ import annotations
from pathlib import Path
from dataclasses import dataclass, field
from typing import Dict, List, Optional
from uuid import uuid4, UUID

import marshmallow_dataclass

from zorro_core.utils.deserialize import load_from_schema

@dataclass
class PluginPaths:
    append: List[str] = field(default_factory=list)
    prepend: List[str] = field(default_factory=list)

@dataclass
class PluginTools:
    commands: List[str] = field(default_factory=list)
    actions: List[str] = field(default_factory=list)
    hooks: List[str] = field(default_factory=list)
    widgets: List[str] = field(default_factory=list)

@dataclass
class ClientConfig:
    name: str
    path: str

@dataclass
class Plugin:
    """
    Plugins register a set of tools, environment variables and clients.
    A set of tools will define what interactions are available or not.
    """

    id: UUID = field(init=False)
    name: str
    label: str = field(default="")
    require: List[str] = field(default_factory=list)
    env: Dict[str, str] = field(default_factory=dict)
    paths: PluginPaths = field(default_factory=PluginPaths)
    tools: PluginTools = field(default_factory=PluginTools)
    clients: List[ClientConfig] = field(default_factory=list)

    def __post_init__(self):
        self.id = uuid4()
        if not self.label:
            self.label = self.name.replace("_", " ").title()

    @staticmethod
    async def load(path: Path) -> Optional[Plugin]:
        return await load_from_schema(path, PluginSchema, Plugin)

PluginSchema = marshmallow_dataclass.class_schema(Plugin)()
