/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package cmd

import (
	"fmt"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of magi",
	Long: `Shows the current version of magi, build date, and commit hash.

Usage:
  magi version

Examples:
  # Default behavior
  magi version

  # Show version in JSON format
  magi version --json

Run 'magi version --help' for more information on a specific command.`,
	Run: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	version, _ := cmd.Flags().GetBool("json")
	if version {
		printVersionJSON()
	} else {
		printVersion()
	}
}

func printVersion() {
	pterm.DefaultSection.Printf("magi CLI Version: %s", version)

	// Create a table for detailed version info
	tableData := pterm.TableData{
		{"Version", version},
		{"Git Commit", commit},
		{"Build Date", date},
		{"Go Version", runtime.Version()},
		{"OS/Arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)},
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}

func printVersionJSON() {
	msg := fmt.Sprintf("magi CLI Version: %s", version)
	data := map[string]any{
		"version":   version,
		"commit":    commit,
		"buildDate": date,
		"goVersion": runtime.Version(),
		"osArch":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	logger := pterm.DefaultLogger.WithLevel(pterm.LogLevelInfo).WithFormatter(pterm.LogFormatterJSON)

	logger.Info(msg, logger.ArgsFromMap(data))
}
