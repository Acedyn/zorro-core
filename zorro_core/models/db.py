from tortoise import Tortoise

async def init_db(url: str):
    await Tortoise.init(
        db_url=url,
        modules={"models": ["zorro_core.models.entity", "zorro_core.models.user"]}
    )

async def migrate():
    await Tortoise.generate_schemas()
