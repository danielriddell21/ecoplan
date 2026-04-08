package service_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/ecoscan/service/internal/providers"
	"github.com/ecoscan/service/internal/service"
)

func newTestService(t *testing.T, resolver providers.BarcodeResolver, classifier providers.ImageClassifier) *service.RecyclingServiceServer {
	t.Helper()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc, err := service.NewRecyclingServiceServer(log, resolver, classifier)
	if err != nil {
		t.Fatalf("NewRecyclingServiceServer: %v", err)
	}
	return svc
}
