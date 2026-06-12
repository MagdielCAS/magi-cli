/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
 **/
package llm

import (
	"fmt"
)

// Capability defines the interface for specific LLM tasks/intents.
type Capability interface {
	Name() string
	BuildPrompt(ctx CapabilityContext) (string, error)
	SystemPrompt() string
}

// CapabilityContext provides the parameters needed to render prompts.
type CapabilityContext struct {
	Diff            string
	PreviousMessage string
	ValidationError string
	RepoRoot        string
}

// GenerateCommitMessageCapability implements Capability for commit message generation.
type GenerateCommitMessageCapability struct{}

func (c *GenerateCommitMessageCapability) Name() string {
	return "generate_commit_message"
}

func (c *GenerateCommitMessageCapability) SystemPrompt() string {
	return commitSystemPrompt
}

func (c *GenerateCommitMessageCapability) BuildPrompt(ctx CapabilityContext) (string, error) {
	if ctx.Diff == "" {
		return "", fmt.Errorf("diff is required")
	}
	return renderCommitPrompt(ctx.Diff)
}

// FixCommitMessageCapability implements Capability for repairing a rejected commit message.
type FixCommitMessageCapability struct{}

func (c *FixCommitMessageCapability) Name() string {
	return "fix_commit_message"
}

func (c *FixCommitMessageCapability) SystemPrompt() string {
	return commitSystemPrompt
}

func (c *FixCommitMessageCapability) BuildPrompt(ctx CapabilityContext) (string, error) {
	if ctx.Diff == "" {
		return "", fmt.Errorf("diff is required")
	}
	var previous = ctx.PreviousMessage
	if previous == "" {
		previous = "N/A"
	}
	var validationErr error
	if ctx.ValidationError != "" {
		validationErr = fmt.Errorf("%s", ctx.ValidationError)
	}
	return renderFixCommitPrompt(ctx.Diff, previous, validationErr)
}
