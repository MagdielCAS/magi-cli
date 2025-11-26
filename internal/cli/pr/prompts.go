package pr

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	openai "github.com/openai/openai-go/v3"
	openaiShared "github.com/openai/openai-go/v3/shared"
)

var (
	AnalysisSchema = &openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openaiShared.ResponseFormatJSONSchemaParam{
			JSONSchema: openaiShared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "analysis_result",
				Description: openai.String("The analysis result of the PR"),
				Schema: interface{}(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"summary":                 map[string]interface{}{"type": "string"},
						"code_smells":             map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"security_concerns":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"agents_guideline_alerts": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"test_recommendations":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"documentation_updates":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"risk_callouts":           map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					},
					"required": []string{
						"summary",
						"code_smells",
						"security_concerns",
						"agents_guideline_alerts",
						"test_recommendations",
						"documentation_updates",
						"risk_callouts",
					},
					"additionalProperties": false,
				}),
				Strict: openai.Bool(true),
			},
		},
	}

	WriterSchema = &openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openaiShared.ResponseFormatJSONSchemaParam{
			JSONSchema: openaiShared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "pr_content",
				Description: openai.String("The PR title and body"),
				Schema: interface{}(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{"type": "string"},
						"body":  map[string]interface{}{"type": "string"},
					},
					"required":             []string{"title", "body"},
					"additionalProperties": false,
				}),
				Strict: openai.Bool(true),
			},
		},
	}
)

const (
	analysisSystemPrompt = `You are "magi-secure-reviewer", a senior engineer tasked with reviewing pull requests for this CLI.
Goals:
1. Identify concrete code smells or correctness risks grounded in the diff.
2. Highlight violations of AGENTS.md policies and security red flags.
3. Recommend regression tests or documentation updates when gaps exist.

Provide analysis in JSON format:
{
  "summary": "<one concise paragraph>",
  "code_smells": ["<issue>: <file>:<line> - <detail>"],
  "security_concerns": ["..."],
  "agents_guideline_alerts": ["..."],
  "test_recommendations": ["..."],
  "documentation_updates": ["..."],
  "risk_callouts": ["..."]
}

Rules:
- Keep responses grounded in the provided diff and guidelines.
- Reference file paths when possible.
- Use empty arrays when a section has no findings.
- Do not emit markdown, prose paragraphs, or additional commentary outside the JSON.
- Make sure the JSON is valid.
- Never add backticks or extra formatting outside of the JSON structure.`

	writerSystemPrompt = `You are "magi-pr-writer", an AI that prepares GitHub pull request descriptions.
Use the JSON analysis from the previous step plus the official pull request template.
Fill every template section with concise, factual content. Mention security, testing, and documentation impacts explicitly.

Provide response in JSON format:
{
  "title": "<short PR title>",
  "body": "<markdown body matching the template verbatim>"
}

Rules:
- Preserve the template headings and checklist syntax.
- Keep the title under 80 characters and avoid trailing punctuation.
- Mention which data, if any, leaves the machine and which safeguards are in place.
- Make sure the JSON is valid.
- Never add backticks or extra formatting outside of the template structure.`
)

var (
	analysisInputTemplate = template.Must(template.New("analysis_input").Parse(
		`Repository branch: {{.Branch}}
Upstream reference: {{.RemoteRef}}

Additional reviewer notes (may be empty):
{{.AdditionalContext}}

Security and workflow requirements (AGENTS.md):
{{.Guidelines}}

Unified git diff between {{.RemoteRef}} and HEAD:
{{.Diff}}

Respond strictly with the JSON schema described in your system prompt.`))

	writerInputTemplate = template.Must(template.New("writer_input").Parse(
		`The secure analysis agent produced this JSON:
{{.AnalysisJSON}}

Use it to fill the pull request template exactly as written:
{{.Template}}

Branch under review: {{.Branch}}

Respond with the JSON schema described in your system prompt.`))
)

// ReviewInput encapsulates the information sent to the analysis workflow.
type ReviewInput struct {
	Diff              string
	Branch            string
	RemoteRef         string
	Guidelines        string
	AdditionalContext string
	Template          string
}

func (ri ReviewInput) validate() error {
	if strings.TrimSpace(ri.Diff) == "" {
		return fmt.Errorf("diff content is required")
	}
	if strings.TrimSpace(ri.Branch) == "" {
		return fmt.Errorf("branch name is required")
	}
	if strings.TrimSpace(ri.RemoteRef) == "" {
		return fmt.Errorf("remote reference is required")
	}
	if strings.TrimSpace(ri.Template) == "" {
		return fmt.Errorf("pull request template content is required")
	}
	return nil
}

func renderAnalysisPayload(input ReviewInput) (string, error) {
	if err := input.validate(); err != nil {
		return "", err
	}

	data := struct {
		Branch            string
		RemoteRef         string
		Guidelines        string
		AdditionalContext string
		Diff              string
	}{
		Branch:            input.Branch,
		RemoteRef:         input.RemoteRef,
		Guidelines:        fallbackText(input.Guidelines, "No AGENTS.md files were detected in this repository."),
		AdditionalContext: fallbackText(input.AdditionalContext, "N/A"),
		Diff:              input.Diff,
	}

	var buf bytes.Buffer
	if err := analysisInputTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render analysis payload: %w", err)
	}
	return buf.String(), nil
}

type writerPayloadParams struct {
	AnalysisJSON string
	Template     string
	Branch       string
}

func renderWriterPayload(params writerPayloadParams) (string, error) {
	if strings.TrimSpace(params.AnalysisJSON) == "" {
		return "", fmt.Errorf("analysis JSON cannot be empty")
	}
	if strings.TrimSpace(params.Template) == "" {
		return "", fmt.Errorf("pull request template is required")
	}

	var buf bytes.Buffer
	if err := writerInputTemplate.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("failed to render writer payload: %w", err)
	}
	return buf.String(), nil
}

func fallbackText(candidate, fallback string) string {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
