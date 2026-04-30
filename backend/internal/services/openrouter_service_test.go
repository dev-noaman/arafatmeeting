package services

import (
	"errors"
	"testing"
)

func TestStringContains_SubstringFound(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		substr  string
		want    bool
	}{
		{"exact match", "status 429", "status 429", true},
		{"substring", "API returned status 429 error", "status 429", true},
		{"not found", "status 200", "status 429", false},
		{"empty substring", "anything", "", true},
		{"empty string", "", "status", false},
		{"both empty", "", "", true},
		{"rate-limited keyword", "request was rate-limited by server", "rate-limited", true},
		{"rate limit keyword", "you hit rate limit", "rate limit", true},
		{"no rate limit", "normal request", "rate limit", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringContains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("stringContains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestIndexOfSubstring_Found(t *testing.T) {
	idx := indexOfSubstring("hello world", "world")
	if idx != 6 {
		t.Errorf("expected index 6, got %d", idx)
	}
}

func TestIndexOfSubstring_NotFound(t *testing.T) {
	idx := indexOfSubstring("hello world", "xyz")
	if idx != -1 {
		t.Errorf("expected index -1, got %d", idx)
	}
}

func TestIndexOfSubstring_AtStart(t *testing.T) {
	idx := indexOfSubstring("hello world", "hello")
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
}

func TestIsRateLimitError_NilError(t *testing.T) {
	if isRateLimitError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsRateLimitError_Status429(t *testing.T) {
	err := errors.New("OpenRouter API returned status 429: rate-limited")
	if !isRateLimitError(err) {
		t.Error("expected true for status 429 error")
	}
}

func TestIsRateLimitError_RateLimitedKeyword(t *testing.T) {
	err := errors.New("request was rate-limited")
	if !isRateLimitError(err) {
		t.Error("expected true for rate-limited keyword")
	}
}

func TestIsRateLimitError_RateLimitKeyword(t *testing.T) {
	err := errors.New("hit rate limit")
	if !isRateLimitError(err) {
		t.Error("expected true for rate limit keyword")
	}
}

func TestIsRateLimitError_OtherError(t *testing.T) {
	err := errors.New("connection refused")
	if isRateLimitError(err) {
		t.Error("expected false for non-rate-limit error")
	}
}

func TestSplitIntoChunks_SmallTranscript(t *testing.T) {
	svc := &OpenRouterService{}
	transcript := "[00:00:00] alice: Hello\n\n[00:00:05] bob: Hi"
	chunks := svc.splitIntoChunks(transcript, 1000)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk for small transcript, got %d", len(chunks))
	}
}

func TestSplitIntoChunks_LargeTranscript(t *testing.T) {
	svc := &OpenRouterService{}
	// Create a transcript with multiple paragraphs
	transcript := ""
	for i := 0; i < 100; i++ {
		transcript += "[00:00:00] speaker: This is a long paragraph with enough text to exceed the token limit when we have many of them. Lorem ipsum dolor sit amet.\n\n"
	}
	chunks := svc.splitIntoChunks(transcript, 500)
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks for large transcript, got %d", len(chunks))
	}
}

func TestSplitIntoChunks_EmptyString(t *testing.T) {
	svc := &OpenRouterService{}
	chunks := svc.splitIntoChunks("", 100)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty string, got %d", len(chunks))
	}
}
