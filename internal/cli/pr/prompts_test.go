package pr

import (
	"strings"
	"testing"
)

func TestRenderAnalysisPayload(t *testing.T) {
	input := ReviewInput{
		Diff:      "diff --git a/file.go b/file.go",
		Branch:    "feature/pr",
		RemoteRef: "origin/feature/pr",
		Guidelines: `# Rules
- Be safe`,
		AdditionalContext: "Need to ship today",
		Template:          "template",
	}

	payload, err := renderAnalysisPayload(input)
	if err != nil {
		t.Fatalf("renderAnalysisPayload() unexpected error: %v", err)
	}

	assertContains(t, payload, input.Diff)
	assertContains(t, payload, input.Branch)
	assertContains(t, payload, input.RemoteRef)
	assertContains(t, payload, input.Guidelines)
	assertContains(t, payload, input.AdditionalContext)
}

func TestRenderWriterPayload(t *testing.T) {
	params := writerPayloadParams{
		AnalysisJSON: `{"summary":"ok"}`,
		Template:     "**Description**",
		Branch:       "feature/pr",
	}

	payload, err := renderWriterPayload(params)
	if err != nil {
		t.Fatalf("renderWriterPayload() unexpected error: %v", err)
	}

	assertContains(t, payload, params.AnalysisJSON)
	assertContains(t, payload, params.Template)
	assertContains(t, payload, params.Branch)
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected payload to contain %q\npayload: %s", needle, haystack)
	}
}
