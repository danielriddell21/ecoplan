package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/ecoscan/service/internal/providers"
	"github.com/ecoscan/service/internal/service"
)

// stubResolver is a test double for BarcodeResolver.
type stubResolver struct {
	result providers.BarcodeResult
	err    error
	calls  int
}

func (s *stubResolver) Resolve(_ context.Context, _ string) (providers.BarcodeResult, error) {
	s.calls++
	return s.result, s.err
}

// stubClassifier is a test double for ImageClassifier.
type stubClassifier struct {
	item string
	err  error
}

func (s *stubClassifier) Classify(_ context.Context, _ []byte) (string, error) {
	return s.item, s.err
}

func newTestService(t *testing.T, resolver providers.BarcodeResolver, classifier providers.ImageClassifier) *service.RecyclingServiceServer {
	t.Helper()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc, err := service.NewRecyclingServiceServer(log, resolver, classifier)
	if err != nil {
		t.Fatalf("NewRecyclingServiceServer: %v", err)
	}
	return svc
}
