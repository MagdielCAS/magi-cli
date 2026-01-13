package pr

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/agent"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
	"github.com/MagdielCAS/magi-cli/pkg/utils"
)

// AgentFindings captures the structured response from the analysis agent.
// AgentFindings captures the structured response from the analysis agent.
type AgentFindings struct {
	Summary               string   `json:"summary"`
	CodeSmells            []string `json:"code_smells"`
	SecurityConcerns      []string `json:"security_concerns"`
	AgentsGuidelineAlerts []string `json:"agents_guideline_alerts"`
	TestRecommendations   []string `json:"test_recommendations"`
	DocumentationUpdates  []string `json:"documentation_updates"`
	RiskCallouts          []string `json:"risk_callouts"`
	NeedsI18n             bool     `json:"needs_i18n"`
	I18nReason            string   `json:"i18n_reason"`
}

// PullRequestPlan stores the generated title and filled template.
type PullRequestPlan struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type TranslationItem struct {
	Key     string `json:"key"`
	ValueEn string `json:"value_en"`
	ValueDe string `json:"value_de"`
}

type I18nResult struct {
	Translations []TranslationItem `json:"translations"`
}

// ReviewArtifacts groups the analysis findings with the final PR plan.
type ReviewArtifacts struct {
	Analysis     AgentFindings
	Plan         PullRequestPlan
	I18nFindings *I18nResult
}

// AgenticReviewer orchestrates the agent workflow for PR prep.
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

	// Render payload for AnalysisAgent
	payload, err := renderAnalysisPayload(input)
	if err != nil {
		return nil, err
	}

	// Initialize AgentManager
	am := agent.NewAgentPool()
	am.WithAgent(NewAnalysisAgent(r.runtime))
	am.WithAgent(NewWriterAgent(r.runtime))
	am.WithAgent(NewI18nAgent(r.runtime))

	// Prepare initial input
	initialInput := map[string]string{
		"payload":  payload,
		"template": input.Template,
		"branch":   input.Branch,
	}

	// Execute agents
	results, err := am.ExecuteAgents(initialInput)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// Parse results
	var artifacts ReviewArtifacts

	analysisOutput := sanitizeLLMJSON(results["AnalysisAgent"])
	if err := json.Unmarshal([]byte(analysisOutput), &artifacts.Analysis); err != nil {
		return nil, fmt.Errorf("analysis agent produced invalid JSON: %w (raw: %s)", err, sanitizeForError(results["AnalysisAgent"]))
	}

	writerOutput := sanitizeLLMJSON(results["WriterAgent"])
	if err := json.Unmarshal([]byte(writerOutput), &artifacts.Plan); err != nil {
		return nil, fmt.Errorf("PR writer agent produced invalid JSON: %w (raw: %s)", err, sanitizeForError(results["WriterAgent"]))
	}

	if artifacts.Analysis.NeedsI18n {
		i18nOutput := sanitizeLLMJSON(results["I18nAgent"])
		if i18nOutput != "" {
			var i18nRes I18nResult
			if err := json.Unmarshal([]byte(i18nOutput), &i18nRes); err != nil {
				// We don't fail the whole PR creation if i18n fails, just log it?
				// Or maybe we should warn. For now let's treat it similarly being strict.
				return nil, fmt.Errorf("i18n agent produced invalid JSON: %w (raw: %s)", err, sanitizeForError(results["I18nAgent"]))
			}
			artifacts.I18nFindings = &i18nRes
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

func sanitizeForError(output string) string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return ""
	}

	// If it looks like JSON but failed parsing, we want to see the structure
	// to debug why it failed (e.g. extra text, markdown blocks).
	// We still truncate to avoid flooding logs, but keep enough context.
	const maxSnippetLen = 4096
	if len(trimmed) > maxSnippetLen {
		return trimmed[:maxSnippetLen] + "... (truncated)"
	}
	return trimmed
}

func sanitizeLLMJSON(raw string) string {
	if raw == "" {
		return ""
	}
	content := utils.RemoveCodeBlock(raw)
	return strings.ReplaceAll(content, "`", "")
}
