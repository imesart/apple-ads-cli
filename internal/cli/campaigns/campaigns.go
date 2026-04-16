package campaigns

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// Command returns the campaigns command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "campaigns",
		ShortUsage: "aads campaigns <subcommand>",
		ShortHelp:  "Manage campaigns.",
		Subcommands: []*ffcli.Command{
			listCmd(),
			getCmd(),
			createCmd(),
			updateCmd(),
			deleteCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
