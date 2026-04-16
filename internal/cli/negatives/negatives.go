package negatives

import (
	"context"
	"flag"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// Command returns the negatives command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "negatives",
		ShortUsage: "aads negatives <subcommand>",
		ShortHelp:  "Manage negative keywords (campaign and ad group level).",
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

// isAdGroupLevel returns true when --adgroup-id was provided (non-empty after trim).
func isAdGroupLevel(adgroupID string) bool {
	return strings.TrimSpace(adgroupID) != ""
}
