package integration_test

import (
	"testing"

	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestRateLimit_imageExhaustsBurst verifies that the image endpoint is rate-limited
// to a burst of 10. The first 10 requests must succeed; the 11th must be rejected.
func TestRateLimit_imageExhaustsBurst(t *testing.T) {
	f := newIntegrationServer(t)
	f.classifier.Item = "plastic bottle"

	ctx := authCtx(t)
	req := &pb.CanItBeRecycledImageRequest{Image: []byte{1}}

	for i := 0; i < 11; i++ {
		_, err := f.client.CanItBeRecycledImage(ctx, req)
		if i < 10 {
			if status.Code(err) == codes.ResourceExhausted {
				t.Fatalf("request %d should not be rate-limited, got ResourceExhausted", i+1)
			}
		} else {
			if status.Code(err) != codes.ResourceExhausted {
				t.Errorf("request %d should be rate-limited, got %v", i+1, err)
			}
		}
	}
}

// TestRateLimit_defaultNotExhaustedByTen verifies that the default limiter (burst 100)
// is not exhausted by 10 search requests.
func TestRateLimit_defaultNotExhaustedByTen(t *testing.T) {
	f := newIntegrationServer(t)
	ctx := authCtx(t)
	req := &pb.CanItBeRecycledSearchRequest{Query: "glass"}

	for i := 0; i < 10; i++ {
		_, err := f.client.CanItBeRecycledSearch(ctx, req)
		if status.Code(err) == codes.ResourceExhausted {
			t.Fatalf("request %d unexpectedly rate-limited: %v", i+1, err)
		}
	}
}

// TestRateLimit_searchDoesNotConsumeImageBucket verifies that the image and default
// rate limiters are independent: exhausting search calls does not deplete the image bucket.
func TestRateLimit_searchDoesNotConsumeImageBucket(t *testing.T) {
	f := newIntegrationServer(t)
	f.classifier.Item = "plastic bottle"
	ctx := authCtx(t)

	// Make 10 search calls (well within default burst of 100).
	for i := 0; i < 10; i++ {
		_, err := f.client.CanItBeRecycledSearch(ctx, &pb.CanItBeRecycledSearchRequest{Query: "glass"})
		if err != nil {
			t.Fatalf("search %d failed: %v", i+1, err)
		}
	}

	// First image call should succeed because the image bucket is still full.
	_, err := f.client.CanItBeRecycledImage(ctx, &pb.CanItBeRecycledImageRequest{Image: []byte{1}})
	if status.Code(err) == codes.ResourceExhausted {
		t.Errorf("image request should not be rate-limited after search calls, got %v", err)
	}
}
