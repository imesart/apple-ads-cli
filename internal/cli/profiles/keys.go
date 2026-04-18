package profiles

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/config"
)

var (
	lookPathFunc = exec.LookPath
	runCommand   = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		return cmd.CombinedOutput()
	}
)

// SetKeygenFuncsForTesting overrides command execution helpers for tests.
func SetKeygenFuncsForTesting(
	lookPath func(string) (string, error),
	run func(context.Context, string, ...string) ([]byte, error),
) func() {
	prevLookPath := lookPathFunc
	prevRunCommand := runCommand
	if lookPath != nil {
		lookPathFunc = lookPath
	}
	if run != nil {
		runCommand = run
	}
	return func() {
		lookPathFunc = prevLookPath
		runCommand = prevRunCommand
	}
}

func defaultPrivateKeyPath(name string) string {
	return filepath.Join(config.DefaultConfigDir(), "keys", name+"-private-key.pem")
}

func expandUserPath(path string) string {
	if path == "" {
		return path
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

func generatePrivateKey(ctx context.Context, path string) error {
	if _, err := lookPathFunc("openssl"); err != nil {
		return fmt.Errorf("openssl is required for 'aads profiles genkey' but was not found in PATH; on macOS you can install it with Homebrew using 'brew install openssl'. Other installation methods also work")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating key directory: %w", err)
	}

	output, err := runCommand(ctx, "openssl", "ecparam", "-genkey", "-name", "prime256v1", "-noout", "-out", path)
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg != "" {
			return fmt.Errorf("generating private key with openssl: %s", msg)
		}
		return fmt.Errorf("generating private key with openssl: %w", err)
	}
	return nil
}

func publicKeyFromPrivateKey(path string) (string, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading private key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("parsing private key: %w", err)
		}
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not an ECDSA key")
	}

	publicDER, err := x509.MarshalPKIXPublicKey(&ecKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("encoding public key: %w", err)
	}

	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}); err != nil {
		return "", fmt.Errorf("encoding public key PEM: %w", err)
	}
	return buf.String(), nil
}

func createMissingPrivateKeyWarning(profileName, path string, explicit bool) string {
	if explicit {
		return fmt.Sprintf(
			"Warning: private key file %q does not exist. Generate a key with 'aads profiles genkey --name %s' to use the default location, or rerun with --private-key-path to point to a different PEM file.",
			path,
			profileName,
		)
	}
	return fmt.Sprintf(
		"Warning: private key file %q does not exist. This is the default key path for profile %q. Generate it with 'aads profiles genkey --name %s' or rerun with --private-key-path to use a different PEM file.",
		path,
		profileName,
		profileName,
	)
}
