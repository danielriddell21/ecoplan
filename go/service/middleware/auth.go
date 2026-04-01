package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryAuth returns a gRPC unary interceptor that validates a Bearer API key.
// The health check method is exempt.
func UnaryAuth(apiKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Exempt gRPC health check.
		if strings.HasSuffix(info.FullMethod, "/Check") || strings.HasSuffix(info.FullMethod, "/Watch") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(values[0], "Bearer ")
		token = strings.TrimPrefix(token, "bearer ")
		if token != apiKey {
			return nil, status.Error(codes.Unauthenticated, "invalid API key")
		}

		return handler(ctx, req)
	}
}
