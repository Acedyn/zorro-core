from __future__ import annotations
from typing import Dict, Union, List
from dataclasses import dataclass, field

Data = Union[None, int, str, bool, List[object], Dict[str, object]]


@dataclass
class Base:
    id: str
    name: str
    label: str
    data: Data = field(repr=False)
