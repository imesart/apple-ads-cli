package testutil

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/imesart/apple-ads-cli/internal/api"
	adgroupsreq "github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	campaignsreq "github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	keywordsreq "github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	"github.com/imesart/apple-ads-cli/internal/auth"
	"github.com/imesart/apple-ads-cli/internal/config"
	"github.com/imesart/apple-ads-cli/internal/types"
)

const liveDiscoveryLimit = 1000

type LiveIDs struct {
	OrgID      string
	AdamID     string
	CampaignID string
	AdGroupID  string
	KeywordID  string
}

var (
	liveIDsOnce sync.Once
	liveIDs     LiveIDs
	liveIDsErr  error
)

// RequireLiveIDs returns a valid campaign/ad group/keyword chain discovered at
// runtime from the real Apple Ads API.
func RequireLiveIDs(t *testing.T) LiveIDs {
	t.Helper()

	if os.Getenv("AADS_INTEGRATION_TEST") != "1" {
		t.Skip("set AADS_INTEGRATION_TEST=1 to run live Apple Ads integration tests")
	}

	liveIDsOnce.Do(func() {
		liveIDs, liveIDsErr = discoverLiveIDs(context.Background())
	})

	if liveIDsErr != nil {
		t.Skipf("skipping live integration test: %v", liveIDsErr)
	}
	return liveIDs
}

func discoverLiveIDs(ctx context.Context) (LiveIDs, error) {
	timeout, err := config.SelectedTimeout(api.DefaultTimeout)
	if err != nil {
		return LiveIDs{}, fmt.Errorf("config validation: %w", err)
	}

	cfg, err := config.Load(config.SelectedProfile(""))
	if err != nil {
		return LiveIDs{}, fmt.Errorf("loading config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return LiveIDs{}, fmt.Errorf("config validation: %w", err)
	}
	if cfg.OrgID == "" {
		return LiveIDs{}, fmt.Errorf("missing org_id for live integration tests")
	}

	tokenStore := auth.NewTokenStore(
		cfg.TeamID,
		cfg.ClientID,
		cfg.KeyID,
		cfg.PrivateKeyPath,
		config.DefaultTokenCachePath(),
	)
	client := api.NewClient(tokenStore.GetToken, tokenStore.Invalidate, cfg.OrgID, false)
	client.SetTimeout(timeout)
	auth.SetHTTPClientTimeout(timeout)

	campaigns, err := listCampaigns(ctx, client)
	if err != nil {
		return LiveIDs{}, err
	}
	for _, campaign := range campaigns {
		campaignID, ok := int64String(campaign.ID)
		if !ok || isDeleted(campaign.Deleted) {
			continue
		}

		adGroups, err := listAdGroups(ctx, client, campaignID)
		if err != nil {
			return LiveIDs{}, fmt.Errorf("listing ad groups for campaign %s: %w", campaignID, err)
		}
		for _, adGroup := range adGroups {
			adGroupID, ok := int64String(adGroup.ID)
			if !ok || isDeleted(adGroup.Deleted) {
				continue
			}

			keywords, err := listKeywords(ctx, client, campaignID, adGroupID)
			if err != nil {
				return LiveIDs{}, fmt.Errorf("listing keywords for campaign %s ad group %s: %w", campaignID, adGroupID, err)
			}
			for _, keyword := range keywords {
				keywordID, ok := int64String(keyword.ID)
				if !ok || isDeleted(keyword.Deleted) {
					continue
				}
				return LiveIDs{
					OrgID:      cfg.OrgID,
					AdamID:     fmt.Sprintf("%d", campaign.AdamID),
					CampaignID: campaignID,
					AdGroupID:  adGroupID,
					KeywordID:  keywordID,
				}, nil
			}
		}
	}

	return LiveIDs{}, fmt.Errorf("no campaign/ad group/keyword chain found via list endpoints")
}

func listCampaigns(ctx context.Context, client *api.Client) ([]types.Campaign, error) {
	var resp types.ListResponse[types.Campaign]
	if err := client.DoList(ctx, campaignsreq.ListRequest{Limit: liveDiscoveryLimit}, &resp); err != nil {
		return nil, fmt.Errorf("listing campaigns: %w", err)
	}
	return resp.Data, nil
}

func listAdGroups(ctx context.Context, client *api.Client, campaignID string) ([]types.AdGroup, error) {
	var resp types.ListResponse[types.AdGroup]
	if err := client.DoList(ctx, adgroupsreq.ListRequest{
		CampaignID: campaignID,
		Limit:      liveDiscoveryLimit,
	}, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func listKeywords(ctx context.Context, client *api.Client, campaignID, adGroupID string) ([]types.Keyword, error) {
	var resp types.ListResponse[types.Keyword]
	if err := client.DoList(ctx, keywordsreq.ListRequest{
		CampaignID: campaignID,
		AdGroupID:  adGroupID,
		Limit:      liveDiscoveryLimit,
	}, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func int64String(v *int64) (string, bool) {
	if v == nil || *v == 0 {
		return "", false
	}
	return fmt.Sprintf("%d", *v), true
}

func isDeleted(v *bool) bool {
	return v != nil && *v
}
