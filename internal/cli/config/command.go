/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package config

import (
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
  init    Initialize a local configuration file

Usage:
  magi config [command]

Examples:
  # Get a configuration value
  magi config get api.key

  # Set a configuration value
  magi config set api.heavy_model gpt-4

  # Set the API provider
  magi config set api.provider custom

  # Set the base URL for the custom provider
  magi config set api.base_url http://localhost:8080

  # List all configuration values
  magi config list

  # Reset the configuration
  magi config reset

  # Initialize a local configuration file
  magi config init

Run 'magi config [command] --help' for more information on a specific command.`,
}

func ConfigCmd() *cobra.Command {
	// Add all subcommands
	configCmd.AddCommand(GetCmd)
	configCmd.AddCommand(SetCmd)
	configCmd.AddCommand(ListCmd)
	configCmd.AddCommand(ResetCmd)
	configCmd.AddCommand(InitCmd)

	return configCmd
}
