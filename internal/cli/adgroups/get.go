package adgroups

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func getCmd() *ffcli.Command {
	return shared.BuildIDGetCommand(shared.IDGetCommandConfig{
		Name:        "get",
		ShortUsage:  "aads adgroups get --campaign-id CID --adgroup-id ID",
		ShortHelp:   "Get an ad group by ID.",
		IDFlag:      "adgroup-id",
		IDUsage:     "Ad Group ID",
		ParentFlags: campaignParent,
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, adgroups.GetRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  id,
			}, &result)
			return result, err
		},
	})
}
