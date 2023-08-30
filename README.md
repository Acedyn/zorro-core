# Zorro core

Zorro core is the main repository of the zorro pipeline. It is responsible for executing code
on various clients like Unreal on Nuke.

## How it works

The user can interact with zorro core via the cli, the rest API or the socket.io API. He can also interact with
various frontends (like Kitsu) wich are actually using the rest API.

All interactions the user will make will be through tools. There is different types of tools, they are declared via
configs and they represent a functionality exposed to the user. They are always bound to a context, when a tool
is triggered a context must be selected first.

## The types of tools

Tools can be composed of each other, there is higher level type of tools than others

- **command**: A command is the smallest type of tool, it is rarely used alone and define a piece of code
  to execute on a given client.
- **action**: Used to group and organise commands into a dependency graph and execute it sequencially.
- **widget**: A group of graphical components bound to a command, used to build interactive GUI
- **hook**: Used to attach commands or actions to a particular event.

## The structure

- **context**: Datatypes and context resolution according to a context query
- **tools**: Datastructures and execution logic related to all the type of tools
- **api**
- **app**:
- **commands**:
- **dcc**:

## Get started

You will need poetry to setup this project

```bash
pip install poetry
```

We usually configure poetry to place the virtualenv locally rather than in your home directory but that's up to you

```bash
poetry config virtualenvs.in-project true
```

Then install the dependencies (including the dev dependencies like mypy and pytest)

```bash
poetry install --with dev
```

From now you can activate the virtualenv manually or prepend all your command with `poetry run` which will run the
command in the environment

## CI - CD

### Generate the json schemas from pydantic classes

This must be done every time some pydantic classes that will be instantiated from JSON files are modified (Action,
Command, Plugin...) This will generate json schemas to have auto completion and validation when writing those JSONs

```bash
poetry run python .\scripts\generate_json_schemas.py
```

### Generate the protobufs from pydantic classes

This is a helper script to generate from protobufs from the pydantic classes that will be sent via GRPC.

> :warning: **Be carefull when modifing the protobufs** This script will regenerate protobufs without
> taking in account the previous protobuf fields's numbers. If you want to modify those classes while maintaining
> backward compatibility you will have to edit the protobufs manually. [see why field numbers are important](https://protobuf.dev/programming-guides/proto3/#assigning)
> You can still use this script but just be aware that it might break compatibility with existing clients implementations

```bash
poetry run python .\scripts\generate_protobufs.py
```

### Generate the grpc endpoints from the protobufs

You must run this script every type the protobufs are modified

```bash
poetry run python -m grpc_tools.protoc -I . --python_out=. --pyi_out=. --grpc_python_out=. zorro_core/network/protos/*.proto
```

### Static type checking

The static type checking is done via mypy

```bash
poetry run mypy zorro-core --ignore-missing-imports --strict .
```

### Unit tests

The unit tests are done via pytest

```bash
poetry run pytest .
```
