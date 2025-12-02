package parsers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTextParser_Process(t *testing.T) {
	parser := NewTextParser()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid input",
			input:   "Create a simple VPC with 2 subnets",
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Short input",
			input:   "Too short",
			wantErr: true,
		},
		{
			name:    "Input with whitespace",
			input:   "   Valid input with whitespace   ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Process(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TextParser.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Errorf("TextParser.Process() returned empty string for valid input")
			}
		})
	}
}

func TestMermaidParser_Validate(t *testing.T) {
	parser := NewMermaidParser()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "Valid graph",
			content: "graph TD\n    A-->B",
			wantErr: false,
		},
		{
			name:    "Valid sequence diagram",
			content: "sequenceDiagram\n    Alice->>John: Hello",
			wantErr: false,
		},
		{
			name:    "Markdown wrapped",
			content: "```mermaid\ngraph TD\n    A-->B\n```",
			wantErr: false,
		},
		{
			name:    "Valid with indicators but no type header",
			content: "A --> B",
			wantErr: false,
		},
		{
			name:    "Empty content",
			content: "",
			wantErr: true,
		},
		{
			name:    "Invalid content",
			content: "This is just some random text",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.Validate(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("MermaidParser.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMermaidParser_ParseFile(t *testing.T) {
	parser := NewMermaidParser()

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mmd")
	content := "graph TD\n    A-->B"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test parsing
	got, err := parser.ParseFile(tmpFile)
	if err != nil {
		t.Errorf("MermaidParser.ParseFile() error = %v", err)
	}
	if got != content {
		t.Errorf("MermaidParser.ParseFile() = %v, want %v", got, content)
	}

	// Test non-existent file
	_, err = parser.ParseFile("non_existent_file.mmd")
	if err == nil {
		t.Errorf("MermaidParser.ParseFile() expected error for non-existent file")
	}
}
