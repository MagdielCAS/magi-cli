/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
 **/
package llm

import (
	"context"
	"time"
)

// IntelligenceProvider defines the common contract for all AI runtime executors,
// whether they are REST APIs (like OpenAI) or command-line agents (like Copilot CLI or Claude Code).
type IntelligenceProvider interface {
	Name() string
	Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error)
}

// ExecutionRequest represents a standardized intent execution request.
type ExecutionRequest struct {
	Capability string
	Prompt     string
	Context    map[string]string
	Options    ExecutionOptions
}

// ExecutionOptions provides tuning parameters for the model execution.
type ExecutionOptions struct {
	Timeout     time.Duration
	MaxTokens   int
	Temperature float32
}

// ExecutionResponse standardizes output returned by any provider.
type ExecutionResponse struct {
	Output string
	Raw    string
	Meta   map[string]string
}
