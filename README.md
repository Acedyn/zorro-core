# Zorro core

Zorro core is the main repository of the zorro ecosystem. It is responsible for sending code execution
request to various processors like Python, Nodejs, even Unreal or Nuke.

## What does it do

Zorro is a framework used to create user friendly **tools**. There is different types of tools,
they are declared via **plugins** and they represent a functionality exposed to the user.
They are always bound to a **context**, when a tool is triggered a context must be selected with it.
Zorro lets you create your own tools that will pilot multiple applications, it serve as a
comunication hub between all your softwares to synchronize operations that you will create.

## The types of tools

Tools can be nested, there is higher level type of tools than others

- **command**: A command is the smallest type of tool, it is rarely used alone and define a piece of code
  to execute on a given client.
- **action**: Used to group and organise commands into a dependency graph and execute it sequencially.
- **widget**: A group of graphical components bound to a command, used to build interactive GUI
- **hook**: Used to attach commands or actions to a particular event.

## How does it works

When creating your plugins you will define:
- **tools**: That will register functionalities exposed to the user
- **processors**: That will define an application able to communicate with zorro

Zorro is using the [grpc](https://grpc.io/) protocol to communicate with processors.
When a tool will be executed, zorro will start the nessesary processors to execute the tool
and will communicate with them to request execution.
The processors can understand each other thanks to [protocol buffers](https://protobuf.dev/)
wich serves as a common data format for all the processors

## Get started

### CI / CD

#### Unit tests

First make sure you fetched the git submodules (used by some tests)

```
git submodule init --remote
```

To run the unit test use:

```bash
go test ./...
```

You might want to use a unit test formater like [gotestsum](https://github.com/gotestyourself/gotestsum)

```bash
gotestsum --format dots
```

#### Formater

The formater used is [gofumpt](https://github.com/mvdan/gofumpt)
Make sure to format your code either by configuring your IDE or executing

```bash
gofumpt -l -w .
```
