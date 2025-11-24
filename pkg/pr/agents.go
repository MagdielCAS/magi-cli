package pr

import (
	"context"
	"fmt"
	"time"

	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// AnalysisAgent performs the initial code analysis
type AnalysisAgent struct {
	runtime *shared.RuntimeContext
}

func NewAnalysisAgent(runtime *shared.RuntimeContext) *AnalysisAgent {
	return &AnalysisAgent{runtime: runtime}
}

func (a *AnalysisAgent) Name() string {
	return "AnalysisAgent"
}

func (a *AnalysisAgent) WaitForResults() []string {
	return []string{}
}

func (a *AnalysisAgent) Execute(input map[string]string) (string, error) {
	// Reconstruct ReviewInput from input map
	// We expect the payload to be pre-rendered or passed as raw components.
	// To keep it simple and consistent with previous logic, let's assume the input contains the rendered payload
	// or we render it here.
	// The AgentManager passes a map.

	payload := input["payload"]
	if payload == "" {
		return "", fmt.Errorf("payload is missing")
	}

	// Build LLM service
	builder := llm.NewServiceBuilder(a.runtime).
		UseHeavyModel() // Analysis uses heavy model (or configured preference)

	// We can allow overriding via input if needed, but for now stick to defaults
	service, err := builder.Build()
	if err != nil {
		return "", fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: analysisSystemPrompt},
			{Role: "user", Content: payload},
		},
		Temperature: 0.2,
		MaxTokens:   4096,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	return service.ChatCompletion(ctx, req)
}

// WriterAgent generates the PR description
type WriterAgent struct {
	runtime *shared.RuntimeContext
}

func NewWriterAgent(runtime *shared.RuntimeContext) *WriterAgent {
	return &WriterAgent{runtime: runtime}
}

func (a *WriterAgent) Name() string {
	return "WriterAgent"
}

func (a *WriterAgent) WaitForResults() []string {
	return []string{"AnalysisAgent"}
}

func (a *WriterAgent) Execute(input map[string]string) (string, error) {
	analysisJSON := input["AnalysisAgent"]
	if analysisJSON == "" {
		return "", fmt.Errorf("analysis result is missing")
	}

	// We need template and branch to render the writer payload.
	// These should be passed in the initial input.
	templateContent := input["template"]
	branch := input["branch"]

	writerPayload, err := renderWriterPayload(writerPayloadParams{
		AnalysisJSON: analysisJSON,
		Template:     templateContent,
		Branch:       branch,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render writer payload: %w", err)
	}

	// Build LLM service - Writer uses light model
	builder := llm.NewServiceBuilder(a.runtime).
		UseLightModel()

	service, err := builder.Build()
	if err != nil {
		return "", fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: writerSystemPrompt},
			{Role: "user", Content: writerPayload},
		},
		Temperature: 0.25,
		MaxTokens:   2048,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return service.ChatCompletion(ctx, req)
}
