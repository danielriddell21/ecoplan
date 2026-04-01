package service

import (
	"context"
	"strings"
	"unicode"

	"github.com/ecoscan/service/internal/mappers"
	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxQueryLen = 200

func (s *RecyclingServiceServer) CanItBeRecycledSearch(ctx context.Context, req *pb.CanItBeRecycledSearchRequest) (*pb.CanItBeRecycledSearchResponse, error) {
	query := strings.TrimSpace(req.GetQuery())

	if query == "" {
		return nil, status.Error(codes.InvalidArgument, "query must not be empty")
	}
	if len(query) > maxQueryLen {
		return nil, status.Errorf(codes.InvalidArgument, "query exceeds maximum length of %d characters", maxQueryLen)
	}
	if containsControlChars(query) {
		return nil, status.Error(codes.InvalidArgument, "query contains invalid characters")
	}

	materials := s.db.SearchMaterials(query)

	items := make([]*pb.RecyclingItem, 0, len(materials))
	for _, m := range materials {
		items = append(items, &pb.RecyclingItem{
			Recyclable: m.Recyclable == "yes",
			Advice:     m.Tips,
			BinColour:  mappers.ToBinColour(m.Bin),
			BinType:    mappers.ToBinType(m.Bin),
		})
	}

	return &pb.CanItBeRecycledSearchResponse{Data: items}, nil
}

func containsControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
