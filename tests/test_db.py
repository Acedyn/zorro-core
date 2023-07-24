import pytest

from zorro_core.db import db_manager
from zorro_core.main.settings import DBSettings


@pytest.mark.asyncio
async def test_db_initialisation():
    await db_manager.init_db(DBSettings(url="sqlite://:memory:"))
    await db_manager.migrate()
