package structure

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	adgroupsreq "github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	appsreq "github.com/imesart/apple-ads-cli/internal/api/requests/apps"
	campaignsreq "github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	keywordsreq "github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	negadgroupreq "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_adgroup"
	negcampaignreq "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_campaign"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/output"
)

func exportCmd() *ffcli.Command {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)

	scope := fs.String("scope", "", "Structure scope: campaigns | adgroups")
	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin)")
	shareable := fs.Bool("shareable", false, "Export a shareable structure preset: omits keywords, negatives, adamId, and times, and redacts names")
	noAdamID := fs.Bool("no-adam-id", false, "Omit campaign adamId from exported structure JSON")
	noBudgets := fs.Bool("no-budgets", false, "Omit budget, bid, CPA, and invoice-related fields unless explicitly requested")
	noTimes := fs.Bool("no-times", false, "Omit campaign/ad group startTime and endTime unless explicitly requested")
	redactNames := fs.Bool("redact-names", false, "Redact campaign/ad group names using %(appName), %(appNameShort), %(countriesOrRegions), and %(campaignName)")
	noNegatives := fs.Bool("no-negatives", false, "Skip campaign and ad group negative keyword export")
	noKeywords := fs.Bool("no-keywords", false, "Skip keyword export")
	campaignsFields := fs.String("campaigns-fields", "", `Campaign fields to export. Omit for normalized defaults; use "" to export all fields.`)
	adgroupsFields := fs.String("adgroups-fields", "", `Ad group fields to export. Omit for normalized defaults; use "" to export all fields.`)
	keywordsFields := fs.String("keywords-fields", "", `Keyword fields to export. Omit for normalized defaults; use "" to export all fields.`)
	campaignNegativesFields := fs.String("campaigns-negatives-fields", "", `Campaign negative keyword fields to export. Omit for normalized defaults; use "" to export all fields.`)
	adgroupNegativesFields := fs.String("adgroups-negatives-fields", "", `Ad group negative keyword fields to export. Omit for normalized defaults; use "" to export all fields.`)

	campaignsSelector := shared.BindNamedSelectorFlags(fs, "campaigns-filter", "campaigns-sort", "campaigns-selector")
	adgroupsSelector := shared.BindNamedSelectorFlags(fs, "adgroups-filter", "adgroups-sort", "adgroups-selector")
	keywordsSelector := shared.BindNamedSelectorFlags(fs, "keywords-filter", "keywords-sort", "keywords-selector")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON even when stdout is not a TTY")

	return &ffcli.Command{
		Name:       "export",
		ShortUsage: "aads structure export --scope campaigns|adgroups [flags]",
		ShortHelp:  "Export a structure JSON for campaigns/ad groups, keywords, and negatives.",
		LongHelp: `Export a structure JSON that can later be consumed by "aads structure import".

Scopes:
  campaigns  Export campaigns, campaign negatives, ad groups, ad group negatives, and keywords.
  adgroups   Export ad groups, ad group negatives, and keywords from matching campaigns.

Selection:
  --campaign-id selects one campaign directly and supports - for stdin pipelines.
  --campaigns-filter / --campaigns-selector choose which campaigns to include.
  Without campaign filters/selectors, all campaigns are included.
  --adgroups-filter / --adgroups-selector choose which ad groups to include within the selected campaigns.
  Without ad group filters/selectors, all ad groups in the selected campaigns are included.
  --keywords-filter / --keywords-selector choose which keywords to include within the selected ad groups.
  Without keyword filters/selectors, all keywords in the selected ad groups are included.

Field export:
  Omit a --*-fields flag to export the normalized creation-oriented field set.
  Pass --*-fields "" to export all fields for that entity type with no default-value omission.
  Pass --*-fields "fieldA,fieldB" to export those fields plus required creation fields.

Example:
  aads structure export --scope campaigns --campaigns-filter "name STARTSWITH FitTrack"

This command always writes JSON to stdout.`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			switch *scope {
			case scopeCampaigns, scopeAdgroups:
			default:
				return shared.UsageError("--scope is required and must be campaigns or adgroups")
			}
			if strings.TrimSpace(*campaignID) != "" && campaignsSelector.HasFlags() {
				return shared.ValidationError("--campaign-id cannot be combined with --campaigns-filter, --campaigns-sort, or --campaigns-selector")
			}

			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("export: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			buildRoot := func() (any, error) {
				root := newStructureRoot(*scope)
				appVarsCache := map[string]map[string]string{}
				effectiveNoAdamID := *noAdamID || *shareable
				effectiveNoBudgets := *noBudgets || *shareable
				effectiveNoTimes := *noTimes || *shareable
				effectiveRedactNames := *redactNames || *shareable
				effectiveNoNegatives := *noNegatives || *shareable
				effectiveNoKeywords := *noKeywords || *shareable
				campaignSelection := parseFieldSelection(fs, "campaigns-fields", *campaignsFields)
				adgroupSelection := parseFieldSelection(fs, "adgroups-fields", *adgroupsFields)
				keywordSelection := parseFieldSelection(fs, "keywords-fields", *keywordsFields)
				campaignNegativeSelection := parseFieldSelection(fs, "campaigns-negatives-fields", *campaignNegativesFields)
				adgroupNegativeSelection := parseFieldSelection(fs, "adgroups-negatives-fields", *adgroupNegativesFields)
				campaignSelectorFields := selectorFieldsForEntity(campaignSelection, campaignAllowedFields, campaignRequiredFields, campaignReadOnlyFields, campaignRelationshipFields)
				adgroupSelectorFields := selectorFieldsForEntity(adgroupSelection, adgroupAllowedFields, adgroupRequiredFields, adgroupReadOnlyFields, adgroupRelationshipFields)

				campaigns, err := fetchCampaignsForExport(ctx, client, strings.TrimSpace(*campaignID), campaignsSelector, campaignSelectorFields)
				if err != nil {
					return nil, err
				}

				if *scope == scopeCampaigns {
					for _, campaign := range campaigns {
						campaignNode, err := buildCampaignNode(ctx, client, campaign, adgroupsSelector, keywordsSelector, campaignSelection, adgroupSelection, keywordSelection, campaignNegativeSelection, adgroupNegativeSelection, appVarsCache, effectiveNoAdamID, effectiveNoBudgets, effectiveNoTimes, effectiveRedactNames, effectiveNoNegatives, effectiveNoKeywords)
						if err != nil {
							return nil, err
						}
						root.Campaigns = append(root.Campaigns, campaignNode)
					}
				} else {
					for _, campaign := range campaigns {
						adgroups, err := fetchAdgroupsForExport(ctx, client, stringValue(campaign["id"]), adgroupsSelector, adgroupSelectorFields)
						if err != nil {
							return nil, err
						}
						for _, adgroup := range adgroups {
							appVars, err := resolveExportAppTemplateVars(ctx, client, stringValue(campaign["adamId"]), appVarsCache, effectiveRedactNames)
							if err != nil {
								return nil, err
							}
							node, err := buildAdgroupNode(ctx, client, stringValue(campaign["id"]), adgroup, stringValue(campaign["name"]), appVars, keywordsSelector, adgroupSelection, keywordSelection, adgroupNegativeSelection, effectiveNoBudgets, effectiveNoTimes, effectiveRedactNames, effectiveNoNegatives, effectiveNoKeywords)
							if err != nil {
								return nil, err
							}
							root.Adgroups = append(root.Adgroups, node)
						}
					}
				}
				return root, nil
			}

			stdinFlags := shared.CollectStdinFlags(shared.StdinFlag{Name: "campaign-id", Ptr: campaignID})
			if len(stdinFlags) > 0 {
				results, dataRows, failures := shared.CollectResultsWithStdin(stdinFlags, buildRoot)
				if failures > 0 {
					return shared.ReportError(fmt.Errorf("%d of %d lines failed", failures, dataRows))
				}
				root := newStructureRoot(*scope)
				for _, result := range results {
					subroot, ok := result.(structureRoot)
					if !ok {
						return fmt.Errorf("export: unexpected stdin result type %T", result)
					}
					mergeStructureRoots(&root, subroot)
				}
				return output.PrintJSON(root, *pretty)
			}

			resp, err := buildRoot()
			if err != nil {
				return fmt.Errorf("export: %w", err)
			}
			return output.PrintJSON(resp, *pretty)
		},
	}
}

