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
				// Wait for dependency to finish
				if doneCh, exists := doneChannels[dep]; exists {
					<-doneCh
				} else {
					// Dependency not found in manager, ignore or error?
					// For now, we assume it might be in initialInput or just missing.
					// If it was an agent, we would have found the channel.
				}

				// Read result safely
				resultsMu.RLock()
				if res, ok := results[dep]; ok {
					dependencyInputs[dep] = res
				}
				resultsMu.RUnlock()
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
