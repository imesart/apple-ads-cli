package reports

import (
	"context"
	"encoding/json"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/reports"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

// Command returns the reports command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "reports",
		ShortUsage: "aads reports <subcommand>",
		ShortHelp:  "Generate reports.",
		Subcommands: []*ffcli.Command{
			campaignsCmd(),
			adgroupsCmd(),
			keywordsCmd(),
			searchtermsCmd(),
			adsCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

var campaignParent = []shared.ParentFlag{
	{Name: "campaign-id", Usage: "Campaign ID", Required: true},
}

// campaignWithOptionalAdGroup is used by keywords and searchterms
// which accept --campaign-id (required) and optionally --adgroup-id.
var campaignWithOptionalAdGroup = []shared.ParentFlag{
	{Name: "campaign-id", Usage: "Campaign ID", Required: true},
	{Name: "adgroup-id", Usage: "Ad Group ID (optional, changes report scope)", Required: false},
}

const sortableMetricsHelp = `Sortable fields:
  impressions, taps, localSpend, tapInstalls, tapNewDownloads,
  tapRedownloads, tapInstallCPI, avgCPT, avgCPM, ttr,
  tapInstallRate, totalInstalls, totalNewDownloads, totalRedownloads,
  totalAvgCPI, totalInstallRate`

func campaignsCmd() *ffcli.Command {
	return shared.BuildReportCommand(shared.ReportCommandConfig{
		Name:       "campaigns",
		ShortUsage: "aads reports campaigns --start DATE --end DATE --sort FIELD:ORDER",
		ShortHelp:  "Campaign-level reports.",
		LongHelp: sortableMetricsHelp + `

Additional filterable fields:
  campaignId, orgId, campaignName, campaignStatus, servingStatus,
  displayStatus, adChannelType, deleted, appAdamId, appAppName,
  countriesOrRegions, deviceClass, gender, ageRange, countryOrRegion,
  adminArea, locality, dailyBudget.amount, dailyBudget.currency,
  totalBudget.amount, totalBudget.currency

Local filter examples (repeatable, combined with AND):
  --filter "campaignId=123"
  --filter "localSpend > 10"
  --filter "localSpend BETWEEN [10, 50]"`,
		DefaultSort:      "impressions:desc",
		DefaultNoMetrics: true,
		Exec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, body json.RawMessage) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, reports.CampaignsRequest{RawBody: body}, &result)
			if err != nil {
				return nil, err
			}
			return flattenReportResponse(result)
		},
	})
}

func adgroupsCmd() *ffcli.Command {
	return shared.BuildReportCommand(shared.ReportCommandConfig{
		Name:       "adgroups",
		ShortUsage: "aads reports adgroups --campaign-id ID [--adgroup-id AGID] --start DATE --end DATE --sort FIELD:ORDER",
		ShortHelp:  "Ad group-level reports.",
		LongHelp: sortableMetricsHelp + `

Additional filterable fields:
  campaignId, adGroupId, orgId, adGroupName, adGroupStatus,
  adGroupServingStatus, adGroupDisplayStatus, deleted,
  automatedKeywordsOptIn, defaultBidAmount.amount, defaultBidAmount.currency,
  cpaGoal.amount, cpaGoal.currency, startTime, endTime, modificationTime,
  deviceClass, gender, ageRange, countryOrRegion, adminArea, locality

Local filter examples (repeatable, combined with AND):
  --filter "adGroupId=5001"
  --filter "localSpend > 10"
  --filter "adGroupName STARTSWITH Brand"

Selector shortcut:
  --adgroup-id 5001
    Adds selector condition: adGroupId EQUALS 5001
    Supports --adgroup-id - for stdin pipelines.`,
		ParentFlags: campaignParent,
		SelectorShortcuts: []shared.ReportSelectorShortcut{
			{FlagName: "adgroup-id", Usage: "Ad Group ID filter shortcut", SelectorField: "adGroupId"},
		},
		DefaultSort:      "impressions:desc",
		DefaultNoMetrics: true,
		Exec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, body json.RawMessage) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, reports.AdGroupsRequest{
				CampaignID: parentIDs["campaign-id"],
				RawBody:    body,
			}, &result)
			if err != nil {
				return nil, err
			}
			return flattenReportResponse(result)
		},
	})
}

