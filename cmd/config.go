/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package cmd

import (
	"github.com/MagdielCAS/magi-cli/cmd/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manages the magi configuration",
	Long: `Manages the magi configuration. You can get, set, list, and reset configuration values.

Available subcommands:
  get     Gets a configuration value
  set     Sets a configuration value
  list    Lists all configuration values
  reset   Resets the configuration

Usage:
  magi config [command]

Examples:
  # Get a configuration value
  magi config get api.key

  # Set a configuration value
  magi config set api.model gpt-4

  # List all configuration values
  magi config list

  # Reset the configuration
  magi config reset

Run 'magi config [command] --help' for more information on a specific command.`,
}

func init() {
	// Add all subcommands
	configCmd.AddCommand(config.GetCmd)
	configCmd.AddCommand(config.SetCmd)
	configCmd.AddCommand(config.ListCmd)
	configCmd.AddCommand(config.ResetCmd)

	rootCmd.AddCommand(configCmd)
}
