from __future__ import annotations
from typing import Dict, Optional, Any, Sequence as CovariantList
from dataclasses import dataclass, field

@dataclass
class Entity:
    id: str
    name: str
    label: str
    parent: Optional[Entity] = field(repr=False)
    children: CovariantList[Entity] = field(repr=False)
    data: Dict[str, Any] = field(repr=False)

    def get_parent_hierarchy(self) -> CovariantList[Entity]:
        """Traverse the parent hierarchy to return a list of it"""

        parent_hierarchy = [self]
        while parent_hierarchy[0].parent is not None:
            parent_hierarchy.insert(0, parent_hierarchy[0].parent)

        return parent_hierarchy

    def get_flattended_data(self) -> Dict[str, Any]:
        """Merge all the datas of this entity hierachy into one"""

        flattened_data: Dict[str, Any] = {}
        for parent in self.get_parent_hierarchy():
            flattened_data.update(parent.data)

        return flattened_data