func fetchCampaignsForExport(ctx context.Context, client *api.Client, campaignID string, selector *shared.SelectorFlags, requiredFields []string) ([]map[string]any, error) {
	if campaignID != "" {
		var raw json.RawMessage
		if err := client.Do(ctx, campaignsreq.GetRequest{CampaignID: campaignID}, &raw); err != nil {
			return nil, err
		}
		row, err := decodeDataObject(raw)
		if err != nil {
			return nil, err
		}
		return []map[string]any{row}, nil
	}
	if selector.HasFlags() {
		return fetchAllFind(ctx, client, selector, func(sel json.RawMessage) (any, error) {
			sel, err := shared.NormalizeStatusSelector(sel, "ENABLED")
			if err != nil {
				return nil, err
			}
			sel, err = ensureSelectorFields(sel, requiredFields)
			if err != nil {
				return nil, err
			}
			var raw json.RawMessage
			if err := client.Do(ctx, campaignsreq.FindRequest{RawBody: sel}, &raw); err != nil {
				return nil, err
			}
			return raw, nil
		})
	}
	return fetchAllList(ctx, client, campaignsreq.ListRequest{})
}

func mergeStructureRoots(dst *structureRoot, src structureRoot) {
	if src.Scope != dst.Scope {
		return
	}
	dst.Campaigns = append(dst.Campaigns, src.Campaigns...)
	dst.Adgroups = append(dst.Adgroups, src.Adgroups...)
}

