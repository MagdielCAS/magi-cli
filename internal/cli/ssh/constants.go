package ssh

import "errors"

const (
	// Config Keys
	ConfigSSHConnections = "ssh_connections"
	ConfigSSHKeys        = "ssh_keys"

	// Defaults
	DefaultPort = 22
	DefaultUser = "ubuntu"

	// Validation Limits
	MinPort = 1
	MaxPort = 65535
)

var (
	// Errors
	ErrAliasExists      = errors.New("connection alias already exists")
	ErrAliasNotFound    = errors.New("connection alias not found")
	ErrInvalidIP        = errors.New("invalid IP address")
	ErrInvalidPort      = errors.New("invalid port number")
	ErrKeyNotFound      = errors.New("SSH key file not found")
	ErrEmptyAlias       = errors.New("alias cannot be empty")
	ErrInvalidAlias     = errors.New("alias contains invalid characters")
	ErrConnectionFailed = errors.New("failed to establish SSH connection")
)

// SSHConnection represents a saved SSH connection configuration
type SSHConnection struct {
	Alias    string `mapstructure:"alias" json:"alias"`
	KeyPath  string `mapstructure:"key_path" json:"key_path"`
	IP       string `mapstructure:"ip" json:"ip"`
	Username string `mapstructure:"username" json:"username"`
	Port     int    `mapstructure:"port" json:"port"`
}
