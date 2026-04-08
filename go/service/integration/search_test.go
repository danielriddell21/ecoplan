package integration_test

import (
	"strings"
	"testing"

	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSearch_happy(t *testing.T) {
	f := newIntegrationServer(t)

	resp, err := f.client.CanItBeRecycledSearch(authCtx(t), &pb.CanItBeRecycledSearchRequest{Query: "glass"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) == 0 {
		t.Error("expected at least one result for 'glass'")
	}
}

func TestSearch_emptyQuery(t *testing.T) {
	f := newIntegrationServer(t)

	_, err := f.client.CanItBeRecycledSearch(authCtx(t), &pb.CanItBeRecycledSearchRequest{Query: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for empty query, got %v", err)
	}
}

func TestSearch_queryTooLong(t *testing.T) {
	f := newIntegrationServer(t)

	_, err := f.client.CanItBeRecycledSearch(authCtx(t), &pb.CanItBeRecycledSearchRequest{Query: strings.Repeat("a", 201)})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for too-long query, got %v", err)
	}
}

func TestSearch_noResults(t *testing.T) {
	f := newIntegrationServer(t)

	resp, err := f.client.CanItBeRecycledSearch(authCtx(t), &pb.CanItBeRecycledSearchRequest{Query: "xyzzy_no_match"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Data))
	}
}
