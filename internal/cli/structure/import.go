package structure

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	adgroupsreq "github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	appsreq "github.com/imesart/apple-ads-cli/internal/api/requests/apps"
	campaignsreq "github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	keywordsreq "github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	negadgroupreq "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_adgroup"
	negcampaignreq "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_campaign"
	adgroupscli "github.com/imesart/apple-ads-cli/internal/cli/adgroups"
	campaignscli "github.com/imesart/apple-ads-cli/internal/cli/campaigns"
	keywordscli "github.com/imesart/apple-ads-cli/internal/cli/keywords"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

type plannedCampaign struct {
	source            map[string]any
	payload           map[string]any
	createdID         any
	createdName       string
	status            string
	errorMessage      string
	campaignNegatives []plannedEntity
	adgroups          []plannedAdgroup
}

type plannedAdgroup struct {
	source           map[string]any
	payload          map[string]any
	createdID        any
	createdName      string
	status           string
	errorMessage     string
	adgroupNegatives []plannedEntity
	keywords         []plannedEntity
}

type plannedEntity struct {
	source       map[string]any
	payload      map[string]any
	createdID    any
	createdName  string
	status       string
	errorMessage string
}

const (
	importStatusCreated      = "created"
	importStatusFailed       = "failed"
	importStatusNotAttempted = "not_attempted"
)

func importCmd() *ffcli.Command {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	flags := &importFlags{}

	fs.StringVar(&flags.fromStructure, "from-structure", "", `Structure JSON input: inline JSON, @file.json, or @- for stdin`)
	fs.StringVar(&flags.outputMapping, "output-mapping", "", `Write mapping JSON to this path; use - for stdout`)
	fs.BoolVar(&flags.check, "check", false, "Validate and emit mapping JSON without sending mutating API requests")
	fs.BoolVar(&flags.noAdgroups, "no-adgroups", false, "Skip ad group creation for campaigns scope imports")
	fs.BoolVar(&flags.noNegatives, "no-negatives", false, "Skip campaign and ad group negative keyword creation")
	fs.BoolVar(&flags.noKeywords, "no-keywords", false, "Skip keyword creation")
	fs.BoolVar(&flags.allowUnmatchedNamePattern, "allow-unmatched-name-pattern", false, "Allow name patterns that do not match the source name")
	fs.StringVar(&flags.campaignID, "campaign-id", "", "Destination campaign ID for adgroups scope")
	fs.StringVar(&flags.campaignsFromJSON, "campaigns-from-json", "", `Campaign override JSON: inline JSON, @file.json, or @- for stdin`)
	fs.StringVar(&flags.adgroupsFromJSON, "adgroups-from-json", "", `Ad group override JSON: inline JSON, @file.json, or @- for stdin`)
	fs.StringVar(&flags.campaignsName, "campaigns-name", "", "Destination campaign name template; accepts %(fieldName), %(FIELD_NAME), %(1), and %1")
	fs.StringVar(&flags.campaignsNamePattern, "campaigns-name-pattern", "", "Basic regexp pattern used to capture source campaign name parts")
	fs.StringVar(&flags.adgroupsName, "adgroups-name", "", "Destination ad group name template; accepts %(fieldName), %(FIELD_NAME), %(1), %1, and %(CAMPAIGN_name)")
	fs.StringVar(&flags.adgroupsNamePattern, "adgroups-name-pattern", "", "Basic regexp pattern used to capture source ad group name parts")
	fs.StringVar(&flags.adamID, "adam-id", "", "Override adamId for created campaigns")
	fs.StringVar(&flags.dailyBudgetAmount, "daily-budget-amount", "", "Override dailyBudgetAmount for created campaigns")
	fs.StringVar(&flags.countriesOrRegions, "countries-or-regions", "", "Override countriesOrRegions for created campaigns")
	fs.StringVar(&flags.budgetAmount, "budget-amount", "", "DEPRECATED: Override budgetAmount for created campaigns")
	fs.StringVar(&flags.targetCpa, "target-cpa", "", "Override campaign targetCpa")
	fs.StringVar(&flags.locInvoiceDetails, "loc-invoice-details", "", `Override locInvoiceDetails JSON: inline JSON, @file.json, or @- for stdin`)
	fs.StringVar(&flags.campaignsStatus, "campaigns-status", "", "Override campaign status")
	fs.StringVar(&flags.campaignsStartTime, "campaigns-start-time", "", "Override campaign startTime")
	fs.StringVar(&flags.campaignsEndTime, "campaigns-end-time", "", "Override campaign endTime")
	fs.StringVar(&flags.adChannelType, "ad-channel-type", "", "Override campaign adChannelType")
	fs.StringVar(&flags.billingEvent, "billing-event", "", "Override campaign billingEvent")
	fs.StringVar(&flags.supplySources, "supply-sources", "", "Override campaign supplySources")
	fs.StringVar(&flags.cpaGoal, "cpa-goal", "", "Override ad group cpaGoal")
	fs.StringVar(&flags.defaultBid, "default-bid", "", "Override ad group defaultBidAmount")
	fs.StringVar(&flags.adgroupsStatus, "adgroups-status", "", "Override ad group status")
	fs.StringVar(&flags.adgroupsStartTime, "adgroups-start-time", "", "Override ad group startTime")
	fs.StringVar(&flags.adgroupsEndTime, "adgroups-end-time", "", "Override ad group endTime")
	fs.BoolVar(&flags.automatedKeywordsOptIn, "automated-keywords-opt-in", false, "Override automatedKeywordsOptIn to true for all created ad groups")
	fs.StringVar(&flags.matchType, "match-type", "", "Override keyword matchType")
	fs.StringVar(&flags.bid, "bid", "", `Override keyword bidAmount; pass "" to clear keyword-level bids`)
	fs.StringVar(&flags.keywordsStatus, "keywords-status", "", "Override keyword status")
	fs.StringVar(&flags.negativesStatus, "negatives-status", "", "Override negative keyword status")
	outputFlags := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "import",
		ShortUsage: "aads structure import --from-structure JSON [flags]",
		ShortHelp:  "Import campaigns/ad groups, keywords, and negatives from structure JSON.",
		LongHelp: `Import a structure JSON previously produced by "aads structure export".

Accepted structure:
  schemaVersion  integer  currently 1
  type           string   must be "structure"
  scope          string   "campaigns" or "adgroups"

Output:
  This command emits mapping JSON with type "mapping".
  Without --output-mapping, the mapping is written to stdout.
  With --output-mapping FILE, the mapping is written only to FILE.
  Use --output-mapping - to explicitly target stdout.

Validation:
  --check runs the same planning and validation path as a live import, including
  collision checks, but uses mock sequential IDs instead of sending mutating API requests.

Skip flags:
  --no-adgroups skips ad group creation for campaigns scope imports. This also skips
  ad group negatives and keywords. Campaign negatives are still created unless
  --no-negatives is also set.

Example:
  aads structure import --from-structure @structure.json --campaign-id 500 --adgroups-name "%(name) Copy" --check`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(flags.fromStructure) == "" {
				return shared.UsageError("--from-structure is required")
			}
			if err := validateImportJSONInputs(flags); err != nil {
				return err
			}

			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}
			cfg, err := shared.LoadConfig()
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}
			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			root, err := parseStructureInput(flags.fromStructure)
			if err != nil {
				return err
			}
			if err := validateScopeFlags(fs, root.Scope, flags); err != nil {
				return err
			}

			plan, campaignsPlan, adgroupsPlan, err := buildImportPlan(ctx, client, cfg, fs, root, flags)
			if err != nil {
				return err
			}

			if flags.check {
				assignMockIDs(campaignsPlan, adgroupsPlan)
				plan.mapping = buildMapping(root.Scope, campaignsPlan, adgroupsPlan, false)
				return writeMappingOutput(plan.mapping, flags.outputMapping, flags.outputMapping == "", *outputFlags.Pretty)
			}

			if err := executeImportPlan(ctx, client, root.Scope, flags.campaignID, campaignsPlan, adgroupsPlan); err != nil {
				mapping := buildMapping(root.Scope, campaignsPlan, adgroupsPlan, true)
				if mapErr := writeMappingOutput(mapping, flags.outputMapping, flags.outputMapping == "", *outputFlags.Pretty); mapErr != nil {
					return fmt.Errorf("import: %v; additionally failed to write mapping: %w", err, mapErr)
				}
				return fmt.Errorf("import: %w", err)
			}

			plan.mapping = buildMapping(root.Scope, campaignsPlan, adgroupsPlan, true)
			return writeMappingOutput(plan.mapping, flags.outputMapping, flags.outputMapping == "", *outputFlags.Pretty)
		},
	}
}

