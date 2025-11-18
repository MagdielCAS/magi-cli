package pr

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LoadPullRequestTemplate returns the contents of the GitHub pull request template.
func LoadPullRequestTemplate(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read pull request template at %s: %w", path, err)
	}
	return string(data), nil
}

// CollectAgentGuidelines aggregates every AGENTS.md file discovered under root.
func CollectAgentGuidelines(root string) (string, error) {
	var sections []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Base(path), "AGENTS.md") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("failed to read %s: %w", path, readErr)
		}
		relPath, relErr := filepath.Rel(root, path)
		if relErr != nil {
			relPath = path
		}
		sections = append(sections, fmt.Sprintf("## %s\n%s", relPath, string(data)))
		return nil
	})
	if err != nil {
		return "", err
	}

	if len(sections) == 0 {
		return "No AGENTS.md files were detected in this repository.", nil
	}

	return strings.Join(sections, "\n\n---\n\n"), nil
}
