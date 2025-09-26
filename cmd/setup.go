/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package cmd

import (
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Starts an interactive setup wizard for magi",
		Long: `The setup command starts an interactive wizard to help you configure magi for first use.
It will guide you through setting up your API key and other preferences.
This command can also be run non-interactively by providing the required flags.

Usage:
  magi setup [flags]

Examples:
  # Run the interactive setup wizard
  magi setup

  # Run setup non-interactively
  magi setup --api-key YOUR_API_KEY --model gpt-4 --format text`,
		Run: runSetup,
	}

	setupCmd.Flags().String("api-key", "", "Your OpenAI API key")
	setupCmd.Flags().String("model", "", "Default AI model (e.g., gpt-4, gpt-3.5-turbo)")
	setupCmd.Flags().String("format", "", "Default output format (e.g., text, json, yaml)")

	rootCmd.AddCommand(setupCmd)
}

func validateOption(option string, validOptions []string) bool {
	for _, o := range validOptions {
		if option == o {
			return true
		}
	}
	return false
}

func runSetup(cmd *cobra.Command, args []string) {
	pterm.Info.Println("Starting magi setup...")

	apiKey, _ := cmd.Flags().GetString("api-key")
	model, _ := cmd.Flags().GetString("model")
	format, _ := cmd.Flags().GetString("format")
	var err error

	// Get API Key
	if apiKey == "" {
		apiKey, err = pterm.DefaultInteractiveTextInput.
			WithMultiLine(false).
			WithMask("*").
			Show("Enter your OpenAI API key")
		if err != nil {
			pterm.Error.Printf("Failed to get API key: %v\n", err)
			return
		}
	}

	// Set default model
	validModels := []string{"gpt-4", "gpt-3.5-turbo"}
	if model == "" {
		model, err = pterm.DefaultInteractiveSelect.
			WithOptions(validModels).
			WithDefaultText("Select default AI model").
			Show()
		if err != nil {
			pterm.Error.Printf("Failed to select model: %v\n", err)
			return
		}
	} else if !validateOption(model, validModels) {
		pterm.Error.Printf("Invalid model: %s. Valid models are: %s\n", model, strings.Join(validModels, ", "))
		return
	}

	// Configure output format
	validFormats := []string{"text", "json", "yaml"}
	if format == "" {
		format, err = pterm.DefaultInteractiveSelect.
			WithOptions(validFormats).
			WithDefaultText("Select default output format").
			Show()
		if err != nil {
			pterm.Error.Printf("Failed to select output format: %v\n", err)
			return
		}
	} else if !validateOption(format, validFormats) {
		pterm.Error.Printf("Invalid format: %s. Valid formats are: %s\n", format, strings.Join(validFormats, ", "))
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
