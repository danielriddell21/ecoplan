package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ecoscan/service/internal/providers"
	"github.com/ecoscan/service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/ecoscan/service/proto"
	"log/slog"
	"os"
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

func TestCanItBeRecycledBarcode_invalidBarcode(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})
	cases := []string{"", "123", "not-a-barcode", "123456789012345"} // too short / non-digit / too long
	for _, bc := range cases {
		_, err := svc.CanItBeRecycledBarcode(context.Background(), &pb.CanItBeRecycledBarcodeRequest{Barcode: bc})
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("barcode %q: expected InvalidArgument, got %v", bc, err)
		}
	}
}

func TestCanItBeRecycledBarcode_notFound(t *testing.T) {
	resolver := &stubResolver{err: providers.ErrNotFound}
	svc := newTestService(t, resolver, &stubClassifier{})

	_, err := svc.CanItBeRecycledBarcode(context.Background(), &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestCanItBeRecycledBarcode_upstreamError(t *testing.T) {
	resolver := &stubResolver{err: errors.New("timeout")}
	svc := newTestService(t, resolver, &stubClassifier{})

	_, err := svc.CanItBeRecycledBarcode(context.Background(), &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"})
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", err)
	}
}

func TestCanItBeRecycledBarcode_cacheHit(t *testing.T) {
	resolver := &stubResolver{
		result: providers.BarcodeResult{
			PackagingTags: []string{"en:plastic-bottle"},
			ProductName:   "Test Product",
			Brand:         "TestBrand",
		},
	}
	svc := newTestService(t, resolver, &stubClassifier{})

	req := &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"}
	_, _ = svc.CanItBeRecycledBarcode(context.Background(), req)
	_, _ = svc.CanItBeRecycledBarcode(context.Background(), req)

	// Resolver should only be called once due to caching.
	if resolver.calls != 1 {
		t.Errorf("expected resolver called once, got %d", resolver.calls)
	}
}
