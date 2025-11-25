# Agent Package

The `agent` package provides a simple orchestration framework for executing AI agents with dependencies.

## Usage

```go
pool := agent.NewAgentPool()
pool.WithAgent(myAgent)
results, err := pool.ExecuteAgents(initialInput)
```

## Features

- **Dependency Management**: Agents can declare dependencies on other agents.
- **Parallel Execution**: Independent agents run in parallel.
- **Error Handling**: Propagates errors from agents and handles missing dependencies.
