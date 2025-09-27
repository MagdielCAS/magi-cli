/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/

package cmd

import (
	"os"
	"os/signal"
	"path/filepath"

	"github.com/MagdielCAS/pcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// These variables are set at build time using ldflags
	version = "dev"
	commit  = "none"
	date    = "unknown"

	rootCmd = &cobra.Command{
		Use:   "magi",
		Short: "A powerful AI-assisted CLI for developers that enhances productivity",
		Long: `magi is a command-line interface tool that leverages AI capabilities
to enhance developer productivity. It provides various commands for code analysis,
documentation, suggestions, and more.

Available Commands:
  setup         Initial setup of magi
  config        Manage magi configuration
  completion    Generate completion script

Usage:
  magi [command]

Examples:
  # Run the setup wizard
  magi setup

  # Configure your API key
  magi config set api-key your-api-key

Run 'magi [command] --help' for more information on a specific command.`,
		Example: `  magi config set api-key your-api-key`,
		Version: "v0.2.0", // <---VERSION---> Updating this version, will also create a new GitHub tag.
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	setupSignalHandler()
	setupCompletionCmd()

	if err := rootCmd.Execute(); err != nil {
		pcli.CheckForUpdates()
		os.Exit(1)
	}

	pcli.CheckForUpdates()
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		pterm.Warning.Println("user interrupt")
		pcli.CheckForUpdates()
		os.Exit(0)
	}()
}

func setupCompletionCmd() {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(magi completion bash)

Zsh:
  # If shell completion is not already enabled in your environment:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session:
  $ magi completion zsh > "${fpath[1]}/_magi"

Fish:
  $ magi completion fish | source

  # To load completions for each session:
  $ magi completion fish > ~/.config/fish/completions/magi.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			}
		},
	}

	rootCmd.AddCommand(completionCmd)
}

func initConfig() {
	setupPTermFlags()
	setupConfigFile()
	setupViper()
}

func setupPTermFlags() {
	rootCmd.PersistentFlags().BoolVarP(&pterm.PrintDebugMessages, "debug", "", false, "enable debug messages")
	rootCmd.PersistentFlags().BoolVarP(&pterm.RawOutput, "raw", "", false, "print unstyled raw output")
	rootCmd.PersistentFlags().BoolVarP(&pcli.DisableUpdateChecking, "disable-update-checks", "", false, "disables update checks")
}

func setupConfigFile() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		return
	}

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	configDir := filepath.Join(home, ".magi")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		pterm.Error.Printf("Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath("$HOME/.magi")
	viper.AddConfigPath(".")
}

func setupViper() {
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err = viper.SafeWriteConfig(); err != nil {
				pterm.Error.Printf("Error creating config file: %v\n", err)
				os.Exit(1)
			}
			pterm.Success.Println("New config file created")
		} else {
			pterm.Error.Printf("Error reading config file: %v\n", err)
			os.Exit(1)
		}
	}

	pterm.Debug.Printf("Using config file: %s\n", viper.ConfigFileUsed())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.magi/config.yaml)")
	rootCmd.PersistentFlags().StringP("author", "", "Magdiel Campelo <github.com/MagdielCAS>", "author name for copyright attribution")

	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.SetDefault("license", "bsd-2")

	// Use https://github.com/pterm/pcli to style the output of cobra.
	pcli.SetRepo("MagdielCAS/magi-cli")
	pcli.SetRootCmd(rootCmd)
	pcli.Setup()

	// Change global PTerm theme
	pterm.ThemeDefault.SectionStyle = *pterm.NewStyle(pterm.FgCyan)
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
