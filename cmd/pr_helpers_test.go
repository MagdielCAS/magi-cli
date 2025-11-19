package cmd

import "testing"

func TestSanitizeCommandOutputTruncates(t *testing.T) {
	long := make([]byte, 600)
	for i := range long {
		long[i] = 'a'
	}
	result := sanitizeCommandOutput(string(long))
	if len(result) >= len(long) {
		t.Fatalf("expected sanitized output to be truncated, length=%d", len(result))
	}
	if wantPrefix := "aaaaa"; result[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("unexpected prefix: %s", result[:10])
	}
	if suffix := "... (truncated)"; len(result) < len(suffix) || result[len(result)-len(suffix):] != suffix {
		t.Fatalf("expected suffix %q, got %q", suffix, result[len(result)-len(suffix):])
	}
}

func TestSanitizeCommandOutputEmpty(t *testing.T) {
	if got := sanitizeCommandOutput("\n  "); got != "no additional details" {
		t.Fatalf("expected fallback message, got %q", got)
	}
}
