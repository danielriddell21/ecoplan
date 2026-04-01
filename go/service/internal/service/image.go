package service

import (
	"context"
	"errors"

	"github.com/ecoscan/service/internal/mappers"
	"github.com/ecoscan/service/internal/providers"
	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxImageBytes = 5 * 1024 * 1024 // 5 MB

func (s *RecyclingServiceServer) CanItBeRecycledImage(ctx context.Context, req *pb.CanItBeRecycledImageRequest) (*pb.CanItBeRecycledImageResponse, error) {
	img := req.GetImage()

	if len(img) == 0 {
		return nil, status.Error(codes.InvalidArgument, "image must not be empty")
	}
	if len(img) > maxImageBytes {
		return nil, status.Errorf(codes.InvalidArgument, "image exceeds maximum size of %d bytes", maxImageBytes)
	}

	itemDesc, err := s.classifier.Classify(ctx, img)
	if err != nil {
		if errors.Is(err, providers.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "could not identify item in image")
		}
		s.log.ErrorContext(ctx, "image classifier failed", "err", err)
		return nil, status.Error(codes.Internal, "failed to classify image")
	}

	materials := s.db.SearchMaterials(itemDesc)
	if len(materials) == 0 {
		return &pb.CanItBeRecycledImageResponse{
			Data: &pb.RecyclingItem{
				Recyclable: false,
				Advice:     "Unable to determine recycling information for: " + itemDesc,
			},
		}, nil
	}

	m := materials[0]
	return &pb.CanItBeRecycledImageResponse{
		Data: &pb.RecyclingItem{
			Recyclable: m.Recyclable == "yes",
			Advice:     m.Tips,
			BinColour:  mappers.ToBinColour(m.Bin),
			BinType:    mappers.ToBinType(m.Bin),
		},
	}, nil
}
