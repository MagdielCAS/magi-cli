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

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long: `Retrieve the value of a specific configuration key.
	
Example keys:
- api.key
- api.model
- output.format
- cache.enabled`,
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
