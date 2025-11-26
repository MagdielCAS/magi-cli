/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestListCmd(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	// Set up a test configuration
	viper.Set("api.model", "gpt-3.5-turbo")
	viper.Set("api.key", "sk-1234567890")
	viper.Set("output.format", "text")
	if err := viper.WriteConfig(); err != nil {
		t.Fatal(err)
	}

	output := execute(t, ListCmd)

	expectedSubstrings := []string{
		"api.model", "gpt-3.5-turbo",
		"api.key", "sk-1...7890",
		"output.format", "text",
	}

	for _, s := range expectedSubstrings {
		assert.Contains(t, output, s)
	}
}
