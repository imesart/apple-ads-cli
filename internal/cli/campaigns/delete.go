package campaigns

import (
	"context"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func deleteCmd() *ffcli.Command {
	return shared.BuildDeleteCommand(shared.DeleteCommandConfig{
		Name:       "delete",
		Resource:   "campaign",
		ShortUsage: "aads campaigns delete --campaign-id ID --confirm",
		ShortHelp:  "Delete a campaign.",
		IDFlag:     "campaign-id",
		IDUsage:    "Campaign ID",
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) error {
			return client.Do(ctx, campaigns.DeleteRequest{CampaignID: id}, nil)
		},
	})
}
