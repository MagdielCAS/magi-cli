/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package llm

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/tiktoken-go/tokenizer"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

const (
	commitSystemPrompt = "You are an expert assistant that writes conventional commit messages."
	commitUserPrompt   = `Analyze the provided git diff and generate a single-line commit message.

Rules:
1. Message format must be <type>(<scope>): <gitmoji> <description>
2. Allowed types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
3. Scope must be a short, meaningful noun (e.g., cli, api, docs)
4. Description must be a short summary of the change in present tense (e.g., add, fix, update). Do not capitalize. Do not end with a period.
5. Pick one appropriate gitmoji from this list: ✨, 🐛, 📚, 🎨, ♻️, ⚡️, ✅, 🔧, 👷, 🔨, ⏪️. Use the unicode emoji, not shortcodes.
6. The entire message must be a single line.
7. Your response must only contain the commit message. Do not include any other text, explanations, or code blocks.

Git diff to analyze:
` + "```diff\n{{.Diff}}\n```"
	fixCommitUserPrompt = `You previously proposed a commit message that failed validation.

Review the original diff and craft a corrected single-line commit message that obeys all rules.

Rules:
1. Message format must be <type>(<scope>): <gitmoji> <description>
2. Allowed types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
3. Scope must be a short, meaningful noun (e.g., cli, api, docs)
4. Description must be a short summary of the change in present tense (e.g., add, fix, update). Do not capitalize. Do not end with a period.
5. Pick one appropriate gitmoji from this list: ✨, 🐛, 📚, 🎨, ♻️, ⚡️, ✅, 🔧, 👷, 🔨, ⏪️. Use the unicode emoji, not shortcodes.
6. The entire message must be a single line and must not include explanations or code blocks.

Context:
` + "```diff\n{{.Diff}}\n```" + `

Previous invalid commit message:
{{.Previous}}

Validation feedback:
{{.ValidationError}}

Respond only with the corrected commit message.`
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
	if runtime.LightModel == "" {
		return "", fmt.Errorf("api.heavy_model must be configured")
	}

	prompt, err := renderCommitPrompt(diff)
	if err != nil {
		return "", err
	}

	service, err := NewServiceBuilder(runtime).UseLightModel().Build()
	if err != nil {
		return "", err
	}

	maxTokens := 100.0
	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		maxTokens = 2048
	}
	if enc != nil {
		count, err := enc.Count(fmt.Sprintf("%s%s", commitSystemPrompt, prompt))
		if err != nil {
			count = 2048
		}
		// commit msg length + an estimative of prompt tokens + 10% error margin
		maxTokens = 100 + float64(count)*1.1
	}

	message, err := service.ChatCompletion(ctx, ChatCompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: commitSystemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.0,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return "", err
	}

	return message, nil
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
	if runtime.LightModel == "" {
		return "", fmt.Errorf("api.heavy_model must be configured")
	}

	prompt, err := renderFixCommitPrompt(diff, previousMessage, validationErr)
	if err != nil {
		return "", err
	}

	service, err := NewServiceBuilder(runtime).UseLightModel().Build()
	if err != nil {
		return "", err
	}

	message, err := service.ChatCompletion(ctx, ChatCompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: commitSystemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.0,
		MaxTokens:   200,
	})
	if err != nil {
		return "", err
	}

	return message, nil
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
