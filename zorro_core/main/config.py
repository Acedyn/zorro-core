from typing import List
from pydantic import BaseModel, Field


class DBConfig(BaseModel):
    url: str = Field(default="sqlite://zorro.db")


class PluginConfig(BaseModel):
    default_require: List[str] = Field(default_factory=list)
    plugin_paths: List[str] = Field(default_factory=list)


class Config(BaseModel):
    db: DBConfig = Field(default_factory=DBConfig)


async def get_config() -> Config:
    return Config()