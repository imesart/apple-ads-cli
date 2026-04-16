package budgetorders

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// Command returns the budgetorders command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "budgetorders",
		ShortUsage: "aads budgetorders <subcommand>",
		ShortHelp:  "Manage budget orders.",
		Subcommands: []*ffcli.Command{
			listCmd(),
			getCmd(),
			createCmd(),
			updateCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
