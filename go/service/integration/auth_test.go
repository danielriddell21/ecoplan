package integration_test

import (
	"testing"

	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuth_missingToken(t *testing.T) {
	f := newIntegrationServer(t)
	ctx := t.Context() // no auth metadata

	calls := []struct {
		name string
		call func() error
	}{
		{"Barcode", func() error {
			_, err := f.client.CanItBeRecycledBarcode(ctx, &pb.CanItBeRecycledBarcodeRequest{Barcode: "12345678"})
			return err
		}},
		{"Search", func() error {
			_, err := f.client.CanItBeRecycledSearch(ctx, &pb.CanItBeRecycledSearchRequest{Query: "glass"})
			return err
		}},
		{"Image", func() error {
			_, err := f.client.CanItBeRecycledImage(ctx, &pb.CanItBeRecycledImageRequest{Image: []byte{1}})
			return err
		}},
	}

	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if status.Code(err) != codes.Unauthenticated {
				t.Errorf("expected Unauthenticated, got %v", err)
			}
		})
	}
}

func TestAuth_wrongToken(t *testing.T) {
	f := newIntegrationServer(t)
	ctx := metadata.NewOutgoingContext(
		t.Context(),
		metadata.Pairs("authorization", "Bearer wrong-key"),
	)

	_, err := f.client.CanItBeRecycledSearch(ctx, &pb.CanItBeRecycledSearchRequest{Query: "glass"})
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated for wrong token, got %v", err)
	}
}

func TestAuth_validToken_passesThrough(t *testing.T) {
	f := newIntegrationServer(t)

	_, err := f.client.CanItBeRecycledSearch(authCtx(t), &pb.CanItBeRecycledSearchRequest{Query: "glass"})
	if status.Code(err) == codes.Unauthenticated {
		t.Errorf("valid token should not return Unauthenticated, got %v", err)
	}
}
