package llm

import (
	"strings"
	"testing"
)

func TestRenderCommitPrompt(t *testing.T) {
	diff := "diff --git a/foo b/foo\n+hello"
	prompt, err := renderCommitPrompt(diff)
	if err != nil {
		t.Fatalf("renderCommitPrompt returned error: %v", err)
	}

	if !strings.Contains(prompt, diff) {
		t.Fatalf("prompt did not include diff content")
	}

	if !strings.Contains(prompt, "Rules:") {
		t.Fatalf("prompt is missing commit rules")
	}
}
