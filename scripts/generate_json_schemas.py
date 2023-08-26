import json
from pathlib import Path

from zorro_core.context.plugin import Plugin
from zorro_core.tools.action import Action
from zorro_core.tools.command import Command
from zorro_core.utils.logger import logger


def main():
    schemas_root = Path(__file__).parent.parent / "schemas"
    schemas_root.mkdir(parents=True, exist_ok=True)

    with open(schemas_root / "plugin.json", "w") as f:
        json.dump(Plugin.model_json_schema(), f, indent=2)
    logger.info("%s written", schemas_root / "plugin.json")

    with open(schemas_root / "command.json", "w") as f:
        json.dump(Command.model_json_schema(), f, indent=2)
    logger.info("%s written", schemas_root / "command.json")

    with open(schemas_root / "action.json", "w") as f:
        json.dump(Action.model_json_schema(), f, indent=2)
    logger.info("%s written", schemas_root / "action.json")


if __name__ == "__main__":
    main()
