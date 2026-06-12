/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
 **/
package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

func TestSanitizeCLIOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean output",
			input:    "feat(cli): add super power",
			expected: "feat(cli): add super power",
		},
		{
			name: "deprecation warnings filtered",
			input: `The gh-copilot extension has been deprecated in favor of the newer GitHub Copilot CLI.
For more information, visit:
- Copilot CLI: https://github.com/github/copilot-cli
- Deprecation announcement: https://github.blog/changelog/2025-09-25-upcoming-deprecation-of-gh-copilot-cli-extension
No commands will be executed.
feat(cli): add super power`,
			expected: "feat(cli): add super power",
		},
		{
			name: "multi line suggestion",
			input: `line 1
line 2
deprecated line`,
			expected: "line 1\nline 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeCLIOutput(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRegistryAndResolve(t *testing.T) {
	runtime := &shared.RuntimeContext{
		Provider:      "copilot_cli",
		CopilotBinary: "copilot",
	}

	provider, err := ResolveProvider(runtime)
	if err != nil {
		t.Fatalf("failed to resolve provider: %v", err)
	}

	if provider.Name() != "copilot_cli" {
		t.Errorf("expected provider 'copilot_cli', got %q", provider.Name())
	}

	runtime.Provider = "claude_code"
	runtime.ClaudeBinary = "claude"
	provider, err = ResolveProvider(runtime)
	if err != nil {
		t.Fatalf("failed to resolve provider: %v", err)
	}

	if provider.Name() != "claude_code" {
		t.Errorf("expected provider 'claude_code', got %q", provider.Name())
	}
}

func TestCopilotCLIProvider_Execute(t *testing.T) {
	// Use 'echo' as the binary to test arguments routing
	runtime := &shared.RuntimeContext{
		Provider:      "copilot_cli",
		CopilotBinary: "echo",
	}

	provider, err := ResolveProvider(runtime)
	if err != nil {
		t.Fatalf("failed to resolve provider: %v", err)
	}

	ctx := context.Background()
	req := &ExecutionRequest{
		Capability: "test",
		Prompt:     "hello-world-prompt",
	}

	resp, err := provider.Execute(ctx, req)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Echo outputs arguments separated by space. With -p "prompt" --silent, output should contain them.
	if !strings.Contains(resp.Output, "hello-world-prompt") {
		t.Errorf("expected output to contain prompt, got %q", resp.Output)
	}
}

func TestClaudeCodeProvider_Execute(t *testing.T) {
	// Use 'echo' as the binary to test arguments routing
	runtime := &shared.RuntimeContext{
		Provider:     "claude_code",
		ClaudeBinary: "echo",
	}

	provider, err := ResolveProvider(runtime)
	if err != nil {
		t.Fatalf("failed to resolve provider: %v", err)
	}

	ctx := context.Background()
	req := &ExecutionRequest{
		Capability: "test",
		Prompt:     "hello-world-prompt",
	}

	resp, err := provider.Execute(ctx, req)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Echo outputs arguments separated by space. With -p --tools "" "prompt", output should contain them.
	if !strings.Contains(resp.Output, "hello-world-prompt") {
		t.Errorf("expected output to contain prompt, got %q", resp.Output)
	}
}
