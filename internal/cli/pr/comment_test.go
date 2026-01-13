package pr

import (
	"strings"
	"testing"
)

func TestFormatFindingsComment(t *testing.T) {
	findings := AgentFindings{
		Summary:               "All checks passed.",
		CodeSmells:            []string{"Unused var in cmd/foo.go:42"},
		SecurityConcerns:      []string{},
		AgentsGuidelineAlerts: []string{"Docs not updated"},
		TestRecommendations:   []string{"Add regression test for pr command"},
		DocumentationUpdates:  []string{},
		RiskCallouts:          []string{"High risk if gh CLI missing"},
	}

	artifacts := ReviewArtifacts{
		Analysis: findings,
		I18nFindings: &I18nResult{
			Translations: []TranslationItem{
				{Key: "hello.world", ValueEn: "Hello World", ValueDe: "Hallo Welt"},
			},
		},
	}

	comment := FormatFindingsComment(artifacts)
	required := []string{
		"Agent Review Summary",
		findings.Summary,
		findings.CodeSmells[0],
		findings.AgentsGuidelineAlerts[0],
		findings.TestRecommendations[0],
		findings.RiskCallouts[0],
		"I18n Recommendations",
		"| `hello.world` | Hello World | Hallo Welt |",
	}

	for _, needle := range required {
		if !strings.Contains(comment, needle) {
			t.Fatalf("expected comment to contain %q\ncomment: %s", needle, comment)
		}
	}
}
