package pr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectAgentGuidelines(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "AGENTS.md")
	content := "# Test Agent\n\nRules"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	guidelines, err := CollectAgentGuidelines(dir)
	if err != nil {
		t.Fatalf("CollectAgentGuidelines() error: %v", err)
	}
	if !strings.Contains(guidelines, "Test Agent") {
		t.Fatalf("expected guidelines to include file content, got: %s", guidelines)
	}
}

func TestLoadPullRequestTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pull_request_template.md")
	content := "**Description**"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := LoadPullRequestTemplate(path)
	if err != nil {
		t.Fatalf("LoadPullRequestTemplate() error: %v", err)
	}
	if result != content {
		t.Fatalf("expected %q, got %q", content, result)
	}
}
