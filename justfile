# List available recipes
default:
    @just --list

# Run the service
# Requires ANTHROPIC_API_KEY and ECOSCAN_API_KEY to be set
run:
    cd go/service && go run ./cmd

# Run all tests with race detection
test:
    cd go/service && go test ./... -race -count=1

# Generate Go code from proto/recycling.proto (run `just setup` first)
proto:
    protoc \
        --proto_path=proto \
        --proto_path=third_party \
        --go_out=go/service/proto \
        --go_opt=paths=source_relative \
        --go-grpc_out=go/service/proto \
        --go-grpc_opt=paths=source_relative \
        --grpc-gateway_out=go/service/proto \
        --grpc-gateway_opt=paths=source_relative \
        proto/recycling.proto

# Install protoc plugins and download third-party proto dependencies
setup:
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
    mkdir -p third_party/google/api
    curl -sSfL -o third_party/google/api/annotations.proto \
        https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
    curl -sSfL -o third_party/google/api/http.proto \
        https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto
