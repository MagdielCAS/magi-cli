package project

import (
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/pterm/pterm"
)

// showDiff generates a unified diff between the original and updated content
// and prints it to the terminal using pterm for syntax highlighting.
func showDiff(original, updated string) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(original),
		B:        difflib.SplitLines(updated),
		FromFile: "Original",
		ToFile:   "Updated",
		Context:  3,
	}

	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil || text == "" {
		pterm.Info.Println("No changes detected.")
		return
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		// go-difflib includes newlines in the strings from SplitLines
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			pterm.FgGreen.Println(line)
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			pterm.FgRed.Println(line)
		case strings.HasPrefix(line, "@@"):
			pterm.FgCyan.Println(line)
		default:
			pterm.Println(line)
		}
	}
}
