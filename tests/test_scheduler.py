# import pytest
# from zorro_core.schedulers.client import Program

# from zorro_core.schedulers.local_scheduler import ClientQuery, LocalScheduler
# from zorro_core.tools.command import Command


# @pytest.mark.asyncio
# async def test_start_client():
#     program = Program(name="python")
#     client = await program.start_as_client()
#     assert client.process is not None


# @pytest.mark.asyncio
# async def test_local_scheduler():
#     client_query = ClientQuery(program_name="python")
#     scheduler = LocalScheduler(query=client_query)
#     command = Command()
#     print(scheduler, command)
