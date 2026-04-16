package creatives

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/creatives"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

// Command returns the creatives command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "creatives",
		ShortUsage: "aads creatives <subcommand>",
		ShortHelp:  "Manage creatives.",
		Subcommands: []*ffcli.Command{
			listCmd(),
			getCmd(),
			createCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func listCmd() *ffcli.Command {
	return shared.BuildSmartListCommand(shared.SmartListCommandConfig{
		Name:         "list",
		ShortUsage:   "aads creatives list [flags]",
		ShortHelp:    "List creatives.",
		EntityIDName: "CREATIVEID",
		LongHelp: `List creatives, with optional filtering.

Filter examples (repeatable):
  --filter "type=CUSTOM_PRODUCT_PAGE"
  --filter "name CONTAINS holiday"
  --filter "state=VALID"
  --sort "name:asc"

Filter operators: EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

		Searchable and filterable fields:
  id, orgId, adamId, name, type, state

Advanced: use --selector for inline JSON.`,
		ListExec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error) {
			req := creatives.ListRequest{Limit: limit, Offset: offset}
			if limit == 0 {
				return api.FetchAllRaw(ctx, client, req)
			}
			var result json.RawMessage
			err := client.Do(ctx, req, &result)
			return result, err
		},
		FindExec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, selector json.RawMessage) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, creatives.FindRequest{RawBody: selector}, &result)
			return result, err
		},
	})
}

func getCmd() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	creativeID := fs.String("creative-id", "", "Creative ID (or - to read IDs from stdin) (required)")
	includeDeleted := fs.Bool("include-deleted-creative-set-assets", false, "Include deleted creative set assets")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "aads creatives get --creative-id ID",
		ShortHelp:  "Get a creative by ID.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(shared.StdinFlag{Name: "creative-id", Ptr: creativeID})
			execOnce := func() (any, error) {
				id := strings.TrimSpace(*creativeID)
				if id == "" {
					return nil, shared.UsageError("--creative-id is required")
				}
				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("get: %w", err)
				}
				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				var result json.RawMessage
				err = client.Do(ctx, creatives.GetRequest{
					CreativeID:                      id,
					IncludeDeletedCreativeSetAssets: *includeDeleted,
				}, &result)
				return result, err
			}
			if len(stdinFlags) > 0 {
				return shared.RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, "CREATIVEID")
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "CREATIVEID")
		},
	}
}

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	adamID := fs.String("adam-id", "", "App Store app ID (required)")
	productPageID := fs.String("product-page-id", "", "Product page ID")
	nameFlag := fs.String("name", "", "Creative name (required)")
	creativeType := fs.String("type", "", "CUSTOM_PRODUCT_PAGE | CREATIVE_SET | DEFAULT_PRODUCT_PAGE (required)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aads creatives create --adam-id ID --name NAME --type TYPE [flags]",
		ShortHelp:  "Create a creative.",
		LongHelp: `Create a creative using shortcut flags or --from-json for full JSON.

JSON keys (for --from-json):
  adamId         integer  (required) App Store app ID
  name           string   (required) Creative name
  type           string   (required) CUSTOM_PRODUCT_PAGE | CREATIVE_SET | DEFAULT_PRODUCT_PAGE
  productPageId  string   Product page ID (for CUSTOM_PRODUCT_PAGE type)

Examples:
  aads creatives create --adam-id 900001 --name "FitTrack Strength Page" --type CUSTOM_PRODUCT_PAGE
  aads creatives create --adam-id 900001 --name "FitTrack Strength Page" --type CUSTOM_PRODUCT_PAGE --product-page-id cpp-fitness-strength
  aads creatives create --from-json creative.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			execOnce := func() (any, error) {
				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				var body json.RawMessage
				if *dataFile != "" {
					body, err = readBodyFile(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("create: reading body: %w", err)
					}
				} else {
					if *adamID == "" {
						return nil, shared.UsageError("--adam-id is required")
					}
					if *nameFlag == "" {
						return nil, shared.UsageError("--name is required")
					}
					if *creativeType == "" {
						return nil, shared.UsageError("--type is required")
					}

					m := map[string]any{
						"adamId": json.Number(*adamID),
						"name":   *nameFlag,
						"type":   strings.ToUpper(*creativeType),
					}
					if *productPageID != "" {
						m["productPageId"] = *productPageID
					}

					body, err = json.Marshal(m)
					if err != nil {
						return nil, fmt.Errorf("create: marshalling body: %w", err)
					}
				}
				if *check {
					return shared.NewMutationCheckSummary("create", "creative", "", body, shared.MutationCheckOptions{}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, creatives.CreateRequest{RawBody: body}, &result)
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
				}
				return result, nil
			}

			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "CREATIVEID")
		},
	}
}

func readBodyFile(path string) (json.RawMessage, error) {
	return shared.ReadJSONInputArg(path)
}
