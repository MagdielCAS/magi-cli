package agent

import (
	"errors"
	"fmt"
	"testing"
)

type mockAgent struct {
	name         string
	dependencies []string
	executeFunc  func(input map[string]string) (string, error)
}

func (m *mockAgent) Name() string {
	return m.name
}

func (m *mockAgent) WaitForResults() []string {
	return m.dependencies
}

func (m *mockAgent) Execute(input map[string]string) (string, error) {
	if m.executeFunc != nil {
		return m.executeFunc(input)
	}
	return "done", nil
}

func TestAgentPool_ExecuteAgents(t *testing.T) {
	t.Run("successful execution with dependencies", func(t *testing.T) {
		pool := NewAgentPool()

		agent1 := &mockAgent{
			name: "agent1",
			executeFunc: func(input map[string]string) (string, error) {
				return "result1", nil
			},
		}

		agent2 := &mockAgent{
			name:         "agent2",
			dependencies: []string{"agent1"},
			executeFunc: func(input map[string]string) (string, error) {
				if input["agent1"] != "result1" {
					return "", fmt.Errorf("expected result1, got %s", input["agent1"])
				}
				return "result2", nil
			},
		}

		pool.WithAgent(agent1)
		pool.WithAgent(agent2)

		results, err := pool.ExecuteAgents(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if results["agent1"] != "result1" {
			t.Errorf("expected agent1 result 'result1', got '%s'", results["agent1"])
		}
		if results["agent2"] != "result2" {
			t.Errorf("expected agent2 result 'result2', got '%s'", results["agent2"])
		}
	})

	t.Run("missing dependency error", func(t *testing.T) {
		pool := NewAgentPool()

		agent2 := &mockAgent{
			name:         "agent2",
			dependencies: []string{"missing_agent"},
		}

		pool.WithAgent(agent2)

		_, err := pool.ExecuteAgents(nil)
		if err == nil {
			t.Fatal("expected error for missing dependency, got nil")
		}
		if err.Error() != `dependency "missing_agent" not found (not an agent and not in initial input) for agent "agent2"` {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("dependency failure propagation", func(t *testing.T) {
		pool := NewAgentPool()

		agent1 := &mockAgent{
			name: "agent1",
			executeFunc: func(input map[string]string) (string, error) {
				return "", errors.New("agent1 failed")
			},
		}

		agent2 := &mockAgent{
			name:         "agent2",
			dependencies: []string{"agent1"},
		}

		pool.WithAgent(agent1)
		pool.WithAgent(agent2)

		_, err := pool.ExecuteAgents(nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// The error could be from agent1 or agent2 (dependency failure)
		// Since they run concurrently (agent1 starts, agent2 waits), agent1 fails first.
		// agent2 wakes up and sees agent1 failed.
		// We expect an error.
	})

	t.Run("initial input dependency", func(t *testing.T) {
		pool := NewAgentPool()

		agent1 := &mockAgent{
			name:         "agent1",
			dependencies: []string{"input_key"},
			executeFunc: func(input map[string]string) (string, error) {
				if input["input_key"] != "input_val" {
					return "", fmt.Errorf("expected input_val, got %s", input["input_key"])
				}
				return "success", nil
			},
		}

		pool.WithAgent(agent1)

		results, err := pool.ExecuteAgents(map[string]string{"input_key": "input_val"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if results["agent1"] != "success" {
			t.Errorf("expected success, got %s", results["agent1"])
		}
	})
}
