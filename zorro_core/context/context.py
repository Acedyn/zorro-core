from pydantic import BaseModel, Field

from zorro_core.context.plugin import Plugin

class Context(BaseModel):
    plugins: list[Plugin] = Field(default_factory=list)
