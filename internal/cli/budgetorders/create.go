package budgetorders

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strconv"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api/requests/budgetorders"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	dataFile := fs.String("from-json", "", "Path to JSON body file (or - for stdin)")
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	nameFlag := fs.String("name", "", "Budget order name (required)")
	startDate := fs.String("start-date", "", "Start date (YYYY-MM-DD, now, or signed offset like -5d)")
	endDate := fs.String("end-date", "", "End date (YYYY-MM-DD, now, or signed offset like -5d)")
	budgetAmt := fs.String("budget-amount", "", "Total budget (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency)")
	orderNumber := fs.String("order-number", "", "Purchase order number")
	primaryBuyerEmail := fs.String("primary-buyer-email", "", "Primary buyer email")
	primaryBuyerName := fs.String("primary-buyer-name", "", "Primary buyer name")
	billingEmail := fs.String("billing-email", "", "Billing contact email")
	clientNameFlag := fs.String("client-name", "", "Client name")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aads budgetorders create --name NAME [flags]",
		ShortHelp:  "Create a budget order.",
		LongHelp: `Create a budget order using shortcut flags or --from-json for full JSON.

Shortcut flags are wrapped in {"orgIds": [ORG], "bo": {...}} using --org-id.

JSON keys (for --from-json):
  orgIds  [integer]  (required) Organization IDs to associate
  bo      object     (required) Budget order details:
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
  aads budgetorders create --name "Q1 Budget" --budget-amount "10000 USD" --start-date 2025-01-01 --end-date 2025-03-31
  aads budgetorders create --name "Q1 Budget" --primary-buyer-email buyer@example.com --primary-buyer-name "Jane Doe"
  aads budgetorders create --from-json budgetorder.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			execOnce := func() (any, error) {
				client, err := shared.GetClient()
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

					bo := map[string]any{
						"name": *nameFlag,
					}
					if *startDate != "" {
						resolvedStartDate, err := shared.ParseDateFlag(*startDate)
						if err != nil {
							return nil, shared.UsageErrorf("--start-date: %v", err)
						}
						bo["startDate"] = resolvedStartDate
					}
					if *endDate != "" {
						resolvedEndDate, err := shared.ParseDateFlag(*endDate)
						if err != nil {
							return nil, shared.UsageErrorf("--end-date: %v", err)
						}
						bo["endDate"] = resolvedEndDate
					}
					if *budgetAmt != "" {
						money, err := shared.ParseMoneyFlag(*budgetAmt)
						if err != nil {
							return nil, err
						}
						bo["budget"] = money
					}
					if *orderNumber != "" {
						bo["orderNumber"] = *orderNumber
					}
					if *primaryBuyerEmail != "" {
						bo["primaryBuyerEmail"] = *primaryBuyerEmail
					}
					if *primaryBuyerName != "" {
						bo["primaryBuyerName"] = *primaryBuyerName
					}
					if *billingEmail != "" {
						bo["billingEmail"] = *billingEmail
					}
					if *clientNameFlag != "" {
						bo["clientName"] = *clientNameFlag
					}

					body, err = json.Marshal(map[string]any{
						"orgIds": []int64{oid},
						"bo":     bo,
					})
					if err != nil {
						return nil, fmt.Errorf("create: marshalling body: %w", err)
					}
				}
				if *check {
					return shared.NewMutationCheckSummary("create", "budget order", "", body, shared.MutationCheckOptions{}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, budgetorders.CreateRequest{RawBody: body}, &result)
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
				}
				return result, nil
			}

			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "BUDGETORDERID")
		},
	}
}
