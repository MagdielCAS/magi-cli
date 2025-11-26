/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestSetCmd(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expected      string
		shouldContain bool
		checkViper    func(t *testing.T)
	}{
		{
			name:          "Set valid key-value pair",
			args:          []string{"api.model", "gpt-4"},
			expected:      "Configuration updated: api.model = gpt-4",
			shouldContain: true,
			checkViper: func(t *testing.T) {
				require.Equal(t, "gpt-4", viper.GetString("api.model"), "api.model should be set to 'gpt-4'")
			},
		},
		{
			name:          "Set invalid key-value pair",
			args:          []string{"output.format", "invalid-format"},
			expected:      "Invalid configuration: output.format = invalid-format",
			shouldContain: true,
			checkViper: func(t *testing.T) {
				// The value should not have been set
				require.Equal(t, "", viper.GetString("output.format"), "output.format should not be set")
			},
		},
		{
			name:          "Set new key",
			args:          []string{"new.key", "new-value"},
			expected:      "Configuration updated: new.key = new-value",
			shouldContain: true,
			checkViper: func(t *testing.T) {
				require.Equal(t, "new-value", viper.GetString("new.key"), "new.key should be set to 'new-value'")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			output := execute(t, SetCmd, tt.args...)

			require.Contains(t, output, tt.expected, "output should contain the expected message")
			tt.checkViper(t)
		})
	}
}
