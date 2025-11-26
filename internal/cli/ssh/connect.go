package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func connectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect [alias]",
		Short: "Connect to a saved SSH server",
		Long: `Connect to a saved SSH server using its alias.

If no alias is provided, an interactive list of available connections will be shown.

Usage:
  magi ssh connect [alias]

Examples:
  # Connect using a specific alias
  magi ssh connect prod-db

  # Select from a list of connections
  magi ssh connect`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := ""
			if len(args) > 0 {
				alias = args[0]
			}

			if err := connectToHost(alias); err != nil {
				pterm.Error.Println(err)
				return err
			}

			return nil
		},
	}
	return cmd
}

func connectToHost(alias string) error {
	var conn SSHConnection
	var err error

	if alias == "" {
		alias, conn, err = selectConnection()
		if err != nil {
			return err
		}
	} else {
		conn, err = getConnection(alias)
		if err != nil {
			return err
		}
	}

	pterm.Info.Printf("Connecting to %s (%s)...\n", alias, conn.IP)

	if err := executeSSHConnection(conn); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	return nil
}

func selectConnection() (string, SSHConnection, error) {
	var connMap map[string]SSHConnection
	if err := viper.UnmarshalKey(ConfigSSHConnections, &connMap); err != nil {
		return "", SSHConnection{}, fmt.Errorf("failed to load connections: %w", err)
	}

	if len(connMap) == 0 {
		return "", SSHConnection{}, fmt.Errorf("no connections found. Run 'magi ssh add' to add one")
	}

	var aliases []string
	for k := range connMap {
		aliases = append(aliases, k)
	}
	sort.Strings(aliases)

	selection, err := pterm.DefaultInteractiveSelect.
		WithDefaultText("Select connection").
		WithOptions(aliases).
		Show()
	if err != nil {
		return "", SSHConnection{}, err
	}

	return selection, connMap[selection], nil
}

func getConnection(alias string) (SSHConnection, error) {
	var connMap map[string]SSHConnection
	if err := viper.UnmarshalKey(ConfigSSHConnections, &connMap); err != nil {
		return SSHConnection{}, fmt.Errorf("failed to load connections: %w", err)
	}

	conn, ok := connMap[alias]
	if !ok {
		return SSHConnection{}, fmt.Errorf("connection '%s' not found", alias)
	}
	return conn, nil
}

func executeSSHConnection(conn SSHConnection) error {
	// Build SSH command
	// ssh -i keyPath user@ip -p port
	args := []string{
		"-i", conn.KeyPath,
		"-p", fmt.Sprintf("%d", conn.Port),
		fmt.Sprintf("%s@%s", conn.Username, conn.IP),
	}

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
