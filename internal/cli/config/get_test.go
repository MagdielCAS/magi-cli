/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package config

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetCmd(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	// Set up a test configuration
	viper.Set("api.model", "gpt-3.5-turbo")
	viper.Set("api.key", "sk-1234567890")
	if err := viper.WriteConfig(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		args          []string
		expected      string
		shouldContain bool
	}{
		{
			name:          "Get existing key",
			args:          []string{"api.model"},
			expected:      "api.model: gpt-3.5-turbo",
			shouldContain: true,
		},
		{
			name:          "Get non-existent key",
			args:          []string{"non.existent.key"},
			expected:      "Configuration key 'non.existent.key' not found",
			shouldContain: true,
		},
		{
			name:          "Get sensitive key",
			args:          []string{"api.key"},
			expected:      "api.key: sk-1...7890",
			shouldContain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := execute(t, GetCmd, tt.args...)
			fmt.Println(output)

			require.Contains(t, output, tt.expected, "output should contain the expected message")
		})
	}
}
