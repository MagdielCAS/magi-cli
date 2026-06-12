/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
 **/
package llm

import (
	"strings"
)

// sanitizeCLIOutput filters out verbose headers, deprecation warnings, and extracts the core response.
func sanitizeCLIOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var cleaned []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Filter out typical terminal deprecation headers/notices
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "deprecated") ||
			strings.Contains(lower, "deprecation") ||
			strings.Contains(lower, "no commands will be executed") ||
			strings.Contains(lower, "visit:") ||
			strings.Contains(lower, "changelog") ||
			strings.Contains(lower, "copilot-cli") ||
			strings.Contains(lower, "copilot cli") ||
			strings.Contains(lower, "gh-copilot") {
			continue
		}

		cleaned = append(cleaned, trimmed)
	}

	if len(cleaned) == 0 {
		return strings.TrimSpace(raw)
	}

	// For simple single-line suggestions, return the first meaningful line.
	// For structured responses, join them.
	// We'll join all cleaned lines so we don't accidentally lose multi-line results (e.g. if the CLI outputs paragraphs).
	return strings.Join(cleaned, "\n")
}
