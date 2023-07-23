from typing import List
import pytest

from zorro_core.tools.action import Action, ActionCommand
from zorro_core.tools.tool_base import ToolBase

@pytest.mark.asyncio
async def test_action_traverse():
    # The names must follow the expected order of execution alphabetically
    # if two children are expected to run concurently you can differentiate
    # them with a '-' as a separator
    dummy_action = Action("0-A", children={
        "00-A": Action("00-A", children={
            "000-A": ActionCommand("000-A"),
            "001-A": ActionCommand("001-A", upstream="00-A"),
            "002-A": ActionCommand("002-A", upstream="01-A"),
            "002-B": Action("002-B", upstream="01-A", children={
                "0020-A": ActionCommand("0020-A"),
                "0020-B": ActionCommand("0020-B"),
                }),
            "002-C": ActionCommand("002-A", upstream="01-A"),
            "002-D": ActionCommand("002-A", upstream="01-A"),
            }),
        "01-A": Action("01-A", upstream="0-A", children={
            "010-A": ActionCommand("010-A"),
            "011-A": ActionCommand("011-A", upstream="10-A"),
            }),
        "01-B": ActionCommand("01-B", upstream="0-A"),
        })

    traversal_history: List[str] = []
    async def traverse_test(tool: ToolBase):
        traversal_history.append(tool.name)
        return tool.name, tool

    await dummy_action.traverse(traverse_test)

    # We make sure the order in wich the action is traversed
    # is correct by checking if the traversed names are
    # in alphabetical order
    assert len(traversal_history) != 0
    for index, traversed_key in enumerate(traversal_history):
        if index == 0:
            assert traversed_key == dummy_action.name
            continue
        # We only care about the first part of the name
        # The other part is for concurent children
        assert traversal_history[index - 1].split("-")[0] < traversed_key.split("-")[0]
