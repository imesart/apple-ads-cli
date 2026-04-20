package profiles

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

func deleteCmd() *ffcli.Command {
	fs := flag.NewFlagSet("profiles delete", flag.ContinueOnError)

	name := fs.String("name", "", "Profile name to delete (required)")
	confirm := fs.Bool("confirm", false, "Confirm deletion")
	deletePrivateKey := fs.Bool("delete-private-key", false, "Delete the configured private key file after deleting the profile")
	check := fs.Bool("check", false, "Validate and summarize without writing config")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "aads profiles delete --name NAME --confirm [--delete-private-key]",
		ShortHelp:  "Delete a profile.",
		LongHelp: `Delete a configuration profile. Requires --confirm.
If the deleted profile was the default, the default is cleared.
Use --delete-private-key to also remove the configured private key file.

Example:
  aads profiles delete --name work --confirm
  aads profiles delete --name work --confirm --delete-private-key`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			profileName := strings.TrimSpace(*name)
			if profileName == "" {
				return shared.UsageError("--name is required")
			}

			cf := config.LoadFile()

			profile, exists := cf.Profiles[profileName]
			if !exists {
				return shared.ReportError(fmt.Errorf("profile %q not found", profileName))
			}
			summaryFields := map[string]any{"name": profileName}
			if *deletePrivateKey {
				summaryFields["deletePrivateKey"] = true
				if strings.TrimSpace(profile.PrivateKeyPath) != "" {
					summaryFields["privateKeyPath"] = strings.TrimSpace(profile.PrivateKeyPath)
				}
			}
			body, err := json.Marshal(summaryFields)
			if err != nil {
				return err
			}
			summary := shared.NewMutationCheckSummary("delete", "profile", shared.FormatTarget("name", profileName), body, shared.MutationCheckOptions{
				Count: 1,
			})
			if *check {
				return shared.PrintOutput(summary, *output.Output, *output.Fields, *output.Pretty)
			}
			if !*confirm {
				if err := shared.PrintOutput(summary, *output.Output, *output.Fields, *output.Pretty); err != nil {
					return err
				}
				return shared.UsageError("--confirm is required to delete a profile")
			}

			delete(cf.Profiles, profileName)

			if cf.DefaultProfile == profileName {
				cf.DefaultProfile = ""
				// If there's exactly one profile left, make it the default
				if len(cf.Profiles) == 1 {
					for remaining := range cf.Profiles {
						cf.DefaultProfile = remaining
					}
				}
			}

			if err := config.SaveFile(cf); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Profile %q deleted from %s\n", profileName, config.DefaultConfigPath())
			if *deletePrivateKey {
				privateKeyPath := strings.TrimSpace(profile.PrivateKeyPath)
				if privateKeyPath == "" {
					fmt.Fprintln(os.Stderr, "Warning: profile had no private key path configured; no key file was deleted.")
				} else if err := os.Remove(expandUserPath(privateKeyPath)); err == nil {
					fmt.Fprintf(os.Stderr, "Private key file %q deleted.\n", privateKeyPath)
				} else if os.IsNotExist(err) {
					fmt.Fprintf(os.Stderr, "Warning: private key file %q does not exist; no key file was deleted.\n", privateKeyPath)
				} else {
					fmt.Fprintf(os.Stderr, "Warning: could not delete private key file %q: %v\n", privateKeyPath, err)
				}
			}
			if cf.DefaultProfile == "" && len(cf.Profiles) > 0 {
				fmt.Fprintf(os.Stderr, "Warning: no default profile set. Use 'aads profiles set-default <name>' to set one.\n")
			}
			return nil
		},
	}
}
