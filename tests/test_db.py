import pytest

from zorro_core.db import db_manager
from zorro_core.db.entity import EntityType, Entity
from zorro_core.main.settings import DBSettings

@pytest.fixture(autouse=True)
async def init_db():
    """
    Initialize the db and clear it right before deletion
    """

    await db_manager.init_db(DBSettings(url="sqlite://:memory:"))
    await db_manager.migrate()

    yield

    await db_manager.drop()


@pytest.mark.asyncio
async def test_db_entities():
    """
    Test the entities and entity types creation and relations
    """

    project_type = EntityType(name="project", label="Project")
    await project_type.save()
    episode_type = EntityType(name="episode", label="Episode")
    await episode_type.save()

    assert len(await EntityType.all()) == 2

    foo = Entity(name="foo", label="Foo", type=project_type)
    await foo.save()
    a = Entity(name="a", label="A", parent=foo, type=episode_type)
    await a.save()
    b = Entity(name="b", label="B", parent=foo, type=episode_type)
    await b.save()
    c = Entity(name="c", label="C", parent=foo, type=episode_type)
    await c.save()
    await c.casting.add(a, b)

    assert len(await Entity.all()) == 4
    assert a.parent == foo
    async for casting in c.casting:
        assert casting in [a, b]
