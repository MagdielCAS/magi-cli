/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	openai "github.com/openai/openai-go/v3"
	openaiShared "github.com/openai/openai-go/v3/shared"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

var (
	CommitSchema = &openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openaiShared.ResponseFormatJSONSchemaParam{
			JSONSchema: openaiShared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "commit_message",
				Description: openai.String("A conventional commit message"),
				Schema: interface{}(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":        map[string]interface{}{"type": "string", "enum": []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"}},
						"scope":       map[string]interface{}{"type": "string"},
						"gitmoji":     map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
					},
					"required":             []string{"type", "scope", "gitmoji", "description"},
					"additionalProperties": false,
				}),
				Strict: openai.Bool(true),
			},
		},
	}
)

const (
	commitSystemPrompt = "You are an expert assistant that writes conventional commit messages."
	commitUserPrompt   = `Analyze the provided git diff and generate a structured commit message.

Rules:
1. Type must be one of: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
2. Scope must be a short, meaningful noun (e.g., cli, api, docs)
3. Description must be a short summary of the change in present tense (e.g., add, fix, update). Do not capitalize. Do not end with a period.
4. Gitmoji must be one appropriate unicode emoji from: ✨, 🐛, 📚, 🎨, ♻️, ⚡️, ✅, 🔧, 👷, 🔨, ⏪️.

Git diff to analyze:
` + "```diff\n{{.Diff}}\n```"
	fixCommitUserPrompt = `You previously proposed a commit message that failed validation.

Review the original diff and craft a corrected commit message that obeys all rules.

Rules:
1. Type must be one of: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
2. Scope must be a short, meaningful noun (e.g., cli, api, docs)
3. Description must be a short summary of the change in present tense (e.g., add, fix, update). Do not capitalize. Do not end with a period.
4. Gitmoji must be one appropriate unicode emoji from: ✨, 🐛, 📚, 🎨, ♻️, ⚡️, ✅, 🔧, 👷, 🔨, ⏪️.

Context:
` + "```diff\n{{.Diff}}\n```" + `

Previous invalid commit message:
{{.Previous}}

Validation feedback:
{{.ValidationError}}

Respond with the corrected commit message structure.`
)

var (
	commitPromptTemplate    = template.Must(template.New("commit_prompt").Parse(commitUserPrompt))
	fixCommitPromptTemplate = template.Must(template.New("fix_commit_prompt").Parse(fixCommitUserPrompt))
)

// GenerateCommitMessage requests an AI-generated conventional commit message for the supplied diff.
func GenerateCommitMessage(ctx context.Context, runtime *shared.RuntimeContext, diff string) (string, error) {
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("diff cannot be empty")
	}
	if runtime == nil {
		return "", fmt.Errorf("runtime context is required")
	}

	cap, err := GetCapability("generate_commit_message")
	if err != nil {
		return "", err
	}

	provider, err := ResolveProvider(runtime)
	if err != nil {
		return "", err
	}

	prompt, err := cap.BuildPrompt(CapabilityContext{
		Diff: diff,
	})
	if err != nil {
		return "", err
	}

	execCtx := map[string]string{
		"system_prompt": cap.SystemPrompt(),
		"model_variant": "light",
	}

	resp, err := provider.Execute(ctx, &ExecutionRequest{
		Capability: cap.Name(),
		Prompt:     prompt,
		Context:    execCtx,
		Options: ExecutionOptions{
			Temperature: 0.0,
			MaxTokens:   500,
		},
	})
	if err != nil {
		return "", err
	}

	return resp.Output, nil
}

func renderCommitPrompt(diff string) (string, error) {
	var buf bytes.Buffer
	if err := commitPromptTemplate.Execute(&buf, struct {
		Diff string
	}{
		Diff: diff,
	}); err != nil {
		return "", fmt.Errorf("failed to render commit prompt: %w", err)
	}
	return buf.String(), nil
}

// FixCommitMessage reparses the diff with guidance about the validation failure and returns a corrected message.
func FixCommitMessage(ctx context.Context, runtime *shared.RuntimeContext, diff, previousMessage string, validationErr error) (string, error) {
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("diff cannot be empty")
	}
	if runtime == nil {
		return "", fmt.Errorf("runtime context is required")
	}

	cap, err := GetCapability("fix_commit_message")
	if err != nil {
		return "", err
	}

	provider, err := ResolveProvider(runtime)
	if err != nil {
		return "", err
	}

	prompt, err := cap.BuildPrompt(CapabilityContext{
		Diff:            diff,
		PreviousMessage: previousMessage,
		ValidationError: formatValidationError(validationErr),
	})
	if err != nil {
		return "", err
	}

	execCtx := map[string]string{
		"system_prompt": cap.SystemPrompt(),
		"model_variant": "light",
	}

	resp, err := provider.Execute(ctx, &ExecutionRequest{
		Capability: cap.Name(),
		Prompt:     prompt,
		Context:    execCtx,
		Options: ExecutionOptions{
			Temperature: 0.0,
			MaxTokens:   500,
		},
	})
	if err != nil {
		return "", err
	}

	return resp.Output, nil
}

func renderFixCommitPrompt(diff, previous string, validationErr error) (string, error) {
	if strings.TrimSpace(previous) == "" {
		previous = "N/A"
	}

	var buf bytes.Buffer
	if err := fixCommitPromptTemplate.Execute(&buf, struct {
		Diff            string
		Previous        string
		ValidationError string
	}{
		Diff:            diff,
		Previous:        previous,
		ValidationError: formatValidationError(validationErr),
	}); err != nil {
		return "", fmt.Errorf("failed to render commit fix prompt: %w", err)
	}
	return buf.String(), nil
}

func formatValidationError(err error) string {
	if err == nil {
		return "Unknown validation error."
	}
	return err.Error()
}

func parseCommitMessage(jsonStr string) (string, error) {
	var result struct {
		Type        string `json:"type"`
		Scope       string `json:"scope"`
		Gitmoji     string `json:"gitmoji"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return "", fmt.Errorf("failed to parse commit message JSON: %w", err)
	}

	// Format: <type>(<scope>): <gitmoji> <description>
	return fmt.Sprintf("%s(%s): %s %s", result.Type, result.Scope, result.Gitmoji, result.Description), nil
}
