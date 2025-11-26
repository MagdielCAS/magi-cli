package ssh

import (
	"sort"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved SSH connections",
		Long: `Display a table of all saved SSH connections.

The table includes:
- Alias
- IP Address
- Username
- Port
- Key Path

Usage:
  magi ssh list

Examples:
  # List all connections
  magi ssh list`,
		Run: func(cmd *cobra.Command, args []string) {
			listConnections()
		},
	}
	return cmd
}

func listConnections() {
	var connMap map[string]SSHConnection
	if err := viper.UnmarshalKey(ConfigSSHConnections, &connMap); err != nil {
		pterm.Error.Printf("Failed to load connections: %v\n", err)
		return
	}

	if len(connMap) == 0 {
		pterm.Info.Println("No SSH connections found. Run 'magi ssh add' to add one.")
		return
	}

	// Prepare table data
	data := [][]string{
		{"Alias", "IP", "User", "Port", "Key Path"},
	}

	var aliases []string
	for k := range connMap {
		aliases = append(aliases, k)
	}
	sort.Strings(aliases)

	for _, alias := range aliases {
		conn := connMap[alias]
		data = append(data, []string{
			conn.Alias,
			conn.IP,
			conn.Username,
			strconv.Itoa(conn.Port),
			conn.KeyPath,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
}
