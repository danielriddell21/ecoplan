package providers

import "context"

// ImageClassifier identifies the item shown in an image and returns a plain-text description.
type ImageClassifier interface {
	Classify(ctx context.Context, imageBytes []byte) (itemDescription string, err error)
}
