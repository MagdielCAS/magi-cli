package project

import (
	"github.com/spf13/cobra"
)

// NewProjectCmd creates the project command which manages project architecture and rules.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [command]",
		Short: "Manage project architecture, rules, and code generation",
		Long: `The project command understands the code architecture using AI agents.
It helps verify project structure, create new features, and manage architectural rules.

Available subcommands:
  init    Initialize project rules and configuration
  create  Create new features/components
  check   Check compliance with project rules
  update  Update existing structures
  redo    Re-analyze project structure

Usage:
  magi project [command]

Examples:
  magi project init
  magi project create slice --name my-feature
  magi project check`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewExecCmd())
	cmd.AddCommand(NewCheckCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewRedoCmd())
    cmd.AddCommand(NewListCmd())

	return cmd
}
