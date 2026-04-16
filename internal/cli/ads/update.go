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

func updateCmd() *ffcli.Command {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID (or - to read IDs from stdin) (required)")
	adID := fs.String("ad-id", "", "Ad ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	status := fs.String("status", "", "ENABLED | PAUSED (also 1/0, enable/pause)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "aads ads update --campaign-id CID --adgroup-id AGID --ad-id ID [flags]",
		ShortHelp:  "Update an ad.",
		LongHelp: `Update an ad. The API accepts partial updates.

Use shortcut flags for quick status changes, or --from-json for arbitrary JSON.

JSON keys (all optional):
  name          string   Ad name
  status        string   ENABLED | PAUSED
  creativeId    integer  Creative ID
  creativeType  string   CUSTOM_PRODUCT_PAGE | CREATIVE_SET | DEFAULT_PRODUCT_PAGE

Examples:
  aads ads update --campaign-id 1 --adgroup-id 2 --ad-id 3 --status 0
  aads ads update --campaign-id 1 --adgroup-id 2 --ad-id 3 --status PAUSED
  aads ads update --campaign-id 1 --adgroup-id 2 --ad-id 3 --from-json changes.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
				shared.StdinFlag{Name: "adgroup-id", Ptr: adgroupID},
				shared.StdinFlag{Name: "ad-id", Ptr: adID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			execOnce := func() (any, error) {
				cid := strings.TrimSpace(*campaignID)
				if cid == "" {
					return nil, shared.UsageError("--campaign-id is required")
				}
				agid := strings.TrimSpace(*adgroupID)
				if agid == "" {
					return nil, shared.UsageError("--adgroup-id is required")
				}
				aid := strings.TrimSpace(*adID)
				if aid == "" {
					return nil, shared.UsageError("--ad-id is required")
				}

				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				var body json.RawMessage
				if *dataFile != "" {
					body, err = readBodyFile(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("update: reading body: %w", err)
					}
				} else if *status != "" {
					s, err := shared.NormalizeStatus(*status, "ENABLED")
					if err != nil {
						return nil, err
					}
					body, err = json.Marshal(map[string]any{"status": s})
					if err != nil {
						return nil, fmt.Errorf("update: marshalling body: %w", err)
					}
				} else {
					return nil, shared.UsageError("--from-json or --status is required")
				}
				if *check {
					return shared.NewMutationCheckSummary("update", "ad", shared.FormatTarget("campaign-id", cid, "adgroup-id", agid, "ad-id", aid), body, shared.MutationCheckOptions{}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, ads.UpdateRequest{
					CampaignID: cid,
					AdGroupID:  agid,
					AdID:       aid,
					RawBody:    body,
				}, &result)
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
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