func validateImportJSONInputs(flags *importFlags) error {
	count := 0
	for _, value := range []string{flags.fromStructure, flags.campaignsFromJSON, flags.adgroupsFromJSON, flags.locInvoiceDetails} {
		if shared.IsStdinJSONInputArg(value) {
			count++
		}
	}
	if count > 1 {
		return shared.ValidationError("only one of --from-structure, --campaigns-from-json, --adgroups-from-json, or --loc-invoice-details may use @-")
	}
	return nil
}

func validateScopeFlags(fs *flag.FlagSet, scope string, flags *importFlags) error {
	if scope == scopeCampaigns {
		if strings.TrimSpace(flags.campaignID) != "" {
			return shared.ValidationError("--campaign-id is only valid for adgroups scope imports")
		}
		return nil
	}
	if flags.noAdgroups {
		return shared.ValidationError("--no-adgroups is only valid for campaigns scope imports")
	}
	if strings.TrimSpace(flags.campaignID) == "" {
		return shared.ValidationError("--campaign-id is required for adgroups scope imports")
	}
	if flagWasSet(fs, "campaigns-from-json") || flagWasSet(fs, "campaigns-name") || flagWasSet(fs, "campaigns-name-pattern") ||
		flagWasSet(fs, "adam-id") || flagWasSet(fs, "daily-budget-amount") || flagWasSet(fs, "countries-or-regions") ||
		flagWasSet(fs, "budget-amount") || flagWasSet(fs, "loc-invoice-details") || flagWasSet(fs, "campaigns-status") ||
		flagWasSet(fs, "campaigns-start-time") || flagWasSet(fs, "campaigns-end-time") || flagWasSet(fs, "ad-channel-type") ||
		flagWasSet(fs, "billing-event") || flagWasSet(fs, "supply-sources") {
		return shared.ValidationError("campaign creation override flags are not valid for adgroups scope imports")
	}
	return nil
}

