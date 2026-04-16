package profiles

import (
	"context"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

func setDefaultCmd() *ffcli.Command {
	return &ffcli.Command{
		Name:       "set-default",
		ShortUsage: "aads profiles set-default <name>",
		ShortHelp:  "Set the default profile.",
		LongHelp: `Set which profile is used by default when --profile is not specified.

Example:
  aads profiles set-default work`,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return shared.UsageError("exactly one profile name is required")
			}
			profileName := args[0]

			cf := config.LoadFile()

			if _, exists := cf.Profiles[profileName]; !exists {
				return shared.ReportError(fmt.Errorf("profile %q not found", profileName))
			}

			cf.DefaultProfile = profileName

			if err := config.SaveFile(cf); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Default profile set to %q\n", profileName)
			return nil
		},
	}
}
