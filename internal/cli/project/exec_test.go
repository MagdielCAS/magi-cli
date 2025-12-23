package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExecCmd(t *testing.T) {
	cmd := NewExecCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "exec [action]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}
