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

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: `Set the value for a specific configuration key.
	
Common configuration keys:
- api.key: Your OpenAI API key
- api.model: AI model to use (e.g., gpt-4, gpt-3.5-turbo)
- output.format: Output format (text, json, yaml)
- cache.enabled: Enable/disable caching (true/false)
- cache.ttl: Cache time-to-live in seconds`,
	Example: `  magi-cli config set api.key your-api-key
  magi-cli config set api.model gpt-4
  magi-cli config set output.format json`,
	Args: cobra.ExactArgs(2),
	Run:  runSet,
}

func runSet(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]

	// Validate key-value pairs
	if !isValidConfig(key, value) {
		pterm.Error.Printf("Invalid configuration: %s = %s\n", key, value)
		return
	}

	viper.Set(key, value)
	if err := viper.WriteConfig(); err != nil {
		pterm.Error.Printf("Failed to save configuration: %v\n", err)
		return
	}

	pterm.Success.Printf("Configuration updated: %s = %v\n", key, value)
}

func isValidConfig(key, value string) bool {
	switch key {
	case "output.format":
		return value == "text" || value == "json" || value == "yaml"
	case "cache.enabled":
		return value == "true" || value == "false"
	case "api.model":
		return value == "gpt-4" || value == "gpt-3.5-turbo"
	case "cache.ttl":
		// Add validation for numeric value if needed
		return true
	default:
		return true
	}
}
