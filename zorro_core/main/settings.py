from dataclasses import dataclass, field


@dataclass
class DBSettings:
    url: str = field(default="sqlite://zorro.db")


@dataclass
class Settings:
    db: DBSettings = field(default_factory=DBSettings)


async def get_settings() -> Settings:
    return Settings()
