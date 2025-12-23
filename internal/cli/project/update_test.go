package project

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdateCmd(t *testing.T) {
	cmd := NewUpdateCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "update [file]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
    // Check args validation
    assert.Error(t, cmd.Args(&cobra.Command{}, []string{"a", "b"})) // Max 1 arg
    assert.NoError(t, cmd.Args(&cobra.Command{}, []string{"a"}))
}
