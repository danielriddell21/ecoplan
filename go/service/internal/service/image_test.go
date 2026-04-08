package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ecoscan/service/internal/providers"
	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCanItBeRecycledImage_invalidRequest(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})
	cases := []struct {
		name  string
		image []byte
	}{
		{"empty image", nil},
		{"image too large", make([]byte, 5*1024*1024+1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CanItBeRecycledImage(context.Background(), &pb.CanItBeRecycledImageRequest{Image: tc.image})
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("expected InvalidArgument, got %v", err)
			}
		})
	}
}

func TestCanItBeRecycledImage_classifierErrors(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{"not found", providers.ErrNotFound, codes.NotFound},
		{"upstream error", errors.New("upstream timeout"), codes.Internal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestService(t, &stubResolver{}, &stubClassifier{err: tc.err})
			_, err := svc.CanItBeRecycledImage(context.Background(), &pb.CanItBeRecycledImageRequest{Image: []byte{1}})
			if status.Code(err) != tc.wantCode {
				t.Errorf("expected %v, got %v", tc.wantCode, err)
			}
		})
	}
}

func TestCanItBeRecycledImage_noMaterialsFound(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{item: "an alien artefact"})
	resp, err := svc.CanItBeRecycledImage(context.Background(), &pb.CanItBeRecycledImageRequest{Image: []byte{1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Recyclable {
		t.Error("expected Recyclable=false for unknown item")
	}
	if !strings.Contains(resp.Data.Advice, "an alien artefact") {
		t.Errorf("advice should mention the item, got %q", resp.Data.Advice)
	}
}

func TestCanItBeRecycledImage_knownRecyclableMaterial(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{item: "plastic bottle"})
	resp, err := svc.CanItBeRecycledImage(context.Background(), &pb.CanItBeRecycledImageRequest{Image: []byte{1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if !resp.Data.Recyclable {
		t.Error("expected Recyclable=true for plastic bottle")
	}
	// Bin colour/type mapping is covered by mappers/bin_test.go; asserting specific
	// values here would couple this test to materials.json data.
}
