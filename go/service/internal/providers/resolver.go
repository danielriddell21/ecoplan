package providers

import "context"

// BarcodeResult holds the packaging tags and product metadata returned by a barcode lookup.
type BarcodeResult struct {
	PackagingTags []string
	ProductName   string
	Brand         string
}

// BarcodeResolver resolves a product barcode to its packaging information.
type BarcodeResolver interface {
	Resolve(ctx context.Context, barcode string) (BarcodeResult, error)
}