func buildImportPlan(ctx context.Context, client *api.Client, cfg *config.Profile, fs *flag.FlagSet, root structureRoot, flags *importFlags) (importPlan, []*plannedCampaign, []*plannedAdgroup, error) {
	plan := importPlan{root: root, mapping: newMappingRoot(root.Scope)}
	bidWasSet := flagWasSet(fs, "bid")
	campaignOverrides, err := readJSONObjectArg(flags.campaignsFromJSON)
	if err != nil {
		return plan, nil, nil, err
	}
	adgroupOverrides, err := readJSONObjectArg(flags.adgroupsFromJSON)
	if err != nil {
		return plan, nil, nil, err
	}
	locInvoiceDetails, err := readJSONObjectArg(flags.locInvoiceDetails)
	if err != nil {
		return plan, nil, nil, err
	}
	if err := validateNoReadOnlyOrRelationshipFields("--campaigns-from-json", campaignOverrides, campaignReadOnlyFields, campaignRelationshipFields); err != nil {
		return plan, nil, nil, err
	}
	if err := validateNoReadOnlyOrRelationshipFields("--adgroups-from-json", adgroupOverrides, adgroupImportReadOnlyFields, adgroupRelationshipFields); err != nil {
		return plan, nil, nil, err
	}
	if err := validateNoReadOnlyOrRelationshipFields("--loc-invoice-details", locInvoiceDetails, map[string]bool{}, map[string]bool{}); err != nil {
		return plan, nil, nil, err
	}

	var targetCampaign map[string]any
	if root.Scope == scopeAdgroups {
		var raw json.RawMessage
		if err := client.Do(ctx, campaignsreq.GetRequest{CampaignID: flags.campaignID}, &raw); err != nil {
			return plan, nil, nil, err
		}
		targetCampaign, err = decodeDataObject(raw)
		if err != nil {
			return plan, nil, nil, err
		}
		plan.targetCampaign = targetCampaign
	}

	var campaignsPlan []*plannedCampaign
	var adgroupsPlan []*plannedAdgroup
	appTemplateVarsCache := map[string]map[string]string{}

	switch root.Scope {
	case scopeCampaigns:
		for _, node := range root.Campaigns {
			campaignPayload, err := resolveCampaignPayload(ctx, client, cfg, node.Campaign, campaignOverrides, locInvoiceDetails, flags, appTemplateVarsCache)
			if err != nil {
				return plan, nil, nil, err
			}
			campaignPlan := &plannedCampaign{
				source:      cloneMap(node.Campaign),
				payload:     campaignPayload,
				createdName: stringValue(campaignPayload["name"]),
			}

			for _, negative := range node.CampaignNegativeKeywords {
				if flags.noNegatives {
					break
				}
				payload, err := resolveNegativePayload(negative, flags)
				if err != nil {
					return plan, nil, nil, err
				}
				campaignPlan.campaignNegatives = append(campaignPlan.campaignNegatives, plannedEntity{
					source:      cloneMap(negative),
					payload:     payload,
					createdName: stringValue(payload["text"]),
				})
			}

			if flags.noAdgroups {
				campaignsPlan = append(campaignsPlan, campaignPlan)
				continue
			}
			for _, adgroupNode := range node.Adgroups {
				payload, err := resolveAdgroupPayload(ctx, client, cfg, adgroupNode.Adgroup, adgroupOverrides, flags, campaignPayload, appTemplateVarsCache)
				if err != nil {
					return plan, nil, nil, err
				}
				adgroupPlan := plannedAdgroup{
					source:      cloneMap(adgroupNode.Adgroup),
					payload:     payload,
					createdName: stringValue(payload["name"]),
				}
				if !flags.noNegatives {
					for _, negative := range adgroupNode.AdgroupNegativeKeywords {
						negPayload, err := resolveNegativePayload(negative, flags)
						if err != nil {
							return plan, nil, nil, err
						}
						adgroupPlan.adgroupNegatives = append(adgroupPlan.adgroupNegatives, plannedEntity{
							source:      cloneMap(negative),
							payload:     negPayload,
							createdName: stringValue(negPayload["text"]),
						})
					}
				}
				if !flags.noKeywords {
					for _, keyword := range adgroupNode.Keywords {
						kwPayload, err := resolveKeywordPayload(keyword, flags, payload["defaultBidAmount"], bidWasSet)
						if err != nil {
							return plan, nil, nil, err
						}
						adgroupPlan.keywords = append(adgroupPlan.keywords, plannedEntity{
							source:      cloneMap(keyword),
							payload:     kwPayload,
							createdName: stringValue(kwPayload["text"]),
						})
					}
				}
				campaignPlan.adgroups = append(campaignPlan.adgroups, adgroupPlan)
			}

			campaignsPlan = append(campaignsPlan, campaignPlan)
		}
	case scopeAdgroups:
		for _, node := range root.Adgroups {
			payload, err := resolveAdgroupPayload(ctx, client, cfg, node.Adgroup, adgroupOverrides, flags, targetCampaign, appTemplateVarsCache)
			if err != nil {
				return plan, nil, nil, err
			}
			adgroupPlan := &plannedAdgroup{
				source:      cloneMap(node.Adgroup),
				payload:     payload,
				createdName: stringValue(payload["name"]),
			}
			if !flags.noNegatives {
				for _, negative := range node.AdgroupNegativeKeywords {
					negPayload, err := resolveNegativePayload(negative, flags)
					if err != nil {
						return plan, nil, nil, err
					}
					adgroupPlan.adgroupNegatives = append(adgroupPlan.adgroupNegatives, plannedEntity{
						source:      cloneMap(negative),
						payload:     negPayload,
						createdName: stringValue(negPayload["text"]),
					})
				}
			}
			if !flags.noKeywords {
				for _, keyword := range node.Keywords {
					kwPayload, err := resolveKeywordPayload(keyword, flags, payload["defaultBidAmount"], bidWasSet)
					if err != nil {
						return plan, nil, nil, err
					}
					adgroupPlan.keywords = append(adgroupPlan.keywords, plannedEntity{
						source:      cloneMap(keyword),
						payload:     kwPayload,
						createdName: stringValue(kwPayload["text"]),
					})
				}
			}
			adgroupsPlan = append(adgroupsPlan, adgroupPlan)
		}
	}

	if err := validateImportCollisions(ctx, client, root.Scope, flags.campaignID, targetCampaign, campaignsPlan, adgroupsPlan); err != nil {
		return plan, nil, nil, err
	}
	if err := validateCPAUsage(root.Scope, targetCampaign, campaignsPlan, adgroupsPlan); err != nil {
		return plan, nil, nil, err
	}
	return plan, campaignsPlan, adgroupsPlan, nil
}

