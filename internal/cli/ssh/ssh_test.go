package ssh

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSSHCmd(t *testing.T) {
	cmd := SSHCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "ssh", cmd.Use)
	assert.True(t, cmd.HasSubCommands())

	subcommands := cmd.Commands()
	subcommandNames := make(map[string]bool)
	for _, sub := range subcommands {
		subcommandNames[sub.Name()] = true
	}

	assert.True(t, subcommandNames["add"])
	assert.True(t, subcommandNames["connect"])
	assert.True(t, subcommandNames["list"])
	assert.True(t, subcommandNames["remove"])
}

func TestSaveAndGetConnection(t *testing.T) {
	// Setup Viper for testing
	viper.Reset()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "config.*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	viper.SetConfigFile(tmpFile.Name())

	viper.Set(ConfigSSHConnections, map[string]interface{}{})

	// Test Data
	conn := SSHConnection{
		Alias:    "test-server",
		KeyPath:  "/tmp/test-key",
		IP:       "192.168.1.1",
		Username: "testuser",
		Port:     2222,
	}

	// Test Save
	err = saveConnection(conn)
	assert.NoError(t, err)

	// Test Get
	retrievedConn, err := getConnection("test-server")
	assert.NoError(t, err)
	assert.Equal(t, conn, retrievedConn)

	// Test Get Non-Existent
	_, err = getConnection("non-existent")
	assert.Error(t, err)
}

func TestGetConnection_EmptyConfig(t *testing.T) {
	viper.Reset()

	_, err := getConnection("any")
	assert.Error(t, err)
}
