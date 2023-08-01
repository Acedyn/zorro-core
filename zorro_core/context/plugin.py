from __future__ import annotations
from pathlib import Path
from typing import Dict, List, Optional, Any
from uuid import uuid4, UUID

from pydantic import BaseModel, Field

from zorro_core.utils.deserialize import patch_model_from_file


class PluginPaths(BaseModel):
    append: List[Path] = Field(default_factory=list)
    prepend: List[Path] = Field(default_factory=list)


class PluginTools(BaseModel):
    commands: List[Path] = Field(default_factory=list)
    actions: List[Path] = Field(default_factory=list)
    hooks: List[Path] = Field(default_factory=list)
    widgets: List[Path] = Field(default_factory=list)


class ClientConfig(BaseModel):
    name: str
    path: str


class Plugin(BaseModel):
    """
    Plugins register a set of tools, environment variables and clients.
    A set of tools will define what interactions are available or not.
    """

    name: str
    version: str
    path: Path
    id: UUID = Field(default_factory=uuid4)
    label: str = Field(default="")
    require: List[str] = Field(default_factory=list, repr=False)
    env: Dict[str, str] = Field(default_factory=dict, repr=False)
    paths: PluginPaths = Field(default_factory=PluginPaths, repr=False)
    tools: PluginTools = Field(default_factory=PluginTools, repr=False)
    clients: List[ClientConfig] = Field(default_factory=list, repr=False)

    def __init__(self, **data: Any):
        super().__init__(**data)
        if not self.label:
            self.label = self.name.replace("_", " ").title()
    
    def __hash__(self):
        return hash((type(self), self.name, self.version, self.path, self.label))

    async def load_full(self):
        """
        Read the config of the plugin and parse it
        """

        # TODO: Currently this function returns a new plugin rather than
        # updating the current one
        if self == await Plugin.load_bare(self.path):
            return
        await patch_model_from_file(self, self.path, Plugin)

    @staticmethod
    async def load_bare(path: Path) -> Plugin:
        """
        Load a minimal verison of a plugin without openning the file
        """
        splited_name = path.parent.name.split("@")
        name, version = splited_name if len(splited_name) == 2 else (path.stem, "0.0.0")
        return Plugin(name=name, version=version, path=path)

    @classmethod
    async def load(cls, path: Path) -> Optional[Plugin]:
        plugin = await cls.load_bare(path)
        await plugin.load_full()
        return plugin
