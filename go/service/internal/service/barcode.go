package service

import (
	"context"
	"errors"
	"regexp"
"github.com/ecoscan/service/internal/mappers"
	"github.com/ecoscan/service/internal/providers"
	pb "github.com/ecoscan/service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var barcodeRE = regexp.MustCompile(`^\d{8,14}$`)

func (s *RecyclingServiceServer) CanItBeRecycledBarcode(ctx context.Context, req *pb.CanItBeRecycledBarcodeRequest) (*pb.CanItBeRecycledBarcodeResponse, error) {
	barcode := req.GetBarcode()

	// Validate barcode: 8–14 digits only.
	if !barcodeRE.MatchString(barcode) {
		return nil, status.Errorf(codes.InvalidArgument, "barcode must be 8–14 digits, got %q", barcode)
	}

	// Check cache first.
	if cached, ok := s.lookupCache(barcode); ok {
		s.log.DebugContext(ctx, "cache hit", "barcode", barcode)
		return s.buildBarcodeResponse(cached), nil
	}

	result, err := s.resolver.Resolve(ctx, barcode)
	if err != nil {
		if errors.Is(err, providers.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "product %q not found", barcode)
		}
		s.log.ErrorContext(ctx, "barcode resolver failed", "barcode", barcode, "err", err)
		return nil, status.Errorf(codes.Internal, "failed to look up product")
	}

	s.storeCache(barcode, result)

	if len(result.PackagingTags) == 0 {
		return nil, status.Errorf(codes.NotFound, "no packaging information for product %q", barcode)
	}

	return s.buildBarcodeResponse(result), nil
}

func (s *RecyclingServiceServer) buildBarcodeResponse(result providers.BarcodeResult) *pb.CanItBeRecycledBarcodeResponse {
	materials := s.db.LookupMaterials(result.PackagingTags)
	if len(materials) == 0 {
		return &pb.CanItBeRecycledBarcodeResponse{
			ProductName: result.ProductName,
			Brand:       result.Brand,
			Data: &pb.RecyclingItem{
				Recyclable: false,
				Advice:     "Unable to determine recycling information for this product's packaging.",
			},
		}
	}

	// Use the first matched material as the primary result.
	m := materials[0]
	return &pb.CanItBeRecycledBarcodeResponse{
		ProductName: result.ProductName,
		Brand:       result.Brand,
		Data: &pb.RecyclingItem{
			Recyclable: m.Recyclable == "yes",
			Advice:     m.Tips,
			BinColour:  mappers.ToBinColour(m.Bin),
			BinType:    mappers.ToBinType(m.Bin),
		},
	}
}
