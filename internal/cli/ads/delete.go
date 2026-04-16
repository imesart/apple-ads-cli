package ads

import (
	"context"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/ads"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func deleteCmd() *ffcli.Command {
	return shared.BuildDeleteCommand(shared.DeleteCommandConfig{
		Name:        "delete",
		Resource:    "ad",
		ShortUsage:  "aads ads delete --campaign-id CID --adgroup-id AGID --ad-id ID --confirm",
		ShortHelp:   "Delete an ad.",
		IDFlag:      "ad-id",
		IDUsage:     "Ad ID",
		ParentFlags: campaignAdGroupParent,
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) error {
			return client.Do(ctx, ads.DeleteRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  parentIDs["adgroup-id"],
				AdID:       id,
			}, nil)
		},
	})
}
