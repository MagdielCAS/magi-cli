package pr

import (
	"fmt"
	"strings"
)

// FormatFindingsComment produces a markdown comment from analysis findings.
func FormatFindingsComment(artifacts ReviewArtifacts) string {
	var b strings.Builder
	findings := artifacts.Analysis

	b.WriteString("## 🤖 Agent Review Summary\n\n")
	if strings.TrimSpace(findings.Summary) != "" {
		b.WriteString(findings.Summary)
		b.WriteString("\n\n")
	} else {
		b.WriteString("No high-level summary was provided.\n\n")
	}

	writeSection(&b, "Code Smells", findings.CodeSmells)
	writeSection(&b, "Security Concerns", findings.SecurityConcerns)
	writeSection(&b, "Policy Alerts (AGENTS.md)", findings.AgentsGuidelineAlerts)
	writeSection(&b, "Suggested Tests", findings.TestRecommendations)
	writeSection(&b, "Documentation Updates", findings.DocumentationUpdates)
	writeSection(&b, "Risk Callouts", findings.RiskCallouts)

	if artifacts.I18nFindings != nil && len(artifacts.I18nFindings.Translations) > 0 {
		b.WriteString("### I18n Recommendations\n")
		b.WriteString("| Key | English | German |\n|---|---|---|\n")
		for _, item := range artifacts.I18nFindings.Translations {
			b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", item.Key, item.ValueEn, item.ValueDe))
		}
		b.WriteString("\n")
	}

	return strings.TrimSpace(b.String())
}

func writeSection(b *strings.Builder, title string, items []string) {
	b.WriteString("### ")
	b.WriteString(title)
	b.WriteString("\n")

	if len(items) == 0 {
		b.WriteString("- None.\n\n")
		return
	}

	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(trimmed)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}
