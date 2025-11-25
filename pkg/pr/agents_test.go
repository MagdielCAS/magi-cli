package pr

import (
	"testing"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

func TestAnalysisAgent_Execute_MissingPayload(t *testing.T) {
	agent := NewAnalysisAgent(&shared.RuntimeContext{})
	_, err := agent.Execute(map[string]string{})
	if err == nil || err.Error() != "payload is missing" {
		t.Errorf("expected 'payload is missing' error, got %v", err)
	}
}

func TestWriterAgent_Execute_MissingInputs(t *testing.T) {
	agent := NewWriterAgent(&shared.RuntimeContext{})

	// Missing analysis result
	_, err := agent.Execute(map[string]string{})
	if err == nil || err.Error() != "analysis result is missing" {
		t.Errorf("expected 'analysis result is missing' error, got %v", err)
	}

	// Missing template
	_, err = agent.Execute(map[string]string{"AnalysisAgent": "{}"})
	if err == nil || err.Error() != "template is missing" {
		t.Errorf("expected 'template is missing' error, got %v", err)
	}

	// Missing branch
	_, err = agent.Execute(map[string]string{"AnalysisAgent": "{}", "template": "tmpl"})
	if err == nil || err.Error() != "branch is missing" {
		t.Errorf("expected 'branch is missing' error, got %v", err)
	}
}