func fetchAdgroupsForExport(ctx context.Context, client *api.Client, campaignID string, selector *shared.SelectorFlags, requiredFields []string) ([]map[string]any, error) {
	if selector.HasFlags() {
		return fetchAllFind(ctx, client, selector, func(sel json.RawMessage) (any, error) {
			sel, err := shared.NormalizeStatusSelector(sel, "ENABLED")
			if err != nil {
				return nil, err
			}
			sel, err = ensureSelectorFields(sel, requiredFields)
			if err != nil {
				return nil, err
			}
			var raw json.RawMessage
			if err := client.Do(ctx, adgroupsreq.FindRequest{CampaignID: campaignID, RawBody: sel}, &raw); err != nil {
				return nil, err
			}
			return raw, nil
		})
	}
	return fetchAllList(ctx, client, adgroupsreq.ListRequest{CampaignID: campaignID})
}

func buildCampaignNode(ctx context.Context, client *api.Client, campaign map[string]any, adgroupsSelector, keywordsSelector *shared.SelectorFlags, campaignSelection, adgroupSelection, keywordSelection, campaignNegativeSelection, adgroupNegativeSelection fieldSelection, appVarsCache map[string]map[string]string, noAdamID, noBudgets, noTimes, redactNames, noNegatives, noKeywords bool) (structureCampaignNode, error) {
	campaignID := stringValue(campaign["id"])
	node := structureCampaignNode{
		Campaign: filterEntityForExport(campaign, campaignSelection, campaignAllowedFields, campaignRequiredFields, campaignReadOnlyFields, campaignRelationshipFields, omitCampaignField),
	}
	appVars, err := resolveExportAppTemplateVars(ctx, client, stringValue(campaign["adamId"]), appVarsCache, redactNames)
	if err != nil {
		return node, err
	}
	if redactNames {
		node.Campaign["name"] = redactCampaignExportName(stringValue(node.Campaign["name"]), campaign, appVars)
	}
	if shouldStripCampaignFieldForExport("adamId", campaignSelection, noAdamID) {
		delete(node.Campaign, "adamId")
	}
	for _, key := range []string{"dailyBudgetAmount", "budgetAmount", "locInvoiceDetails", "budgetOrders", "targetCpa"} {
		if shouldStripCampaignFieldForExport(key, campaignSelection, noBudgets) {
			delete(node.Campaign, key)
		}
	}
	if shouldStripTimeFieldForExport("startTime", campaignSelection, noTimes) {
		delete(node.Campaign, "startTime")
	}
	if shouldStripTimeFieldForExport("endTime", campaignSelection, noTimes) {
		delete(node.Campaign, "endTime")
	}

	if !shouldSkipChildEntityExport(noNegatives, campaignNegativeSelection) {
		campaignNegatives, err := fetchAllList(ctx, client, negcampaignreq.ListRequest{CampaignID: campaignID})
		if err != nil {
			return node, err
		}
		for _, negative := range campaignNegatives {
			node.CampaignNegativeKeywords = append(node.CampaignNegativeKeywords,
				filterEntityForExport(negative, campaignNegativeSelection, negativeAllowedFields, negativeRequiredFields, negativeReadOnlyFields, negativeRelationshipFields, omitNegativeField))
		}
	}

	adgroups, err := fetchAdgroupsForExport(ctx, client, campaignID, adgroupsSelector, selectorFieldsForEntity(adgroupSelection, adgroupAllowedFields, adgroupRequiredFields, adgroupReadOnlyFields, adgroupRelationshipFields))
	if err != nil {
		return node, err
	}
	for _, adgroup := range adgroups {
		child, err := buildAdgroupNode(ctx, client, campaignID, adgroup, stringValue(campaign["name"]), appVars, keywordsSelector, adgroupSelection, keywordSelection, adgroupNegativeSelection, noBudgets, noTimes, redactNames, noNegatives, noKeywords)
		if err != nil {
			return node, err
		}
		node.Adgroups = append(node.Adgroups, child)
	}
	return node, nil
}

