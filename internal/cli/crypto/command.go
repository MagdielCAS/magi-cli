package crypto

import (
	"github.com/spf13/cobra"
)

var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "Cryptographic utilities",
	Long: `Cryptographic utilities for generating secure keys, salts, and keyfiles.

Available subcommands:
  salt        Generate a random salt key
  keyfile     Generate a MongoDB keyfile
  keypair     Generate a public/private key pair

Usage:
  magi crypto [command]

Examples:
  # Default behavior (generates a salt)
  magi crypto

  # Generate a salt
  magi crypto salt

  # Generate a MongoDB keyfile
  magi crypto keyfile

  # Generate a key pair
  magi crypto keypair

Run 'magi crypto [command] --help' for more information on a specific command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Default to salt
			SaltCmd.Run(cmd, args)
		}
	},
}

func CryptoCmd() *cobra.Command {
	cryptoCmd.AddCommand(SaltCmd)
	cryptoCmd.AddCommand(KeyfileCmd)
	cryptoCmd.AddCommand(KeypairCmd)
	return cryptoCmd
}
