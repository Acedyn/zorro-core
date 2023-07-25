from pathlib import Path
from typing import TypeVar, Optional, Type

import aiofiles
from pydantic import BaseModel, TypeAdapter

T = TypeVar("T", bound=BaseModel)


async def load_from_schema(path: Path, model: Type[T]) -> Optional[T]:
    """
    Load a json or a yaml file into its dataclass using pydantic schema.
    """
    if not path.exists():
        return None

    # YAML and JSON are supported
    loaded_data = ""
    if path.suffix in [".yml", ".yaml"]:
        pass
    elif path.suffix in [".json"]:
        async with aiofiles.open(path) as config:
            loaded_data = await config.read()
    else:
        return None

    return TypeAdapter(model).validate_json(loaded_data)
