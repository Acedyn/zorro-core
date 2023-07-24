import asyncio
from typing import Coroutine

from .settings import get_settings
from zorro_core.db.db_manager import init_db


async def async_init_app(app_task: Coroutine):
    app_settings = await get_settings()
    await init_db(app_settings.db)

    await app_task


def init_app(app_task: Coroutine):
    asyncio.run(async_init_app(async_init_app(app_task)))
