package acls

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/acls"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

// Command returns the orgs command group backed by Apple Ads ACL endpoints.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "orgs",
		ShortUsage: "aads orgs <subcommand>",
		ShortHelp:  "Manage organizations and user context (Apple Ads ACLs).",
		Subcommands: []*ffcli.Command{
			listCmd(),
			userCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func listCmd() *ffcli.Command {
	return shared.BuildListCommand(shared.ListCommandConfig{
		Name:             "list",
		ShortUsage:       "aads orgs list",
		ShortHelp:        "List organizations (Apple Ads ACLs).",
		EnablePagination: false,
		EnableLocalSort:  true,
		Exec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, acls.ListRequest{}, &result)
			return result, err
		},
	})
}

func userCmd() *ffcli.Command {
	fs := flag.NewFlagSet("user", flag.ContinueOnError)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "user",
		ShortUsage: "aads orgs user",
		ShortHelp:  "Get current user details from Apple Ads ACL context.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("user: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			var result json.RawMessage
			err = client.Do(ctx, acls.MeRequest{}, &result)
			if err != nil {
				return fmt.Errorf("user: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}
