# Metaballs

This is a Metaballs implementation in Go, rendered in the browser with WASM.

## Compilation

Compile the Go code targeting WASM:
```shell
GOOS=js GOARCH=wasm go build -o main.wasm main.go
```

## Run the file server

It's necessary to run a separated file server to execute the web assembly binaries:
```shell
go run cmd/server/server.go
```

