package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "embed"

	"github.com/ecoscan/service/internal/models"
	"github.com/ecoscan/service/internal/providers"
	pb "github.com/ecoscan/service/proto"
)

//go:embed materials.json
var materialsData []byte

// cacheEntry holds a cached barcode result with its expiry time.
type cacheEntry struct {
	result    providers.BarcodeResult
	expiresAt time.Time
}

// RecyclingServiceServer implements the gRPC RecyclingService.
type RecyclingServiceServer struct {
	pb.UnimplementedRecyclingServiceServer
	log        *slog.Logger
	db         models.MaterialsDB
	resolver   providers.BarcodeResolver
	classifier providers.ImageClassifier

	cacheMu sync.Mutex
	cache   map[string]cacheEntry
}

type materialsFile struct {
	Materials map[string]models.Material `json:"materials"`
	Bins      map[string]models.Bin      `json:"bins"`
}

// NewRecyclingServiceServer creates the service, loading the embedded materials DB.
func NewRecyclingServiceServer(
	log *slog.Logger,
	resolver providers.BarcodeResolver,
	classifier providers.ImageClassifier,
) (*RecyclingServiceServer, error) {
	var mf materialsFile
	if err := json.Unmarshal(materialsData, &mf); err != nil {
		return nil, fmt.Errorf("loading materials: %w", err)
	}

	return &RecyclingServiceServer{
		log: log,
		db: models.MaterialsDB{
			Materials: mf.Materials,
			Bins:      mf.Bins,
		},
		resolver:   resolver,
		classifier: classifier,
		cache:      make(map[string]cacheEntry),
	}, nil
}

func (s *RecyclingServiceServer) storeCache(barcode string, result providers.BarcodeResult) {
	s.cacheMu.Lock()
	s.cache[barcode] = cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(24 * time.Hour),
	}
	s.cacheMu.Unlock()
}

func (s *RecyclingServiceServer) lookupCache(barcode string) (providers.BarcodeResult, bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	entry, ok := s.cache[barcode]
	if !ok || time.Now().After(entry.expiresAt) {
		return providers.BarcodeResult{}, false
	}
	return entry.result, true
}
