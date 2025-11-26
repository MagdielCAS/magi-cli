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

  # Run setup non-interactively with OpenAI
  magi setup --api-provider openai --api-key YOUR_API_KEY --heavy-model gpt-4

  # Run setup non-interactively with a custom provider
  magi setup --api-provider custom --base-url http://localhost:8080 --api-key YOUR_API_KEY --heavy-model custom-model`,
		Run: runSetup,
	}

	setupCmd.Flags().String("api-provider", "", "API provider (e.g., openai, custom)")
	setupCmd.Flags().String("base-url", "", "Base URL for custom OpenAI compatible API")
	setupCmd.Flags().String("api-key", "", "Your OpenAI API key")
	setupCmd.Flags().String("light-model", "", "Model for light tasks (e.g., gpt-3.5-turbo)")
	setupCmd.Flags().String("heavy-model", "", "Model for heavy tasks (e.g., gpt-4)")
	setupCmd.Flags().String("fallback-model", "", "Fallback model (e.g., gpt-3.5-turbo)")
	setupCmd.Flags().String("format", "", "Default output format (e.g., text, json, yaml)")
	setupCmd.Flags().Bool("ci", false, "Run setup in CI mode (non-interactive, uses defaults)")

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

	isCI, _ := cmd.Flags().GetBool("ci")
	apiProvider, _ := cmd.Flags().GetString("api-provider")
	baseURL, _ := cmd.Flags().GetString("base-url")
	apiKey, _ := cmd.Flags().GetString("api-key")
	lightModel, _ := cmd.Flags().GetString("light-model")
	heavyModel, _ := cmd.Flags().GetString("heavy-model")
	fallbackModel, _ := cmd.Flags().GetString("fallback-model")
	format, _ := cmd.Flags().GetString("format")
	var err error

	if isCI {
		pterm.Info.Println("Running in CI mode...")
		if apiProvider == "" {
			apiProvider = "openai"
		}
		if apiKey == "" {
			apiKey = "ci-dummy-key"
		}
		if lightModel == "" {
			lightModel = "gpt-3.5-turbo"
		}
		if heavyModel == "" {
			heavyModel = "gpt-4"
		}
		if fallbackModel == "" {
			fallbackModel = "gpt-3.5-turbo"
		}
		if format == "" {
			format = "text"
		}
	}

	// Select API provider
	validProviders := []string{"openai", "custom"}
	if apiProvider == "" {
		apiProvider, err = pterm.DefaultInteractiveSelect.
			WithOptions(validProviders).
			WithDefaultText("Select your API provider").
			Show()
		if err != nil {
			pterm.Error.Printf("Failed to select API provider: %v\n", err)
			return
		}
	} else if !validateOption(apiProvider, validProviders) {
		pterm.Error.Printf("Invalid API provider: %s. Valid providers are: %s\n", apiProvider, strings.Join(validProviders, ", "))
		return
	}

	// Get Base URL for custom provider
	if apiProvider == "custom" {
		if baseURL == "" {
			if isCI {
				// Should have been provided via flag if needed, or we can set a default
				pterm.Warning.Println("Base URL not provided in CI mode for custom provider")
			} else {
				baseURL, err = pterm.DefaultInteractiveTextInput.
					WithMultiLine(false).
					Show("Enter the base URL for the custom API")
				if err != nil {
					pterm.Error.Printf("Failed to get base URL: %v\n", err)
					return
				}
			}
		}
	}

	// Get API Key
	if apiKey == "" {
		apiKey, err = pterm.DefaultInteractiveTextInput.
			WithMultiLine(false).
			WithMask("*").
			Show("Enter your API key")
		if err != nil {
			pterm.Error.Printf("Failed to get API key: %v\n", err)
			return
		}
	}

	// Set models
	if lightModel == "" {
		lightModel, err = pterm.DefaultInteractiveTextInput.
			WithDefaultValue("gpt-3.5-turbo").
			Show("Enter the model for light tasks")
		if err != nil {
			pterm.Error.Printf("Failed to get light model: %v\n", err)
			return
		}
	}
	if heavyModel == "" {
		heavyModel, err = pterm.DefaultInteractiveTextInput.
			WithDefaultValue("gpt-4").
			Show("Enter the model for heavy tasks")
		if err != nil {
			pterm.Error.Printf("Failed to get heavy model: %v\n", err)
			return
		}
	}
	if fallbackModel == "" {
		fallbackModel, err = pterm.DefaultInteractiveTextInput.
			WithDefaultValue("gpt-3.5-turbo").
			Show("Enter the fallback model")
		if err != nil {
			pterm.Error.Printf("Failed to get fallback model: %v\n", err)
			return
		}
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
	viper.Set("api.provider", apiProvider)
	if apiProvider == "custom" {
		viper.Set("api.base_url", baseURL)
	}
	viper.Set("api.key", apiKey)
	viper.Set("api.light_model", lightModel)
	viper.Set("api.heavy_model", heavyModel)
	viper.Set("api.fallback_model", fallbackModel)
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
