# ZORRO CORE

## CI / CD

### Testing

#### Native unit tests

You can run the tests with the native build for your environment

```bash
go test ./...
```

#### Wasm tests

The tests depends on some binaries that must be compiled first

```bash
export GOOS="js"
export GOARCH="wasm"
go build -o .\testdata\processors\wasm_worker.wasm .\testdata\processors\wasm_worker.go
```

You can run the tests in the browser with [wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest?tab=readme-ov-file#quickstart)

```bash
export GOOS="js"
export GOARCH="wasm"
go test ./...
```
