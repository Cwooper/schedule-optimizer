# Schedule Optimizer

This is the backend for the Schedule Optimizer Web Application.

## Prerequisites

- Go 1.16 or later
- Protocol Buffer Compiler (protoc) 3.0.0 or later
- Go Protocol Buffers plugin

## Setup

1. Install the Protocol Buffer Compiler (protoc) if you haven't already:
   - MacOS: `brew install protobuf`
   - Linux: `sudo apt install protobuf-compiler`
   - Windows: Download from https://github.com/protocolbuffers/protobuf/releases

2. Install the `Go` Protocol Buffers plugin:

   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   ```

3. Make sure your PATH includes the `Go` bin directory:

   ```bash
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```

## Compiling Protocol Buffers

To compile the Protocol Buffer files and generate the corresponding `Go` code, run:

```bash
make proto
```

This will compile all `.proto` files in the `internal/models/proto` directory and
output the generated `Go` files to `internal/proto/generated`.

To remove all generated Protocol Buffer Go files, run:

```bash
make clean
```

For a list of available make targets, run:

```bash
make help
```

## Project Backend Structure

```bash
backend
├── data
│   └──protobufs
├── go.mod
├── go.sum
├── internal
│   ├── generator
│   │   └── generator.go
│   ├── models
│   │   ├── course.go
│   │   ├── proto
│   │   │   └── course.proto
│   │   └── schedule.go
│   ├── proto
│   │   └── generated
│   │       └── course.pb.go
│   └── utils
│       └── utils.go
├── main.go
├── Makefile
├── pkg
│   └── protoutils
│       └── convertor.go
└── README.md
```
