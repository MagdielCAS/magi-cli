package pr

import (
	"context"
	"fmt"
	"strings"

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

	service, err := buildServiceWithFallback(a.runtime, []llm.ModelVariant{
		llm.ModelVariantHeavy,
		llm.ModelVariantFallback,
		llm.ModelVariantLight,
	})
	if err != nil {
		return "", fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: analysisSystemPrompt},
			{Role: "user", Content: payload},
		},
		Temperature:    0.2,
		MaxTokens:      4096,
		ResponseFormat: AnalysisSchema,
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.runtime.AnalysisTimeout)
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
	if templateContent == "" {
		return "", fmt.Errorf("template is missing")
	}
	branch := input["branch"]
	if branch == "" {
		return "", fmt.Errorf("branch is missing")
	}

	writerPayload, err := renderWriterPayload(writerPayloadParams{
		AnalysisJSON: analysisJSON,
		Template:     templateContent,
		Branch:       branch,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render writer payload: %w", err)
	}

	service, err := buildServiceWithFallback(a.runtime, []llm.ModelVariant{
		llm.ModelVariantLight,
		llm.ModelVariantHeavy,
		llm.ModelVariantFallback,
	})
	if err != nil {
		return "", fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: writerSystemPrompt},
			{Role: "user", Content: writerPayload},
		},
		Temperature:    0.25,
		MaxTokens:      2048,
		ResponseFormat: WriterSchema,
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.runtime.WriterTimeout)
	defer cancel()

	return service.ChatCompletion(ctx, req)
}

func buildServiceWithFallback(runtime *shared.RuntimeContext, variants []llm.ModelVariant) (*llm.Service, error) {
	var firstErr error

	for _, variant := range variants {
		if !variantConfigured(runtime, variant) {
			continue
		}

		builder := llm.NewServiceBuilder(runtime)
		switch variant {
		case llm.ModelVariantLight:
			builder.UseLightModel()
		case llm.ModelVariantFallback:
			builder.UseFallbackModel()
		default:
			builder.UseHeavyModel()
		}

		service, err := builder.Build()
		if err == nil {
			return service, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return nil, fmt.Errorf("no configured model found for variants %s", stringifyVariants(variants))
}

func variantConfigured(runtime *shared.RuntimeContext, variant llm.ModelVariant) bool {
	switch variant {
	case llm.ModelVariantLight:
		return strings.TrimSpace(runtime.LightModel) != ""
	case llm.ModelVariantFallback:
		return strings.TrimSpace(runtime.Fallback) != ""
	default:
		return strings.TrimSpace(runtime.HeavyModel) != ""
	}
}

func stringifyVariants(variants []llm.ModelVariant) string {
	labels := make([]string, 0, len(variants))
	for _, variant := range variants {
		switch variant {
		case llm.ModelVariantLight:
			labels = append(labels, "light")
		case llm.ModelVariantFallback:
			labels = append(labels, "fallback")
		default:
			labels = append(labels, "heavy")
		}
	}

	return strings.Join(labels, ", ")
}
