/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
 **/
package llm

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// ClaudeCodeProvider runs local Claude Code tasks.
type ClaudeCodeProvider struct {
	binaryPath string
}

// NewClaudeCodeProvider creates a new instance of ClaudeCodeProvider.
func NewClaudeCodeProvider(runtime *shared.RuntimeContext) *ClaudeCodeProvider {
	binary := runtime.ClaudeBinary
	if binary == "" {
		binary = "claude"
	}
	return &ClaudeCodeProvider{binaryPath: binary}
}

// Name returns the provider identifier.
func (p *ClaudeCodeProvider) Name() string {
	return "claude_code"
}

// Execute runs the local claude binary using non-interactive print mode.
func (p *ClaudeCodeProvider) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	prompt := req.Prompt
	if sysPrompt, ok := req.Context["system_prompt"]; ok && sysPrompt != "" {
		prompt = sysPrompt + "\n\n" + prompt
	}

	// Execute prompt in non-interactive print mode with tools disabled for security
	args := []string{"-p", "--tools", "", prompt}

	fields := strings.Fields(p.binaryPath)
	if len(fields) == 0 {
		return nil, fmt.Errorf("invalid claude binary path: %q", p.binaryPath)
	}

	cmdArgs := append(fields[1:], args...)
	cmd := exec.CommandContext(ctx, fields[0], cmdArgs...)
	if repoRoot, ok := req.Context["repo_root"]; ok && repoRoot != "" {
		cmd.Dir = repoRoot
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("claude code error: %w\nstderr: %s\nstdout: %s", err, stderr.String(), stdout.String())
	}

	output := sanitizeCLIOutput(stdout.String())
	return &ExecutionResponse{
		Output: output,
		Raw:    stdout.String(),
	}, nil
}
