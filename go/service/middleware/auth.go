package middleware

import (
	"context"
	"strings"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryAuth returns a gRPC unary interceptor that validates a Bearer API key.
// The health check methods are exempt.
func UnaryAuth(apiKey string) grpc.UnaryServerInterceptor {
	return grpcauth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
		method, _ := grpc.Method(ctx)
		if strings.HasSuffix(method, "/Check") || strings.HasSuffix(method, "/Watch") {
			return ctx, nil
		}

		token, err := grpcauth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}
		if token != apiKey {
			return nil, status.Error(codes.Unauthenticated, "invalid API key")
		}
		return ctx, nil
	})
}
