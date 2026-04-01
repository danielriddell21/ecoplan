# List available recipes
default:
    @just --list

# Run the service
# Requires ANTHROPIC_API_KEY and ECOSCAN_API_KEY to be set
run:
    go run ./go/service/cmd

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
