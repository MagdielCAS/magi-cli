package parsers

import (
	"fmt"
	"os"
	"strings"
)

// MermaidParser handles parsing and validation of Mermaid diagrams
type MermaidParser struct{}

// NewMermaidParser creates a new Mermaid parser
func NewMermaidParser() *MermaidParser {
	return &MermaidParser{}
}

// ParseFile reads and validates a Mermaid file
func (p *MermaidParser) ParseFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read mermaid file: %w", err)
	}

	mermaidContent := string(content)
	if err := p.Validate(mermaidContent); err != nil {
		return "", fmt.Errorf("invalid mermaid diagram: %w", err)
	}

	return mermaidContent, nil
}

// Validate performs basic validation of Mermaid syntax
func (p *MermaidParser) Validate(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return fmt.Errorf("empty content")
	}

	// Basic check for common Mermaid diagram types
	validTypes := []string{
		"graph", "flowchart", "sequenceDiagram", "classDiagram",
		"stateDiagram", "erDiagram", "gantt", "pie", "gitGraph",
	}

	hasValidType := false
	for _, t := range validTypes {
		if strings.HasPrefix(trimmed, t) {
			hasValidType = true
			break
		}
	}

	if !hasValidType {
		// Check if it's wrapped in markdown block
		if strings.Contains(trimmed, "```mermaid") {
			return nil
		}

		// If it doesn't start with a known type and isn't a markdown block,
		// check if it contains at least some mermaid syntax indicators
		indicators := []string{"-->", "---", "subgraph", "participant", "class ", "state "}
		hasIndicator := false
		for _, ind := range indicators {
			if strings.Contains(trimmed, ind) {
				hasIndicator = true
				break
			}
		}

		if !hasIndicator {
			return fmt.Errorf("content does not appear to be a valid Mermaid diagram (missing diagram type or common syntax)")
		}
	}

	return nil
}
