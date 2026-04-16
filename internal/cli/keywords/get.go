package keywords

import (
	"context"
	"encoding/json"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func getCmd() *ffcli.Command {
	return shared.BuildIDGetCommand(shared.IDGetCommandConfig{
		Name:        "get",
		ShortUsage:  "aads keywords get --campaign-id CID --adgroup-id AGID --keyword-id ID",
		ShortHelp:   "Get a targeting keyword.",
		IDFlag:      "keyword-id",
		IDUsage:     "Keyword ID",
		ParentFlags: campaignAdGroupParent,
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, keywords.GetRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  parentIDs["adgroup-id"],
				KeywordID:  id,
			}, &result)
			return result, err
		},
	})
}
