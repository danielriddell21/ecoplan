FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /service ./cmd/main.go

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /service /service

EXPOSE 50051 9090

ENTRYPOINT ["/service"]
