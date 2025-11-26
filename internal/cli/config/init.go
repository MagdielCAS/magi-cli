package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a local configuration file",
	Long: `Initialize a local configuration file (.magi.yaml) in the current directory.
This file will override the global configuration (but only the ones with the same key).
The goal is to have custom envs like models or keys for different workspaces.

Usage:
  magi config init

Examples:
  # Initialize a local configuration file
  magi config init
`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			pterm.Error.Printf("Error getting current working directory: %v\n", err)
			return
		}

		configPath := filepath.Join(cwd, ".magi.yaml")
		if _, err := os.Stat(configPath); err == nil {
			pterm.Warning.Println("Local configuration file already exists")
			return
		}

		// Create a minimal template with the keys required configs defined at the cmd/setup.go command
		template := `# Magi CLI Configuration

api:
  # API provider (e.g., openai, custom)
  provider: openai
  # Base URL for custom OpenAI compatible API (optional)
  base_url: ""
  # Your API key
  key: ""
  # Model for light tasks (e.g., gpt-3.5-turbo)
  light_model: gpt-3.5-turbo
  # Model for heavy tasks (e.g., gpt-4)
  heavy_model: gpt-4
  # Fallback model (e.g., gpt-3.5-turbo)
  fallback_model: gpt-3.5-turbo

output:
  # Default output format (e.g., text, json, yaml)
  format: text
  # Enable color output
  color: true

cache:
  # Enable caching
  enabled: true
  # Cache TTL in seconds
  ttl: 3600
`

		if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
			pterm.Error.Printf("Error creating local configuration file: %v\n", err)
			return
		}

		pterm.Success.Printf("Local configuration file created at %s\n", configPath)

		// Add to .gitignore
		gitignorePath := filepath.Join(cwd, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			if err := addToGitignore(gitignorePath, ".magi.yaml"); err != nil {
				pterm.Error.Printf("Error updating .gitignore: %v\n", err)
			} else {
				pterm.Success.Println("Added .magi.yaml to .gitignore")
			}
		}
	},
}

func addToGitignore(path, content string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == content {
			return nil // Already exists
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Re-open for appending
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// We can't easily check if the last line has a newline without reading the whole file or seeking.
	// For simplicity, let's just append a newline before the content if we are appending.
	// Or better, just write "\ncontent\n" but that might create extra empty lines.
	// Let's just write "content\n". If the file didn't end in newline, it might be on same line.
	// A safer bet is reading the file fully or seeking to end - 1 to check char.

	// Let's use the ReadFile approach for simplicity as .gitignore is usually small.
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	prefix := ""
	if len(data) > 0 && data[len(data)-1] != '\n' {
		prefix = "\n"
	}

	if _, err := f.WriteString(prefix + content + "\n"); err != nil {
		return err
	}

	return nil
}
