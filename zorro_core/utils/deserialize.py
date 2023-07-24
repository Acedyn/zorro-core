from pathlib import Path
import json
from typing import TypeVar, cast, Optional, Type

import aiofiles
from marshmallow import Schema

T = TypeVar("T")

async def load_from_schema(path: Path, schema: Schema, _: Type[T]) -> Optional[T]:
    """
    Load a json or a yaml file into its dataclass using marshmallow schema.
    """
    if not path.exists():
        return None

    # YAML and JSON are supported
    loaded_config = {}
    if path.suffix in [".yml", ".yaml"]:
        pass
    elif path.suffix in [".json"]:
        async with aiofiles.open(path) as config:
            loaded_config = json.loads(await config.read())
    else:
        return None

    return cast(T, schema.load(loaded_config))
