package adrejections

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	adrejections "github.com/imesart/apple-ads-cli/internal/api/requests/ad_rejections"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

// Command returns the ad-rejections command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "ad-rejections",
		ShortUsage: "aads ad-rejections <subcommand>",
		ShortHelp:  "Manage ad creative rejection reasons.",
		Subcommands: []*ffcli.Command{
			listCmd(),
			getCmd(),
			assetsCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func listCmd() *ffcli.Command {
	return shared.BuildSmartListCommand(shared.SmartListCommandConfig{
		Name:         "list",
		ShortUsage:   "aads ad-rejections list [flags]",
		ShortHelp:    "List ad creative rejection reasons.",
		EntityIDName: "ID",
		LongHelp: `List ad creative rejection reasons, with optional filtering.

Filter examples (repeatable):
  --filter "adamId=900001"
  --filter "countryOrRegion=US"
  --filter "reasonLevel IN [CUSTOM_PRODUCT_PAGE, DEFAULT_PRODUCT_PAGE]"
  --sort "id:desc"

Filter operators: EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, adamId, productPageId, assetGenId, languageCode, reasonCode,
  reasonType, reasonLevel, supplySource, countryOrRegion

Advanced: use --selector for inline JSON.`,
		FindExec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, selector json.RawMessage) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, adrejections.FindRequest{RawBody: selector}, &result)
			return result, err
		},
	})
}

func getCmd() *ffcli.Command {
	return shared.BuildIDGetCommand(shared.IDGetCommandConfig{
		Name:       "get",
		ShortUsage: "aads ad-rejections get --reason-id ID",
		ShortHelp:  "Get rejection reasons by product page reason ID.",
		IDFlag:     "reason-id",
		IDUsage:    "Product Page Reason ID",
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, adrejections.GetRequest{ID: id}, &result)
			return result, err
		},
	})
}

func assetsCmd() *ffcli.Command {
	fs := flag.NewFlagSet("assets", flag.ContinueOnError)
	adamID := fs.String("adam-id", "", "App Adam ID")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "assets",
		ShortUsage: "aads ad-rejections assets --adam-id ID",
		ShortHelp:  "List app assets.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			aid := strings.TrimSpace(*adamID)
			if aid == "" {
				return shared.UsageErrorf("--adam-id is required")
			}

			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("assets: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			var result json.RawMessage
			err = client.Do(ctx, adrejections.FindAssetsRequest{AdamID: aid}, &result)
			if err != nil {
				return fmt.Errorf("assets: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}
