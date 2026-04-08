package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeClassifier implements ImageClassifier using the Anthropic Claude API.
type ClaudeClassifier struct {
	client  anthropic.Client
	model   string
	timeout time.Duration
	log     *slog.Logger
}

func NewClaudeClassifier(apiKey, model string, timeout time.Duration, log *slog.Logger) *ClaudeClassifier {
	return &ClaudeClassifier{
		client:  anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:   model,
		timeout: timeout,
		log:     log,
	}
}

func (c *ClaudeClassifier) Classify(ctx context.Context, imageBytes []byte) (string, error) {
	mediaType := detectMediaType(imageBytes)
	encoded := base64.StdEncoding.EncodeToString(imageBytes)

	reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.Messages.New(reqCtx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64(mediaType, encoded),
				anthropic.NewTextBlock(`Identify the primary item in this image. Reply with a single JSON object containing one field "item" with a short plain-English description (e.g. "plastic bottle", "glass jar", "cardboard box"). No other text.`),
			),
		},
	})
	if err != nil {
		c.log.ErrorContext(ctx, "claude request failed", "err", err)
		return "", fmt.Errorf("claude: %w", err)
	}

	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			return parseItemResponse(text.Text), nil
		}
	}

	return "", fmt.Errorf("empty response from claude")
}

// parseItemResponse extracts the item description from Claude's text response.
// It expects a JSON object with an "item" field; if parsing fails it returns the
// trimmed raw text as a fallback.
func parseItemResponse(text string) string {
	text = strings.TrimSpace(text)
	var result struct {
		Item string `json:"item"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return text
	}
	return result.Item
}

// detectMediaType inspects the first bytes of the image to determine its MIME type.
func detectMediaType(data []byte) string {
	if len(data) < 4 {
		return "image/jpeg"
	}
	switch {
	case bytes.HasPrefix(data, []byte("\x89PNG")):
		return "image/png"
	case bytes.HasPrefix(data, []byte("GIF8")):
		return "image/gif"
	case bytes.HasPrefix(data, []byte("RIFF")) && len(data) > 8 && string(data[8:12]) == "WEBP":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
