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

// CopilotCLIProvider runs local Copilot CLI tasks.
type CopilotCLIProvider struct {
	binaryPath string
}

// NewCopilotCLIProvider creates a new instance of CopilotCLIProvider.
func NewCopilotCLIProvider(runtime *shared.RuntimeContext) *CopilotCLIProvider {
	binary := runtime.CopilotBinary
	if binary == "" {
		binary = "copilot"
	}
	return &CopilotCLIProvider{binaryPath: binary}
}

// Name returns the provider identifier.
func (p *CopilotCLIProvider) Name() string {
	return "copilot_cli"
}

// Execute runs the local copilot binary using non-interactive mode.
func (p *CopilotCLIProvider) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	prompt := req.Prompt
	if sysPrompt, ok := req.Context["system_prompt"]; ok && sysPrompt != "" {
		prompt = sysPrompt + "\n\n" + prompt
	}

	// Execute prompt in non-interactive mode and output only response (-s/--silent)
	args := []string{"-p", prompt, "--silent"}

	fields := strings.Fields(p.binaryPath)
	if len(fields) == 0 {
		return nil, fmt.Errorf("invalid copilot binary path: %q", p.binaryPath)
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
		return nil, fmt.Errorf("copilot CLI error: %w\nstderr: %s\nstdout: %s", err, stderr.String(), stdout.String())
	}

	output := sanitizeCLIOutput(stdout.String())
	return &ExecutionResponse{
		Output: output,
		Raw:    stdout.String(),
	}, nil
}
