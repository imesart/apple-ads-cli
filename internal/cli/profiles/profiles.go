package profiles

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// Command returns the profiles command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "profiles",
		ShortUsage: "aads profiles <subcommand>",
		ShortHelp:  "Manage configuration profiles.",
		Subcommands: []*ffcli.Command{
			listCmd(),
			getCmd(),
			genkeyCmd(),
			createCmd(),
			updateCmd(),
			deleteCmd(),
			setDefaultCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
