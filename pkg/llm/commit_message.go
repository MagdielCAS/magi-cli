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
	commitSystemPrompt = "You are an expert software assistant that writes concise, conventional commit messages."
	commitUserPrompt   = `Analyze the provided git diff and craft a single-line commit message.

Rules:
1. Message format must be <type>(<scope>): <gitmoji> <description>
2. Allowed types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
3. Select a short, meaningful scope (e.g., cli, api, docs)
4. Pick an appropriate gitmoji (✨, 🐛, 📚, 🎨, ♻️, ⚡️, ✅, 🔧, 👷, 🔨, ⏪️). Use unicode emoji, not shortcodes.
5. The description must summarize why the change exists instead of restating file names.
6. Only answer with the commit line. No explanations or code blocks.

Git diff to analyze:
{{.Diff}}`
)

var commitPromptTemplate = template.Must(template.New("commit_prompt").Parse(commitUserPrompt))

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
		Temperature: 0.2,
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
