package middleware

import (
	"context"

	grpcratelimit "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// perMethodLimiter applies different rate limits per gRPC method.
type perMethodLimiter struct {
	defaultLimiter *rate.Limiter
	imageLimiter   *rate.Limiter
	imageMethod    string
}

func (l *perMethodLimiter) Limit(ctx context.Context) error {
	method, _ := grpc.Method(ctx)
	lim := l.defaultLimiter
	if method == l.imageMethod {
		lim = l.imageLimiter
	}
	if !lim.Allow() {
		return status.Error(codes.ResourceExhausted, "rate limit exceeded — please slow down")
	}
	return nil
}

// UnaryRateLimit returns a gRPC unary interceptor with per-method rate limits.
// imageMethod receives a stricter limit (imageRPM) than the default (defaultRPM).
func UnaryRateLimit(defaultRPM float64, imageMethod string, imageRPM float64) grpc.UnaryServerInterceptor {
	limiter := &perMethodLimiter{
		defaultLimiter: rate.NewLimiter(rate.Limit(defaultRPM/60.0), int(defaultRPM)),
		imageLimiter:   rate.NewLimiter(rate.Limit(imageRPM/60.0), int(imageRPM)),
		imageMethod:    imageMethod,
	}
	return grpcratelimit.UnaryServerInterceptor(limiter)
}
