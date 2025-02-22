/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package config

import (
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage magi-cli configuration",
	Long: `Configure magi-cli settings and preferences.
	
You can view, set, or reset various configuration options including:
- API settings (key, endpoint, model)
- Output preferences (format, color, verbosity)
- Cache settings (enable/disable, TTL)`,
	Example: `  magi-cli config get api.key
  magi-cli config set api.model gpt-4
  magi-cli config list
  magi-cli config reset`,
}

func init() {
	// Add all subcommands
	ConfigCmd.AddCommand(getCmd)
	ConfigCmd.AddCommand(setCmd)
	ConfigCmd.AddCommand(listCmd)
	ConfigCmd.AddCommand(resetCmd)
}
