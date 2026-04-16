package budgetorders

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/budgetorders"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func listCmd() *ffcli.Command {
	return shared.BuildListCommand(shared.ListCommandConfig{
		Name:         "list",
		ShortUsage:   "aads budgetorders list",
		ShortHelp:    "List all budget orders.",
		EntityIDName: "BUDGETORDERID",
		Exec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error) {
			req := budgetorders.ListRequest{Limit: limit, Offset: offset}
			if limit == 0 {
				return api.FetchAllRaw(ctx, client, req)
			}
			var result json.RawMessage
			err := client.Do(ctx, req, &result)
			return result, err
		},
	})
}
