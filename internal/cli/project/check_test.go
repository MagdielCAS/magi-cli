package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCheckCmd(t *testing.T) {
	cmd := NewCheckCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "check", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}
