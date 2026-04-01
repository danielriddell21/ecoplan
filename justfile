# List available recipes
default:
    @just --list

# Run services — `just run` starts both, `just run grpc` or `just run rest` starts one
run type="all":
    #!/usr/bin/env bash
    set -e
    case "{{type}}" in
        grpc) just _run-grpc ;;
        rest) just _run-rest ;;
        all)
            trap 'kill 0' EXIT
            just _run-grpc &
            just _run-rest &
            wait ;;
        *) echo "unknown type: {{type}}"; exit 1 ;;
    esac

[private]
_run-grpc:
    go run ./go/service/cmd/grpc

[private]
_run-rest:
    go run ./go/service/cmd/gateway

# Run all tests with race detection
test:
    go test ./go/service/... -race -count=1

# Generate Go code from proto/recycling.proto (run `just setup` first)
proto:
    buf generate

# Install buf and protoc plugins required for code generation
setup:
    go install github.com/bufbuild/buf/cmd/buf@latest
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
