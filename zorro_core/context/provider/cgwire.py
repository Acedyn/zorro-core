from typing import List, cast, Optional, Dict, Tuple, Union, Sequence as CovariantList
from datetime import datetime, timedelta
from collections import defaultdict

import gazu
from timecode import Timecode

from zorro_core.context.entities import (
    Decor,
    Entity,
    Project,
    Episode,
    Sequence,
    Shot,
    Range,
)
from zorro_core.context.provider.provider import ContextProvider
from zorro_core.utils.logger import logger


class CgwireProvider(ContextProvider):
    # How much time do we consider that we need to fetch again the data from the database
    CACHE_PERIOD = timedelta(minutes=5)
    # The frame rate should be specified on the project or episode's config on the database
    # this value only serves as a failback in case it is not specified
    DEFAULT_FPS = 25

    def __init__(
        self,
        project_name: str,
        episode_name: Optional[str],
        selection: Optional[List[str]] = None,
        host: Optional[str] = None,
        login: Optional[str] = None,
        password: Optional[str] = None,
    ):
        gazu.set_host(host)
        gazu.log_in(login, password)

        self.project_name = project_name
        self.episode_name = episode_name
        self.selection = selection or []
        self.cache: Dict[
            str,
            Dict[str, Tuple[Optional[Union[Entity, CovariantList[Entity]]], datetime]],
        ] = defaultdict(dict)

    def is_entity_selected(self, entity: Entity) -> bool:
        # We consider that when nothing is specified as selected we selected everything
        if not self.selection:
            return True
        # If any parent is selected then this child is selected too
        if any(
            parent.id not in self.selection for parent in entity.get_parent_hierarchy()
        ):
            return True
        if any(self.is_entity_selected(child) for child in entity.children):
            return True
        return False

    def store_cache(
        self,
        entity_type: str,
        entity_name: str,
        value: Optional[Union[Entity, CovariantList[Entity]]],
    ) -> None:
        """Store the value with a timestamp into the cache"""

        self.cache[entity_type][entity_name] = (value, datetime.now())

    def clear_cached(
        self, entity_type: Optional[str] = None, entity_name: Optional[str] = None
    ) -> None:
        """Clear the specified entry on the cache, all values are clear if not entries are specified"""

        if entity_type is None:
            self.cache = defaultdict(dict)
            return
        if entity_name is None:
            del self.cache[entity_type]
            return

        del self.cache[entity_type][entity_name]

    def get_cached(
        self, entity_type: str, entity_name: str
    ) -> Optional[Union[Entity, CovariantList[Entity]]]:
        """Get the value is present on the cache"""

        entry = self.cache[entity_type].get(entity_name)
        if entry is None:
            return None
        if entry[1] - datetime.now() > self.CACHE_PERIOD:
            return None
        return entry[0]

    def get_project(self) -> Optional[Project]:
        """Get the selected project from the cache or cgwire database"""

        cached_project = self.get_cached("project", self.project_name)
        if cached_project is None:
            return self.fetch_project()

        return cast(Project, cached_project)

    def fetch_project(self) -> Optional[Project]:
        """Get the selected project from the cgwire database"""

        cgwire_project = cast(dict, gazu.project.get_project_by_name(self.project_name))
        if not cgwire_project:
            logger.error(
                "Project %s does not exists on cgwire database", self.project_name
            )
            return None

        project = Project(
            cgwire_project["id"],
            cgwire_project["name"],
            (cgwire_project["data"] or {}).get("config", {}),
            None,
            [],
            cgwire_project["data"] or {},
        )
        project.children = self.get_episodes(project) or []
        return project

    def get_episodes(self, parent: Project) -> Optional[List[Episode]]:
        """Get the selected episode from the cache or cgwire database"""

        cached_episode = self.get_cached(
            "episode", f"{self.project_name}-{self.episode_name}"
        )
        if cached_episode is None:
            fetched_episodes = self.fetch_episodes(parent)
            self.store_cache(
                "episode", f"{self.project_name}-{self.episode_name}", fetched_episodes
            )
            return fetched_episodes

        return cast(List[Episode], cached_episode)

    def fetch_episodes(self, parent: Project) -> Optional[List[Episode]]:
        """Get the selected episode from the cgwire database"""

        # The user might not want to work on any episode, to run some global actions
        # on the project
        if self.episode_name is None:
            return []

        cgwire_episode = cast(
            dict, gazu.shot.get_episode_by_name(parent.id, self.episode_name)
        )
        if not cgwire_episode:
            logger.error(
                "Episode %s does not exists on cgwire database", self.episode_name
            )
            return None

        episode = Episode(
            cgwire_episode["id"],
            cgwire_episode["name"],
            (cgwire_episode["data"] or {}).get("config", {}),
            parent,
            [],
            {**parent.raw_data, **(cgwire_episode["data"] or {})},
        )
        episode.children = self.get_sequences(episode) or []

        # This provider does not support working on multiple episode at the same time
        return [episode]

    def get_sequences(self, parent: Episode) -> Optional[List[Sequence]]:
        """Get the selected sequence from the cache or cgwire database"""

        cached_sequence = self.get_cached(
            "sequences", f"{self.project_name}-{self.episode_name}-{parent.name}"
        )
        if cached_sequence is None:
            fetched_sequences = self.fetch_sequences(parent)
            self.store_cache(
                "sequences",
                f"{self.project_name}-{self.episode_name}-{parent.name}",
                fetched_sequences,
            )
            return fetched_sequences

        return cast(List[Sequence], cached_sequence)

    def fetch_sequences(self, parent: Episode) -> Optional[List[Sequence]]:
        """Get the selected sequence from the cgwire database"""

        sequences = []
        cgwire_sequences = cast(dict, gazu.shot.all_sequences_for_episode(parent.id))
        for cgwire_sequence in cgwire_sequences:
            if cgwire_sequence["canceled"]:
                continue
            sequence = Sequence(
                cgwire_sequence["id"],
                cgwire_sequence["name"],
                (cgwire_sequence["data"] or {}).get("config", {}),
                parent,
                [],
                {**parent.raw_data, **(cgwire_sequence["data"] or {})},
            )
            sequence.children = self.get_shots(sequence) or []

            if self.is_entity_selected(sequence):
                sequences.append(sequence)

        return sequences

    def get_shots(self, parent: Sequence) -> Optional[List[Shot]]:
        """Get the selected shot from the cache or cgwire database"""

        cached_shot = self.get_cached(
            "shots", f"{self.project_name}-{self.episode_name}-{parent.name}"
        )
        if cached_shot is None:
            fetched_shots = self.fetch_shots(parent)
            self.store_cache(
                "shots",
                f"{self.project_name}-{self.episode_name}-{parent.name}",
                fetched_shots,
            )
            return fetched_shots

        return cast(List[Shot], cached_shot)

    def fetch_shots(self, parent: Sequence) -> Optional[List[Shot]]:
        """Get the selected shot from the cgwire database"""

        shots = []
        cgwire_shots = cast(dict, gazu.shot.all_shots_for_sequence(parent.id))
        for cgwire_shot in cgwire_shots:
            if cgwire_shot["canceled"]:
                continue
            merged_shot_data = {**parent.raw_data, **(cgwire_shot["data"] or {})}

            # The data field of the range must have specific data
            required_shot_datas = [
                "pelure_video",
                "previz_video",
                "source_video",
                "decor",
                "level",
                "template",
                "subdec",
            ]
            missing_data = [
                required_data
                for required_data in required_shot_datas
                if required_data not in merged_shot_data
            ]
            if missing_data:
                logger.warning(
                    "Shot %s of sequence %s is missing required data: %s",
                    cgwire_shot["name"],
                    parent.name,
                    missing_data,
                )
                continue

            decor = self.get_decor(
                merged_shot_data["decor"],
                merged_shot_data["subdec"],
                merged_shot_data["level"],
                merged_shot_data["template"],
            )
            if decor is None:
                logger.error("Could not resolve decor of shot %s", cgwire_shot)
                continue

            shot = Shot(
                cgwire_shot["id"],
                cgwire_shot["name"],
                (cgwire_shot["data"] or {}).get("config", {}),
                parent,
                [],
                merged_shot_data,
                decor,
                merged_shot_data["pelure_video"],
                merged_shot_data["previz_video"],
                merged_shot_data["source_video"],
            )
            shot.children = self.get_ranges(shot) or []

            if self.is_entity_selected(shot):
                shots.append(shot)

        return shots

    def get_ranges(self, parent: Shot) -> Optional[List[Range]]:
        """Get the selected range from the cache or cgwire database"""

        cached_range = self.get_cached(
            "ranges", f"{self.project_name}-{self.episode_name}-{parent.name}"
        )
        if cached_range is None:
            fetched_ranges = self.fetch_ranges(parent)
            self.store_cache(
                "ranges",
                f"{self.project_name}-{self.episode_name}-{parent.name}",
                fetched_ranges,
            )
            return fetched_ranges

        return cast(List[Range], cached_range)

    def fetch_ranges(self, parent: Shot) -> Optional[List[Range]]:
        """Get the selected range from the cgwire database"""

        ranges = []
        cgwire_ranges = cast(dict, gazu.shot.all_ranges_for_shot(parent.id))
        for cgwire_range in cgwire_ranges:
            if cgwire_range["canceled"]:
                continue
            merged_range_data = {**parent.raw_data, **(cgwire_range["data"] or {})}

            # The data field of the range must have specific data
            required_range_datas = [
                "timecode_source_in",
                "timecode_source_out",
                "timecode_edit_in",
                "timecode_edit_out",
            ]
            missing_data = [
                required_data
                for required_data in required_range_datas
                if required_data not in merged_range_data
            ]
            if missing_data:
                logger.warning(
                    "Range %s of shot %s is missing required data: %s",
                    cgwire_range["name"],
                    parent.name,
                    missing_data,
                )
                continue

            frame_rate_config = parent.resolve_config_entry("frame_rate")
            if frame_rate_config is None:
                logger.warning(
                    "Could not find frame rate config for entity %s: Using default value %s",
                    parent,
                    self.DEFAULT_FPS,
                )
            frame_rate = (
                int(frame_rate_config)
                if frame_rate_config
                and (isinstance(frame_rate_config, int) or frame_rate_config.isdigit())
                else self.DEFAULT_FPS
            )

            range = Range(
                cgwire_range["id"],
                cgwire_range["name"],
                (cgwire_range["data"] or {}).get("config", {}),
                parent,
                [],
                merged_range_data,
                Timecode(frame_rate, merged_range_data["timecode_edit_in"]),
                Timecode(frame_rate, merged_range_data["timecode_edit_out"]),
                Timecode(frame_rate, merged_range_data["timecode_source_in"]),
                Timecode(frame_rate, merged_range_data["timecode_source_out"]),
            )

            if self.is_entity_selected(range):
                ranges.append(range)

        return ranges

    def get_decor(
        self, name: str, subdecor: str, level: str, template: str
    ) -> Optional[Decor]:
        return Decor("", name, {}, None, [], {}, "", subdecor, level, template)
