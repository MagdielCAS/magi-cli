package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedoCmd(t *testing.T) {
	cmd := NewRedoCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "redo", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}
