package adgroups

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func listCmd() *ffcli.Command {
	return shared.BuildSmartListCommand(shared.SmartListCommandConfig{
		Name:         "list",
		ShortUsage:   "aads adgroups list [--campaign-id ID] [flags]",
		ShortHelp:    "List ad groups.",
		EntityIDName: "ADGROUPID",
		LongHelp: `List ad groups, with optional filtering.

With --campaign-id, lists ad groups in that campaign.
Without --campaign-id, searches across all campaigns.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, campaignId, name, status, servingStatus, displayStatus,
  pricingModel, defaultCpcBid, startTime, endTime

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads adgroups list
  aads adgroups list --campaign-id 123
  aads adgroups list --campaign-id 123 --filter "status=ENABLED" --sort "name:asc"
  aads adgroups list --filter "name STARTSWITH Search"`,
		ParentFlags: []shared.ParentFlag{
			{Name: "campaign-id", Usage: "Campaign ID (omit to search all campaigns)", Required: false},
		},
		FindWhenMissingParents: []string{"campaign-id"},
		ListExec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error) {
			req := adgroups.ListRequest{
				CampaignID: parentIDs["campaign-id"],
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
			err = client.Do(ctx, adgroups.FindRequest{
				CampaignID: parentIDs["campaign-id"],
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
			err = client.Do(ctx, adgroups.FindAllRequest{RawBody: selector}, &result)
			return result, err
		},
	})
}
