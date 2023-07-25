import asyncio
from typing import Coroutine

from .config import get_config
from zorro_core.db.db_manager import init_db


async def async_init_app(app_task: Coroutine):
    app_config = await get_config()
    await init_db(app_config.db)

    await app_task


def init_app(app_task: Coroutine):
    asyncio.run(async_init_app(async_init_app(app_task)))
