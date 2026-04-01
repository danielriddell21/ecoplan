package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"
)

type claudeContent struct {
	Type   string        `json:"type"`
	Text   string        `json:"text,omitempty"`
	Source *claudeSource `json:"source,omitempty"`
}

type claudeSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type claudeMessage struct {
	Role    string          `json:"role"`
	Content []claudeContent `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeResponse struct {
	Content []claudeContent `json:"content"`
}

// ClaudeClassifier implements ImageClassifier using the Anthropic Claude API.
type ClaudeClassifier struct {
	apiKey  string
	model   string
	timeout time.Duration
	client  *http.Client
	log     *slog.Logger
}

func NewClaudeClassifier(apiKey, model string, timeout time.Duration, log *slog.Logger) *ClaudeClassifier {
	return &ClaudeClassifier{
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
		client:  &http.Client{},
		log:     log,
	}
}

func (c *ClaudeClassifier) Classify(ctx context.Context, imageBytes []byte) (string, error) {
	mediaType := detectMediaType(imageBytes)
	encoded := base64.StdEncoding.EncodeToString(imageBytes)

	body := claudeRequest{
		Model:     c.model,
		MaxTokens: 256,
		Messages: []claudeMessage{
			{
				Role: "user",
				Content: []claudeContent{
					{
						Type: "image",
						Source: &claudeSource{
							Type:      "base64",
							MediaType: mediaType,
							Data:      encoded,
						},
					},
					{
						Type: "text",
						Text: `Identify the primary item in this image. Reply with a single JSON object containing one field "item" with a short plain-English description (e.g. "plastic bottle", "glass jar", "cardboard box"). No other text.`,
					},
				},
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshalling request: %w", err)
	}

	var resp *http.Response
	var doErr error

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			wait := time.Duration(math.Pow(2, float64(attempt))) * 200 * time.Millisecond
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(wait):
			}
		}

		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(payload))
		if reqErr != nil {
			cancel()
			return "", fmt.Errorf("building request: %w", reqErr)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		resp, doErr = c.client.Do(req)
		cancel()

		if doErr != nil {
			c.log.WarnContext(ctx, "claude request failed", "attempt", attempt+1, "err", doErr)
			continue
		}
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			c.log.WarnContext(ctx, "claude 5xx", "attempt", attempt+1, "status", resp.StatusCode)
			doErr = fmt.Errorf("upstream returned %d", resp.StatusCode)
			continue
		}
		break
	}

	if doErr != nil {
		return "", fmt.Errorf("claude: %w", doErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude returned %d", resp.StatusCode)
	}

	var cr claudeResponse
	if decErr := json.NewDecoder(resp.Body).Decode(&cr); decErr != nil {
		return "", fmt.Errorf("decoding claude response: %w", decErr)
	}

	if len(cr.Content) == 0 {
		return "", fmt.Errorf("empty response from claude")
	}

	// Extract the "item" field from the JSON response.
	text := strings.TrimSpace(cr.Content[0].Text)
	var result struct {
		Item string `json:"item"`
	}
	if jsonErr := json.Unmarshal([]byte(text), &result); jsonErr != nil {
		// Fallback: use the raw text as the description.
		return text, nil
	}
	return result.Item, nil
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
