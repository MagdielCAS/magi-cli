package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new SSH connection",
		Long: `Interactive wizard to add a new SSH connection configuration.

This command will prompt you for:
- Connection Alias (unique name)
- SSH Key (select existing or add new)
- Server IP
- Username (default: ubuntu)
- Port (default: 22)

Usage:
  magi ssh add

Examples:
  # Start the interactive add wizard
  magi ssh add`,
		Run: func(cmd *cobra.Command, args []string) {
			addConnection()
		},
	}
	return cmd
}

func addConnection() {
	pterm.DefaultHeader.WithFullWidth().Println("SSH Connection Add")

	// 1. Alias
	alias, err := promptForAlias()
	if err != nil {
		pterm.Error.Println(err)
		return
	}

	// 2. SSH Key
	keyPath, err := selectOrAddSSHKey()
	if err != nil {
		pterm.Error.Println(err)
		return
	}

	// 3. Connection Details
	conn, err := collectConnectionConfig(alias, keyPath)
	if err != nil {
		pterm.Error.Println(err)
		return
	}

	// 4. Save
	if err := saveConnection(conn); err != nil {
		pterm.Error.Printf("Failed to save connection: %v\n", err)
		return
	}

	pterm.Success.Printf("Connection '%s' saved successfully!\n", alias)
}

func promptForAlias() (string, error) {
	var alias string
	var err error

	existing := viper.GetStringMap(ConfigSSHConnections)

	aliasRegex := regexp.MustCompile("^[a-zA-Z0-9_-]+$")

	for {
		alias, err = pterm.DefaultInteractiveTextInput.WithDefaultText("Connection Alias").Show()
		if err != nil {
			return "", err
		}

		alias = strings.TrimSpace(alias)
		if alias == "" {
			pterm.Warning.Println("Alias cannot be empty")
			continue
		}

		// Check for valid characters (alphanumeric, -, _)
		if !aliasRegex.MatchString(alias) {
			pterm.Warning.Println("Alias can only contain letters, numbers, hyphens, and underscores")
			continue
		}

		if _, exists := existing[alias]; exists {
			overwrite, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Alias '%s' already exists. Overwrite?", alias)).Show()
			if !overwrite {
				continue
			}
		}
		break
	}
	return alias, nil
}

func selectOrAddSSHKey() (string, error) {
	keys := viper.GetStringSlice(ConfigSSHKeys)
	options := append([]string{"Add new key path"}, keys...)

	selection, err := pterm.DefaultInteractiveSelect.
		WithDefaultText("Select SSH Key").
		WithOptions(options).
		Show()
	if err != nil {
		return "", err
	}

	if selection == "Add new key path" {
		return promptForNewKey()
	}

	return selection, nil
}

func promptForNewKey() (string, error) {
	for {
		path, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Enter absolute path to private key").Show()
		if err != nil {
			return "", err
		}

		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		// Expand ~ if present
		if strings.HasPrefix(path, "~/") {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, path[2:])
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			pterm.Warning.Printf("File does not exist: %s\n", path)
			continue
		}

		// Save new key to list if not exists
		keys := viper.GetStringSlice(ConfigSSHKeys)
		exists := false
		for _, k := range keys {
			if k == path {
				exists = true
				break
			}
		}
		if !exists {
			keys = append(keys, path)
			viper.Set(ConfigSSHKeys, keys)
		}

		return path, nil
	}
}

func collectConnectionConfig(alias, keyPath string) (SSHConnection, error) {
	// IP
	var ip string
	var err error
	for {
		ip, err = pterm.DefaultInteractiveTextInput.WithDefaultText("Server IP").Show()
		if err != nil {
			return SSHConnection{}, err
		}
		if net.ParseIP(ip) == nil {
			pterm.Warning.Println("Invalid IP address")
			continue
		}
		break
	}

	// Username
	username, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("Username").
		WithDefaultValue(DefaultUser).
		Show()
	if err != nil {
		return SSHConnection{}, err
	}
	if username == "" {
		username = DefaultUser
	}

	// Port
	var port int
	for {
		portStr, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("Port").
			WithDefaultValue(strconv.Itoa(DefaultPort)).
			Show()
		if err != nil {
			return SSHConnection{}, err
		}

		p, err := strconv.Atoi(portStr)
		if err != nil || p < MinPort || p > MaxPort {
			pterm.Warning.Printf("Port must be between %d and %d\n", MinPort, MaxPort)
			continue
		}
		port = p
		break
	}

	return SSHConnection{
		Alias:    alias,
		KeyPath:  keyPath,
		IP:       ip,
		Username: username,
		Port:     port,
	}, nil
}

func saveConnection(conn SSHConnection) error {
	var connMap map[string]SSHConnection
	if err := viper.UnmarshalKey(ConfigSSHConnections, &connMap); err != nil {
		// If unmarshal fails (e.g. empty), initialize map
		connMap = make(map[string]SSHConnection)
	}
	if connMap == nil {
		connMap = make(map[string]SSHConnection)
	}

	connMap[conn.Alias] = conn
	viper.Set(ConfigSSHConnections, connMap)

	return viper.WriteConfig()
}
