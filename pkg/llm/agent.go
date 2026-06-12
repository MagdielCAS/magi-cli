package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

const (
	AIModelKey_HEAVY  = "heavy"
	DefaultHeavyModel = "gpt-4"
)

// Agent represents a generic AI agent
type Agent struct {
	Name              string
	Task              string
	Personality       string
	CompletionRequest CompletionRequest
	Runtime           *shared.RuntimeContext
}

// CompletionRequest wraps ChatCompletionRequest with additional fields
type CompletionRequest struct {
	ChatCompletionRequest
	ApiKey string
}

// Analyze performs the agent's analysis based on input
func (a *Agent) Analyze(input map[string]string) (string, error) {
	if a.Runtime == nil {
		return "", fmt.Errorf("runtime context is required for agent %s", a.Name)
	}

	// Build prompt from Task, Personality, and Input
	systemPrompt := fmt.Sprintf("You are %s. %s\n\nTask: %s", a.Name, a.Personality, a.Task)

	var userPromptBuilder strings.Builder
	for k, v := range input {
		userPromptBuilder.WriteString(fmt.Sprintf("%s:\n%s\n\n", k, v))
	}
	userPrompt := userPromptBuilder.String()

	provider, err := ResolveProvider(a.Runtime)
	if err != nil {
		return "", err
	}

	execCtx := map[string]string{
		"system_prompt": systemPrompt,
		"model_variant": "heavy", // Agents default to heavy reasoning model
	}

	if a.CompletionRequest.ApiKey != "" {
		execCtx["api_key_override"] = a.CompletionRequest.ApiKey
	}

	resp, err := provider.Execute(context.Background(), &ExecutionRequest{
		Capability: fmt.Sprintf("agent_%s", strings.ToLower(a.Name)),
		Prompt:     userPrompt,
		Context:    execCtx,
		Options: ExecutionOptions{
			MaxTokens:   int(a.CompletionRequest.MaxTokens),
			Temperature: float32(a.CompletionRequest.Temperature),
		},
	})
	if err != nil {
		return "", err
	}

	return resp.Output, nil
}
