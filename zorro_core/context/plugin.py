from __future__ import annotations
from pathlib import Path
from typing import Dict, List, Optional, Any
from uuid import uuid4, UUID

from pydantic import BaseModel, Field, field_validator

from zorro_core.utils.deserialize import patch_model_from_file


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

    name: str
    version: str
    id: UUID = Field(default_factory=uuid4)
    label: str = Field(default="")
    require: List[str] = Field(default_factory=list)
    env: Dict[str, str] = Field(default_factory=dict)
    paths: PluginPaths = Field(default_factory=PluginPaths)
    tools: PluginTools = Field(default_factory=PluginTools)
    clients: List[ClientConfig] = Field(default_factory=list)

    def __init__(self, **data: Any):
        super().__init__(**data)
        if not self.label:
            self.label = self.name.replace("_", " ").title()

    @field_validator("version")
    def validate_version(cls, value: str):
        """
        The version should follow the pattern <MAJOR>.<MINOR>.<PATCH>
        """

        splitted_value = value.split(".")
        if any(not value.isdigit() for value in splitted_value):
            raise ValueError(
                "Version should be composed of digits separated by dots (ex: 3.8.12)"
            )
        if len(splitted_value) > 3:
            raise ValueError(
                "Version should not have more than than 3 components (ex: 3.8.12)"
            )

        if len(splitted_value) < 2:
            splitted_value.append("0")

        if len(splitted_value) < 3:
            splitted_value.append("0")

        return ".".join(splitted_value)

    @staticmethod
    async def load_bare(path: Path) -> Plugin:
        """
        Load a minimal verison of a plugin without openning the file
        """
        splited_name = path.stem.split("@")
        name, version = splited_name if len(splited_name) == 2 else (path.stem, "0.0.0")
        return Plugin(name=name, version=version)

    @classmethod
    async def load(cls, path: Path) -> Optional[Plugin]:
        plugin = await cls.load_bare(path)
        return await patch_model_from_file(plugin, path, Plugin)
