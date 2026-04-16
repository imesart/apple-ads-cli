package structure

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// Command returns the structure command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "structure",
		ShortUsage: "aads structure <subcommand>",
		ShortHelp:  "Export and import campaign/ad group structures as JSON.",
		Subcommands: []*ffcli.Command{
			exportCmd(),
			importCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
