from __future__ import annotations
from dataclasses import dataclass
from typing import List, Optional

from zorro_core.context.entities import Project, Episode, Sequence, Shot, Range
from zorro_core.context.provider.provider import ContextProvider
from zorro_core.utils.logger import logger


@dataclass
class Context:
    """
    Every actions must be bound to a context.
    It stores the entities and parameters of the action
    """

    project: Project
    episodes: List[Episode]
    sequences: List[Sequence]
    shots: List[Shot]
    ranges: List[Range]

    @classmethod
    def build_with_provider(cls, provider: ContextProvider) -> Optional[Context]:
        """Fetch the entities data to resolve a complete context"""
        project = provider.get_project()
        if project is None:
            logger.error(
                "Could not build context with provider %s: Project could not be resolved"
            )
            return None
        episodes = provider.get_episodes(project) or []
        sequences = [
            sequence
            for episode in episodes
            for sequence in provider.get_sequences(episode) or []
        ]
        shots = [
            shot
            for sequence in sequences
            for shot in provider.get_shots(sequence) or []
        ]
        ranges = [range for shot in shots for range in provider.get_ranges(shot) or []]

        return cls(project, episodes, sequences, shots, ranges)

    def update_with_provider(self, provider: ContextProvider) -> None:
        """Update the current context using a provider as a source"""
        pass
