/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Initial setup of magi",
		Long: `Setup command helps you configure magi for first use.
It will guide you through setting up your API key and other preferences.`,
		Run: runSetup,
	}

	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	pterm.Info.Println("Starting magi setup...")

	// Get API Key
	apiKey, err := pterm.DefaultInteractiveTextInput.
		WithMultiLine(false).
		WithMask("*").
		Show("Enter your OpenAI API key")

	if err != nil {
		pterm.Error.Printf("Failed to get API key: %v\n", err)
		return
	}

	// Set default model
	model, err := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"gpt-4", "gpt-3.5-turbo"}).
		WithDefaultText("Select default AI model").
		Show()

	if err != nil {
		pterm.Error.Printf("Failed to select model: %v\n", err)
		return
	}

	// Configure output format
	format, err := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"text", "json", "yaml"}).
		WithDefaultText("Select default output format").
		Show()

	if err != nil {
		pterm.Error.Printf("Failed to select output format: %v\n", err)
		return
	}

	// Save configuration
	viper.Set("api.key", apiKey)
	viper.Set("api.model", model)
	viper.Set("output.format", format)
	viper.Set("output.color", true)
	viper.Set("cache.enabled", true)
	viper.Set("cache.ttl", 3600)

	if err := viper.WriteConfig(); err != nil {
		pterm.Error.Printf("Failed to save configuration: %v\n", err)
		return
	}

	pterm.Success.Println("Setup completed successfully!")
	pterm.Info.Println("You can modify these settings anytime using 'magi config' commands")
}
