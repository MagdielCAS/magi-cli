/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
 **/
package llm

import (
	"context"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// OpenAIProvider executes tasks against OpenAI-compatible APIs.
type OpenAIProvider struct {
	runtime *shared.RuntimeContext
}

// NewOpenAIProvider creates a new instance of OpenAIProvider.
func NewOpenAIProvider(runtime *shared.RuntimeContext) *OpenAIProvider {
	return &OpenAIProvider{runtime: runtime}
}

// Name returns the provider identifier.
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Execute performs the chat completion and returns the response.
func (p *OpenAIProvider) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	builder := NewServiceBuilder(p.runtime)

	switch req.Context["model_variant"] {
	case "light":
		builder.UseLightModel()
	case "fallback":
		builder.UseFallbackModel()
	default:
		builder.UseHeavyModel()
	}

	if keyOverride, ok := req.Context["api_key_override"]; ok && keyOverride != "" {
		builder.WithAPIKey(keyOverride)
	}

	service, err := builder.Build()
	if err != nil {
		return nil, err
	}

	chatReq := ChatCompletionRequest{
		Messages: []ChatMessage{
			{Role: "user", Content: req.Prompt},
		},
		Temperature: 0.2, // Default to a low temperature for deterministic tasks
	}

	if req.Options.Temperature != 0 {
		chatReq.Temperature = float64(req.Options.Temperature)
	}
	if req.Options.MaxTokens != 0 {
		chatReq.MaxTokens = float64(req.Options.MaxTokens)
	}

	if sysPrompt, ok := req.Context["system_prompt"]; ok && sysPrompt != "" {
		chatReq.Messages = append([]ChatMessage{{Role: "system", Content: sysPrompt}}, chatReq.Messages...)
	}

	if req.Capability == "generate_commit_message" || req.Capability == "fix_commit_message" {
		chatReq.ResponseFormat = CommitSchema
	}

	rawResp, err := service.ChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	var output = rawResp
	if req.Capability == "generate_commit_message" || req.Capability == "fix_commit_message" {
		parsed, err := parseCommitMessage(rawResp)
		if err == nil {
			output = parsed
		}
	}

	return &ExecutionResponse{
		Output: output,
		Raw:    rawResp,
	}, nil
}
