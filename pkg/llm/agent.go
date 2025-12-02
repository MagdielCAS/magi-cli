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

	// Use ServiceBuilder
	// We default to Heavy model for agents as they usually require more reasoning
	builder := NewServiceBuilder(a.Runtime).UseHeavyModel()
	if a.CompletionRequest.ApiKey != "" {
		builder.WithAPIKey(a.CompletionRequest.ApiKey)
	}

	service, err := builder.Build()
	if err != nil {
		return "", err
	}

	req := a.CompletionRequest.ChatCompletionRequest
	req.Messages = []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	return service.ChatCompletion(context.Background(), req)
}
