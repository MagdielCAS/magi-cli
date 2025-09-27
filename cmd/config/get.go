/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package config

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var GetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Gets a configuration value",
	Long: `Gets a configuration value.

Usage:
  magi config get [key]

Examples:
  # Get the value of a key
  magi config get api.model

Run 'magi config get --help' for more information on a specific command.`,
	Args: cobra.ExactArgs(1),
	Run:  runGet,
}

func runGet(cmd *cobra.Command, args []string) {
	key := args[0]

	if !viper.IsSet(key) {
		pterm.Error.Printf("Configuration key '%s' not found\n", key)
		return
	}

	value := viper.Get(key)

	// Handle sensitive data
	if key == "api.key" {
		fmt.Printf("%s: %s\n", key, maskAPIKey(fmt.Sprintf("%v", value)))
		return
	}

	fmt.Printf("%s: %v\n", key, value)
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
