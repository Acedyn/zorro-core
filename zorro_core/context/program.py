from pydantic import BaseModel, Field


class Program(BaseModel):
    """
    A program defines an installed program an how to
    launch it
    """

    name: str
    subsets: list[str] = Field(default_factory=list)
    launch_template: list[str] = Field(default_factory=lambda: ["{name}"])
    launch_client_template: list[str] = Field(default_factory=lambda: ["{name}"])
