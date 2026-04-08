package service_test

import (
	"context"
	"strings"
	"testing"

	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCanItBeRecycledSearch_validation(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})
	cases := []struct {
		name  string
		query string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"too long", strings.Repeat("a", 201)},
		{"control chars", "bottle\x00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: tc.query})
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("expected InvalidArgument, got %v", err)
			}
		})
	}
}

func TestCanItBeRecycledSearch_noResults(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})
	resp, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: "xyzzy_no_match"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Data))
	}
}

func TestCanItBeRecycledSearch_caseInsensitive(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})

	lower, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: "glass"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	upper, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: "GLASS"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lower.Data) == 0 {
		t.Fatal("expected results for 'glass'")
	}
	if len(lower.Data) != len(upper.Data) {
		t.Errorf("case-insensitive: expected same count, got %d vs %d", len(lower.Data), len(upper.Data))
	}
}

// TestCanItBeRecycledSearch_deterministicOrdering verifies that ordinex produces a
// stable result order across repeated calls (map iteration is otherwise non-deterministic).
func TestCanItBeRecycledSearch_deterministicOrdering(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})

	resp1, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: "plastic"})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	resp2, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: "plastic"})
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if len(resp1.Data) == 0 {
		t.Fatal("expected results for 'plastic'")
	}
	if len(resp1.Data) != len(resp2.Data) {
		t.Fatalf("expected same result count, got %d and %d", len(resp1.Data), len(resp2.Data))
	}
	for i := range resp1.Data {
		if resp1.Data[i].Advice != resp2.Data[i].Advice {
			t.Errorf("index %d: result ordering not deterministic between calls", i)
		}
	}
}

// TestCanItBeRecycledSearch_noDuplicates verifies that retrievium's binary search
// deduplication guard prevents the same material appearing more than once.
func TestCanItBeRecycledSearch_noDuplicates(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})

	resp, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: "plastic"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seen := make(map[string]bool)
	for _, item := range resp.Data {
		if seen[item.Advice] {
			t.Errorf("duplicate material in results: %q", item.Advice)
		}
		seen[item.Advice] = true
	}
}

func TestCanItBeRecycledSearch_recyclableGlass(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})

	resp, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: "glass bottles"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) == 0 {
		t.Fatal("expected at least one result for 'glass bottles'")
	}

	var foundRecyclable bool
	for _, item := range resp.Data {
		if item.Recyclable {
			foundRecyclable = true
			break
		}
	}
	if !foundRecyclable {
		t.Error("expected at least one recyclable result for 'glass bottles'")
	}
}

func TestCanItBeRecycledSearch_nonRecyclableItem(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})

	resp, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: "black plastic"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) == 0 {
		t.Fatal("expected at least one result for 'black plastic'")
	}
	for _, item := range resp.Data {
		if item.Recyclable {
			t.Errorf("'Black Plastic' should not be recyclable, got Recyclable=true")
		}
	}
}

func TestCanItBeRecycledSearch_maxLengthQueryIsAccepted(t *testing.T) {
	svc := newTestService(t, &stubResolver{}, &stubClassifier{})

	resp, err := svc.CanItBeRecycledSearch(context.Background(), &pb.CanItBeRecycledSearchRequest{Query: strings.Repeat("a", 200)})
	if err != nil {
		t.Fatalf("max-length query should be accepted: %v", err)
	}
	// No materials match 200 'a's; response is empty but not an error.
	if resp == nil {
		t.Error("expected non-nil response")
	}
}
