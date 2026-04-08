package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ecoscan/service/internal/providers"
	"github.com/ecoscan/service/internal/providers/providerstest"
	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCanItBeRecycledBarcode_invalidBarcode(t *testing.T) {
	svc := newTestService(t, &providerstest.StubResolver{}, &providerstest.StubClassifier{})
	cases := []string{"", "123", "not-a-barcode", "123456789012345"} // too short / non-digit / too long
	for _, bc := range cases {
		_, err := svc.CanItBeRecycledBarcode(context.Background(), &pb.CanItBeRecycledBarcodeRequest{Barcode: bc})
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("barcode %q: expected InvalidArgument, got %v", bc, err)
		}
	}
}

func TestCanItBeRecycledBarcode_notFound(t *testing.T) {
	resolver := &providerstest.StubResolver{Err: providers.ErrNotFound}
	svc := newTestService(t, resolver, &providerstest.StubClassifier{})

	_, err := svc.CanItBeRecycledBarcode(context.Background(), &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestCanItBeRecycledBarcode_upstreamError(t *testing.T) {
	resolver := &providerstest.StubResolver{Err: errors.New("timeout")}
	svc := newTestService(t, resolver, &providerstest.StubClassifier{})

	_, err := svc.CanItBeRecycledBarcode(context.Background(), &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"})
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", err)
	}
}

func TestCanItBeRecycledBarcode_cacheHit(t *testing.T) {
	resolver := &providerstest.StubResolver{
		Result: providers.BarcodeResult{
			PackagingTags: []string{"en:plastic-bottle"},
			ProductName:   "Test Product",
			Brand:         "TestBrand",
		},
	}
	svc := newTestService(t, resolver, &providerstest.StubClassifier{})

	req := &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"}
	_, _ = svc.CanItBeRecycledBarcode(context.Background(), req)
	_, _ = svc.CanItBeRecycledBarcode(context.Background(), req)

	// Resolver should only be called once due to caching.
	if resolver.Calls.Load() != 1 {
		t.Errorf("expected resolver called once, got %d", resolver.Calls.Load())
	}
}
