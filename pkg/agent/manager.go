// Package agent provides a simple agent orchestration framework.
//
// It allows defining agents with dependencies and executing them in parallel
// while respecting the dependency graph.
//
// Example Usage:
//
//	pool := agent.NewAgentPool()
//	pool.WithAgent(myAgent)
//	results, err := pool.ExecuteAgents(initialInput)
package agent

import (
	"fmt"
	"sync"
)

// AgentInstance interface for extensibility
type AgentInstance interface {
	Name() string
	WaitForResults() []string
	Execute(input map[string]string) (string, error)
}

// AgentPool to handle agent execution
type AgentPool struct {
	agents map[string]AgentInstance
}

// NewAgentPool initializes a new AgentPool
func NewAgentPool() *AgentPool {
	return &AgentPool{
		agents: make(map[string]AgentInstance),
	}
}

// WithAgent adds an agent to the manager
func (am *AgentPool) WithAgent(agent AgentInstance) {
	am.agents[agent.Name()] = agent
}

// ExecuteAgents runs all agents, respecting dependencies
func (am *AgentPool) ExecuteAgents(initialInput map[string]string) (map[string]string, error) {
	errors := make(chan error, len(am.agents))
	results := make(map[string]string)
	resultsMu := sync.RWMutex{}
	var wg sync.WaitGroup

	// doneChannels signals when an agent has finished and its result is available
	doneChannels := make(map[string]chan struct{})
	for name := range am.agents {
		doneChannels[name] = make(chan struct{})
	}

	// Launch agents
	for name, agent := range am.agents {
		wg.Add(1)
		go func(name string, agent AgentInstance) {
			defer wg.Done()

			// Wait for dependencies
			dependencyInputs := make(map[string]string)
			// Copy initial inputs
			for k, v := range initialInput {
				dependencyInputs[k] = v
			}

			for _, dep := range agent.WaitForResults() {
				// Check if dependency is an agent
				if doneCh, exists := doneChannels[dep]; exists {
					// Wait for agent to finish
					<-doneCh

					// Check if agent produced a result (it might have failed)
					resultsMu.RLock()
					res, ok := results[dep]
					resultsMu.RUnlock()

					if !ok {
						// Agent failed or didn't produce output
						errors <- fmt.Errorf("dependency %q failed or produced no output for agent %q", dep, name)
						// Close channel to prevent deadlocks in dependents (though we are returning)
						close(doneChannels[name])
						return
					}
					dependencyInputs[dep] = res
				} else {
					// Not an agent, must be in initialInput
					if _, ok := initialInput[dep]; !ok {
						errors <- fmt.Errorf("dependency %q not found (not an agent and not in initial input) for agent %q", dep, name)
						close(doneChannels[name])
						return
					}
					// It's already in dependencyInputs because we copied initialInput
				}
			}

			// Execute agent actions
			result, err := agent.Execute(dependencyInputs)
			if err != nil {
				errors <- fmt.Errorf("error in agent %s: %v", name, err)
				// We still close the channel so dependents don't hang, but they might get partial data
				close(doneChannels[name])
				return
			}

			// Store result
			resultsMu.Lock()
			results[name] = result
			resultsMu.Unlock()

			// Signal completion
			close(doneChannels[name])
		}(name, agent)
	}

	// Wait for all agents to complete
	wg.Wait()
	close(errors)

	// Check for errors
	if len(errors) > 0 {
		return nil, <-errors
	}

	return results, nil
}
