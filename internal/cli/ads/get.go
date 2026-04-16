package ads

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/ads"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func getCmd() *ffcli.Command {
	return shared.BuildIDGetCommand(shared.IDGetCommandConfig{
		Name:        "get",
		ShortUsage:  "aads ads get --campaign-id CID --adgroup-id AGID --ad-id ID",
		ShortHelp:   "Get an ad by ID.",
		IDFlag:      "ad-id",
		IDUsage:     "Ad ID",
		ParentFlags: campaignAdGroupParent,
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, ads.GetRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  parentIDs["adgroup-id"],
				AdID:       id,
			}, &result)
			return result, err
		},
	})
}
