package pr

import (
	"strings"
)

// FormatFindingsComment produces a markdown comment from analysis findings.
func FormatFindingsComment(findings AgentFindings) string {
	var b strings.Builder

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
