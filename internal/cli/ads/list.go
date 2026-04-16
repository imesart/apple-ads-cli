package ads

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/ads"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func listCmd() *ffcli.Command {
	return shared.BuildSmartListCommand(shared.SmartListCommandConfig{
		Name:         "list",
		ShortUsage:   "aads ads list [--campaign-id CID --adgroup-id AGID] [flags]",
		ShortHelp:    "List ads.",
		EntityIDName: "ADID",
		LongHelp: `List ads, with optional filtering.

With --campaign-id and --adgroup-id, lists ads in that ad group.
Without them, searches across all campaigns.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, orgId, campaignId, adGroupId, name, status, servingStatus,
  creativeType

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads ads list
  aads ads list --campaign-id 123 --adgroup-id 456
  aads ads list --filter "status=ENABLED"
  aads ads list --campaign-id 123 --adgroup-id 456 --filter "name CONTAINS promo" --filter "creativeType=CUSTOM_PRODUCT_PAGE" --sort "name:asc"`,
		ParentFlags: []shared.ParentFlag{
			{Name: "campaign-id", Usage: "Campaign ID (omit both to search all)", Required: false},
			{Name: "adgroup-id", Usage: "Ad Group ID (omit both to search all)", Required: false},
		},
		FindWhenMissingParents: []string{"campaign-id", "adgroup-id"},
		ListExec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error) {
			req := ads.ListRequest{
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
			selector, err := shared.NormalizeStatusSelector(selector, "ENABLED")
			if err != nil {
				return nil, err
			}
			var result json.RawMessage
			err = client.Do(ctx, ads.FindRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  parentIDs["adgroup-id"],
				RawBody:    selector,
			}, &result)
			return result, err
		},
		FindAllExec: func(ctx context.Context, client *api.Client, selector json.RawMessage) (any, error) {
			selector, err := shared.NormalizeStatusSelector(selector, "ENABLED")
			if err != nil {
				return nil, err
			}
			var result json.RawMessage
			err = client.Do(ctx, ads.FindAllRequest{RawBody: selector}, &result)
			return result, err
		},
	})
}
