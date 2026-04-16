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

func getCmd() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID; omit for campaign-level (or - to read IDs from stdin)")
	keywordID := fs.String("keyword-id", "", "Negative Keyword ID (or - to read IDs from stdin) (required)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "aads negatives get --campaign-id CID [--adgroup-id AGID] --keyword-id ID",
		ShortHelp:  "Get a negative keyword.",
		LongHelp: `Get a negative keyword by ID.

Without --adgroup-id, fetches a campaign-level negative keyword.
With --adgroup-id, fetches an ad group-level negative keyword.`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
				shared.StdinFlag{Name: "adgroup-id", Ptr: adgroupID},
				shared.StdinFlag{Name: "keyword-id", Ptr: keywordID},
			)

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
					return nil, fmt.Errorf("get: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				if isAdGroupLevel(*adgroupID) {
					var result json.RawMessage
					err := client.Do(ctx, negadgroup.GetRequest{
						CampaignID: cid,
						AdGroupID:  strings.TrimSpace(*adgroupID),
						KeywordID:  kwid,
					}, &result)
					return result, err
				}
				var result json.RawMessage
				err = client.Do(ctx, negcampaign.GetRequest{
					CampaignID: cid,
					KeywordID:  kwid,
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