func resolveCampaignPayload(ctx context.Context, client *api.Client, cfg *config.Profile, source map[string]any, overrides map[string]any, locInvoiceDetails map[string]any, flags *importFlags, appTemplateVarsCache map[string]map[string]string) (map[string]any, error) {
	payload := normalizeCampaignForImport(source)
	payload = mergeMaps(payload, overrides)
	if err := resolveImportedScheduleFields(payload, cfg, "campaign"); err != nil {
		return nil, err
	}
	if flags.adamID != "" {
		payload["adamId"] = json.Number(flags.adamID)
	}
	if flags.dailyBudgetAmount != "" {
		money, err := shared.ParseMoneyFlag(flags.dailyBudgetAmount)
		if err != nil {
			return nil, err
		}
		payload["dailyBudgetAmount"] = money
	}
	if flagValue := strings.TrimSpace(flags.countriesOrRegions); flagValue != "" {
		payload["countriesOrRegions"] = parseCSVList(flagValue, true)
	}
	if flags.budgetAmount != "" {
		money, err := shared.ParseMoneyFlag(flags.budgetAmount)
		if err != nil {
			return nil, err
		}
		payload["budgetAmount"] = money
	}
	if flags.targetCpa != "" {
		money, err := shared.ParseMoneyFlag(flags.targetCpa)
		if err != nil {
			return nil, err
		}
		payload["targetCpa"] = money
	}
	if locInvoiceDetails != nil {
		payload["locInvoiceDetails"] = locInvoiceDetails
	}
	if flags.campaignsStatus != "" {
		status, err := shared.NormalizeStatus(flags.campaignsStatus, "ENABLED")
		if err != nil {
			return nil, err
		}
		payload["status"] = status
	}
	if flags.campaignsStartTime != "" {
		value, err := shared.ResolveDateTimeFlag(flags.campaignsStartTime, cfg)
		if err != nil {
			return nil, err
		}
		payload["startTime"] = value
	}
	if flags.campaignsEndTime != "" {
		value, err := shared.ResolveDateTimeFlag(flags.campaignsEndTime, cfg)
		if err != nil {
			return nil, err
		}
		payload["endTime"] = value
	}
	if err := validateScheduleOrdering(payload, "campaign"); err != nil {
		return nil, err
	}
	if flags.adChannelType != "" {
		value, err := shared.NormalizeAdChannelType(flags.adChannelType)
		if err != nil {
			return nil, err
		}
		payload["adChannelType"] = value
	}
	if flags.billingEvent != "" {
		value, err := shared.NormalizeBillingEvent(flags.billingEvent)
		if err != nil {
			return nil, err
		}
		payload["billingEvent"] = value
	}
	if strings.TrimSpace(flags.supplySources) != "" {
		payload["supplySources"] = parseCSVList(flags.supplySources, true)
	}
	payload = applyCampaignCreateDefaults(payload)
	nameTemplate := mapStringValue(source, "name")
	extraVars, err := resolveCampaignTemplateVars(ctx, client, payload, nameTemplate, flags.campaignsName, appTemplateVarsCache)
	if err != nil {
		return nil, err
	}
	renderedName, err := applyImportedNameTemplate(nameTemplate, payload, extraVars)
	if err != nil {
		return nil, err
	}
	payload["name"] = renderedName
	renderedName, err = applyNameTemplate(renderedName, payload, flags.campaignsName, flags.campaignsNamePattern, flags.allowUnmatchedNamePattern, extraVars)
	if err != nil {
		return nil, err
	}
	payload["name"] = renderedName
	if err := ensureRequiredFieldsWithFlags("campaign", payload, campaignRequiredFields, campaignRequiredFieldFlags); err != nil {
		return nil, err
	}
	if err := campaignscli.ValidatePayload(ctx, client, mustMarshalRaw(payload), ""); err != nil {
		return nil, err
	}
	return payload, nil
}

func resolveAdgroupPayload(ctx context.Context, client *api.Client, cfg *config.Profile, source map[string]any, overrides map[string]any, flags *importFlags, campaign map[string]any, appTemplateVarsCache map[string]map[string]string) (map[string]any, error) {
	payload := normalizeAdgroupForImport(source)
	payload = mergeMaps(payload, overrides)
	if err := resolveImportedScheduleFields(payload, cfg, "adgroup"); err != nil {
		return nil, err
	}
	fields := adgroupscli.Fields{
		Status:                 flags.adgroupsStatus,
		StartTime:              flags.adgroupsStartTime,
		EndTime:                flags.adgroupsEndTime,
		AutomatedKeywordsOptIn: flags.automatedKeywordsOptIn,
	}
	if flags.defaultBid != "" {
		money, err := shared.ParseMoneyFlag(flags.defaultBid)
		if err != nil {
			return nil, err
		}
		fields.DefaultBidAmount = money
	}
	if flags.cpaGoal != "" {
		money, err := shared.ParseMoneyFlag(flags.cpaGoal)
		if err != nil {
			return nil, err
		}
		fields.CPAGoal = money
	}
	if err := adgroupscli.ApplyFields(payload, fields, cfg, adgroupscli.FieldLabels{
		StartTime: "--adgroups-start-time",
		EndTime:   "--adgroups-end-time",
	}); err != nil {
		return nil, err
	}
	if err := adgroupscli.EnsureCreateStartTime(payload, cfg); err != nil {
		return nil, err
	}
	if err := validateScheduleOrdering(payload, "adgroup"); err != nil {
		return nil, err
	}
	payload = applyAdgroupCreateDefaults(payload)

	extraVars := map[string]string{}
	if campaign != nil {
		for key, value := range buildTemplateVariables(campaign) {
			extraVars["CAMPAIGN_"+key] = value
		}
		if name := stringValue(campaign["name"]); strings.TrimSpace(name) != "" {
			extraVars["campaignName"] = name
			extraVars["CAMPAIGN_NAME"] = name
		}
		if id := stringValue(campaign["id"]); strings.TrimSpace(id) != "" && id != "<nil>" {
			extraVars["campaignId"] = id
			extraVars["CAMPAIGN_ID"] = id
		}
		appVars, err := resolveAppTemplateVars(ctx, client, stringValue(campaign["adamId"]), mapStringValue(source, "name"), flags.adgroupsName, appTemplateVarsCache)
		if err != nil {
			return nil, err
		}
		for key, value := range appVars {
			extraVars[key] = value
		}
	}
	renderedName, err := applyImportedNameTemplate(mapStringValue(source, "name"), payload, extraVars)
	if err != nil {
		return nil, err
	}
	payload["name"] = renderedName
	renderedName, err = applyNameTemplate(renderedName, payload, flags.adgroupsName, flags.adgroupsNamePattern, flags.allowUnmatchedNamePattern, extraVars)
	if err != nil {
		return nil, err
	}
	payload["name"] = renderedName
	if err := ensureRequiredFieldsWithFlags("adgroup", payload, adgroupRequiredFields, adgroupRequiredFieldFlags); err != nil {
		return nil, err
	}
	body := mustMarshalRaw(payload)
	hasCPAGoal, err := adgroupscli.PayloadHasCPAGoal(body)
	if err != nil {
		return nil, err
	}
	if err := adgroupscli.ValidatePayload(ctx, client, stringValue(campaign["id"]), stringValue(campaign["adChannelType"]), body, "cpaGoal", hasCPAGoal); err != nil {
		return nil, err
	}
	return payload, nil
}

