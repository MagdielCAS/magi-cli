package project

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewInitCmd(t *testing.T) {
	tests := []struct {
		name string
		want *cobra.Command
	}{
		{
			name: "Init Command Creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewInitCmd()
			assert.NotNil(t, cmd)
			assert.Equal(t, "init", cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotEmpty(t, cmd.Long)
			assert.NotNil(t, cmd.RunE)
		})
	}
}
