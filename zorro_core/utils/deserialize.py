import json
from pathlib import Path
from typing import TypeVar, Optional, Type

import aiofiles
from pydantic import BaseModel, TypeAdapter

T = TypeVar("T", bound=BaseModel)


def load_model(data: dict, model: Type[T]) -> Optional[T]:
    """Abstraction layer on top if pydantic's validate_python"""
    return TypeAdapter(model).validate_python(data)


async def load_file(path: Path) -> Optional[dict]:
    """
    Load a json, toml or a yaml file.
    """
    if not path.exists():
        return None

    loaded_data = {}
    if path.suffix in [".yml", ".yaml"]:
        pass
    elif path.suffix in [".json"]:
        async with aiofiles.open(path) as config:
            loaded_data = json.loads(await config.read())
    elif path.suffix in [".toml"]:
        pass
    else:
        return None

    return loaded_data


async def load_model_from_file(path: Path, model: Type[T]) -> Optional[T]:
    """
    Load a json or a yaml file into its dataclass using pydantic schema.
    """
    data = await load_file(path)
    if data is None:
        return None

    return load_model(data, model)


async def patch_model_from_file(source: T, path: Path, model: Type[T]) -> Optional[T]:
    """
    Load a json or a yaml file into its dataclass using pydantic schema.
    """
    data = await load_file(path)
    if data is None:
        return None

    return load_model(source.model_dump() | data, model)
