package ssh

import (
	"fmt"
	"sort"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func removeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [alias]",
		Short: "Remove a saved SSH connection",
		Long: `Remove a saved SSH connection by its alias.

If no alias is provided, an interactive list of connections will be shown to select from.
You will be prompted for confirmation before the connection is deleted.

Usage:
  magi ssh remove [alias]

Examples:
  # Remove a specific connection
  magi ssh remove prod-db

  # Select a connection to remove
  magi ssh remove`,
		Run: func(cmd *cobra.Command, args []string) {
			alias := ""
			if len(args) > 0 {
				alias = args[0]
			}
			removeConnection(alias)
		},
	}
	return cmd
}

func removeConnection(alias string) {
	var connMap map[string]SSHConnection
	if err := viper.UnmarshalKey(ConfigSSHConnections, &connMap); err != nil {
		pterm.Error.Printf("Failed to load connections: %v\n", err)
		return
	}

	if len(connMap) == 0 {
		pterm.Info.Println("No SSH connections found.")
		return
	}

	if alias == "" {
		// Interactive selection
		var aliases []string
		for k := range connMap {
			aliases = append(aliases, k)
		}
		sort.Strings(aliases)

		var err error
		alias, err = pterm.DefaultInteractiveSelect.
			WithDefaultText("Select connection to remove").
			WithOptions(aliases).
			Show()
		if err != nil {
			pterm.Error.Println(err)
			return
		}
	}

	if _, exists := connMap[alias]; !exists {
		pterm.Error.Printf("Connection '%s' not found\n", alias)
		return
	}

	confirm, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultText(fmt.Sprintf("Are you sure you want to remove '%s'?", alias)).
		Show()

	if !confirm {
		pterm.Info.Println("Operation cancelled")
		return
	}

	delete(connMap, alias)
	viper.Set(ConfigSSHConnections, connMap)

	if err := viper.WriteConfig(); err != nil {
		pterm.Error.Printf("Failed to save config: %v\n", err)
		return
	}

	pterm.Success.Printf("Connection '%s' removed successfully\n", alias)
}
