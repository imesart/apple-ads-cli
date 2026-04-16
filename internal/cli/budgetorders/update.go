package budgetorders

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api/requests/budgetorders"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func updateCmd() *ffcli.Command {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)

	budgetOrderID := fs.String("budget-order-id", "", "Budget Order ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	name := fs.String("name", "", "Budget order name")
	startDate := fs.String("start-date", "", "Start date (YYYY-MM-DD, now, or signed offset like -5d)")
	endDate := fs.String("end-date", "", "End date (YYYY-MM-DD, now, or signed offset like -5d)")
	budgetAmt := fs.String("budget-amount", "", "Total budget (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "aads budgetorders update --budget-order-id ID [flags]",
		ShortHelp:  "Update a budget order.",
		LongHelp: `Update a budget order. Use shortcut flags or --from-json for the full body.

Shortcut flags are wrapped in {"orgIds": [ORG], "bo": {...}} using --org-id.

JSON keys (for --from-json):
  orgIds  [integer]  (required) Organization IDs
  bo      object     (required) Fields to update:
    name               string  Budget order name
    budget             Money   Total budget amount
    startDate          string  Start date (YYYY-MM-DD)
    endDate            string  End date (YYYY-MM-DD)
    primaryBuyerEmail  string  Primary buyer email
    primaryBuyerName   string  Primary buyer name
    billingEmail       string  Billing contact email
    clientName         string  Client name
    orderNumber        string  Purchase order number

Money object: {"amount": "10000.00", "currency": "USD"}

Examples:
  aads budgetorders update --budget-order-id 123 --name "Q2 Budget"
  aads budgetorders update --budget-order-id 123 --budget-amount "15000 USD"
  aads budgetorders update --budget-order-id 123 --end-date 2025-06-30
  aads budgetorders update --budget-order-id 123 --from-json changes.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "budget-order-id", Ptr: budgetOrderID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			execOnce := func() (any, error) {
				boid := strings.TrimSpace(*budgetOrderID)
				if boid == "" {
					return nil, shared.UsageError("--budget-order-id is required")
				}

				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				hasShortcuts := *name != "" || *startDate != "" || *endDate != "" || *budgetAmt != ""

				var body json.RawMessage
				if *dataFile != "" {
					body, err = readBodyFile(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("update: reading body: %w", err)
					}
				} else if hasShortcuts {
					bo := make(map[string]any)
					if err := applyBudgetOrderShortcuts(bo, *name, *startDate, *endDate, *budgetAmt); err != nil {
						return nil, err
					}

					orgID := shared.OrgID()
					if orgID == "" {
						cfg, cfgErr := shared.GetConfig()
						if cfgErr == nil && cfg.OrgID != "" {
							orgID = cfg.OrgID
						}
					}
					if orgID == "" {
						return nil, shared.UsageError("--org-id is required when using shortcut flags (or set org_id in config)")
					}
					oid, err := strconv.ParseInt(orgID, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid org-id: %w", err)
					}

					body, err = json.Marshal(map[string]any{
						"orgIds": []int64{oid},
						"bo":     bo,
					})
					if err != nil {
						return nil, fmt.Errorf("update: marshalling body: %w", err)
					}
				} else {
					return nil, shared.UsageError("--from-json or shortcut flags (--name, --start-date, --end-date, --budget-amount) required")
				}
				if *check {
					return shared.NewMutationCheckSummary("update", "budget order", shared.FormatTarget("budget-order-id", boid), body, shared.MutationCheckOptions{}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, budgetorders.UpdateRequest{BudgetOrderID: boid, RawBody: body}, &result)
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}

				return result, nil
			}

			if len(stdinFlags) > 0 {
				return shared.RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, "BUDGETORDERID")
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "BUDGETORDERID")
		},
	}
}

func applyBudgetOrderShortcuts(bo map[string]any, name, startDate, endDate, budgetAmt string) error {
	if name != "" {
		bo["name"] = name
	}
	if startDate != "" {
		resolvedStartDate, err := shared.ParseDateFlag(startDate)
		if err != nil {
			return shared.UsageErrorf("--start-date: %v", err)
		}
		bo["startDate"] = resolvedStartDate
	}
	if endDate != "" {
		resolvedEndDate, err := shared.ParseDateFlag(endDate)
		if err != nil {
			return shared.UsageErrorf("--end-date: %v", err)
		}
		bo["endDate"] = resolvedEndDate
	}
	if budgetAmt != "" {
		money, err := shared.ParseMoneyFlag(budgetAmt)
		if err != nil {
			return err
		}
		bo["budget"] = money
	}
	return nil
}
