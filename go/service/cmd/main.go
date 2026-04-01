package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/ecoscan/service/internal/config"
	"github.com/ecoscan/service/internal/providers"
	"github.com/ecoscan/service/internal/service"
	"github.com/ecoscan/service/middleware"
	pb "github.com/ecoscan/service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.SlogLevel(),
	}))

	resolver := providers.NewOpenFoodFactsResolver(cfg.OFFBaseURL, cfg.OFFTimeout, log)
	classifier := providers.NewClaudeClassifier(cfg.AnthropicKey, cfg.ClaudeModel, cfg.ClaudeTimeout, log)

	svc, err := service.NewRecyclingServiceServer(log, resolver, classifier)
	if err != nil {
		return fmt.Errorf("initialising service: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryLogging(log),
			middleware.UnaryAuth(cfg.APIKey),
			middleware.UnaryRateLimit(
				100, // default: 100 req/min
				"/recycling.RecyclingService/CanItBeRecycledImage",
				10, // image: 10 req/min (Claude is expensive)
			),
		),
	)

	pb.RegisterRecyclingServiceServer(grpcServer, svc)

	// Register gRPC health check service.
	healthSvc := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSvc)
	healthSvc.SetServingStatus("recycling.RecyclingService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Start HTTP server for /healthz and /metrics on MetricsPort.
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "ok")
		})
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			// Placeholder: wire Prometheus exporter here when adding OTel metrics.
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "# metrics endpoint — Prometheus exporter not yet configured\n")
		})
		addr := fmt.Sprintf(":%d", cfg.MetricsPort)
		log.Info("starting metrics/health HTTP server", "addr", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Error("metrics server failed", "err", err)
		}
	}()

	// Start gRPC server.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("listening on port %d: %w", cfg.GRPCPort, err)
	}

	log.Info("starting gRPC server", "port", cfg.GRPCPort)

	// Graceful shutdown on context cancellation (extend later with signal handling).
	_ = context.Background()

	return grpcServer.Serve(lis)
}
