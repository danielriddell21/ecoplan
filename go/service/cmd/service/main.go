package main

import (
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

	"github.com/joho/godotenv"
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
	_ = godotenv.Load()

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
				100,
				"/recycling.v1.RecyclingService/CanItBeRecycledImage",
				10,
			),
		),
	)

	pb.RegisterRecyclingServiceServer(grpcServer, svc)

	healthSvc := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSvc)
	healthSvc.SetServingStatus("recycling.v1.RecyclingService", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "ok")
		})
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "# metrics endpoint — Prometheus exporter not yet configured\n")
		})
		addr := fmt.Sprintf(":%d", cfg.MetricsPort)
		log.Info("starting metrics/health HTTP server", "addr", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Error("metrics server failed", "err", err)
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("listening on port %d: %w", cfg.GRPCPort, err)
	}

	log.Info("starting gRPC server", "port", cfg.GRPCPort)
	return grpcServer.Serve(lis)
}
