# ecoscan

gRPC API for recycling advice. Looks up items by barcode, text search, or image. Currently covers Portsmouth City Council rules.

## Setup

Copy `.env.example` to `.env` and fill in the required values.

| Variable | Default | Notes |
|---|---|---|
| `ANTHROPIC_API_KEY` | — | Required |
| `ECOSCAN_API_KEY` | — | Required — bearer token for API auth |
| `GRPC_PORT` | `50051` | |
| `GATEWAY_PORT` | `8080` | HTTP/JSON gateway |
| `METRICS_PORT` | `9090` | `/healthz` and `/metrics` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `GRPC_ADDR` | `localhost:50051` | Gateway only — address of the gRPC service |
| `CLAUDE_MODEL` | `claude-sonnet-4-6` | |
| `OFF_TIMEOUT` | `5s` | |
| `CLAUDE_TIMEOUT` | `15s` | |

## Running

```bash
just run           # both service and gateway
just run service   # gRPC service only
just run gateway   # HTTP gateway only
```

The gateway connects to the gRPC service via `GRPC_ADDR`. Run the service first if starting them separately.

## Docker

```bash
docker build -f Dockerfile.service -t ecoscan/service .
docker build -f Dockerfile.gateway -t ecoscan/gateway .
```

Run from `go/service/`.

## Calling the API

All requests require `authorization: bearer <ECOSCAN_API_KEY>`.

### gRPC

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

# Image (base64-encoded JPEG/PNG)
grpcurl -plaintext \
  -H "authorization: bearer your-secret" \
  -d "{\"image\": \"$(base64 -w0 item.jpg)\"}" \
  localhost:50051 recycling.RecyclingService/CanItBeRecycledImage
```

### HTTP

```bash
curl -X POST http://localhost:8080/v1/recycle/barcode \
  -H "authorization: bearer your-secret" \
  -d '{"barcode": "5000112546415"}'

curl "http://localhost:8080/v1/recycle/search?query=glass+bottle" \
  -H "authorization: bearer your-secret"

curl -X POST http://localhost:8080/v1/recycle/image \
  -H "authorization: bearer your-secret" \
  -d "{\"image\": \"$(base64 -w0 item.jpg)\"}"
```

## Rate limits

| Method | Limit |
|---|---|
| `CanItBeRecycled`, `CanItBeRecycledSearch` | 100 req/min |
| `CanItBeRecycledImage` | 10 req/min |

## Proto

```bash
just proto   # regenerate Go code from proto/recycling.proto
just setup   # install buf and protoc plugins (first time)
```

## Adding a council

1. Add a JSON file alongside `materials.json` using the same schema.
2. Load it based on the `council_id` field on each request.
3. Clients pass `council_id` in the request; omitting it defaults to `portsmouth`.
