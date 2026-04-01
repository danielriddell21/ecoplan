package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/ecoscan/service/internal/models"
)

// OpenFoodFactsResolver implements BarcodeResolver using the Open Food Facts API.
type OpenFoodFactsResolver struct {
	baseURL string
	timeout time.Duration
	client  *http.Client
	log     *slog.Logger
}

func NewOpenFoodFactsResolver(baseURL string, timeout time.Duration, log *slog.Logger) *OpenFoodFactsResolver {
	return &OpenFoodFactsResolver{
		baseURL: baseURL,
		timeout: timeout,
		client:  &http.Client{},
		log:     log,
	}
}

func (r *OpenFoodFactsResolver) Resolve(ctx context.Context, barcode string) (BarcodeResult, error) {
	url := fmt.Sprintf(
		"%s/api/v2/product/%s?fields=product_name,brands,packaging_tags,packaging_materials_tags,packaging_shapes_tags,packaging_text",
		r.baseURL, barcode,
	)

	var resp *http.Response
	var err error

	// Retry up to 3 times with exponential backoff on transient errors.
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			wait := time.Duration(math.Pow(2, float64(attempt))) * 200 * time.Millisecond
			select {
			case <-ctx.Done():
				return BarcodeResult{}, ctx.Err()
			case <-time.After(wait):
			}
		}

		reqCtx, cancel := context.WithTimeout(ctx, r.timeout)
		req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
		if reqErr != nil {
			cancel()
			return BarcodeResult{}, fmt.Errorf("building request: %w", reqErr)
		}

		resp, err = r.client.Do(req)
		cancel()

		if err != nil {
			r.log.WarnContext(ctx, "openfoodfacts request failed", "attempt", attempt+1, "err", err)
			continue
		}
		if resp.StatusCode >= 500 {
			resp.Body.Close() //nolint:errcheck
			r.log.WarnContext(ctx, "openfoodfacts 5xx", "attempt", attempt+1, "status", resp.StatusCode)
			err = fmt.Errorf("upstream returned %d", resp.StatusCode)
			continue
		}
		break
	}

	if err != nil {
		return BarcodeResult{}, fmt.Errorf("openfoodfacts: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return BarcodeResult{}, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return BarcodeResult{}, fmt.Errorf("openfoodfacts returned %d", resp.StatusCode)
	}

	var off models.OFFResponse
	if decErr := json.NewDecoder(resp.Body).Decode(&off); decErr != nil {
		return BarcodeResult{}, fmt.Errorf("decoding response: %w", decErr)
	}

	// Status 0 means product not found in Open Food Facts.
	if off.Status == 0 {
		return BarcodeResult{}, ErrNotFound
	}

	tags := append(off.Product.PackagingTags, off.Product.PackagingMaterialsTags...)
	tags = append(tags, off.Product.PackagingShapesTags...)

	return BarcodeResult{
		PackagingTags: tags,
		ProductName:   off.Product.ProductName,
		Brand:         off.Product.Brands,
	}, nil
}
