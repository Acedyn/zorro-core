from __future__ import annotations
from pathlib import Path
from typing import Dict, List, Optional, Any, Union, cast, TYPE_CHECKING
from uuid import uuid4, UUID

from pydantic import BaseModel, Field

from zorro_core.utils.deserialize import patch_model_from_file

if TYPE_CHECKING:
    from zorro_core.context.resolver import VersionQuery


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
        return await patch_model_from_file(plugin, plugin.path, Plugin)

    async def as_bare(self):
        return await self.load_bare(self.path)

    async def reload(self):
        return await self.load(self.path) or self

    def __hash__(self):
        return hash((type(self), self.name, self.version, self.path, self.label))

    def __lt__(self, plugin: Union[Plugin, VersionQuery]):
        """
        The comparison is based on plugin versions
        """
        # The versions are compared parts by parts
        for self_version, plugin_version in zip(
            self.version.split("."), plugin.version.split(".")
        ):
            # The versions can either by strings (like beta, alpha) or numbers
            if self_version.isdigit() and self_version.isdigit():
                self_version = int(self_version)
                plugin_version = int(plugin_version)

            # HACK: The python typing system does not allow us to compare
            # two unions even if we are sure they will be of the same type
            self_version = cast(str, self_version)
            plugin_version = cast(str, plugin_version)

            if self_version < plugin_version:
                return True
            elif self_version > plugin_version:
                break

        # If none of the parts are lower the the other plugin version
        # this plugin is higher or equal the other
        return False

    def __le__(self, plugin: Union[Plugin, VersionQuery]):
        """
        The comparison is based on plugin version
        """
        if self < plugin:
            return True

        # The versions are compared parts by parts
        for self_version, plugin_version in zip(
            self.version.split("."), plugin.version.split(".")
        ):
            # The versions can either by strings (like beta, alpha) or numbers
            if self_version.isdigit() and self_version.isdigit():
                self_version = int(self_version)
                plugin_version = int(plugin_version)

            # HACK: The python typing system does not allow us to compare
            # two unions even if we are sure they will be of the same type
            self_version = cast(str, self_version)
            plugin_version = cast(str, plugin_version)

            if self_version != plugin_version:
                return False

        return True

    def __gt__(self, plugin: Union[Plugin, VersionQuery]):
        """
        The comparison is based on plugin version
        """
        return not self <= plugin

    def __ge__(self, plugin: Union[Plugin, VersionQuery]):
        """
        The comparison is based on plugin version
        """
        return not self < plugin
