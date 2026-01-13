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

func TestI18nAgent_Execute(t *testing.T) {
	agent := NewI18nAgent(&shared.RuntimeContext{})

	// Case 1: Missing AnalysisAgent output
	_, err := agent.Execute(map[string]string{})
	if err == nil || err.Error() != "analysis result is missing" {
		t.Errorf("expected 'analysis result is missing' error, got %v", err)
	}

	// Case 2: NeedsI18n is false -> should skip (return empty string)
	analysisNoI18n := `{"needs_i18n": false}`
	res, err := agent.Execute(map[string]string{
		"AnalysisAgent": analysisNoI18n,
	})
	if err != nil {
		t.Errorf("expected no error when i18n is not needed, got %v", err)
	}
	if res != "" {
		t.Errorf("expected empty result (skipped), got %q", res)
	}

	// Case 3: NeedsI18n is true, but payload (diff) is missing -> should error
	analysisNeedsI18n := `{"needs_i18n": true}`
	_, err = agent.Execute(map[string]string{
		"AnalysisAgent": analysisNeedsI18n,
	})
	if err == nil || err.Error() != "payload (diff) is missing" {
		t.Errorf("expected 'payload (diff) is missing' error, got %v", err)
	}

	// Case 4: Invalid JSON in AnalysisAgent output -> should verify silent skip/safety
	// The implementation swallows JSON errors to prevent crashing the flow.
	_, err = agent.Execute(map[string]string{
		"AnalysisAgent": "{invalid-json",
	})
	if err != nil {
		t.Errorf("expected no error (silent skip) for invalid JSON, got %v", err)
	}
}
