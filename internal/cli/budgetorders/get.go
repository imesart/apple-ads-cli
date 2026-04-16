package budgetorders

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/budgetorders"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func getCmd() *ffcli.Command {
	return shared.BuildIDGetCommand(shared.IDGetCommandConfig{
		Name:       "get",
		ShortUsage: "aads budgetorders get --budget-order-id ID",
		ShortHelp:  "Get a budget order.",
		IDFlag:     "budget-order-id",
		IDUsage:    "Budget Order ID",
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, budgetorders.GetRequest{BudgetOrderID: id}, &result)
			return result, err
		},
	})
}