func buildAdgroupNode(ctx context.Context, client *api.Client, campaignID string, adgroup map[string]any, originalCampaignName string, appVars map[string]string, keywordsSelector *shared.SelectorFlags, adgroupSelection, keywordSelection, adgroupNegativeSelection fieldSelection, noBudgets, noTimes, redactNames, noNegatives, noKeywords bool) (structureAdGroupNode, error) {
	adgroupID := stringValue(adgroup["id"])
	node := structureAdGroupNode{
		Adgroup: filterEntityForExport(adgroup, adgroupSelection, adgroupAllowedFields, adgroupRequiredFields, adgroupReadOnlyFields, adgroupRelationshipFields, omitAdgroupField),
	}
	if redactNames {
		node.Adgroup["name"] = redactAdgroupExportName(stringValue(node.Adgroup["name"]), originalCampaignName, appVars)
	}
	for _, key := range []string{"defaultBidAmount", "cpaGoal"} {
		if shouldStripCampaignFieldForExport(key, adgroupSelection, noBudgets) {
			delete(node.Adgroup, key)
		}
	}
	if shouldStripTimeFieldForExport("startTime", adgroupSelection, noTimes) {
		delete(node.Adgroup, "startTime")
	}
	if shouldStripTimeFieldForExport("endTime", adgroupSelection, noTimes) {
		delete(node.Adgroup, "endTime")
	}

	if !shouldSkipChildEntityExport(noNegatives, adgroupNegativeSelection) {
		adgroupNegatives, err := fetchAllList(ctx, client, negadgroupreq.ListRequest{CampaignID: campaignID, AdGroupID: adgroupID})
		if err != nil {
			return node, err
		}
		for _, negative := range adgroupNegatives {
			node.AdgroupNegativeKeywords = append(node.AdgroupNegativeKeywords,
				filterEntityForExport(negative, adgroupNegativeSelection, negativeAllowedFields, negativeRequiredFields, negativeReadOnlyFields, negativeRelationshipFields, omitNegativeField))
		}
	}

	if !shouldSkipChildEntityExport(noKeywords, keywordSelection) {
		keywords, err := fetchKeywordsForExport(ctx, client, campaignID, adgroupID, keywordsSelector, selectorFieldsForEntity(keywordSelection, keywordAllowedFields, keywordRequiredFields, keywordReadOnlyFields, keywordRelationshipFields))
		if err != nil {
			return node, err
		}
		for _, keyword := range keywords {
			keywordWithDefault := cloneMap(keyword)
			if defaultBid, ok := adgroup["defaultBidAmount"]; ok {
				keywordWithDefault["defaultBidAmount"] = defaultBid
			}
			filteredKeyword := filterEntityForExport(keywordWithDefault, keywordSelection, keywordAllowedFields, keywordRequiredFields, keywordReadOnlyFields, keywordRelationshipFields, omitKeywordField)
			if shouldStripCampaignFieldForExport("bidAmount", keywordSelection, noBudgets) {
				delete(filteredKeyword, "bidAmount")
			}
			node.Keywords = append(node.Keywords, filteredKeyword)
		}
	}
	return node, nil
}

func fetchKeywordsForExport(ctx context.Context, client *api.Client, campaignID, adgroupID string, selector *shared.SelectorFlags, requiredFields []string) ([]map[string]any, error) {
	if selector != nil && selector.HasFlags() {
		return fetchAllFind(ctx, client, selector, func(sel json.RawMessage) (any, error) {
			sel, err := shared.AddSelectorEqualsCondition(sel, "adGroupId", adgroupID)
			if err != nil {
				return nil, err
			}
			sel, err = shared.NormalizeStatusSelector(sel, "ACTIVE")
			if err != nil {
				return nil, err
			}
			sel, err = ensureSelectorFields(sel, requiredFields)
			if err != nil {
				return nil, err
			}
			var raw json.RawMessage
			if err := client.Do(ctx, keywordsreq.FindRequest{CampaignID: campaignID, RawBody: sel}, &raw); err != nil {
				return nil, err
			}
			return raw, nil
		})
	}
	return fetchAllList(ctx, client, keywordsreq.ListRequest{CampaignID: campaignID, AdGroupID: adgroupID})
}

func resolveExportAppTemplateVars(ctx context.Context, client *api.Client, adamID string, cache map[string]map[string]string, redactNames bool) (map[string]string, error) {
	if !redactNames {
		return nil, nil
	}
	adamID = strings.TrimSpace(adamID)
	if adamID == "" || adamID == "<nil>" {
		return nil, shared.ValidationError("--redact-names requires campaign adamId")
	}
	if cached, ok := cache[adamID]; ok {
		return cloneStringMap(cached), nil
	}
	var raw json.RawMessage
	if err := client.Do(ctx, appsreq.DetailsRequest{AdamID: adamID}, &raw); err != nil {
		return nil, err
	}
	app, err := decodeDataObject(raw)
	if err != nil {
		return nil, err
	}
	vars := buildTemplateVariables(app)
	if appName := vars["appName"]; strings.TrimSpace(appName) != "" {
		if short := deriveAppNameShortValue(appName); short != "" {
			vars["appNameShort"] = short
			vars["APP_NAME_SHORT"] = short
		}
	}
	cache[adamID] = cloneStringMap(vars)
	return vars, nil
}
