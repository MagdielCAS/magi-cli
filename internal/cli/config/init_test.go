package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitCmd(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "magi-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change working directory to tmpDir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Reset viper
	viper.Reset()

	// Create a dummy .gitignore
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignorePath, []byte("node_modules\n"), 0644)

	// Run InitCmd
	InitCmd.Run(InitCmd, []string{})

	// Check if .magi.yaml exists
	configPath := filepath.Join(tmpDir, ".magi.yaml")
	assert.FileExists(t, configPath)

	// Check if .magi.yaml was added to .gitignore
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, string(content), ".magi.yaml")

	// Run InitCmd again (should not fail, but warn)
	InitCmd.Run(InitCmd, []string{})
	assert.FileExists(t, configPath)

	// Check idempotency of .gitignore
	content, err = os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	// Count occurrences
	count := 0
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == ".magi.yaml" {
			count++
		}
	}
	assert.Equal(t, 1, count, ".magi.yaml should appear exactly once in .gitignore")
}
