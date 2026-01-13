# AI Architecture in magi-cli

This document outlines the architecture and implementation patterns for Artificial Intelligence (AI) and Large Language Model (LLM) integration within the `magi-cli` project.

## Overview

`magi-cli` treats AI as a core capability, enabling features like intelligent project analysis, automated code generation, and PR description writing. The architecture is designed to be:

*   **Provider-Agnostic:** While currently optimized for OpenAI-compatible APIs, the core abstractions allow swapping providers.
*   **Resilient:** Implements fallback mechanisms to ensure continuity even if the primary model fails.
*   **Efficient:** Uses different model "variants" (Heavy, Light, Fallback) to balance cost, latency, and intelligence.
*   **Agentic:** Supports both direct LLM calls and complex, multi-agent orchestration.

## Core Components (`pkg/llm`)

The `pkg/llm` package provides the foundational building blocks for all AI interactions.

### 1. Service Builder & Model Variants

To prevent hardcoding model names and ensure consistent configuration, we use a `ServiceBuilder`. It allows commands to request a "class" of model rather than a specific one.

*   **Heavy (`ModelVariantHeavy`):** used for complex reasoning, code generation, and architectural analysis. (e.g., GPT-4 class).
*   **Light (`ModelVariantLight`):** used for simple tasks, planning, summarization, or high-volume operations. (e.g., GPT-3.5/4o-mini class).
*   **Fallback (`ModelVariantFallback`):** a safety net model used if the primary models fail or for basic redundancy.

**Usage:**

```go
// Create a builder with the runtime context
builder := llm.NewServiceBuilder(runtimeContext)

// Select a variant
builder.UseHeavyModel() 
// or 
builder.UseLightModel()

// Build the service
service, err := builder.Build()
```

### 2. LLM Service

The `Service` struct acts as the client for the AI provider. It handles authentication, base URL configuration, and HTTP client usage (including redaction and timeout policies defined in `shared.RuntimeContext`).

The primary method is `ChatCompletion`, which accepts a structured request and returns the assistant's text response.

## Agent Framework (`pkg/agent`)

For complex workflows requiring multiple steps or parallel execution, we use the `pkg/agent` orchestration framework.

### 1. Agent Instance

Any struct can be an agent by implementing the `AgentInstance` interface:

```go
type AgentInstance interface {
    Name() string
    WaitForResults() []string // Returns names of agents this agent depends on
    Execute(input map[string]string) (string, error)
}
```

### 2. Agent Pool

The `AgentPool` manages the lifecycle and execution of agents. It constructs a dependency graph based on `WaitForResults()` and executes independent agents in parallel.

**Data Flow:**
*   **Input:** Initial map of data.
*   **Inter-Agent:** Outputs from dependencies are injected into the input map of dependent agents.
*   **Result:** A map containing the outputs of all executed agents.

**Example Flow (PR Reviewer):**
1.  **AnalysisAgent:** Reads code changes -> Outputs JSON Analysis.
2.  **WriterAgent:** Waits for `AnalysisAgent` -> Reads JSON Analysis -> Writes PR Description.

## Implementation Patterns

### Pattern 1: Orchestrated Agents

Use this pattern when you have a multi-step workflow where some steps can run in parallel or strictly depend on previous steps. This promotes separation of concerns.

**Example:** `internal/cli/pr/agents.go`

```go
// 1. Define Agents
type AnalysisAgent struct { ... }
type WriterAgent struct { ... }

// 2. Register in Pool
pool := agent.NewAgentPool()
pool.WithAgent(NewAnalysisAgent(runtime))
pool.WithAgent(NewWriterAgent(runtime))

// 3. Execute
results, err := pool.ExecuteAgents(initialPayload)
```

### Pattern 2: Direct Service Usage

Use this pattern for linear, synchronous interactions or when you need tight control over the prompt loop (e.g., Validation loops, interactive generation).

**Example:** `internal/cli/project/agents.go` (`ArchitectureAgent`, `ValidatorAgent`)

```go
func (a *ArchitectureAgent) Analyze(rootPath string) {
    // 1. Prepare Prompt
    prompt := buildPrompt(rootPath)

    // 2. Build Service
    service := llm.NewServiceBuilder(a.runtime).UseHeavyModel().Build()

    // 3. Call LLM
    resp, err := service.ChatCompletion(...)
    
    // 4. Parse (e.g., JSON Unmarshal)
    // ...
}
```

## Best Practices

1.  **Always use `ServiceBuilder`:** Never instantiate clients manually. This ensures global configuration (API keys, timeouts) is respected.
2.  **Effective Prompting:**
    *   Use **System Prompts** to define the role, constraints, and output format.
    *   Use **JSON Schemas** in prompts to enforce structured output, making parsing reliable.
    *   Be explicit about what *not* to include (e.g., "Return ONLY JSON, no markdown blocks").
3.  **Error Handling:**
    *   Implement **Fallback Logic**: If a `Heavy` model call fails, consider retrying with `Fallback` or `Light` depending on the task.
    *   **Validate LLM Output**: Always unmarshal and validate JSON responses. LLMs can hallucinate formats. The `ValidatorAgent` pattern (asking the LLM to fix its own output) is highly effective.
4.  **Token Management:**
    *   Be mindful of context windows.
    *   Truncate or summarize large inputs (like file trees or generic file contents) before sending to the LLM.
5.  **Security:**
    *   Follow the security rules in `AGENTS.md`.
    *   Do not send sensitive secrets (API keys, passwords) in prompts.

## How to Add AI to a New Command

1.  **Define the Goal:** What problem is the AI solving?
2.  **Choose the Pattern:**
    *   *Complex/Multi-step?* -> Implement `pkg/agent.AgentInstance`.
    *   *Linear/Simple?* -> Use `llm.Service` directly in your command logic.
3.  **Draft Prompts:** Create `prompts.go` in your package to keep prompt text separate from logic.
4.  **Implement Logic:**
    *   Construct the context/input.
    *   Call the `service.ChatCompletion`.
    *   Parse and use the result.
5.  **User Experience:** Use `pterm` to show spinners or progress indicators while the AI is "thinking".
