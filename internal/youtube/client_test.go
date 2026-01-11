package youtube

import (
	"testing"

	"google.golang.org/api/youtube/v3"
)

func TestGetBestThumbnail(t *testing.T) {
	tests := []struct {
		name     string
		input    *youtube.ThumbnailDetails
		expected string
	}{
		{
			name:     "nil thumbnails",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty thumbnails",
			input:    &youtube.ThumbnailDetails{},
			expected: "",
		},
		{
			name: "only default",
			input: &youtube.ThumbnailDetails{
				Default: &youtube.Thumbnail{Url: "https://example.com/default.jpg"},
			},
			expected: "https://example.com/default.jpg",
		},
		{
			name: "only high",
			input: &youtube.ThumbnailDetails{
				High: &youtube.Thumbnail{Url: "https://example.com/high.jpg"},
			},
			expected: "https://example.com/high.jpg",
		},
		{
			name: "only medium",
			input: &youtube.ThumbnailDetails{
				Medium: &youtube.Thumbnail{Url: "https://example.com/medium.jpg"},
			},
			expected: "https://example.com/medium.jpg",
		},
		{
			name: "all available prefers medium",
			input: &youtube.ThumbnailDetails{
				Default: &youtube.Thumbnail{Url: "https://example.com/default.jpg"},
				Medium:  &youtube.Thumbnail{Url: "https://example.com/medium.jpg"},
				High:    &youtube.Thumbnail{Url: "https://example.com/high.jpg"},
			},
			expected: "https://example.com/medium.jpg",
		},
		{
			name: "high and default prefers high",
			input: &youtube.ThumbnailDetails{
				Default: &youtube.Thumbnail{Url: "https://example.com/default.jpg"},
				High:    &youtube.Thumbnail{Url: "https://example.com/high.jpg"},
			},
			expected: "https://example.com/high.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBestThumbnail(tt.input)
			if result != tt.expected {
				t.Errorf("getBestThumbnail() = %q, want %q", result, tt.expected)
			}
		})
	}
}
