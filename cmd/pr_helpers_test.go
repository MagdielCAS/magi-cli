package cmd

import (
	"strings"
	"testing"

	"github.com/MagdielCAS/magi-cli/pkg/pr"
)

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

func TestGenerateMarkdownReport(t *testing.T) {
	artifacts := pr.ReviewArtifacts{
		Plan: pr.PullRequestPlan{
			Title: "Test PR",
			Body:  "Test Body",
		},
		Analysis: pr.AgentFindings{
			Summary:    "Test Summary",
			CodeSmells: []string{"Smell 1"},
		},
	}

	report := generateMarkdownReport(artifacts)

	if !strings.Contains(report, "# Pull Request Plan") {
		t.Error("Report missing plan header")
	}
	if !strings.Contains(report, "Test PR") {
		t.Error("Report missing title")
	}
	if !strings.Contains(report, "Test Body") {
		t.Error("Report missing body")
	}
	if !strings.Contains(report, "Test Summary") {
		t.Error("Report missing summary")
	}
	if !strings.Contains(report, "Smell 1") {
		t.Error("Report missing code smell")
	}
}
