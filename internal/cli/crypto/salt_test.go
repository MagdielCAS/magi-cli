package crypto

import (
	"bytes"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSaltCmd(t *testing.T) {
	cmd := SaltCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "salt", cmd.Use)
}

func TestRunGenerateSalt(t *testing.T) {
	tests := []struct {
		name       string
		length     int
		expectSkip bool
	}{
		{"DefaultLength", 32, false},
		{"CustomLength", 16, false},
		{"LargeLength", 64, false},
		{"InvalidLength", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Restore stdout at the end
			defer func() {
				os.Stdout = oldStdout
			}()

			// Set flag value
			saltLength = tt.length

			// Execute
			runGenerateSalt(&cobra.Command{}, []string{})

			// Close write end of pipe
			w.Close()

			// Read captured output
			var buf bytes.Buffer
_, _ = buf.ReadFrom(r)
output := buf.String()

if tt.expectSkip {
assert.Empty(t, strings.TrimSpace(output))
return
}

			// Verify
			lines := strings.Split(strings.TrimSpace(output), "\n")
			// The last line should be the encoded salt
			// pterm might add colors/formatting, so we look for the last non-empty line
			var encodedSalt string
			for i := len(lines) - 1; i >= 0; i-- {
				if strings.TrimSpace(lines[i]) != "" {
					encodedSalt = strings.TrimSpace(lines[i])
					break
				}
			}

			// If pterm output is mixed in, we might need to be more careful.
			// But fmt.Println(encoded) should be clean if pterm writes to stderr or if we just look at the last line.
			// However, pterm writes to stdout by default.
			// Let's try to decode the last line.

			// Clean up any potential ANSI codes if pterm added them to the salt line (unlikely for fmt.Println)
			decoded, err := base64.StdEncoding.DecodeString(encodedSalt)
			if err != nil {
				// Try the line before, maybe there's a trailing newline issue
				if len(lines) > 1 {
					encodedSalt = strings.TrimSpace(lines[len(lines)-2])
					decoded, err = base64.StdEncoding.DecodeString(encodedSalt)
				}
			}

			assert.NoError(t, err, "Failed to decode salt: %s", encodedSalt)
			assert.Equal(t, tt.length, len(decoded))
		})
	}
}
