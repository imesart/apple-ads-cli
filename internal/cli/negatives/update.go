package negatives

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	negadgroup "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_adgroup"
	negcampaign "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_campaign"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func updateCmd() *ffcli.Command {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID; omit for campaign-level (or - to read IDs from stdin)")
	keywordID := fs.String("keyword-id", "", "Negative Keyword ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	status := fs.String("status", "", "ACTIVE | PAUSED (also 1/0, enable/pause)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "aads negatives update --campaign-id CID [--adgroup-id AGID] --keyword-id ID [flags]",
		ShortHelp:  "Update negative keywords.",
		LongHelp: `Update negative keywords.

Without --adgroup-id, updates campaign-level negative keywords.
With --adgroup-id, updates ad group-level negative keywords.

Use shortcut flags for quick status changes, or --from-json for arbitrary JSON array updates.

JSON keys per negative keyword (all optional except id):
  id         integer  (required) Negative keyword ID to update
  text       string   Keyword text
  matchType  string   BROAD | EXACT
  status     string   ACTIVE | PAUSED

Examples:
  aads negatives update --campaign-id 1 --keyword-id 2 --status 0
  aads negatives update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --status PAUSED
  aads negatives update --campaign-id 1 --keyword-id 2 --from-json updates.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
				shared.StdinFlag{Name: "adgroup-id", Ptr: adgroupID},
				shared.StdinFlag{Name: "keyword-id", Ptr: keywordID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			execOnce := func() (any, error) {
				cid := strings.TrimSpace(*campaignID)
				if cid == "" {
					return nil, shared.UsageError("--campaign-id is required")
				}
				kwid := strings.TrimSpace(*keywordID)
				if kwid == "" {
					return nil, shared.UsageError("--keyword-id is required")
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
					s, err := shared.NormalizeStatus(*status, "ACTIVE")
					if err != nil {
						return nil, err
					}
					body, err = json.Marshal([]any{map[string]any{"id": kwid, "status": s}})
					if err != nil {
						return nil, fmt.Errorf("update: marshalling body: %w", err)
					}
				} else {
					return nil, shared.UsageError("--from-json or --status is required")
				}
				target := shared.FormatTarget("campaign-id", cid, "keyword-id", kwid)
				if isAdGroupLevel(*adgroupID) {
					target = shared.FormatTarget("campaign-id", cid, "adgroup-id", strings.TrimSpace(*adgroupID), "keyword-id", kwid)
				}
				if *check {
					return shared.NewMutationCheckSummary("update", "negative keyword", target, body, shared.MutationCheckOptions{}), nil
				}

				if isAdGroupLevel(*adgroupID) {
					var result json.RawMessage
					err := client.Do(ctx, negadgroup.UpdateRequest{
						CampaignID: cid,
						AdGroupID:  strings.TrimSpace(*adgroupID),
						RawBody:    body,
					}, &result)
					return result, err
				}
				var result json.RawMessage
				err = client.Do(ctx, negcampaign.UpdateRequest{
					CampaignID: cid,
					RawBody:    body,
				}, &result)
				return result, err
			}

			if len(stdinFlags) > 0 {
				return shared.RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, "KEYWORDID")
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "KEYWORDID")
		},
	}
}
