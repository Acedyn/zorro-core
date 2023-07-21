from __future__ import annotations
from typing import Dict, Optional, Any, Sequence as CovariantList
from dataclasses import dataclass, field

from timecode import Timecode

Config = Dict[str, str]


@dataclass
class Entity:
    id: str
    name: str
    config: Config = field(repr=False)
    parent: Optional[Entity] = field(repr=False)
    children: CovariantList[Entity] = field(repr=False)
    raw_data: Dict[str, Any] = field(repr=False)

    def get_parent_hierarchy(self) -> CovariantList[Entity]:
        """Traverse the parent hierarchy to return a list of it"""

        parent_hierarchy = [self]
        while parent_hierarchy[0].parent is not None:
            parent_hierarchy.insert(0, parent_hierarchy[0].parent)

        return parent_hierarchy

    def get_flattended_config(self) -> Config:
        """Merge all the configs of this entity hierachy into one"""

        flattened_config: Config = {}
        for parent in self.get_parent_hierarchy():
            flattened_config.update(parent.config)

        return flattened_config

    def resolve_config_entry(
        self, entry: str, extra_config: Optional[Config] = None
    ) -> Optional[str]:
        """Find and resolve by interpolation the value of the specified entrie in the flattened config"""

        flattened_config = self.get_flattended_config()
        flattened_config.update(extra_config or {})
        raw_value = flattened_config.get(entry)
        if raw_value is None:
            return None

        # The interpolation must be done recursively because config entries are interpolated with
        # other config entries
        previous_value = raw_value
        interpolated_value = raw_value.format(**flattened_config)
        while previous_value != interpolated_value:
            interpolated_value = raw_value.format(**flattened_config)

        return interpolated_value


@dataclass
class Movie(Entity):
    path: str
    frame_rate: int
    timecode_start: Timecode
    timecode_end: Timecode


@dataclass
class Decor(Entity):
    path: str
    subdecor: str
    template: str
    level: str


@dataclass
class Project(Entity):
    children: CovariantList[Episode] = field(repr=False)


@dataclass
class Episode(Entity):
    children: CovariantList[Sequence] = field(repr=False)


@dataclass
class Sequence(Entity):
    children: CovariantList[Shot] = field(repr=False)


@dataclass
class Shot(Entity):
    children: CovariantList[Range] = field(repr=False)
    decor: Decor
    pelure_video: Movie
    previz_video: Movie
    source_video: Movie


@dataclass
class Range(Entity):
    timecode_edit_in: Timecode = field(default=Timecode(25, "00:00:00:00"))
    timecode_edit_out: Timecode = field(default=Timecode(25, "00:00:00:00"))
    timecode_source_in: Timecode = field(default=Timecode(25, "00:00:00:00"))
    timecode_source_out: Timecode = field(default=Timecode(25, "00:00:00:00"))