func keywordsCmd() *ffcli.Command {
	return shared.BuildReportCommand(shared.ReportCommandConfig{
		Name:       "keywords",
		ShortUsage: "aads reports keywords --campaign-id ID [--adgroup-id AGID] --start DATE --end DATE --sort FIELD:ORDER",
		ShortHelp:  "Keyword-level reports.",
		LongHelp: sortableMetricsHelp + `

Additional filterable fields:
  campaignId, adGroupId, keywordId, orgId, adGroupName, keyword,
  keywordStatus, keywordDisplayStatus, matchType, deleted, adGroupDeleted,
  bidAmount.amount, bidAmount.currency, modificationTime,
  deviceClass, gender, ageRange, countryOrRegion, adminArea, locality

Selector shortcut:
  --adgroup-id 5001
    Adds selector condition: adGroupId EQUALS 5001
    Supports --adgroup-id - for stdin pipelines.`,
		ParentFlags: campaignWithOptionalAdGroup,
		DefaultSort: "impressions:desc",
		Exec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, body json.RawMessage) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, reports.KeywordsRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  parentIDs["adgroup-id"],
				RawBody:    body,
			}, &result)
			if err != nil {
				return nil, err
			}
			return flattenReportResponse(result)
		},
	})
}

func searchtermsCmd() *ffcli.Command {
	return shared.BuildReportCommand(shared.ReportCommandConfig{
		Name:       "searchterms",
		ShortUsage: "aads reports searchterms --campaign-id ID [--adgroup-id AGID] --start DATE --end DATE --sort FIELD:ORDER",
		ShortHelp:  "Search term-level reports.",
		LongHelp: sortableMetricsHelp + `

Additional filterable fields:
  campaignId, adGroupId, keywordId, orgId, adGroupName, keyword,
  keywordStatus, keywordDisplayStatus, matchType, deleted, adGroupDeleted,
  bidAmount.amount, bidAmount.currency, searchTermSource, searchTermText,
  modificationTime, deviceClass, gender, ageRange, countryOrRegion,
  adminArea, locality

Selector shortcut:
  --adgroup-id 5001
    Adds selector condition: adGroupId EQUALS 5001
    Supports --adgroup-id - for stdin pipelines.`,
		ParentFlags:   campaignWithOptionalAdGroup,
		DefaultSort:   "impressions:desc",
		ForceTimezone: "ORTZ",
		Exec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, body json.RawMessage) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, reports.SearchTermsRequest{
				CampaignID: parentIDs["campaign-id"],
				AdGroupID:  parentIDs["adgroup-id"],
				RawBody:    body,
			}, &result)
			if err != nil {
				return nil, err
			}
			return flattenReportResponse(result)
		},
	})
}

func adsCmd() *ffcli.Command {
	return shared.BuildReportCommand(shared.ReportCommandConfig{
		Name:       "ads",
		ShortUsage: "aads reports ads --campaign-id ID --start DATE --end DATE --sort FIELD:ORDER",
		ShortHelp:  "Ad-level reports.",
		LongHelp: sortableMetricsHelp + `

Additional filterable fields:
  campaignId, adGroupId, adId, creativeId, orgId, productPageId,
  adName, creativeType, status, adDisplayStatus, adServingStateReasons,
  language, deleted, creationTime, modificationTime, deviceClass,
  gender, ageRange, countryOrRegion, adminArea, locality`,
		ParentFlags: campaignParent,
		DefaultSort: "impressions:desc",
		Exec: func(ctx context.Context, client *api.Client, parentIDs map[string]string, body json.RawMessage) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, reports.AdsRequest{
				CampaignID: parentIDs["campaign-id"],
				RawBody:    body,
			}, &result)
			if err != nil {
				return nil, err
			}
			return flattenReportResponse(result)
		},
	})
}
