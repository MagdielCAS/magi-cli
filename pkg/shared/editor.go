package shared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// OpenEditor opens the user's preferred editor with the given content.
// It returns the edited content and any error derived from the operation.
func OpenEditor(initialContent, extension string) (string, error) {
	editorEnv := os.Getenv("VISUAL")
	if editorEnv == "" {
		editorEnv = os.Getenv("EDITOR")
	}
	if editorEnv == "" {
		if runtime.GOOS == "windows" {
			editorEnv = "notepad"
		} else {
			editorEnv = "vim"
		}
	}

	tmpFile, err := os.CreateTemp("", "magi-editor-*"+extension)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(initialContent); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Split the editor command string into command and arguments
	parts := strings.Fields(editorEnv)
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid editor environment variable")
	}

	editorCmd := parts[0]
	editorArgs := parts[1:]

	// Heuristic: If editor is VS Code, ensure --wait is passed
	if isVSCode(editorCmd) {
		hasWait := false
		for _, arg := range editorArgs {
			if arg == "--wait" || arg == "-w" {
				hasWait = true
				break
			}
		}
		if !hasWait {
			editorArgs = append(editorArgs, "--wait")
		}
	}

	editorArgs = append(editorArgs, tmpFile.Name())

	cmd := exec.Command(editorCmd, editorArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor execution failed: %w", err)
	}

	raw, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read back temp file: %w", err)
	}

	return string(raw), nil
}

func isVSCode(cmd string) bool {
	base := filepath.Base(cmd)
	return base == "code" || base == "code-insiders" || base == "cursor"
}
