/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/

package cmd

import (
	"os"
	"os/signal"
	"path/filepath"

	cliCommit "github.com/MagdielCAS/magi-cli/internal/cli/commit"
	"github.com/MagdielCAS/magi-cli/internal/cli/config"
	"github.com/MagdielCAS/magi-cli/internal/cli/crypto"
	"github.com/MagdielCAS/magi-cli/internal/cli/docker"
	"github.com/MagdielCAS/magi-cli/internal/cli/i18n"
	"github.com/MagdielCAS/magi-cli/internal/cli/pr"
	"github.com/MagdielCAS/magi-cli/internal/cli/project"
	"github.com/MagdielCAS/magi-cli/internal/cli/pulumi"
	"github.com/MagdielCAS/magi-cli/internal/cli/push"
	"github.com/MagdielCAS/magi-cli/internal/cli/ssh"
	"github.com/MagdielCAS/magi-cli/internal/cli/update"
	"github.com/MagdielCAS/magi-cli/pkg/utils"
	"github.com/MagdielCAS/pcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// These variables are set at build time using ldflags
	version = "v0.8.1" // <---VERSION---> Updating this version, will also create a new GitHub tag.
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
  magi config set api.key your-api-key

Run 'magi [command] --help' for more information on a specific command.`,
		Example: `  magi config set api.key your-api-key`,
		Version: version,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	setupSignalHandler()
	setupCompletionCmd()

	if err := rootCmd.Execute(); err != nil {
		utils.CheckForUpdates(rootCmd)
		os.Exit(1)
	}

	utils.CheckForUpdates(rootCmd)
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
	loadConfiguration()
}

func setupPTermFlags() {
	rootCmd.PersistentFlags().BoolVarP(&pterm.PrintDebugMessages, "debug", "", false, "enable debug messages")
	rootCmd.PersistentFlags().BoolVarP(&pterm.RawOutput, "raw", "", false, "print unstyled raw output")
	rootCmd.PersistentFlags().BoolVarP(&pcli.DisableUpdateChecking, "disable-update-checks", "", false, "disables update checks")
}
func loadConfiguration() {
	viper.AutomaticEnv()

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	globalConfigDir := filepath.Join(home, ".magi")
	globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")

	// Ensure global config directory exists
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		pterm.Error.Printf("Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	// 1. Load Global Config
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigFile(globalConfigPath)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok || os.IsNotExist(err) {
			// If the config file is not found, create it if it's the default global config path
			// or if the user explicitly specified the default global config path.
			if cfgFile == "" || cfgFile == globalConfigPath {
				if err = viper.WriteConfigAs(globalConfigPath); err != nil {
					pterm.Error.Printf("Error creating config file: %v\n", err)
					os.Exit(1)
				}
				pterm.Success.Println("New config file created at " + globalConfigPath)
			} else {
				// Custom config file not found
				pterm.Error.Printf("Config file not found: %s\n", cfgFile)
				os.Exit(1)
			}
		} else {
			pterm.Error.Printf("Error reading config file: %v\n", err)
			os.Exit(1)
		}
	}

	pterm.Debug.Printf("Using global config file: %s\n", viper.ConfigFileUsed())

	// 2. Merge Local Config (.magi.yaml)
	// Only check for local config if we are not forcing a specific config file via flag
	if cfgFile == "" {
		cwd, err := os.Getwd()
		if err == nil {
			localConfigPath := filepath.Join(cwd, ".magi.yaml")
			if _, err := os.Stat(localConfigPath); err == nil {
				// Set the config file to the local one so MergeInConfig uses it
				// And subsequent WriteConfig calls (like 'magi config set') will write to it
				viper.SetConfigFile(localConfigPath)
				if err := viper.MergeInConfig(); err != nil {
					pterm.Warning.Printf("Error merging local config file: %v\n", err)
				} else {
					pterm.Debug.Printf("Merged local config file: %s\n", localConfigPath)
				}
			}
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.magi/config.yaml)")
	rootCmd.PersistentFlags().StringP("author", "", "Magdiel Campelo <github.com/MagdielCAS>", "author name for copyright attribution")

	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.SetDefault("license", "bsd-2")
	viper.SetDefault("api.provider", "openai")
	viper.SetDefault("api.base_url", "")
	viper.SetDefault("api.light_model", "gpt-3.5-turbo")
	viper.SetDefault("api.heavy_model", "gpt-4")
	viper.SetDefault("api.fallback_model", "gpt-3.5-turbo")

	// Change global PTerm theme
	pterm.ThemeDefault.SectionStyle = *pterm.NewStyle(pterm.FgCyan)

	rootCmd.AddCommand(push.PushCmd())
	rootCmd.AddCommand(pr.PRCmd())
	rootCmd.AddCommand(cliCommit.CommitCmd())
	rootCmd.AddCommand(config.ConfigCmd())
	rootCmd.AddCommand(crypto.CryptoCmd())
	rootCmd.AddCommand(ssh.SSHCmd())
	rootCmd.AddCommand(i18n.I18nCmd())
	rootCmd.AddCommand(docker.NewDockerCommand())
	rootCmd.AddCommand(pulumi.NewPulumiCommand())
	rootCmd.AddCommand(project.NewProjectCmd())
	rootCmd.AddCommand(update.UpdateCmd())

	// Use https://github.com/pterm/pcli to style the output of cobra.
	pcli.SetRepo("MagdielCAS/magi-cli")
	pcli.SetRootCmd(rootCmd)
	pcli.Setup()
}
