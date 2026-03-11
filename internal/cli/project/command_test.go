package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProjectCmd(t *testing.T) {
	cmd := NewProjectCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "project [command]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.HasSubCommands())

	// Check for expected subcommands
	expectedParams := []string{"init", "exec", "check", "update", "redo", "list"}
    foundCount := 0
	for _, sub := range cmd.Commands() {
		for _, expected := range expectedParams {
			if sub.Name() == expected {
				foundCount++
                break
			}
		}
	}
    assert.Equal(t, len(expectedParams), foundCount, "Not all expected subcommands found")
}
