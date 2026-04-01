package middleware

import (
	"context"
	"log/slog"

	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
)

// UnaryLogging returns a gRPC unary interceptor that logs each request with method, duration, and status.
func UnaryLogging(log *slog.Logger) grpc.UnaryServerInterceptor {
	logger := grpclogging.LoggerFunc(func(ctx context.Context, level grpclogging.Level, msg string, fields ...any) {
		switch level {
		case grpclogging.LevelDebug:
			log.DebugContext(ctx, msg, fields...)
		case grpclogging.LevelWarn:
			log.WarnContext(ctx, msg, fields...)
		case grpclogging.LevelError:
			log.ErrorContext(ctx, msg, fields...)
		default:
			log.InfoContext(ctx, msg, fields...)
		}
	})
	return grpclogging.UnaryServerInterceptor(logger)
}
