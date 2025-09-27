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

func TestResetCmd(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	// Set up a test configuration
	viper.Set("api.model", "gpt-4")
	viper.Set("api.key", "sk-1234567890")
	viper.Set("output.format", "json")
	require.NoError(t, viper.WriteConfig())

	execute(t, ResetCmd)

	require.Equal(t, "gpt-3.5-turbo", viper.GetString("api.model"), "api.model should be reset to 'gpt-3.5-turbo'")
	require.Equal(t, "text", viper.GetString("output.format"), "output.format should be reset to 'text'")
	require.Equal(t, "sk-1234567890", viper.GetString("api.key"), "api.key should not be reset")
}
