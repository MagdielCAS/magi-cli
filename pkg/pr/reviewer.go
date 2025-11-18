package pr

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/azureopenai"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/openai"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

const (
	analysisStepName = "secure_analysis"
	writerStepName   = "pr_template_writer"
)

// AgentFindings captures the structured response from the analysis agent.
type AgentFindings struct {
	Summary               string   `json:"summary"`
	CodeSmells            []string `json:"code_smells"`
	SecurityConcerns      []string `json:"security_concerns"`
	AgentsGuidelineAlerts []string `json:"agents_guideline_alerts"`
	TestRecommendations   []string `json:"test_recommendations"`
	DocumentationUpdates  []string `json:"documentation_updates"`
	RiskCallouts          []string `json:"risk_callouts"`
}

// PullRequestPlan stores the generated title and filled template.
type PullRequestPlan struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// ReviewArtifacts groups the analysis findings with the final PR plan.
type ReviewArtifacts struct {
	Analysis AgentFindings
	Plan     PullRequestPlan
}

// AgenticReviewer orchestrates the AgenticGoKit workflow for PR prep.
type AgenticReviewer struct {
	runtime *shared.RuntimeContext
}

// NewAgenticReviewer creates a reviewer bound to the shared runtime context.
func NewAgenticReviewer(runtime *shared.RuntimeContext) *AgenticReviewer {
	return &AgenticReviewer{runtime: runtime}
}

// Review executes the multi-agent workflow and returns structured artifacts.
func (r *AgenticReviewer) Review(ctx context.Context, input ReviewInput) (*ReviewArtifacts, error) {
	if r == nil || r.runtime == nil {
		return nil, fmt.Errorf("runtime context is required")
	}

	payload, err := renderAnalysisPayload(input)
	if err != nil {
		return nil, err
	}

	analysisAgent, err := r.buildAgent("magi-secure-reviewer", analysisSystemPrompt, agentPreferences{
		preferLightModel: false,
		temperature:      0.2,
		maxTokens:        4096,
		timeout:          3 * time.Minute,
	})
	if err != nil {
		return nil, err
	}

	writerAgent, err := r.buildAgent("magi-pr-writer", writerSystemPrompt, agentPreferences{
		preferLightModel: true,
		temperature:      0.25,
		maxTokens:        2048,
		timeout:          2 * time.Minute,
	})
	if err != nil {
		return nil, err
	}

	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Timeout: 3 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	if err := workflow.AddStep(vnext.WorkflowStep{
		Name:  analysisStepName,
		Agent: analysisAgent,
	}); err != nil {
		return nil, fmt.Errorf("failed to add analysis step: %w", err)
	}

	if err := workflow.AddStep(vnext.WorkflowStep{
		Name:  writerStepName,
		Agent: writerAgent,
		Transform: func(previousOutput string) string {
			writerPayload, err := renderWriterPayload(writerPayloadParams{
				AnalysisJSON: previousOutput,
				Template:     input.Template,
				Branch:       input.Branch,
			})
			if err != nil {
				return fmt.Sprintf("Unable to render writer payload (%v). Last analysis output:\n%s", err, previousOutput)
			}
			return writerPayload
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to add writer step: %w", err)
	}

	result, err := workflow.Run(ctx, payload)
	if err != nil {
		return nil, err
	}

	var artifacts ReviewArtifacts
	for _, step := range result.StepResults {
		switch step.StepName {
		case analysisStepName:
			if err := json.Unmarshal([]byte(step.Output), &artifacts.Analysis); err != nil {
				return nil, fmt.Errorf("analysis agent produced invalid JSON: %w (raw: %s)", err, sanitizeForError(step.Output))
			}
		case writerStepName:
			if err := json.Unmarshal([]byte(step.Output), &artifacts.Plan); err != nil {
				return nil, fmt.Errorf("PR writer agent produced invalid JSON: %w (raw: %s)", err, sanitizeForError(step.Output))
			}
		}
	}

	if strings.TrimSpace(artifacts.Plan.Title) == "" {
		return nil, fmt.Errorf("PR writer did not return a title")
	}
	if strings.TrimSpace(artifacts.Plan.Body) == "" {
		return nil, fmt.Errorf("PR writer did not return a body")
	}

	return &artifacts, nil
}

type agentPreferences struct {
	preferLightModel bool
	temperature      float64
	maxTokens        int
	timeout          time.Duration
}

type modelSelection struct {
	provider string
	model    string
	apiKey   string
	baseURL  string
}

func (r *AgenticReviewer) buildAgent(name, prompt string, prefs agentPreferences) (vnext.Agent, error) {
	selection, err := r.selectModel(prefs.preferLightModel)
	if err != nil {
		return nil, err
	}

	opts := []vnext.Option{
		vnext.WithSystemPrompt(prompt),
		vnext.WithLLMConfig(selection.provider, selection.model, prefs.temperature, prefs.maxTokens),
		withAPIKey(selection.apiKey),
		withBaseURL(selection.baseURL),
		withTimeout(prefs.timeout),
	}

	return vnext.NewDataAgent(name, opts...)
}

func (r *AgenticReviewer) selectModel(preferLight bool) (modelSelection, error) {
	type candidate struct {
		model    string
		endpoint shared.ModelEndpoint
	}

	var pick candidate
	if preferLight && strings.TrimSpace(r.runtime.LightModel) != "" {
		pick = candidate{
			model:    strings.TrimSpace(r.runtime.LightModel),
			endpoint: r.runtime.LightEndpoint,
		}
	} else if strings.TrimSpace(r.runtime.HeavyModel) != "" {
		pick = candidate{
			model:    strings.TrimSpace(r.runtime.HeavyModel),
			endpoint: r.runtime.HeavyEndpoint,
		}
	} else if strings.TrimSpace(r.runtime.Fallback) != "" {
		pick = candidate{
			model:    strings.TrimSpace(r.runtime.Fallback),
			endpoint: r.runtime.FallbackEndpoint,
		}
	}

	if pick.model == "" {
		return modelSelection{}, fmt.Errorf("no LLM model configured")
	}

	apiKey := strings.TrimSpace(pick.endpoint.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(r.runtime.APIKey)
	}
	if apiKey == "" {
		return modelSelection{}, fmt.Errorf("API key is required for agent execution")
	}

	baseURL := strings.TrimSpace(pick.endpoint.BaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(r.runtime.BaseURL)
	}

	return modelSelection{
		provider: strings.TrimSpace(strings.ToLower(r.runtime.Provider)),
		model:    pick.model,
		apiKey:   apiKey,
		baseURL:  baseURL,
	}, nil
}

func withAPIKey(key string) vnext.Option {
	return func(cfg *vnext.Config) {
		cfg.LLM.APIKey = strings.TrimSpace(key)
	}
}

func withBaseURL(baseURL string) vnext.Option {
	return func(cfg *vnext.Config) {
		cfg.LLM.BaseURL = strings.TrimSpace(baseURL)
	}
}

func withTimeout(duration time.Duration) vnext.Option {
	return func(cfg *vnext.Config) {
		if duration > 0 {
			cfg.Timeout = duration
		}
	}
}

func sanitizeForError(output string) string {
	trimmed := strings.TrimSpace(output)
	if len(trimmed) > 512 {
		return trimmed[:512] + "..."
	}
	return trimmed
}
