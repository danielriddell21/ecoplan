// Package providerstest provides test doubles for the providers interfaces.
package providerstest

import (
	"context"
	"sync/atomic"

	"github.com/ecoscan/service/internal/providers"
)

// StubResolver is a test double for providers.BarcodeResolver.
type StubResolver struct {
	Result providers.BarcodeResult
	Err    error
	Calls  atomic.Int32
}

func (s *StubResolver) Resolve(_ context.Context, _ string) (providers.BarcodeResult, error) {
	s.Calls.Add(1)
	return s.Result, s.Err
}

// StubClassifier is a test double for providers.ImageClassifier.
type StubClassifier struct {
	Item string
	Err  error
}

func (s *StubClassifier) Classify(_ context.Context, _ []byte) (string, error) {
	return s.Item, s.Err
}
