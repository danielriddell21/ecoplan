package providers

import (
	"testing"
)

func TestDetectMediaType(t *testing.T) {
	webp := make([]byte, 12)
	copy(webp[0:], "RIFF")
	copy(webp[8:], "WEBP")

	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"png", []byte("\x89PNG extra"), "image/png"},
		{"gif", []byte("GIF8 extra"), "image/gif"},
		{"webp", webp, "image/webp"},
		{"jpeg fallback", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "image/jpeg"},
		{"short (3 bytes)", []byte{0x89, 0x50, 0x4E}, "image/jpeg"},
		{"empty", []byte{}, "image/jpeg"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := detectMediaType(tc.data)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseItemResponse(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{"json with item", `{"item":"plastic bottle"}`, "plastic bottle"},
		{"json with whitespace", "  {\"item\":\"glass jar\"}  ", "glass jar"},
		{"non-json falls back to raw text", "cardboard box", "cardboard box"},
		{"invalid json falls back to raw text", `{"item":}`, `{"item":}`},
		{"json missing item field returns empty string", `{"type":"bottle"}`, ""},
		{"empty string", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseItemResponse(tc.text)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
