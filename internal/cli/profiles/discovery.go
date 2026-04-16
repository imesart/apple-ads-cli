package profiles

import (
	"context"
	"fmt"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/acls"
	"github.com/imesart/apple-ads-cli/internal/auth"
	"github.com/imesart/apple-ads-cli/internal/config"
	"github.com/imesart/apple-ads-cli/internal/types"
)

type discoveredProfileDefaults struct {
	OrgID           string
	DefaultCurrency string
	DefaultTimezone string
	Warnings        []string
}

func discoverProfileDefaults(ctx context.Context, cfg config.Profile, explicitOrgID string) (discoveredProfileDefaults, error) {
	result := discoveredProfileDefaults{}

	timeout, err := config.SelectedTimeout(api.DefaultTimeout)
	if err != nil {
		return result, err
	}

	bootstrapCfg := config.Profile{
		ClientID:       strings.TrimSpace(cfg.ClientID),
		TeamID:         strings.TrimSpace(cfg.TeamID),
		KeyID:          strings.TrimSpace(cfg.KeyID),
		PrivateKeyPath: strings.TrimSpace(cfg.PrivateKeyPath),
	}
	bootstrapCfg.ApplyEnv()
	if err := bootstrapCfg.Validate(); err != nil {
		return result, err
	}

	tokenStore := auth.NewTokenStore(
		bootstrapCfg.TeamID,
		bootstrapCfg.ClientID,
		bootstrapCfg.KeyID,
		bootstrapCfg.PrivateKeyPath,
		config.DefaultTokenCachePath(),
	)
	client := api.NewClient(tokenStore.GetToken, tokenStore.Invalidate, "", false)
	client.SetTimeout(timeout)
	auth.SetHTTPClientTimeout(timeout)

	resolvedOrgID := strings.TrimSpace(explicitOrgID)
	if resolvedOrgID == "" {
		var meResp types.DataResponse[types.MeDetail]
		if err := client.Do(ctx, acls.MeRequest{}, &meResp); err != nil {
			return result, fmt.Errorf("discovering org from orgs user (Apple ACLs): %w", err)
		}
		if meResp.Data.ParentOrgID == 0 {
			return result, fmt.Errorf("discovering org from orgs user (Apple ACLs): parentOrgId is empty")
		}
		resolvedOrgID = fmt.Sprintf("%d", meResp.Data.ParentOrgID)
	}
	result.OrgID = resolvedOrgID

	var aclResp types.ListResponse[types.UserACL]
	if err := client.Do(ctx, acls.ListRequest{}, &aclResp); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("could not inspect orgs list (Apple ACLs) for org %s: %v", resolvedOrgID, err))
		return result, nil
	}

	for _, acl := range aclResp.Data {
		if fmt.Sprintf("%d", acl.OrgID) != resolvedOrgID {
			continue
		}
		result.DefaultCurrency = strings.TrimSpace(acl.Currency)
		result.DefaultTimezone = strings.TrimSpace(acl.TimeZone)
		return result, nil
	}

	result.Warnings = append(result.Warnings, fmt.Sprintf("could not find org %s in orgs list (Apple ACLs); default currency and timezone were not inferred", resolvedOrgID))
	return result, nil
}
