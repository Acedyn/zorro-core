from abc import ABC, abstractmethod
from typing import List, Optional

from zorro_core.context.entities import Project, Episode, Sequence, Shot, Range


class ContextProvider(ABC):
    """A context provider will fill the context of data from the user input"""

    @abstractmethod
    def __init__(self) -> None:
        pass

    @abstractmethod
    def get_project(self) -> Optional[Project]:
        pass

    @abstractmethod
    def get_episodes(self, parent: Project) -> Optional[List[Episode]]:
        pass

    @abstractmethod
    def get_sequences(self, parent: Episode) -> Optional[List[Sequence]]:
        pass

    @abstractmethod
    def get_shots(self, parent: Sequence) -> Optional[List[Shot]]:
        pass

    @abstractmethod
    def get_ranges(self, parent: Shot) -> Optional[List[Range]]:
        pass
