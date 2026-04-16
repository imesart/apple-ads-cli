package campaigns

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/types"
)

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	dataFile := fs.String("from-json", "", "Path to JSON body file (or - for stdin)")
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	shared.BindForceFlag(fs)
	nameFlag := fs.String("name", "", "Campaign name (accepts template variables like %(fieldName) or %(FIELD_NAME)) (required)")
	adamID := fs.String("adam-id", "", "App Store app ID (required)")
	dailyBudget := fs.String("daily-budget-amount", "", "Daily budget (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency) (required)")
	budgetAmt := fs.String("budget-amount", "", "DEPRECATED: Total budget (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency)")
	locInvoiceDetails := fs.String("loc-invoice-details", "", `LOC invoice details JSON: inline JSON, @file.json, or @- for stdin`)
	countries := fs.String("countries-or-regions", "", "Comma-separated country codes (required)")
	adChannelType := fs.String("ad-channel-type", "SEARCH", "SEARCH (default) | DISPLAY")
	supplySources := fs.String("supply-sources", string(types.SupplySourceAppStoreSearchResults), "APPSTORE_SEARCH_RESULTS (default) | APPSTORE_SEARCH_TAB | APPSTORE_PRODUCT_PAGES_BROWSE | APPSTORE_TODAY_TAB")
	billingEvent := fs.String("billing-event", "TAPS", "TAPS (default) | IMPRESSIONS")
	status := fs.String("status", "", "ENABLED (default) | PAUSED (also 1/0, enable/pause)")
	startTime := fs.String("start-time", "", "Start time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)")
	endTime := fs.String("end-time", "", "End time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aads campaigns create --name NAME --adam-id ID --daily-budget-amount AMT --countries-or-regions CC [flags]",
		ShortHelp:  "Create a campaign.",
		LongHelp: `Use shortcut flags or --from-json for the full JSON body.

JSON keys (for --from-json):
  adamId              integer   (required) App Store app ID
  name                string    (required) Campaign name
  adChannelType       string    SEARCH (default) | DISPLAY
  supplySources       [string]  APPSTORE_SEARCH_RESULTS (default) |
                                APPSTORE_SEARCH_TAB |
                                APPSTORE_PRODUCT_PAGES_BROWSE | APPSTORE_TODAY_TAB
  countriesOrRegions  [string]  (required) ISO alpha-2 country codes
  billingEvent        string    TAPS (default) | IMPRESSIONS
  dailyBudgetAmount   Money     (required) Daily budget cap
  budgetAmount        Money     DEPRECATED: Total (lifetime) budget cap
  locInvoiceDetails   object    {billingContactEmail, buyerEmail, buyerName,
                                 clientName, orderNumber}
  status              string    ENABLED (default) | PAUSED
  startTime           string    ISO 8601 datetime
  endTime             string    ISO 8601 datetime

Money object: {"amount": "10.00", "currency": "USD"}

Examples:
  aads campaigns create --name "FitTrack US Search" --adam-id 900001 --daily-budget-amount 50 --countries-or-regions US
  aads campaigns create --name "FitTrack EU Search" --adam-id 900001 --daily-budget-amount "50 EUR" --countries-or-regions "GB,DE,FR"
  aads campaigns create --name "FitTrack %(COUNTRIES_OR_REGIONS) %(adChannelType)" --adam-id 900001 --daily-budget-amount 50 --countries-or-regions "DE,FR"
  aads campaigns create --name "FitTrack LOC" --adam-id 900001 --daily-budget-amount 50 --countries-or-regions US --loc-invoice-details '{"orderNumber":"PO-123"}'
  aads campaigns create --from-json campaign.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			execOnce := func() (any, error) {
				if shared.IsStdinJSONInputArg(*dataFile) && shared.IsStdinJSONInputArg(*locInvoiceDetails) {
					return nil, shared.UsageError("cannot use --from-json @- with --loc-invoice-details @-")
				}

				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
				}
				cfg, err := shared.LoadConfig()
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				var body json.RawMessage
				if *dataFile != "" {
					body, err = readBodyFile(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("create: reading body: %w", err)
					}
				} else {
					if *nameFlag == "" {
						return nil, shared.UsageError("--name is required")
					}
					if *adamID == "" {
						return nil, shared.UsageError("--adam-id is required")
					}
					if *dailyBudget == "" {
						return nil, shared.UsageError("--daily-budget-amount is required")
					}
					if *countries == "" {
						return nil, shared.UsageError("--countries-or-regions is required")
					}
					acht, err := shared.NormalizeAdChannelType(*adChannelType)
					if err != nil {
						return nil, err
					}
					be, err := shared.NormalizeBillingEvent(*billingEvent)
					if err != nil {
						return nil, err
					}

					m := map[string]any{
						"adamId":        json.Number(*adamID),
						"adChannelType": acht,
						"billingEvent":  be,
					}
					ss, err := parseSupplySourcesFlag(*supplySources)
					if err != nil {
						return nil, err
					}
					m["supplySources"] = ss

					daily, err := shared.ParseMoneyFlag(*dailyBudget)
					if err != nil {
						return nil, err
					}
					m["dailyBudgetAmount"] = daily

					if *budgetAmt != "" {
						ba, err := shared.ParseMoneyFlag(*budgetAmt)
						if err != nil {
							return nil, err
						}
						m["budgetAmount"] = ba
					}
					if *locInvoiceDetails != "" {
						loc, err := readJSONObjectArg(*locInvoiceDetails)
						if err != nil {
							return nil, fmt.Errorf("--loc-invoice-details: %w", err)
						}
						m["locInvoiceDetails"] = loc
					}

					codes := strings.Split(*countries, ",")
					cc := make([]string, 0, len(codes))
					for _, c := range codes {
						c = strings.TrimSpace(c)
						if c != "" {
							cc = append(cc, strings.ToUpper(c))
						}
					}
					m["countriesOrRegions"] = cc

					if *status != "" {
						s, err := shared.NormalizeStatus(*status, "ENABLED")
						if err != nil {
							return nil, err
						}
						m["status"] = s
					}
					if *startTime != "" {
						st, err := shared.ResolveDateTimeFlag(*startTime, cfg)
						if err != nil {
							return nil, fmt.Errorf("--start-time: %w", err)
						}
						m["startTime"] = st
					}
					if *endTime != "" {
						et, err := shared.ResolveDateTimeFlag(*endTime, cfg)
						if err != nil {
							return nil, fmt.Errorf("--end-time: %w", err)
						}
						m["endTime"] = et
					}

					renderedName, err := shared.RenderNameTemplate(*nameFlag, m)
					if err != nil {
						return nil, fmt.Errorf("--name: %w", err)
					}
					m["name"] = renderedName

					body, err = json.Marshal(m)
					if err != nil {
						return nil, fmt.Errorf("create: marshalling body: %w", err)
					}
				}

				if err := shared.CheckBudgetLimitJSON(body); err != nil {
					return nil, err
				}
				if *check {
					return shared.NewMutationCheckSummary("create", "campaign", "", body, shared.MutationCheckOptions{
						Safety: []string{"budget limits ok"},
					}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, campaigns.CreateRequest{RawBody: body}, &result)
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
				}
				return result, nil
			}

			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "CAMPAIGNID")
		},
	}
}

func parseSupplySourcesFlag(raw string) ([]string, error) {
	valid := map[string]bool{
		string(types.SupplySourceAppStoreSearchResults):      true,
		string(types.SupplySourceAppStoreSearchTab):          true,
		string(types.SupplySourceAppStoreProductPagesBrowse): true,
		string(types.SupplySourceAppStoreTodayTab):           true,
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.ToUpper(strings.TrimSpace(part))
		if value == "" {
			continue
		}
		if !valid[value] {
			return nil, shared.ValidationErrorf("invalid supply source %q: use APPSTORE_SEARCH_RESULTS, APPSTORE_SEARCH_TAB, APPSTORE_PRODUCT_PAGES_BROWSE, or APPSTORE_TODAY_TAB", part)
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil, shared.ValidationError("--supply-sources must not be empty")
	}
	return out, nil
}
