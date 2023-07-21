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

- **context**: Datatypes and providers for context resolution according to the context query
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

### Static type checking

The static type checking is done via mypy

```bash
poetry run mypy zorro-core --ignore-missing-imports --strict
```

### Unit tests

The unit tests are done via pytest

```bash
poetry run mypy pytest
```
