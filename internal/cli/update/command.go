/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/

package update

import (
	"github.com/MagdielCAS/magi-cli/pkg/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update magi-cli to the latest version",
	Long:  `Update magi-cli to the latest version using the self-update script.`,
	Run: func(cmd *cobra.Command, args []string) {
		available, tagName, err := utils.IsUpdateAvailable(cmd.Root().Version)
		if err != nil {
			pterm.Error.Printf("Failed to check for updates: %v\n", err)
			return
		}

		if !available {
			pterm.Info.Println("magi-cli is already up to date")
			return
		}

		pterm.Info.Printfln("Updating magi-cli to %s...", tagName)
		if err := runUpdateScript(); err != nil {
			pterm.Error.Printf("Failed to update magi-cli: %v\n", err)
			return
		}

		pterm.Success.Println("magi-cli has been updated successfully!")
	},
}

func UpdateCmd() *cobra.Command {
	return updateCmd
}
