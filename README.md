# Zorro core

Zorro core is the main repository of the zorro ecosystem. It is responsible for executing code
on various clients like Unreal or Nuke.

## How it works

The user can interact with zorro via the cli, the REST API or the gRPC API. He can also interact with
various frontends (like kitsu or shotgrid) wich are actually using the REST API.

All interactions the user will make will be through tools. There is different types of tools, they are declared via
configs and they represent a functionality exposed to the user. They are always bound to a context, when a tool
is triggered a context must be selected with it.

## The types of tools

Tools can be composed of each other, there is higher level type of tools than others

- **command**: A command is the smallest type of tool, it is rarely used alone and define a piece of code
  to execute on a given client.
- **action**: Used to group and organise commands into a dependency graph and execute it sequencially.
- **widget**: A group of graphical components bound to a command, used to build interactive GUI
- **hook**: Used to attach commands or actions to a particular event.

## Get started

### Protobufs

#### Install protoc

- Install the (protoc executable)[https://protobuf.dev/downloads]
- Install the golang protoc generator

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

#### Generate structs from protofiles

This project is using [protobufs](https://protobuf.dev/) for a lots of struct definitions. After every
modifications if a proto buffer make sure to regenerate the structs

```bash
protoc --go_out=. --go_opt=paths=source_relative .\internal\tools\*.proto
protoc --go_out=. --go_opt=paths=source_relative .\internal\context\*.proto
```

Since the proto files are importing each other it is a good idea to regenerate everything when
modifying a protofile
