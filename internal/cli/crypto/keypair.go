package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	keypairFilename        string
	keypairPath            string
	keypairAlgorithm       string
	keypairYes             bool
	keypairGeneratePublic  bool
	keypairGeneratePrivate bool
	keypairPrivateKeyPath  string
)

var KeypairCmd = &cobra.Command{
	Use:   "keypair",
	Short: "Generate a public/private key pair",
	Long: `Generate a public/private key pair using RSA, ECDSA, or Ed25519 algorithms.
Keys are saved in PEM format.
Private keys are saved with 0600 permissions.
Public keys are saved with 0644 permissions.`,
	Example: `  # Generate RSA key pair (default)
  magi crypto keypair

  # Generate ECDSA key pair
  magi crypto keypair --algorithm ecdsa

  # Generate Ed25519 key pair
  magi crypto keypair --algorithm ed25519

  # Generate only public key from existing private key
  magi crypto keypair --public --private-key-path ./private.pem`,
	Run: runGenerateKeypair,
}

func init() {
	KeypairCmd.Flags().StringVarP(&keypairFilename, "filename", "f", "key", "Key filename (no extension)")
	KeypairCmd.Flags().StringVarP(&keypairPath, "path", "p", ".", "Output directory")
	KeypairCmd.Flags().StringVarP(&keypairAlgorithm, "algorithm", "a", "rsa", "Key algorithm (rsa, ecdsa, ed25519)")
	KeypairCmd.Flags().BoolVarP(&keypairYes, "yes", "y", false, "Skip prompts")
	KeypairCmd.Flags().BoolVar(&keypairGeneratePublic, "public", false, "Generate only public key")
	KeypairCmd.Flags().BoolVar(&keypairGeneratePrivate, "private", false, "Generate only private key")
	KeypairCmd.Flags().StringVar(&keypairPrivateKeyPath, "private-key-path", "", "Existing private key path (for public key generation)")
}

func runGenerateKeypair(cmd *cobra.Command, args []string) {
	// Validate flags
	if keypairGeneratePublic && keypairGeneratePrivate {
		pterm.Error.Println("Cannot specify both --public and --private")
		return
	}

	if keypairGeneratePublic && keypairPrivateKeyPath == "" {
		pterm.Error.Println("Must specify --private-key-path when using --public")
		return
	}

	// Interactive prompts if not skipped
	if !keypairYes && !keypairGeneratePublic {
		var err error
		keypairAlgorithm, err = pterm.DefaultInteractiveSelect.
			WithOptions([]string{"rsa", "ecdsa", "ed25519"}).
			WithDefaultOption(keypairAlgorithm).
			Show("Select key algorithm")
		if err != nil {
			return
		}

		keypairFilename, err = pterm.DefaultInteractiveTextInput.
			WithDefaultText(keypairFilename).
			Show("Enter key filename (no extension)")
		if err != nil {
			return
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(keypairPath, 0755); err != nil {
		pterm.Error.Printf("Failed to create directory: %v\n", err)
		return
	}

	privateKeyPath := filepath.Join(keypairPath, keypairFilename+".pem")
	publicKeyPath := filepath.Join(keypairPath, keypairFilename+".pub")

	if keypairGeneratePublic {
		// Generate public key from existing private key
		if err := generatePublicKeyFromPrivate(keypairPrivateKeyPath, publicKeyPath); err != nil {
			pterm.Error.Printf("Failed to generate public key: %v\n", err)
			return
		}
		pterm.Success.Printf("Public key generated at %s\n", publicKeyPath)
		return
	}

	// Generate new key pair
	priv, pub, err := generateKeys(keypairAlgorithm)
	if err != nil {
		pterm.Error.Printf("Failed to generate keys: %v\n", err)
		return
	}

	// Save private key
	if !keypairGeneratePublic {
		if err := savePrivateKey(priv, privateKeyPath); err != nil {
			pterm.Error.Printf("Failed to save private key: %v\n", err)
			return
		}
		pterm.Success.Printf("Private key generated at %s\n", privateKeyPath)
	}

	// Save public key
	if !keypairGeneratePrivate {
		if err := savePublicKey(pub, publicKeyPath); err != nil {
			pterm.Error.Printf("Failed to save public key: %v\n", err)
			return
		}
		pterm.Success.Printf("Public key generated at %s\n", publicKeyPath)
	}
}

func generateKeys(algo string) (interface{}, interface{}, error) {
	switch strings.ToLower(algo) {
	case "rsa":
		priv, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, nil, err
		}
		return priv, &priv.PublicKey, nil
	case "ecdsa":
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		return priv, &priv.PublicKey, nil
	case "ed25519":
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		return priv, pub, nil
	default:
		return nil, nil, fmt.Errorf("unsupported algorithm: %s", algo)
	}
}

func savePrivateKey(key interface{}, path string) error {
	var bytes []byte
	var err error
	var block *pem.Block

	switch k := key.(type) {
	case *rsa.PrivateKey:
		bytes = x509.MarshalPKCS1PrivateKey(k)
		block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: bytes}
	case *ecdsa.PrivateKey:
		bytes, err = x509.MarshalECPrivateKey(k)
		if err != nil {
			return err
		}
		block = &pem.Block{Type: "EC PRIVATE KEY", Bytes: bytes}
	case ed25519.PrivateKey:
		bytes, err = x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			return err
		}
		block = &pem.Block{Type: "PRIVATE KEY", Bytes: bytes}
	default:
		return fmt.Errorf("unsupported private key type")
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

func savePublicKey(key interface{}, path string) error {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return err
	}

	block := &pem.Block{Type: "PUBLIC KEY", Bytes: bytes}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

func generatePublicKeyFromPrivate(privatePath, publicPath string) error {
	data, err := os.ReadFile(privatePath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block containing private key")
	}

	var pub interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
		pub = &priv.PublicKey
	case "EC PRIVATE KEY":
		priv, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return err
		}
		pub = &priv.PublicKey
	case "PRIVATE KEY":
		priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
		switch k := priv.(type) {
		case ed25519.PrivateKey:
			pub = k.Public()
		default:
			return fmt.Errorf("unsupported private key type in PKCS8")
		}
	default:
		return fmt.Errorf("unsupported private key type: %s", block.Type)
	}

	return savePublicKey(pub, publicPath)
}