func resolveCampaignTemplateVars(ctx context.Context, client *api.Client, payload map[string]any, sourceNameTemplate string, overrideNameTemplate string, appTemplateVarsCache map[string]map[string]string) (map[string]string, error) {
	return resolveAppTemplateVars(ctx, client, stringValue(payload["adamId"]), sourceNameTemplate, overrideNameTemplate, appTemplateVarsCache)
}

func resolveAppTemplateVars(ctx context.Context, client *api.Client, adamID string, sourceTemplate string, overrideTemplate string, cache map[string]map[string]string) (map[string]string, error) {
	if !templatesReferenceAnyVariable([]string{sourceTemplate, overrideTemplate}, "appName", "APP_NAME", "appNameShort", "APP_NAME_SHORT") {
		return nil, nil
	}
	adamID = strings.TrimSpace(adamID)
	if adamID == "" || adamID == "<nil>" {
		return nil, shared.ValidationError("appName template variable requires adamId")
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

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func resolveImportedScheduleFields(payload map[string]any, cfg *config.Profile, entity string) error {
	for _, key := range []string{"startTime", "endTime"} {
		raw, ok := payload[key]
		if !ok || raw == nil {
			continue
		}
		value, ok := raw.(string)
		if !ok {
			return shared.ValidationErrorf("%s %s must be a string", entity, key)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			delete(payload, key)
			continue
		}
		resolved, err := shared.ResolveDateTimeFlag(value, cfg)
		if err != nil {
			return shared.ValidationErrorf("invalid %s %s: %v", entity, key, err)
		}
		payload[key] = resolved
	}
	return nil
}

func validateScheduleOrdering(payload map[string]any, entity string) error {
	startRaw, hasStart := scheduleStringValue(payload, "startTime")
	endRaw, hasEnd := scheduleStringValue(payload, "endTime")
	if !hasStart || !hasEnd {
		return nil
	}
	start, err := time.Parse("2006-01-02T15:04:05.000", startRaw)
	if err != nil {
		return shared.ValidationErrorf("invalid %s startTime: %v", entity, err)
	}
	end, err := time.Parse("2006-01-02T15:04:05.000", endRaw)
	if err != nil {
		return shared.ValidationErrorf("invalid %s endTime: %v", entity, err)
	}
	if end.Before(start) {
		return shared.ValidationErrorf("%s endTime must not be earlier than startTime", entity)
	}
	return nil
}

func scheduleStringValue(payload map[string]any, key string) (string, bool) {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return "", false
	}
	value, ok := raw.(string)
	if !ok {
		return "", false
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return value, true
}

func resolveKeywordPayload(source map[string]any, flags *importFlags, defaultBid any, bidWasSet bool) (map[string]any, error) {
	payload := normalizeKeywordForImport(source)
	fields := keywordscli.Fields{
		Status:    flags.keywordsStatus,
		MatchType: flags.matchType,
	}
	if bidWasSet {
		if flags.bid == "" {
			delete(payload, "bidAmount")
		} else {
			money, err := shared.ParseMoneyFlag(flags.bid)
			if err != nil {
				return nil, err
			}
			fields.Bid = money
		}
	}
	if err := keywordscli.ApplyFields(payload, fields); err != nil {
		return nil, err
	}
	if moneyEquals(payload["bidAmount"], defaultBid) {
		delete(payload, "bidAmount")
	}
	if err := ensureRequiredFields("keyword", payload, keywordRequiredFields); err != nil {
		return nil, err
	}
	if err := keywordscli.ValidatePayload(mustMarshalRaw([]map[string]any{payload})); err != nil {
		return nil, err
	}
	return payload, nil
}

func resolveNegativePayload(source map[string]any, flags *importFlags) (map[string]any, error) {
	payload := normalizeNegativeForImport(source)
	if flags.negativesStatus != "" {
		value, err := shared.NormalizeStatus(flags.negativesStatus, "ACTIVE")
		if err != nil {
			return nil, err
		}
		payload["status"] = value
	}
	if err := ensureRequiredFields("negative keyword", payload, negativeRequiredFields); err != nil {
		return nil, err
	}
	return payload, nil
}

func parseCSVList(value string, uppercase bool) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if uppercase {
			part = strings.ToUpper(part)
		}
		out = append(out, part)
	}
	return out
}

func validateImportCollisions(ctx context.Context, client *api.Client, scope, campaignID string, targetCampaign map[string]any, campaignsPlan []*plannedCampaign, adgroupsPlan []*plannedAdgroup) error {
	var collisions []string
	if scope == scopeCampaigns {
		var names []string
		for _, plan := range campaignsPlan {
			names = append(names, plan.createdName)
		}
		for _, name := range collectDuplicateNames(names) {
			collisions = append(collisions, fmt.Sprintf("duplicate campaign name in import batch: %s", name))
		}

		existing, err := fetchAllList(ctx, client, campaignsreq.ListRequest{})
		if err != nil {
			return err
		}
		existingNames := map[string]string{}
		for _, campaign := range existing {
			existingNames[canonicalName(stringValue(campaign["name"]))] = stringValue(campaign["name"])
		}
		for _, plan := range campaignsPlan {
			if match, ok := existingNames[canonicalName(plan.createdName)]; ok {
				collisions = append(collisions, fmt.Sprintf("campaign name already exists: %s", match))
			}
		}
		for _, campaign := range campaignsPlan {
			if len(campaign.adgroups) == 0 {
				continue
			}
			var names []string
			for _, adgroup := range campaign.adgroups {
				names = append(names, adgroup.createdName)
			}
			for _, name := range collectDuplicateNames(names) {
				collisions = append(collisions, fmt.Sprintf("duplicate adgroup name in import batch for campaign %s: %s", campaign.createdName, name))
			}
		}
	} else {
		var names []string
		for _, adgroup := range adgroupsPlan {
			names = append(names, adgroup.createdName)
		}
		for _, name := range collectDuplicateNames(names) {
			collisions = append(collisions, fmt.Sprintf("duplicate adgroup name in import batch for campaign %s: %s", stringValue(targetCampaign["name"]), name))
		}

		existing, err := fetchAllList(ctx, client, adgroupsreq.ListRequest{CampaignID: campaignID})
		if err != nil {
			return err
		}
		existingNames := map[string]string{}
		for _, adgroup := range existing {
			existingNames[canonicalName(stringValue(adgroup["name"]))] = stringValue(adgroup["name"])
		}
		for _, adgroup := range adgroupsPlan {
			if match, ok := existingNames[canonicalName(adgroup.createdName)]; ok {
				collisions = append(collisions, fmt.Sprintf("adgroup name already exists in campaign %s: %s", stringValue(targetCampaign["name"]), match))
			}
		}
	}

	if len(collisions) == 0 {
		return nil
	}
	sort.Strings(collisions)
	return shared.ValidationError(strings.Join(compactSortedStrings(collisions), "\n"))
}

