package middleware

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// tokenBucket is a simple token-bucket rate limiter.
type tokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	capacity float64
	rate     float64 // tokens per second
	lastFill time.Time
}

func newTokenBucket(capacity float64, ratePerMinute float64) *tokenBucket {
	return &tokenBucket{
		tokens:   capacity,
		capacity: capacity,
		rate:     ratePerMinute / 60.0,
		lastFill: time.Now(),
	}
}

func (b *tokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens = min(b.capacity, b.tokens+elapsed*b.rate)
	b.lastFill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// UnaryRateLimit returns a gRPC unary interceptor with per-method rate limits.
// imageMethod receives a stricter limit than the default.
func UnaryRateLimit(defaultRPM float64, imageMethod string, imageRPM float64) grpc.UnaryServerInterceptor {
	defaultBucket := newTokenBucket(defaultRPM, defaultRPM)
	imageBucket := newTokenBucket(imageRPM, imageRPM)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		bucket := defaultBucket
		if info.FullMethod == imageMethod {
			bucket = imageBucket
		}

		if !bucket.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded — please slow down")
		}

		return handler(ctx, req)
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
