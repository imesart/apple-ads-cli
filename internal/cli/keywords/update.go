package keywords

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func updateCmd() *ffcli.Command {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID (or - to read IDs from stdin) (required)")
	keywordID := fs.String("keyword-id", "", "Keyword ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	shared.BindForceFlag(fs)
	cpcBid := fs.String("bid", "", "Keyword bid (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency), delta (+1), multiplier (x1.1), or percent (+10%)")
	status := fs.String("status", "", "ACTIVE | PAUSED (also 1/0, enable/pause)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "aads keywords update --campaign-id CID --adgroup-id AGID --keyword-id ID [flags]",
		ShortHelp:  "Update a targeting keyword.",
		LongHelp: `Update a targeting keyword. The API accepts partial updates.

Use shortcut flags for common changes, or --from-json for arbitrary JSON.
Shortcut flags can be combined with each other.

The --bid flag also accepts math expressions:
  Delta:       +1, -0.50, "+1 USD" (adjusts current value)
  Multiplier:  x1.1, x0.9 (multiplies current value)
  Percent:     +10%, -15% (percent adjustment)
Relative expressions fetch the current keyword from the API first.

With --from-json, the body is a JSON array of keyword update objects.

JSON keys per keyword (all optional except id):
  id         integer  (required) Keyword ID to update
  text       string   Keyword text
  matchType  string   BROAD | EXACT
  status     string   ACTIVE | PAUSED
  bidAmount  Money    Keyword-level bid override

Money object: {"amount": "1.50", "currency": "USD"}

Examples:
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid 1.50
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid "1.50 EUR"
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid +0.25
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid x1.1
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid -10%
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --status 0
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --from-json updates.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
				shared.StdinFlag{Name: "adgroup-id", Ptr: adgroupID},
				shared.StdinFlag{Name: "keyword-id", Ptr: keywordID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			if *dataFile != "" {
				conflicts := shared.VisitedFlagNames(fs, "bid", "status")
				if len(conflicts) > 0 {
					return shared.UsageErrorf("--from-json cannot be combined with --%s (shortcut flags are ignored under --from-json)", conflicts[0])
				}
			}

			execOnce := func() (any, error) {
				cid := strings.TrimSpace(*campaignID)
				if cid == "" {
					return nil, shared.UsageError("--campaign-id is required")
				}
				agid := strings.TrimSpace(*adgroupID)
				if agid == "" {
					return nil, shared.UsageError("--adgroup-id is required")
				}
				kwid := strings.TrimSpace(*keywordID)
				if kwid == "" {
					return nil, shared.UsageError("--keyword-id is required")
				}

				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				// Parse bid expression
				var bidExpr *shared.BidExpr
				if *cpcBid != "" {
					bidExpr, err = shared.ParseBidExpr(*cpcBid)
					if err != nil {
						return nil, fmt.Errorf("--bid: %w", err)
					}
				}

				hasShortcuts := bidExpr != nil || *status != ""
				readOnlyChecks := []string{}

				// Resolve bid expression
				var resolvedBid map[string]string
				if bidExpr != nil {
					if bidExpr.IsRelative() {
						readOnlyChecks = append(readOnlyChecks, "fetched current keyword")
						// Fetch current keyword for relative expression
						var current json.RawMessage
						err = client.Do(ctx, keywords.GetRequest{
							CampaignID: cid,
							AdGroupID:  agid,
							KeywordID:  kwid,
						}, &current)
						if err != nil {
							return nil, fmt.Errorf("update: fetching current keyword: %w", err)
						}
						var currentMap map[string]any
						if err = json.Unmarshal(current, &currentMap); err != nil {
							return nil, fmt.Errorf("update: parsing current keyword: %w", err)
						}
						if data, ok := currentMap["data"].(map[string]any); ok {
							currentMap = data
						}
						resolvedBid, err = bidExpr.Resolve(shared.ExtractMoney(currentMap, "bidAmount"))
					} else {
						resolvedBid, err = bidExpr.Resolve(nil)
					}
					if err != nil {
						return nil, fmt.Errorf("--bid: %w", err)
					}
				}

				var body json.RawMessage
				if *dataFile != "" {
					body, err = readBodyFile(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("update: reading body: %w", err)
					}
				} else if hasShortcuts {
					update := map[string]any{"id": kwid}
					if err = ApplyFields(update, Fields{
						Bid:    resolvedBid,
						Status: *status,
					}); err != nil {
						return nil, err
					}
					body, err = json.Marshal([]any{update})
					if err != nil {
						return nil, fmt.Errorf("update: marshalling body: %w", err)
					}
				} else {
					return nil, shared.UsageError("--from-json, --bid, or --status is required")
				}

				if err = ValidatePayload(body); err != nil {
					return nil, err
				}
				if *check {
					return shared.NewMutationCheckSummary("update", "keyword", shared.FormatTarget("campaign-id", cid, "adgroup-id", agid, "keyword-id", kwid), body, shared.MutationCheckOptions{
						Safety:         []string{"bid limits ok"},
						ReadOnlyChecks: readOnlyChecks,
					}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, keywords.UpdateRequest{
					CampaignID: cid,
					AdGroupID:  agid,
					RawBody:    body,
				}, &result)
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}

				return result, nil
			}

			if len(stdinFlags) > 0 {
				return shared.RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, "KEYWORDID")
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "KEYWORDID")
		},
	}
}