func validateCPAUsage(scope string, targetCampaign map[string]any, campaignsPlan []*plannedCampaign, adgroupsPlan []*plannedAdgroup) error {
	if scope == scopeCampaigns {
		for _, campaign := range campaignsPlan {
			if !strings.EqualFold(stringValue(campaign.payload["adChannelType"]), "SEARCH") {
				for _, adgroup := range campaign.adgroups {
					if _, ok := adgroup.payload["cpaGoal"]; ok {
						return shared.ValidationErrorf("campaign %q is not SEARCH but adgroup %q sets cpaGoal", campaign.createdName, adgroup.createdName)
					}
				}
			}
		}
		return nil
	}
	if !strings.EqualFold(stringValue(targetCampaign["adChannelType"]), "SEARCH") {
		for _, adgroup := range adgroupsPlan {
			if _, ok := adgroup.payload["cpaGoal"]; ok {
				return shared.ValidationErrorf("destination campaign %q is not SEARCH but adgroup %q sets cpaGoal", stringValue(targetCampaign["name"]), adgroup.createdName)
			}
		}
	}
	return nil
}

func assignMockIDs(campaignsPlan []*plannedCampaign, adgroupsPlan []*plannedAdgroup) {
	var ids mockIDGenerator
	for _, campaign := range campaignsPlan {
		campaign.createdID = ids.Next()
		for i := range campaign.campaignNegatives {
			campaign.campaignNegatives[i].createdID = ids.Next()
		}
		for j := range campaign.adgroups {
			campaign.adgroups[j].createdID = ids.Next()
			for i := range campaign.adgroups[j].adgroupNegatives {
				campaign.adgroups[j].adgroupNegatives[i].createdID = ids.Next()
			}
			for i := range campaign.adgroups[j].keywords {
				campaign.adgroups[j].keywords[i].createdID = ids.Next()
			}
		}
	}
	for _, adgroup := range adgroupsPlan {
		adgroup.createdID = ids.Next()
		for i := range adgroup.adgroupNegatives {
			adgroup.adgroupNegatives[i].createdID = ids.Next()
		}
		for i := range adgroup.keywords {
			adgroup.keywords[i].createdID = ids.Next()
		}
	}
}

