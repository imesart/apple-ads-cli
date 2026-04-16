package keywords

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func listCmd() *ffcli.Command {
	return shared.BuildSmartListCommand(shared.SmartListCommandConfig{
		Name:         "list",
		ShortUsage:   "aads keywords list --campaign-id CID --adgroup-id AGID [flags]",
		ShortHelp:    "List targeting keywords.",
		EntityIDName: "KEYWORDID",
		LongHelp: `List targeting keywords, with optional filtering.

Lists keywords in a specific ad group.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, campaignId, adGroupId, text, matchType, status, bidAmount

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads keywords list --campaign-id 123 --adgroup-id 456
  aads keywords list --campaign-id 123 --adgroup-id 456 --filter "matchType=EXACT"
  aads keywords list --campaign-id 123 --adgroup-id 456 --filter "status IN [ACTIVE, PAUSED]"
    --sort "text:asc"`,
		ParentFlags: []shared.ParentFlag{
			{Name: "campaign-id", Usage: "Campaign ID", Required: true},
			{Name: "adgroup-id", Usage: "Ad Group ID", Required: true},
		},
		ListExec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error) {
			req := keywords.ListRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  parentIDs["adgroup-id"],
				Limit:      limit,
				Offset:     offset,
			}
			if limit == 0 {
				return api.FetchAllRaw(ctx, client, req)
			}

			var result json.RawMessage
			err := client.Do(ctx, req, &result)
			return result, err
		},
		FindExec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, selector json.RawMessage) (any, error) {
			selector, err := shared.AddSelectorEqualsCondition(selector, "adGroupId", parentIDs["adgroup-id"])
			if err != nil {
				return nil, err
			}
			selector, err = shared.NormalizeStatusSelector(selector, "ACTIVE")
			if err != nil {
				return nil, err
			}
			var result json.RawMessage
			err = client.Do(ctx, keywords.FindRequest{
				CampaignID: parentIDs["campaign-id"],
				RawBody:    selector,
			}, &result)
			return result, err
		},
	})
}
