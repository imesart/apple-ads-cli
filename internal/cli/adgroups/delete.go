package adgroups

import (
	"context"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func deleteCmd() *ffcli.Command {
	return shared.BuildDeleteCommand(shared.DeleteCommandConfig{
		Name:        "delete",
		Resource:    "adgroup",
		ShortUsage:  "aads adgroups delete --campaign-id CID --adgroup-id ID --confirm",
		ShortHelp:   "Delete an ad group.",
		IDFlag:      "adgroup-id",
		IDUsage:     "Ad Group ID",
		ParentFlags: campaignParent,
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) error {
			return client.Do(ctx, adgroups.DeleteRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  id,
			}, nil)
		},
	})
}
