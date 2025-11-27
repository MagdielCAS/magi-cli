package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	keyfileYes         bool
	keyfileFilename    string
	keyfilePath        string
	keyfileInteractive bool
)

var KeyfileCmd = &cobra.Command{
	Use:   "keyfile",
	Short: "Generate a MongoDB keyfile",
	Long: `Generate a MongoDB keyfile for replica set authentication.
The keyfile contains 768 bytes of random data (1024 base64 characters), base64 encoded.
File permissions are set to 0400 (read-only for owner) for security.`,
	Example: `  # Generate keyfile with default settings (prompts for confirmation)
  magi crypto keyfile

  # Generate keyfile non-interactively
  magi crypto keyfile --yes

  # Generate keyfile with custom name and path
  magi crypto keyfile --filename my-key --path ./secrets --yes

  # Interactive mode
  magi crypto keyfile --interactive`,
	Run: runGenerateKeyfile,
}

func init() {
	KeyfileCmd.Flags().BoolVarP(&keyfileYes, "yes", "y", false, "Skip prompts")
	KeyfileCmd.Flags().StringVarP(&keyfileFilename, "filename", "f", "keyfile", "Filename")
	KeyfileCmd.Flags().StringVarP(&keyfilePath, "path", "p", ".", "Directory path")
	KeyfileCmd.Flags().BoolVarP(&keyfileInteractive, "interactive", "i", false, "Interactive mode")
}

func runGenerateKeyfile(cmd *cobra.Command, args []string) {
	if keyfileInteractive && keyfileYes {
		pterm.Error.Println("Cannot use both --interactive and --yes flags")
		return
	}

	if keyfileInteractive {
		var err error
		keyfileFilename, err = pterm.DefaultInteractiveTextInput.WithDefaultText(keyfileFilename).Show("Enter filename")
		if err != nil {
			return
		}
		keyfilePath, err = pterm.DefaultInteractiveTextInput.WithDefaultText(keyfilePath).Show("Enter directory path")
		if err != nil {
			return
		}
	} else if !keyfileYes {
		confirm, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).Show("Generate MongoDB keyfile?")
		if !confirm {
			pterm.Warning.Println("Operation cancelled")
			return
		}
	}

	generateFileWithDirCheck(keyfilePath, keyfileFilename)
}

func generateFileWithDirCheck(path, filename string) {
	fullPath := filepath.Join(path, filename)

	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		pterm.Error.Printf("Failed to create directory: %v\n", err)
		return
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); err == nil {
		overwrite, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).Show(fmt.Sprintf("File %s already exists. Overwrite?", fullPath))
		if !overwrite {
			pterm.Warning.Println("Operation cancelled")
			return
		}
	}

	// Generate key (MongoDB limits keyfiles to 1024 characters, so we use 768 raw bytes)
	key := make([]byte, 768)
	_, err := rand.Read(key)
	if err != nil {
		pterm.Error.Printf("Failed to generate key: %v\n", err)
		return
	}

	encoded := base64.StdEncoding.EncodeToString(key)

	// Write file
	if err := os.WriteFile(fullPath, []byte(encoded), 0400); err != nil {
		pterm.Error.Printf("Failed to write keyfile: %v\n", err)
		return
	}

	pterm.Success.Printf("Keyfile generated at %s\n", fullPath)
}
