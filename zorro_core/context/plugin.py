from __future__ import annotations
from pathlib import Path
from typing import Dict, List, Optional
from uuid import uuid4, UUID

from pydantic import BaseModel, Field

from zorro_core.utils.deserialize import load_from_schema


class PluginPaths(BaseModel):
    append: List[str] = Field(default_factory=list)
    prepend: List[str] = Field(default_factory=list)


class PluginTools(BaseModel):
    commands: List[str] = Field(default_factory=list)
    actions: List[str] = Field(default_factory=list)
    hooks: List[str] = Field(default_factory=list)
    widgets: List[str] = Field(default_factory=list)


class ClientConfig(BaseModel):
    name: str
    path: str


class Plugin(BaseModel):
    """
    Plugins register a set of tools, environment variables and clients.
    A set of tools will define what interactions are available or not.
    """

    id: UUID = Field(default_factory=uuid4)
    name: str
    label: str = Field(default="")
    require: List[str] = Field(default_factory=list)
    env: Dict[str, str] = Field(default_factory=dict)
    paths: PluginPaths = Field(default_factory=PluginPaths)
    tools: PluginTools = Field(default_factory=PluginTools)
    clients: List[ClientConfig] = Field(default_factory=list)

    def __post_init__(self):
        self.id = uuid4()
        if not self.label:
            self.label = self.name.replace("_", " ").title()

    @staticmethod
    async def load(path: Path) -> Optional[Plugin]:
        return await load_from_schema(path, Plugin)
