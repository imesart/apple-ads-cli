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

func deleteCmd() *ffcli.Command {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID; omit for campaign-level (or - to read IDs from stdin)")
	keywordID := fs.String("keyword-id", "", "Negative Keyword ID, or comma-separated negative keyword IDs (or - to read IDs from stdin)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	confirm := fs.Bool("confirm", false, "Confirm deletion")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "aads negatives delete --campaign-id CID [--adgroup-id AGID] (--keyword-id ID | --from-json FILE) --confirm",
		ShortHelp:  "Delete negative keywords.",
		LongHelp: `Delete negative keywords by ID or from a JSON array of keyword IDs.

Without --adgroup-id, deletes campaign-level negative keywords.
With --adgroup-id, deletes ad group-level negative keywords.
Use --keyword-id for one ID or a comma-separated list of IDs.
Use --from-json for inline JSON, @file.json, or @- for stdin. The body is a JSON array of keyword IDs to delete.

Requires --confirm to execute. Without --confirm, the command prints a check summary and exits with an error.`,
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

				hasKeywordID := strings.TrimSpace(*keywordID) != ""
				hasJSON := strings.TrimSpace(*dataFile) != ""
				if hasKeywordID && hasJSON {
					return nil, shared.UsageError("only one of --keyword-id or --from-json may be used")
				}
				if !hasKeywordID && !hasJSON {
					return nil, shared.UsageError("--keyword-id or --from-json is required")
				}

				var body json.RawMessage
				var ids []string
				if hasKeywordID {
					var err error
					ids, err = shared.ParseIDList(*keywordID)
					if err != nil {
						return nil, shared.UsageError(err.Error())
					}
					body, err = json.Marshal(ids)
					if err != nil {
						return nil, fmt.Errorf("delete: marshalling body: %w", err)
					}
				} else {
					var err error
					body, err = readBodyFile(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("delete: reading body: %w", err)
					}
				}
				target := shared.FormatTarget("campaign-id", cid)
				if isAdGroupLevel(*adgroupID) {
					target = shared.FormatTarget("campaign-id", cid, "adgroup-id", strings.TrimSpace(*adgroupID))
				}
				if hasKeywordID && len(ids) == 1 {
					if isAdGroupLevel(*adgroupID) {
						target = shared.FormatTarget("campaign-id", cid, "adgroup-id", strings.TrimSpace(*adgroupID), "keyword-id", ids[0])
					} else {
						target = shared.FormatTarget("campaign-id", cid, "keyword-id", ids[0])
					}
				}
				summary := shared.NewMutationCheckSummary("delete", "negative keyword", target, body, shared.MutationCheckOptions{})
				if *check {
					return summary, nil
				}
				if !*confirm {
					if err := shared.PrintOutput(summary, *output.Output, *output.Fields, *output.Pretty, "KEYWORDID"); err != nil {
						return nil, err
					}
					return nil, shared.UsageError("--confirm is required for deletion")
				}

				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("delete: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				if isAdGroupLevel(*adgroupID) {
					var result json.RawMessage
					err := client.Do(ctx, negadgroup.DeleteBulkRequest{
						CampaignID: cid,
						AdGroupID:  strings.TrimSpace(*adgroupID),
						RawBody:    body,
					}, &result)
					return result, err
				}
				var result json.RawMessage
				err = client.Do(ctx, negcampaign.DeleteBulkRequest{
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
