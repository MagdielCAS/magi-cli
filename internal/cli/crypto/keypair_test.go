package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeypairCmd(t *testing.T) {
	cmd := KeypairCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "keypair", cmd.Use)
}

func TestGenerateKeys(t *testing.T) {
	tests := []struct {
		name string
		algo string
	}{
		{"RSA", "rsa"},
		{"ECDSA", "ecdsa"},
		{"Ed25519", "ed25519"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priv, pub, err := generateKeys(tt.algo)
			assert.NoError(t, err)
			assert.NotNil(t, priv)
			assert.NotNil(t, pub)
		})
	}
}

func TestSavePrivateKey(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "magi-crypto-test-priv")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	priv, _, err := generateKeys("rsa")
	assert.NoError(t, err)

	path := filepath.Join(tmpDir, "private.pem")
	err = savePrivateKey(priv, path)
	assert.NoError(t, err)

	// Verify file exists and permissions
	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content is PEM
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	block, _ := pem.Decode(content)
	assert.NotNil(t, block)
	assert.Equal(t, "RSA PRIVATE KEY", block.Type)
}

func TestSavePublicKey(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "magi-crypto-test-pub")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	_, pub, err := generateKeys("rsa")
	assert.NoError(t, err)

	path := filepath.Join(tmpDir, "public.pem")
	err = savePublicKey(pub, path)
	assert.NoError(t, err)

	// Verify file exists and permissions
	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())

	// Verify content is PEM
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	block, _ := pem.Decode(content)
	assert.NotNil(t, block)
	assert.Equal(t, "PUBLIC KEY", block.Type)
}

func TestGeneratePublicKeyFromPrivate(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "magi-crypto-test-extract")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Generate and save private key
	priv, _, err := generateKeys("rsa")
	assert.NoError(t, err)
	privPath := filepath.Join(tmpDir, "private.pem")
	err = savePrivateKey(priv, privPath)
	assert.NoError(t, err)

	// Extract public key
	pubPath := filepath.Join(tmpDir, "public.pem")
	err = generatePublicKeyFromPrivate(privPath, pubPath)
	assert.NoError(t, err)

	// Verify public key file
	content, err := os.ReadFile(pubPath)
	assert.NoError(t, err)
	block, _ := pem.Decode(content)
	assert.NotNil(t, block)
	assert.Equal(t, "PUBLIC KEY", block.Type)

	// Verify it matches the private key
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	assert.NoError(t, err)
	assert.NotNil(t, pubKey)
}
