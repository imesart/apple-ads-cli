package adgroups

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

// Command returns the adgroups command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "adgroups",
		ShortUsage: "aads adgroups <subcommand>",
		ShortHelp:  "Manage ad groups.",
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

var campaignParent = []shared.ParentFlag{
	{Name: "campaign-id", Usage: "Campaign ID", Required: true},
}
