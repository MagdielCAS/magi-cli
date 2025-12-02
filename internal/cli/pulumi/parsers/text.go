package parsers

import (
	"fmt"
	"strings"
)

// TextParser handles processing of natural language input
type TextParser struct{}

// NewTextParser creates a new Text parser
func NewTextParser() *TextParser {
	return &TextParser{}
}

// Process cleans and validates natural language input
func (p *TextParser) Process(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("input text cannot be empty")
	}

	// Remove common filler words that might confuse the analyzer if they appear in isolation
	// This is a basic cleanup; the LLM is robust enough to handle most natural language
	if len(trimmed) < 10 {
		return "", fmt.Errorf("input text is too short to be meaningful (minimum 10 characters)")
	}

	return trimmed, nil
}
