from tortoise import Tortoise
from typing import Literal

from zorro_core.main.config import DBConfig


async def init_db(settigns: DBConfig):
    await Tortoise.init(
        db_url=settigns.url,
        modules={"models": ["zorro_core.db.entity", "zorro_core.db.user"]},
    )


async def migrate():
    await Tortoise.generate_schemas()


async def drop():
    await Tortoise._drop_databases()


async def cli_handler(operation: Literal["migrate", "reset", "create-admin"]):
    if operation == "migrate":
        await migrate()
