from __future__ import annotations
from pathlib import Path
from dataclasses import dataclass, field
from typing import Dict, List, cast, Optional
from uuid import uuid4, UUID
import json

import marshmallow_dataclass
import aiofiles

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
class Plugin:
    id: UUID = field(init=False)
    name: str
    label: str = field(default="")
    env: Dict[str, str] = field(default_factory=dict)
    paths: PluginPaths = field(default_factory=PluginPaths)
    tools: PluginTools = field(default_factory=PluginTools)

    def __post_init__(self):
        self.id = uuid4()
        if not self.label:
            self.label = self.name.replace("_", " ").title()

    @staticmethod
    async def load(path: Path) -> Optional[Plugin]:
        if not path.exists():
            return Plugin("")

        loaded_config = {}
        if path.suffix in [".yml", ".yaml"]:
            pass
        elif path.suffix in [".json"]:
            async with aiofiles.open(path) as config:
                loaded_config = json.loads(await config.read())
        else:
            return Plugin("")

        loaded_plugin = cast(Optional[Plugin], PluginSchema.load(loaded_config))
        return loaded_plugin

PluginSchema = marshmallow_dataclass.class_schema(Plugin)()
