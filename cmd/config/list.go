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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	Long: `Display all current configuration settings.
Note: Sensitive information like API keys will be masked.`,
	Run: runList,
}

func runList(cmd *cobra.Command, args []string) {
	settings := viper.AllSettings()

	// Create a table
	tableData := pterm.TableData{
		{"Key", "Value"},
	}

	// Add all settings to the table
	for key, value := range flattenMap(settings, "") {
		if key == "api.key" {
			value = maskAPIKey(fmt.Sprintf("%v", value))
		}
		tableData = append(tableData, []string{key, fmt.Sprintf("%v", value)})
	}

	// Print the table
	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}

func flattenMap(m map[string]interface{}, prefix string) map[string]interface{} {
	flattened := make(map[string]interface{})

	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch v := v.(type) {
		case map[string]interface{}:
			nested := flattenMap(v, key)
			for nk, nv := range nested {
				flattened[nk] = nv
			}
		default:
			flattened[key] = v
		}
	}

	return flattened
}