func executeImportPlan(ctx context.Context, client *api.Client, scope string, campaignID string, campaignsPlan []*plannedCampaign, adgroupsPlan []*plannedAdgroup) error {
	if scope == scopeCampaigns {
		for _, campaign := range campaignsPlan {
			if err := createCampaign(ctx, client, campaign); err != nil {
				return err
			}
			if len(campaign.campaignNegatives) > 0 {
				if err := createCampaignNegatives(ctx, client, campaign); err != nil {
					return err
				}
			}
			for i := range campaign.adgroups {
				if err := createAdgroup(ctx, client, campaign, &campaign.adgroups[i]); err != nil {
					return err
				}
				if len(campaign.adgroups[i].adgroupNegatives) > 0 {
					if err := createAdgroupNegatives(ctx, client, campaign, &campaign.adgroups[i]); err != nil {
						return err
					}
				}
				if len(campaign.adgroups[i].keywords) > 0 {
					if err := createKeywords(ctx, client, campaign, &campaign.adgroups[i]); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	for _, adgroup := range adgroupsPlan {
		campaign := &plannedCampaign{createdID: json.Number(campaignID), createdName: ""}
		if err := createAdgroup(ctx, client, campaign, adgroup); err != nil {
			return err
		}
		if len(adgroup.adgroupNegatives) > 0 {
			if err := createAdgroupNegatives(ctx, client, campaign, adgroup); err != nil {
				return err
			}
		}
		if len(adgroup.keywords) > 0 {
			if err := createKeywords(ctx, client, campaign, adgroup); err != nil {
				return err
			}
		}
	}
	return nil
}

func createCampaign(ctx context.Context, client *api.Client, campaign *plannedCampaign) error {
	var raw json.RawMessage
	if err := client.Do(ctx, campaignsreq.CreateRequest{RawBody: mustMarshalRaw(campaign.payload)}, &raw); err != nil {
		markCampaignFailed(campaign, err)
		return err
	}
	obj, err := decodeDataObject(raw)
	if err != nil {
		markCampaignFailed(campaign, err)
		return err
	}
	campaign.createdID = obj["id"]
	if name := stringValue(obj["name"]); strings.TrimSpace(name) != "" {
		campaign.createdName = name
	}
	campaign.status = importStatusCreated
	campaign.errorMessage = ""
	return nil
}

func createCampaignNegatives(ctx context.Context, client *api.Client, campaign *plannedCampaign) error {
	payloads := make([]map[string]any, 0, len(campaign.campaignNegatives))
	for _, item := range campaign.campaignNegatives {
		payloads = append(payloads, item.payload)
	}
	var raw json.RawMessage
	if err := client.Do(ctx, negcampaignreq.CreateRequest{
		CampaignID: stringValue(campaign.createdID),
		RawBody:    mustMarshalRaw(payloads),
	}, &raw); err != nil {
		markEntitiesFailed(campaign.campaignNegatives, err)
		return err
	}
	rows, err := decodeDataList(raw)
	if err != nil {
		markEntitiesFailed(campaign.campaignNegatives, err)
		return err
	}
	for i := range campaign.campaignNegatives {
		if i < len(rows) {
			campaign.campaignNegatives[i].createdID = rows[i]["id"]
			if text := stringValue(rows[i]["text"]); strings.TrimSpace(text) != "" {
				campaign.campaignNegatives[i].createdName = text
			}
		}
		campaign.campaignNegatives[i].status = importStatusCreated
		campaign.campaignNegatives[i].errorMessage = ""
	}
	return nil
}

func createAdgroup(ctx context.Context, client *api.Client, campaign *plannedCampaign, adgroup *plannedAdgroup) error {
	var raw json.RawMessage
	if err := client.Do(ctx, adgroupsreq.CreateRequest{
		CampaignID: stringValue(campaign.createdID),
		RawBody:    mustMarshalRaw(adgroup.payload),
	}, &raw); err != nil {
		markAdgroupFailed(adgroup, err)
		return err
	}
	obj, err := decodeDataObject(raw)
	if err != nil {
		markAdgroupFailed(adgroup, err)
		return err
	}
	adgroup.createdID = obj["id"]
	if name := stringValue(obj["name"]); strings.TrimSpace(name) != "" {
		adgroup.createdName = name
	}
	adgroup.status = importStatusCreated
	adgroup.errorMessage = ""
	return nil
}

func createAdgroupNegatives(ctx context.Context, client *api.Client, campaign *plannedCampaign, adgroup *plannedAdgroup) error {
	payloads := make([]map[string]any, 0, len(adgroup.adgroupNegatives))
	for _, item := range adgroup.adgroupNegatives {
		payloads = append(payloads, item.payload)
	}
	var raw json.RawMessage
	if err := client.Do(ctx, negadgroupreq.CreateRequest{
		CampaignID: stringValue(campaign.createdID),
		AdGroupID:  stringValue(adgroup.createdID),
		RawBody:    mustMarshalRaw(payloads),
	}, &raw); err != nil {
		markEntitiesFailed(adgroup.adgroupNegatives, err)
		return err
	}
	rows, err := decodeDataList(raw)
	if err != nil {
		markEntitiesFailed(adgroup.adgroupNegatives, err)
		return err
	}
	for i := range adgroup.adgroupNegatives {
		if i < len(rows) {
			adgroup.adgroupNegatives[i].createdID = rows[i]["id"]
			if text := stringValue(rows[i]["text"]); strings.TrimSpace(text) != "" {
				adgroup.adgroupNegatives[i].createdName = text
			}
		}
		adgroup.adgroupNegatives[i].status = importStatusCreated
		adgroup.adgroupNegatives[i].errorMessage = ""
	}
	return nil
}

func createKeywords(ctx context.Context, client *api.Client, campaign *plannedCampaign, adgroup *plannedAdgroup) error {
	payloads := make([]map[string]any, 0, len(adgroup.keywords))
	for _, item := range adgroup.keywords {
		payloads = append(payloads, item.payload)
	}
	var raw json.RawMessage
	if err := client.Do(ctx, keywordsreq.CreateRequest{
		CampaignID: stringValue(campaign.createdID),
		AdGroupID:  stringValue(adgroup.createdID),
		RawBody:    mustMarshalRaw(payloads),
	}, &raw); err != nil {
		markEntitiesFailed(adgroup.keywords, err)
		return err
	}
	rows, err := decodeDataList(raw)
	if err != nil {
		markEntitiesFailed(adgroup.keywords, err)
		return err
	}
	for i := range adgroup.keywords {
		if i < len(rows) {
			adgroup.keywords[i].createdID = rows[i]["id"]
			if text := stringValue(rows[i]["text"]); strings.TrimSpace(text) != "" {
				adgroup.keywords[i].createdName = text
			}
		}
		adgroup.keywords[i].status = importStatusCreated
		adgroup.keywords[i].errorMessage = ""
	}
	return nil
}

func buildMapping(scope string, campaignsPlan []*plannedCampaign, adgroupsPlan []*plannedAdgroup, includeExecutionStatus bool) mappingRoot {
	root := newMappingRoot(scope)
	for _, campaign := range campaignsPlan {
		campaignCreated := createdNameRef(campaign.createdID, campaign.createdName)
		campaignStatus := ""
		campaignError := ""
		if includeExecutionStatus {
			campaignStatus = statusOrDefault(campaign.status, true)
			campaignError = errorForStatus(campaign.errorMessage, campaign.status, true)
			campaignCreated = createdNameRefForStatus(campaign.createdID, campaign.createdName, campaignStatus)
		}
		node := mappingCampaignNode{
			Campaign: entityMapping{
				Source:  sourceNameRef(campaign.source),
				Created: campaignCreated,
				Status:  campaignStatus,
				Error:   campaignError,
			},
		}
		for _, negative := range campaign.campaignNegatives {
			created := createdTextRef(negative.createdID, negative.createdName)
			status := ""
			errText := ""
			if includeExecutionStatus {
				status = statusOrDefault(negative.status, true)
				errText = errorForStatus(negative.errorMessage, negative.status, true)
				created = createdTextRefForStatus(negative.createdID, negative.createdName, status)
			}
			node.CampaignNegativeKeywords = append(node.CampaignNegativeKeywords, entityMapping{
				Source:  sourceTextRef(negative.source),
				Created: created,
				Status:  status,
				Error:   errText,
			})
		}
		for _, adgroup := range campaign.adgroups {
			adgroupCreated := createdNameRef(adgroup.createdID, adgroup.createdName)
			adgroupStatus := ""
			adgroupError := ""
			if includeExecutionStatus {
				adgroupStatus = statusOrDefault(adgroup.status, true)
				adgroupError = errorForStatus(adgroup.errorMessage, adgroup.status, true)
				adgroupCreated = createdNameRefForStatus(adgroup.createdID, adgroup.createdName, adgroupStatus)
			}
			child := mappingAdGroupNode{
				Adgroup: entityMapping{
					Source:  sourceNameRef(adgroup.source),
					Created: adgroupCreated,
					Status:  adgroupStatus,
					Error:   adgroupError,
				},
			}
			for _, negative := range adgroup.adgroupNegatives {
				created := createdTextRef(negative.createdID, negative.createdName)
				status := ""
				errText := ""
				if includeExecutionStatus {
					status = statusOrDefault(negative.status, true)
					errText = errorForStatus(negative.errorMessage, negative.status, true)
					created = createdTextRefForStatus(negative.createdID, negative.createdName, status)
				}
				child.AdgroupNegativeKeywords = append(child.AdgroupNegativeKeywords, entityMapping{
					Source:  sourceTextRef(negative.source),
					Created: created,
					Status:  status,
					Error:   errText,
				})
			}
			for _, keyword := range adgroup.keywords {
				created := createdTextRef(keyword.createdID, keyword.createdName)
				status := ""
				errText := ""
				if includeExecutionStatus {
					status = statusOrDefault(keyword.status, true)
					errText = errorForStatus(keyword.errorMessage, keyword.status, true)
					created = createdTextRefForStatus(keyword.createdID, keyword.createdName, status)
				}
				child.Keywords = append(child.Keywords, entityMapping{
					Source:  sourceTextRef(keyword.source),
					Created: created,
					Status:  status,
					Error:   errText,
				})
			}
			node.Adgroups = append(node.Adgroups, child)
		}
		root.Campaigns = append(root.Campaigns, node)
	}
	for _, adgroup := range adgroupsPlan {
		adgroupCreated := createdNameRef(adgroup.createdID, adgroup.createdName)
		adgroupStatus := ""
		adgroupError := ""
		if includeExecutionStatus {
			adgroupStatus = statusOrDefault(adgroup.status, true)
			adgroupError = errorForStatus(adgroup.errorMessage, adgroup.status, true)
			adgroupCreated = createdNameRefForStatus(adgroup.createdID, adgroup.createdName, adgroupStatus)
		}
		node := mappingAdGroupNode{
			Adgroup: entityMapping{
				Source:  sourceNameRef(adgroup.source),
				Created: adgroupCreated,
				Status:  adgroupStatus,
				Error:   adgroupError,
			},
		}
		for _, negative := range adgroup.adgroupNegatives {
			created := createdTextRef(negative.createdID, negative.createdName)
			status := ""
			errText := ""
			if includeExecutionStatus {
				status = statusOrDefault(negative.status, true)
				errText = errorForStatus(negative.errorMessage, negative.status, true)
				created = createdTextRefForStatus(negative.createdID, negative.createdName, status)
			}
			node.AdgroupNegativeKeywords = append(node.AdgroupNegativeKeywords, entityMapping{
				Source:  sourceTextRef(negative.source),
				Created: created,
				Status:  status,
				Error:   errText,
			})
		}
		for _, keyword := range adgroup.keywords {
			created := createdTextRef(keyword.createdID, keyword.createdName)
			status := ""
			errText := ""
			if includeExecutionStatus {
				status = statusOrDefault(keyword.status, true)
				errText = errorForStatus(keyword.errorMessage, keyword.status, true)
				created = createdTextRefForStatus(keyword.createdID, keyword.createdName, status)
			}
			node.Keywords = append(node.Keywords, entityMapping{
				Source:  sourceTextRef(keyword.source),
				Created: created,
				Status:  status,
				Error:   errText,
			})
		}
		root.Adgroups = append(root.Adgroups, node)
	}
	return root
}

func markCampaignFailed(campaign *plannedCampaign, err error) {
	campaign.status = importStatusFailed
	campaign.errorMessage = err.Error()
}

func markAdgroupFailed(adgroup *plannedAdgroup, err error) {
	adgroup.status = importStatusFailed
	adgroup.errorMessage = err.Error()
}

func markEntitiesFailed(items []plannedEntity, err error) {
	for i := range items {
		items[i].status = importStatusFailed
		items[i].errorMessage = err.Error()
	}
}

func statusOrDefault(status string, includeExecutionStatus bool) string {
	if !includeExecutionStatus {
		return ""
	}
	if strings.TrimSpace(status) == "" {
		return importStatusNotAttempted
	}
	return status
}

func errorForStatus(message, status string, includeExecutionStatus bool) string {
	if !includeExecutionStatus || status != importStatusFailed {
		return ""
	}
	return message
}

func sourceNameRef(source map[string]any) map[string]any {
	out := map[string]any{"id": source["id"], "name": source["name"]}
	if out["id"] == nil {
		delete(out, "id")
	}
	return out
}

func createdNameRef(id any, name string) map[string]any {
	out := map[string]any{"id": id, "name": name}
	if out["id"] == nil {
		delete(out, "id")
	}
	return out
}

func createdNameRefForStatus(id any, name string, status string) map[string]any {
	switch status {
	case importStatusCreated:
		out := createdNameRef(id, name)
		if len(out) == 0 {
			return nil
		}
		if out["id"] == nil {
			delete(out, "id")
		}
		if strings.TrimSpace(stringValue(out["name"])) == "" || stringValue(out["name"]) == "<nil>" {
			delete(out, "name")
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case importStatusFailed:
		if id == nil {
			return nil
		}
		return createdNameRef(id, name)
	default:
		return nil
	}
}

func sourceTextRef(source map[string]any) map[string]any {
	out := map[string]any{"id": source["id"], "text": source["text"]}
	if out["id"] == nil {
		delete(out, "id")
	}
	return out
}

func createdTextRef(id any, text string) map[string]any {
	out := map[string]any{"id": id, "text": text}
	if out["id"] == nil {
		delete(out, "id")
	}
	return out
}

func createdTextRefForStatus(id any, text string, status string) map[string]any {
	switch status {
	case importStatusCreated:
		out := createdTextRef(id, text)
		if len(out) == 0 {
			return nil
		}
		if out["id"] == nil {
			delete(out, "id")
		}
		if strings.TrimSpace(stringValue(out["text"])) == "" || stringValue(out["text"]) == "<nil>" {
			delete(out, "text")
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case importStatusFailed:
		if id == nil {
			return nil
		}
		return createdTextRef(id, text)
	default:
		return nil
	}
}
