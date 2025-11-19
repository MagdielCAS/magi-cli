// Package llm centralizes AI service construction for magi commands.
// The service builder enforces shared defaults (timeouts, provider selection,
// base URL overrides, and hardened HTTP clients) that are consumed by
// cmd/commit.go, cmd/pr.go, and future commands that need to talk to LLMs.
package llm
