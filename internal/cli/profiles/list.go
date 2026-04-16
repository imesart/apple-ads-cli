package profiles

import (
	"context"
	"flag"
	"fmt"
	"sort"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/config"
)

func listCmd() *ffcli.Command {
	fs := flag.NewFlagSet("profiles list", flag.ContinueOnError)
	showCreds := fs.Bool("show-credentials", false, "Show client ID, team ID, key ID, and private key path")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "aads profiles list [--show-credentials]",
		ShortHelp:  "List all configured profiles.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			cf := config.LoadFile()

			if len(cf.Profiles) == 0 {
				fmt.Println("No profiles configured. Use 'aads profiles create' to add one.")
				return nil
			}

			names := make([]string, 0, len(cf.Profiles))
			for name := range cf.Profiles {
				names = append(names, name)
			}
			sort.Strings(names)

			var rows [][]string
			for _, name := range names {
				p := cf.Profiles[name]
				rows = append(rows, profileRow(name, name == cf.DefaultProfile, &p, *showCreds))
			}
			renderTable(profileHeaders(*showCreds), rows)
			return nil
		},
	}
}
