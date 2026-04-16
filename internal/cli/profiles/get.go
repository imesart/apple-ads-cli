package profiles

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

func getCmd() *ffcli.Command {
	fs := flag.NewFlagSet("profiles get", flag.ContinueOnError)
	name := fs.String("name", "", "Profile name (default current default profile)")
	showCreds := fs.Bool("show-credentials", false, "Show client ID, team ID, key ID, and private key path")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "aads profiles get [--name NAME] [--show-credentials]",
		ShortHelp:  "Show profile details.",
		LongHelp: `Show details for a profile. Without --name, shows the current default profile.

Example:
  aads profiles get
  aads profiles get --name work`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			cf := config.LoadFile()

			if len(cf.Profiles) == 0 {
				return shared.ReportError(fmt.Errorf("no profiles configured; use 'aads profiles create' to add one"))
			}

			profileName := *name
			if profileName == "" {
				profileName = cf.DefaultProfile
			}
			if profileName == "" {
				profileName = "default"
			}

			p, ok := cf.Profiles[profileName]
			if !ok {
				return shared.ReportError(fmt.Errorf("profile %q not found", profileName))
			}

			row := profileRow(profileName, profileName == cf.DefaultProfile, &p, *showCreds)
			renderTable(profileHeaders(*showCreds), [][]string{row})
			return nil
		},
	}
}
