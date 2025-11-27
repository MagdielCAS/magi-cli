package crypto

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyfileCmd(t *testing.T) {
	cmd := KeyfileCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "keyfile", cmd.Use)
}

func TestGenerateFileWithDirCheck(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "magi-crypto-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filename := "test-keyfile"
	fullPath := filepath.Join(tmpDir, filename)

	// Generate keyfile
	generateFileWithDirCheck(tmpDir, filename)

	// Verify file exists
	info, err := os.Stat(fullPath)
	assert.NoError(t, err)

	if runtime.GOOS != "windows" {
		// Verify permissions (0400)
		assert.Equal(t, os.FileMode(0400), info.Mode().Perm())
	}

	// Verify content
	content, err := os.ReadFile(fullPath)
	assert.NoError(t, err)

	// Should be base64 encoded
	decoded, err := base64.StdEncoding.DecodeString(string(content))
	assert.NoError(t, err)

	// Should be 768 bytes to stay under MongoDB's 1024-character limit
	assert.Equal(t, 768, len(decoded))
}
