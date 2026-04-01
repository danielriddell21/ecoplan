# ecoscan

A production-ready gRPC API for recycling advice. Identifies how to recycle items via barcode lookup, text search, or AI image classification. Currently covers Portsmouth City Council recycling rules.

## Usage

### Requirements

| Variable | Description |
|---|---|
| `ANTHROPIC_API_KEY` | Anthropic API key (required) |
| `ECOSCAN_API_KEY` | Shared secret for API authentication (required) |
| `GRPC_PORT` | gRPC listen port (default: `50051`) |
| `METRICS_PORT` | HTTP health/metrics port (default: `9090`) |
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` (default: `info`) |
| `CLAUDE_MODEL` | Claude model ID (default: `claude-sonnet-4-6`) |
| `OFF_TIMEOUT` | OpenFoodFacts request timeout (default: `5s`) |
| `CLAUDE_TIMEOUT` | Claude request timeout (default: `15s`) |

### Run locally

```bash
export ANTHROPIC_API_KEY=sk-ant-...
export ECOSCAN_API_KEY=your-secret
cd go/service
go run ./cmd/main.go
```

### Run with Docker

```bash
docker build -t ecoscan .
docker run -p 50051:50051 -p 9090:9090 \
  -e ANTHROPIC_API_KEY=sk-ant-... \
  -e ECOSCAN_API_KEY=your-secret \
  ecoscan
```

### Health check

```bash
curl http://localhost:9090/healthz
```

### Calling the API

All RPCs require an `authorization: bearer <ECOSCAN_API_KEY>` metadata header.

```bash
# Barcode lookup
grpcurl -plaintext \
  -H "authorization: bearer your-secret" \
  -d '{"barcode": "5000112546415"}' \
  localhost:50051 recycling.RecyclingService/CanItBeRecycled

# Text search
grpcurl -plaintext \
  -H "authorization: bearer your-secret" \
  -d '{"query": "glass bottle"}' \
  localhost:50051 recycling.RecyclingService/CanItBeRecycledSearch

# Image classification (base64-encoded JPEG/PNG)
grpcurl -plaintext \
  -H "authorization: bearer your-secret" \
  -d "{\"image\": \"$(base64 -w0 item.jpg)\"}" \
  localhost:50051 recycling.RecyclingService/CanItBeRecycledImage
```

### Rate limits

| Method | Limit |
|---|---|
| `CanItBeRecycled`, `CanItBeRecycledSearch` | 100 req/min |
| `CanItBeRecycledImage` | 10 req/min |

## Structure

```
ecoscan/
├── proto/
│   └── recycling.proto               # Service definition
├── go/service/
│   ├── cmd/main.go                   # Entrypoint — wires config, providers, server
│   ├── internal/
│   │   ├── config/config.go          # Typed config loaded from env vars
│   │   ├── models/                   # Material, MaterialsDB, OFFResponse types
│   │   ├── mappers/bin.go            # Bin name → proto enum conversion
│   │   ├── providers/                # External integration interfaces + implementations
│   │   │   ├── resolver.go           # BarcodeResolver interface
│   │   │   ├── classifier.go         # ImageClassifier interface
│   │   │   ├── openfoodfacts.go      # Barcode → packaging tags (Open Food Facts API)
│   │   │   └── claude.go             # Image → item description (Claude API)
│   │   └── service/
│   │       ├── service.go            # Server struct, cache, DI wiring
│   │       ├── barcode.go            # CanItBeRecycled handler
│   │       ├── image.go              # CanItBeRecycledImage handler
│   │       ├── search.go             # CanItBeRecycledSearch handler
│   │       └── materials.json        # Embedded Portsmouth recycling rules
│   └── middleware/
│       ├── logging.go                # Request logging interceptor
│       ├── auth.go                   # Bearer token auth interceptor
│       └── ratelimit.go              # Token-bucket rate limiting interceptor
├── Dockerfile                        # Multi-stage distroless build
└── .github/workflows/ci.yml          # CI: vet, test, docker build
```

## Adding a new council

1. Add a new JSON file alongside `materials.json` (same schema, different rules).
2. Load it based on the `council_id` field present on all request messages.
3. Pass the correct `council_id` from your client. Omitting it defaults to `portsmouth`.

## Regenerating proto code

```bash
protoc \
  --proto_path=proto \
  --go_out=go/service --go_opt=paths=source_relative \
  --go-grpc_out=go/service --go-grpc_opt=paths=source_relative \
  proto/recycling.proto
mv go/service/recycling*.pb.go go/service/proto/
```
