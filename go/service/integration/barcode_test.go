package integration_test

import (
	"errors"
	"testing"

	"github.com/ecoscan/service/internal/providers"
	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBarcode_happy(t *testing.T) {
	f := newIntegrationServer(t)
	f.resolver.Result = providers.BarcodeResult{
		PackagingTags: []string{"en:plastic-bottle"},
		ProductName:   "Evian Water",
		Brand:         "Evian",
	}

	resp, err := f.client.CanItBeRecycledBarcode(authCtx(t), &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ProductName != "Evian Water" {
		t.Errorf("expected ProductName %q, got %q", "Evian Water", resp.ProductName)
	}
	if resp.Data == nil {
		t.Fatal("expected non-nil Data")
	}
}

func TestBarcode_invalidBarcode(t *testing.T) {
	f := newIntegrationServer(t)

	_, err := f.client.CanItBeRecycledBarcode(authCtx(t), &pb.CanItBeRecycledBarcodeRequest{Barcode: "short"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestBarcode_notFound(t *testing.T) {
	f := newIntegrationServer(t)
	f.resolver.Err = providers.ErrNotFound

	_, err := f.client.CanItBeRecycledBarcode(authCtx(t), &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestBarcode_upstreamError(t *testing.T) {
	f := newIntegrationServer(t)
	f.resolver.Err = errors.New("timeout")

	_, err := f.client.CanItBeRecycledBarcode(authCtx(t), &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"})
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", err)
	}
}

func TestBarcode_cacheHit(t *testing.T) {
	f := newIntegrationServer(t)
	f.resolver.Result = providers.BarcodeResult{
		PackagingTags: []string{"en:plastic-bottle"},
		ProductName:   "Evian Water",
		Brand:         "Evian",
	}

	req := &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"}
	ctx := authCtx(t)
	_, _ = f.client.CanItBeRecycledBarcode(ctx, req)
	_, _ = f.client.CanItBeRecycledBarcode(ctx, req)

	if f.resolver.Calls.Load() != 1 {
		t.Errorf("expected resolver called once due to cache, got %d", f.resolver.Calls.Load())
	}
}
