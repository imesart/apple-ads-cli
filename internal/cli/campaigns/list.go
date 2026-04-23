package campaigns

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func listCmd() *ffcli.Command {
	return shared.BuildSmartListCommand(shared.SmartListCommandConfig{
		Name:         "list",
		ShortUsage:   "aads campaigns list [flags]",
		ShortHelp:    "List campaigns.",
		EntityIDName: "CAMPAIGNID",
		LongHelp: `List campaigns, with optional filtering.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, orgId, name, adamId, status, servingStatus, displayStatus,
  adChannelType, supplySources, billingEvent, paymentModel,
  countriesOrRegions, budgetAmount, dailyBudgetAmount, targetCpa

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads campaigns list
  aads campaigns list --filter "status=ENABLED"
  aads campaigns list --filter "name STARTSWITH MyApp" --filter "adamId IN [123, 456]" --filter "dailyBudgetAmount BETWEEN [10, 100]" --sort "name:asc" --sort "id:desc"`,
		ListExec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error) {
			req := campaigns.ListRequest{Limit: limit, Offset: offset}
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
			err = client.Do(ctx, campaigns.FindRequest{RawBody: selector}, &result)
			return result, err
		},
	})
}
