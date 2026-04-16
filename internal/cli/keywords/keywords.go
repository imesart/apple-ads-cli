package keywords

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

// Command returns the keywords command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "keywords",
		ShortUsage: "aads keywords <subcommand>",
		ShortHelp:  "Manage targeting keywords.",
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

var campaignAdGroupParent = []shared.ParentFlag{
	{Name: "campaign-id", Usage: "Campaign ID", Required: true},
	{Name: "adgroup-id", Usage: "Ad Group ID", Required: true},
}
