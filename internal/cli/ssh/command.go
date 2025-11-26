package ssh

import (
	"github.com/spf13/cobra"
)

// SSHCmd represents the base ssh command
func SSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Manage and connect to SSH servers",
		Long: `A comprehensive SSH connection management system.
Allows you to add, connect, list, and remove SSH connections with ease.

Available subcommands:
  add         Add a new SSH connection
  connect     Connect to a saved SSH server
  list        List all saved SSH connections
  remove      Remove a saved SSH connection

Usage:
  magi ssh [command]`,
		Example: `  # Add a new connection
  magi ssh add

  # Connect to a saved server
  magi ssh connect my-server

  # List all connections
  magi ssh list`,
	}

	// Register subcommands
	cmd.AddCommand(addCmd())
	cmd.AddCommand(connectCmd())
	cmd.AddCommand(listCmd())
	cmd.AddCommand(removeCmd())

	return cmd
}
