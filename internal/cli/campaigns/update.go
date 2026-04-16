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
)

func updateCmd() *ffcli.Command {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	shared.BindForceFlag(fs)
	status := fs.String("status", "", "ENABLED | PAUSED (also 1/0, enable/pause)")
	name := fs.String("name", "", "Campaign name")
	budgetAmount := fs.String("budget-amount", "", "DEPRECATED: Total budget (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency)")
	dailyBudgetAmount := fs.String("daily-budget-amount", "", "Daily budget (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency)")
	locInvoiceDetails := fs.String("loc-invoice-details", "", `LOC invoice details JSON: inline JSON, @file.json, or @- for stdin`)
	countriesOrRegions := fs.String("countries-or-regions", "", "Comma-separated country codes (e.g. US,GB)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "aads campaigns update --campaign-id ID [flags]",
		ShortHelp:  "Update a campaign.",
		LongHelp: `Update a campaign. The API accepts partial updates.
The CLI wraps the body in a {"campaign": ...} envelope automatically.

Use shortcut flags for common changes, or --from-json for arbitrary JSON.
Shortcut flags can be combined with each other.

JSON keys (all optional):
  name                string    Campaign name
  status              string    ENABLED | PAUSED
  budgetAmount        Money     DEPRECATED: Total (lifetime) budget cap
  dailyBudgetAmount   Money     Daily budget cap
  countriesOrRegions  [string]  ISO alpha-2 country codes
  budgetOrders        [integer] Budget order IDs (LOC only)
  locInvoiceDetails   object    {billingContactEmail, buyerEmail, buyerName,
                                clientName, orderNumber}
  startTime           string    ISO 8601 datetime
  endTime             string    ISO 8601 datetime

Money object: {"amount": "10.00", "currency": "USD"}

Examples:
  aads campaigns update --campaign-id 123 --status 0
  aads campaigns update --campaign-id 123 --daily-budget-amount 75.00
  aads campaigns update --campaign-id 123 --daily-budget-amount "75.00 EUR"
  aads campaigns update --campaign-id 123 --name "New Name" --status ENABLED
  aads campaigns update --campaign-id 123 --loc-invoice-details '{"orderNumber":"PO-123"}'
  aads campaigns update --campaign-id 123 --countries-or-regions "US,GB,CA"
  aads campaigns update --campaign-id 123 --from-json changes.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			execOnce := func() (any, error) {
				cid := strings.TrimSpace(*campaignID)
				if cid == "" {
					return nil, shared.UsageError("--campaign-id is required")
				}

				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				hasShortcuts := *status != "" || *name != "" || *budgetAmount != "" || *dailyBudgetAmount != "" || *locInvoiceDetails != "" || *countriesOrRegions != ""

				var body json.RawMessage
				if *dataFile != "" {
					body, err = readBodyFile(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("update: reading body: %w", err)
					}
				} else if hasShortcuts {
					update := make(map[string]any)
					if err := applyCampaignShortcuts(update, *status, *name, *budgetAmount, *dailyBudgetAmount, *locInvoiceDetails, *countriesOrRegions); err != nil {
						return nil, err
					}
					body, err = json.Marshal(update)
					if err != nil {
						return nil, fmt.Errorf("update: marshalling body: %w", err)
					}
				} else {
					return nil, shared.UsageError("--from-json or shortcut flags (--status, --name, --budget-amount, --daily-budget-amount, --loc-invoice-details, --countries-or-regions) required")
				}

				if err := shared.CheckBudgetLimitJSON(body); err != nil {
					return nil, err
				}
				if *check {
					return shared.NewMutationCheckSummary("update", "campaign", shared.FormatTarget("campaign-id", cid), body, shared.MutationCheckOptions{
						Safety: []string{"budget limits ok"},
					}), nil
				}

				// Wrap body in campaign envelope: {"campaign": <body>}
				envelope := json.RawMessage(fmt.Sprintf(`{"campaign":%s}`, body))
				var result json.RawMessage
				err = client.Do(ctx, campaigns.UpdateRequest{CampaignID: cid, RawBody: envelope}, &result)
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}

				return result, nil
			}

			if len(stdinFlags) > 0 {
				return shared.RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, "CAMPAIGNID")
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "CAMPAIGNID")
		},
	}
}

func applyCampaignShortcuts(m map[string]any, status, name, budgetAmt, dailyBudgetAmt, locInvoiceDetails, countries string) error {
	if status != "" {
		s, err := shared.NormalizeStatus(status, "ENABLED")
		if err != nil {
			return err
		}
		m["status"] = s
	}
	if name != "" {
		m["name"] = name
	}
	if budgetAmt != "" {
		money, err := shared.ParseMoneyFlag(budgetAmt)
		if err != nil {
			return err
		}
		m["budgetAmount"] = money
	}
	if dailyBudgetAmt != "" {
		money, err := shared.ParseMoneyFlag(dailyBudgetAmt)
		if err != nil {
			return err
		}
		m["dailyBudgetAmount"] = money
	}
	if locInvoiceDetails != "" {
		loc, err := readJSONObjectArg(locInvoiceDetails)
		if err != nil {
			return fmt.Errorf("--loc-invoice-details: %w", err)
		}
		m["locInvoiceDetails"] = loc
	}
	if countries != "" {
		parts := strings.Split(countries, ",")
		codes := make([]string, 0, len(parts))
		for _, p := range parts {
			c := strings.TrimSpace(p)
			if c != "" {
				codes = append(codes, strings.ToUpper(c))
			}
		}
		m["countriesOrRegions"] = codes
	}
	return nil
}
