package integration_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"

	"github.com/ecoscan/service/internal/providers/providerstest"
	"github.com/ecoscan/service/internal/service"
	"github.com/ecoscan/service/middleware"
	pb "github.com/ecoscan/service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const (
	testAPIKey  = "test-api-key-integration"
	imageMethod = "/recycling.v1.RecyclingService/CanItBeRecycledImage"
	bufSize     = 1 << 20 // 1 MiB in-memory buffer
)

// serverFixture holds the gRPC client and configurable stubs for a test server.
type serverFixture struct {
	client     pb.RecyclingServiceClient
	conn       *grpc.ClientConn
	resolver   *providerstest.StubResolver
	classifier *providerstest.StubClassifier
}

// newIntegrationServer spins up a real gRPC server with the full middleware chain
// (identical to main.go) backed by a bufconn listener and returns a connected client.
// The server is stopped automatically via t.Cleanup.
func newIntegrationServer(t *testing.T) *serverFixture {
	t.Helper()

	resolver := &providerstest.StubResolver{}
	classifier := &providerstest.StubClassifier{}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc, err := service.NewRecyclingServiceServer(log, resolver, classifier)
	if err != nil {
		t.Fatalf("NewRecyclingServiceServer: %v", err)
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryLogging(log),
			middleware.UnaryAuth(testAPIKey),
			middleware.UnaryRateLimit(100, imageMethod, 10),
		),
	)
	pb.RegisterRecyclingServiceServer(grpcSrv, svc)

	lis := bufconn.Listen(bufSize)
	go func() { _ = grpcSrv.Serve(lis) }()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}

	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("closing gRPC connection: %v", err)
		}
		grpcSrv.GracefulStop()
	})

	return &serverFixture{
		client:     pb.NewRecyclingServiceClient(conn),
		conn:       conn,
		resolver:   resolver,
		classifier: classifier,
	}
}

// authCtx returns a context carrying the correct Bearer token for the test server.
// It is bound to the test lifetime via t.Context() so any in-flight RPC is
// cancelled automatically if the test ends.
func authCtx(t *testing.T) context.Context {
	t.Helper()
	return metadata.NewOutgoingContext(
		t.Context(),
		metadata.Pairs("authorization", "Bearer "+testAPIKey),
	)
}
