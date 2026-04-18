package profiles

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

func genkeyCmd() *ffcli.Command {
	fs := flag.NewFlagSet("profiles genkey", flag.ContinueOnError)
	name := fs.String("name", "", "Profile name or key name (required)")
	confirm := fs.Bool("confirm", false, "Overwrite the existing private key file")

	return &ffcli.Command{
		Name:       "genkey",
		ShortUsage: "aads profiles genkey --name NAME [--confirm]",
		ShortHelp:  "Generate a P-256 private key and print its public key.",
		LongHelp: `Generate a P-256 (ES256) private key using openssl and print the
corresponding public key to stdout. The private key path is always:

  ~/.aads/keys/NAME-private-key.pem

If NAME matches an existing profile, the command updates that profile's
private_key_path after successful generation.

Example:
  aads profiles genkey --name default
  aads profiles genkey --name work --confirm`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			keyName := strings.TrimSpace(*name)
			if keyName == "" {
				return shared.UsageError("--name is required")
			}

			targetPath := defaultPrivateKeyPath(keyName)
			cf := config.LoadFile()

			if err := warnOnProfileKeyReplacement(cf, keyName, targetPath); err != nil {
				return err
			}

			if _, err := os.Stat(targetPath); err == nil {
				publicKey, pubErr := publicKeyFromPrivateKey(targetPath)
				if pubErr == nil {
					fmt.Fprint(os.Stdout, publicKey)
				}
				if !*confirm {
					return shared.ReportError(fmt.Errorf("private key file %q already exists; rerun with --confirm to overwrite", targetPath))
				}
			} else if !os.IsNotExist(err) {
				return shared.ReportError(fmt.Errorf("checking private key path %q: %w", targetPath, err))
			}

			if err := generatePrivateKey(ctx, targetPath); err != nil {
				return shared.ReportError(err)
			}

			publicKey, err := publicKeyFromPrivateKey(targetPath)
			if err != nil {
				return shared.ReportError(err)
			}
			fmt.Fprint(os.Stdout, publicKey)

			if profile, ok := cf.Profiles[keyName]; ok {
				profile.PrivateKeyPath = targetPath
				cf.Profiles[keyName] = profile
				if err := config.SaveFile(cf); err != nil {
					return shared.ReportError(fmt.Errorf("saving config: %w", err))
				}
				fmt.Fprintf(os.Stderr, "Profile %q updated in %s\n", keyName, config.DefaultConfigPath())
			}
			return nil
		},
	}
}

func warnOnProfileKeyReplacement(cf *config.ConfigFile, profileName, targetPath string) error {
	profile, ok := cf.Profiles[profileName]
	if !ok {
		return nil
	}

	currentPath := expandUserPath(strings.TrimSpace(profile.PrivateKeyPath))
	if currentPath == "" || currentPath == targetPath {
		return nil
	}
	if _, err := os.Stat(currentPath); err == nil {
		fmt.Fprintf(os.Stderr, "Warning: profile %q currently points to existing key file %q; updating it to %q\n", profileName, currentPath, targetPath)
		return nil
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return shared.ReportError(fmt.Errorf("checking profile private key path %q: %w", currentPath, err))
	}
}
