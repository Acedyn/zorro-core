# Zorro - Proto

This repository centralize the proto files used for the zorro application

## Generate the proto files's code

First you must install the [protoc executable](https://protobuf.dev/downloads)

### Golang

- Install the golang protoc generator

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

- Generate the go code

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./**/*.proto
```

Since go imports are importing from this repository directly, you have to commit and
push your changes to make them usable on the other zorro repositories

### Python

- Install the python protoc generator

```bash
python -m pip install grpcio-tools
```

- Generate the python code

```bash
python -m grpc_tools.protoc -I. --python_out=. --pyi_out=. --grpc_python_out=. ./**/*.proto
```
