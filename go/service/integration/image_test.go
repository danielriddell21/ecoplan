package integration_test

import (
	"errors"
	"testing"

	"github.com/ecoscan/service/internal/providers"
	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImage_happy(t *testing.T) {
	f := newIntegrationServer(t)
	f.classifier.Item = "plastic bottle"

	resp, err := f.client.CanItBeRecycledImage(authCtx(t), &pb.CanItBeRecycledImageRequest{Image: []byte{1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if !resp.Data.Recyclable {
		t.Error("expected Recyclable=true for plastic bottle")
	}
}

func TestImage_emptyImage(t *testing.T) {
	f := newIntegrationServer(t)

	_, err := f.client.CanItBeRecycledImage(authCtx(t), &pb.CanItBeRecycledImageRequest{Image: nil})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for empty image, got %v", err)
	}
}

func TestImage_classifierNotFound(t *testing.T) {
	f := newIntegrationServer(t)
	f.classifier.Err = providers.ErrNotFound

	_, err := f.client.CanItBeRecycledImage(authCtx(t), &pb.CanItBeRecycledImageRequest{Image: []byte{1}})
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestImage_classifierError(t *testing.T) {
	f := newIntegrationServer(t)
	f.classifier.Err = errors.New("upstream error")

	_, err := f.client.CanItBeRecycledImage(authCtx(t), &pb.CanItBeRecycledImageRequest{Image: []byte{1}})
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal, got %v", err)
	}
}
