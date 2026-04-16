package negatives

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	negadgroup "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_adgroup"
	negcampaign "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_campaign"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func listCmd() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID; omit for campaign-level (or - to read IDs from stdin)")
	limit := fs.Int("limit", 0, "Maximum results; 0 fetches all pages")
	offset := fs.Int("offset", 0, "Starting offset")
	sel := shared.BindSelectorFlags(fs)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "aads negatives list --campaign-id CID [--adgroup-id AGID] [flags]",
		ShortHelp:  "List negative keywords.",
		LongHelp: `List negative keywords, with optional filtering.

Without --adgroup-id, lists campaign-level negative keywords.
With --adgroup-id, lists ad group-level negative keywords.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, campaignId, adGroupId, text, matchType, status

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads negatives list --campaign-id 123
  aads negatives list --campaign-id 123 --adgroup-id 456
  aads negatives list --campaign-id 123 --filter "text CONTAINS free" --sort "text:asc"
  aads negatives list --campaign-id 123 --adgroup-id 456 --filter "matchType=EXACT"`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
				shared.StdinFlag{Name: "adgroup-id", Ptr: adgroupID},
			)

			execOnce := func() (any, error) {
				cid := strings.TrimSpace(*campaignID)
				if cid == "" {
					return nil, shared.UsageError("--campaign-id is required")
				}

				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("list: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				if sel.HasFlags() {
					findLimit := *limit
					if findLimit == 0 {
						findLimit = 1000
					}
					selector, err := sel.Build(findLimit, *offset, "KEYWORDID")
					if err != nil {
						return nil, fmt.Errorf("list: %w", err)
					}
					selector, err = shared.NormalizeStatusSelector(selector, "ACTIVE")
					if err != nil {
						return nil, err
					}
					if *limit != 0 {
						selector, err = shared.SetSelectorPagination(selector, *offset, *limit)
						if err != nil {
							return nil, err
						}
					}
					if isAdGroupLevel(*adgroupID) {
						selector, err = shared.AddSelectorEqualsCondition(selector, "adGroupId", strings.TrimSpace(*adgroupID))
						if err != nil {
							return nil, err
						}
						if *limit == 0 {
							return shared.FetchAllSelectorPages(ctx, selector, func(pageSelector json.RawMessage) (any, error) {
								var result json.RawMessage
								err := client.Do(ctx, negadgroup.FindRequest{
									CampaignID: cid,
									RawBody:    pageSelector,
								}, &result)
								return result, err
							})
						}
						var result json.RawMessage
						err := client.Do(ctx, negadgroup.FindRequest{
							CampaignID: cid,
							RawBody:    selector,
						}, &result)
						return result, err
					}
					if *limit == 0 {
						return shared.FetchAllSelectorPages(ctx, selector, func(pageSelector json.RawMessage) (any, error) {
							var result json.RawMessage
							err := client.Do(ctx, negcampaign.FindRequest{
								CampaignID: cid,
								RawBody:    pageSelector,
							}, &result)
							return result, err
						})
					}
					var result json.RawMessage
					err = client.Do(ctx, negcampaign.FindRequest{
						CampaignID: cid,
						RawBody:    selector,
					}, &result)
					return result, err
				}

				if isAdGroupLevel(*adgroupID) {
					req := negadgroup.ListRequest{
						CampaignID: cid,
						AdGroupID:  strings.TrimSpace(*adgroupID),
						Limit:      *limit,
						Offset:     *offset,
					}
					if *limit == 0 {
						return api.FetchAllRaw(ctx, client, req)
					}
					var result json.RawMessage
					err := client.Do(ctx, req, &result)
					return result, err
				}
				req := negcampaign.ListRequest{
					CampaignID: cid,
					Limit:      *limit,
					Offset:     *offset,
				}
				if *limit == 0 {
					return api.FetchAllRaw(ctx, client, req)
				}
				var result json.RawMessage
				err = client.Do(ctx, req, &result)
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
