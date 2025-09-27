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

var SetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Sets a configuration value",
	Long: `Sets a configuration value.

Usage:
  magi config set [key] [value]

Examples:
  # Set the value of a key
  magi config set api.model gpt-4

Run 'magi config set --help' for more information on a specific command.`,
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
