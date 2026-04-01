# List available recipes
default:
    @just --list

# Run services — `just run` starts both, `just run service` or `just run gateway` starts one
run type="all":
    #!/usr/bin/env bash
    set -e
    case "{{type}}" in
        service) just _run-service ;;
        gateway) just _run-gateway ;;
        all)
            trap 'kill 0' EXIT
            just _run-service &
            just _run-gateway &
            wait ;;
        *) echo "unknown type: {{type}}"; exit 1 ;;
    esac

[private]
_run-service:
    go run ./go/service/cmd/service

[private]
_run-gateway:
    go run ./go/service/cmd/gateway

# Lint Go code
lint:
    cd go/service && golangci-lint run ./...

# Run all tests
test:
    go test ./go/service/... -count=1

# Build Docker images — `just docker` builds both, `just docker service` or `just docker gateway` builds one
docker type="all":
    #!/usr/bin/env bash
    set -e
    case "{{type}}" in
        service) just _docker-service ;;
        gateway) just _docker-gateway ;;
        all)
            just _docker-service
            just _docker-gateway ;;
        *) echo "unknown type: {{type}}"; exit 1 ;;
    esac

[private]
_docker-service:
    cd go/service && docker build -f Dockerfile.service -t ecoscan/service .

[private]
_docker-gateway:
    cd go/service && docker build -f Dockerfile.gateway -t ecoscan/gateway .

# Generate Go code from proto/recycling.proto (run `just setup` first)
gen:
    buf generate

# Install buf and protoc plugins required for code generation
setup:
    go install github.com/bufbuild/buf/cmd/buf@latest
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
