/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package config

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Resets the configuration",
	Long: `Resets the configuration to its default values.

Usage:
  magi config reset

Examples:
  # Reset the configuration
  magi config reset

Run 'magi config reset --help' for more information on a specific command.`,
	Run: runReset,
}

func runReset(cmd *cobra.Command, args []string) {
	// Store API key temporarily
	apiKey := viper.GetString("api.key")

	// Set default values
	defaults := map[string]interface{}{
		"api": map[string]interface{}{
			"key":      apiKey,
			"model":    "gpt-3.5-turbo",
			"endpoint": "https://api.openai.com/v1",
		},
		"output": map[string]interface{}{
			"format": "text",
			"color":  true,
		},
		"cache": map[string]interface{}{
			"enabled":  true,
			"ttl":      3600,
			"max_size": "100MB",
		},
	}

	// Clear current config
	viper.Reset()

	// Set new defaults
	for k, v := range defaults {
		viper.Set(k, v)
	}

	// Save the configuration
	if err := viper.WriteConfig(); err != nil {
		pterm.Error.Printf("Failed to save configuration: %v\n", err)
		return
	}

	pterm.Success.Println("Configuration reset to defaults")
	runList(cmd, args) // Show new configuration
}
