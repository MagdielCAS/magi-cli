package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var saltLength int

var SaltCmd = &cobra.Command{
	Use:   "salt",
	Short: "Generate a random salt key",
	Long:  `Generate a random salt key of a specified length.`,
	Example: `  # Generate a 32-byte salt (default)
  magi crypto salt

  # Generate a 64-byte salt
  magi crypto salt --length 64`,
	Run: runGenerateSalt,
}

func init() {
	SaltCmd.Flags().IntVarP(&saltLength, "length", "l", 32, "Salt length")
}

func runGenerateSalt(cmd *cobra.Command, args []string) {
	key := make([]byte, saltLength)
	_, err := rand.Read(key)
	if err != nil {
		pterm.Error.Printf("Failed to generate salt: %v\n", err)
		return
	}

	encoded := base64.StdEncoding.EncodeToString(key)
	pterm.Success.Printf("Generated salt (%d bytes):\n", saltLength)
	fmt.Println(encoded)
}
