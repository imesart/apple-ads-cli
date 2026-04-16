package ads

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api/requests/ads"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	campID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	agID := fs.String("adgroup-id", "", "Ad Group ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	nameFlag := fs.String("name", "", "Ad name (required)")
	creativeID := fs.String("creative-id", "", "Creative ID (required)")
	statusFlag := fs.String("status", "", "ENABLED (default) | PAUSED (also 1/0, enable/pause)")
	metadataJSON := fs.String("metadata-json", "", "Inline creative metadata JSON")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aads ads create --campaign-id CID --adgroup-id AGID --name NAME --creative-id ID [flags]",
		ShortHelp:  "Create an ad.",
		LongHelp: `Create an ad using shortcut flags or --from-json for full JSON.

JSON keys (for --from-json):
  creativeId    integer  (required) Creative ID to use for this ad
  name          string   (required) Ad name
  status        string   ENABLED (default) | PAUSED
  creativeType  string   CUSTOM_PRODUCT_PAGE | CREATIVE_SET | DEFAULT_PRODUCT_PAGE

Examples:
  aads ads create --campaign-id 1 --adgroup-id 2 --name "My Ad" --creative-id 123456
  aads ads create --campaign-id 1 --adgroup-id 2 --name "My Ad" --creative-id 123456 --status 0
  aads ads create --campaign-id 1 --adgroup-id 2 --from-json ad.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campID},
				shared.StdinFlag{Name: "adgroup-id", Ptr: agID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			execOnce := func() (any, error) {
				cid := strings.TrimSpace(*campID)
				if cid == "" {
					return nil, shared.UsageError("--campaign-id is required")
				}
				agid := strings.TrimSpace(*agID)
				if agid == "" {
					return nil, shared.UsageError("--adgroup-id is required")
				}

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
					if *nameFlag == "" {
						return nil, shared.UsageError("--name is required")
					}
					if *creativeID == "" {
						return nil, shared.UsageError("--creative-id is required")
					}

					m := map[string]any{
						"name":       *nameFlag,
						"creativeId": json.Number(*creativeID),
					}
					if *statusFlag != "" {
						s, err := shared.NormalizeStatus(*statusFlag, "ENABLED")
						if err != nil {
							return nil, err
						}
						m["status"] = s
					}
					if *metadataJSON != "" {
						var meta json.RawMessage
						if err := json.Unmarshal([]byte(*metadataJSON), &meta); err != nil {
							return nil, fmt.Errorf("invalid --metadata-json: %w", err)
						}
						m["creativeMetadata"] = meta
					}

					body, err = json.Marshal(m)
					if err != nil {
						return nil, fmt.Errorf("create: marshalling body: %w", err)
					}
				}
				if *check {
					return shared.NewMutationCheckSummary("create", "ad", shared.FormatTarget("campaign-id", cid, "adgroup-id", agid), body, shared.MutationCheckOptions{}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, ads.CreateRequest{
					CampaignID: cid,
					AdGroupID:  agid,
					RawBody:    body,
				}, &result)
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
				}
				return result, nil
			}

			if len(stdinFlags) > 0 {
				return shared.RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, "ADID")
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "ADID")
		},
	}
}
