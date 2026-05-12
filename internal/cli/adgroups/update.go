package adgroups

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	keywordsreq "github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/types"
)

func updateCmd() *ffcli.Command {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	shared.BindForceFlag(fs)
	merge := fs.Bool("merge", false, "Fetch current ad group and merge changes first (for targeting dimensions)")
	cpcBid := fs.String("default-bid", "", "Default CPC bid (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency), delta (+1), multiplier (x1.1), or percent (+10%)")
	updateInheritedBids := fs.Bool("update-inherited-bids", false, "Also update keyword bids that currently inherit this ad group's default bid")
	cpaGoal := fs.String("cpa-goal", "", "CPA goal (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency), delta (+1), multiplier (x1.1), or percent (+10%); Search only")
	status := fs.String("status", "", "ENABLED | PAUSED (also 1/0, enable/pause)")
	name := fs.String("name", "", "Ad group name")
	startTime := fs.String("start-time", "", "Start time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)")
	endTime := fs.String("end-time", "", "End time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "aads adgroups update --campaign-id CID --adgroup-id ID [flags]",
		ShortHelp:  "Update an ad group.",
		LongHelp: `Update an ad group. The API accepts partial updates.

Use shortcut flags for common changes, or --from-json for arbitrary JSON fields,
or --merge to fetch the current state first (needed for targeting dimensions).
Shortcut flags can be combined with each other.

Bid/CPA flags also accept math expressions:
  Delta:       +1, -0.50, "+1 USD" (adjusts current value)
  Multiplier:  x1.1, x0.9 (multiplies current value)
  Percent:     +10%, -15% (percent adjustment)
Relative expressions fetch the current value from the API first.

JSON keys (all optional):
  name                    string  Ad group name
  status                  string  ENABLED | PAUSED
  defaultBidAmount        Money   Default bid
  cpaGoal                 Money   Target cost per acquisition
  automatedKeywordsOptIn  bool    Enable automated keywords
  startTime               string  ISO 8601 datetime
  endTime                 string  ISO 8601 datetime
  targetingDimensions     object  Audience targeting (see adgroups create --help)

Money object: {"amount": "1.50", "currency": "USD"}

The --cpa-goal flag requires the campaign's adChannelType to be SEARCH;
the campaign is fetched automatically to verify this.

Use --update-inherited-bids with --default-bid to also update keyword-level
bid overrides that currently match the ad group's old default bid.

Examples:
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid 1.50
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid "1.50 EUR"
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid +0.50
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid x1.1
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid +10%
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid +10% --update-inherited-bids
  aads adgroups update --campaign-id 123 --adgroup-id 456 --cpa-goal 2.00
  aads adgroups update --campaign-id 123 --adgroup-id 456 --status 0
  aads adgroups update --campaign-id 123 --adgroup-id 456 --name "New Name" --status ENABLED
  aads adgroups update --campaign-id 123 --adgroup-id 456 --start-time 2025-06-01T00:00:00.000
  aads adgroups update --campaign-id 123 --adgroup-id 456 --from-json changes.json
  aads adgroups update --campaign-id 123 --adgroup-id 456 --merge --from-json targeting.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
				shared.StdinFlag{Name: "adgroup-id", Ptr: adgroupID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			if *dataFile != "" && !*merge {
				conflicts := shared.VisitedFlagNames(fs,
					"default-bid", "cpa-goal", "status", "name", "start-time", "end-time",
				)
				if len(conflicts) > 0 {
					return shared.UsageErrorf("--from-json cannot be combined with --%s (shortcut flags are ignored under --from-json; use --merge to overlay shortcuts on top of JSON)", conflicts[0])
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
				if *updateInheritedBids && *cpcBid == "" {
					return nil, shared.UsageError("--update-inherited-bids requires --default-bid")
				}

				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}
				cfg, err := shared.LoadConfig()
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				// Parse bid expressions
				var bidExpr, cpaExpr *shared.BidExpr
				if *cpcBid != "" {
					bidExpr, err = shared.ParseBidExpr(*cpcBid)
					if err != nil {
						return nil, fmt.Errorf("--default-bid: %w", err)
					}
				}
				if *cpaGoal != "" {
					cpaExpr, err = shared.ParseBidExpr(*cpaGoal)
					if err != nil {
						return nil, fmt.Errorf("--cpa-goal: %w", err)
					}
				}

				hasShortcuts := bidExpr != nil || cpaExpr != nil || *status != "" || *name != "" || *startTime != "" || *endTime != ""
				needsFetch := (bidExpr != nil && bidExpr.IsRelative()) || (cpaExpr != nil && cpaExpr.IsRelative())
				readOnlyChecks := []string{}
				warnings := []string{}

				// Fetch current ad group if needed for --merge or relative expressions
				var currentMap map[string]any
				if *merge || needsFetch || *updateInheritedBids {
					readOnlyChecks = append(readOnlyChecks, "fetched current adgroup")
					var current json.RawMessage
					err = client.Do(ctx, adgroups.GetRequest{
						CampaignID: cid,
						AdGroupID:  agid,
					}, &current)
					if err != nil {
						return nil, fmt.Errorf("update: fetching current ad group: %w", err)
					}
					if err = json.Unmarshal(current, &currentMap); err != nil {
						return nil, fmt.Errorf("update: parsing current ad group: %w", err)
					}
					if data, ok := currentMap["data"].(map[string]any); ok {
						currentMap = data
					}
				}

				// Resolve bid expressions against current values
				var resolvedBid, resolvedCPA map[string]string
				var previousDefaultBid *types.Money
				if *updateInheritedBids {
					previousDefaultBid = shared.ExtractMoney(currentMap, "defaultBidAmount")
					if previousDefaultBid == nil {
						return nil, shared.ValidationError("current ad group has no defaultBidAmount; cannot use --update-inherited-bids")
					}
				}
				if bidExpr != nil {
					resolvedBid, err = bidExpr.Resolve(shared.ExtractMoney(currentMap, "defaultBidAmount"))
					if err != nil {
						return nil, fmt.Errorf("--default-bid: %w", err)
					}
				}
				if cpaExpr != nil {
					resolvedCPA, err = cpaExpr.Resolve(shared.ExtractMoney(currentMap, "cpaGoal"))
					if err != nil {
						return nil, fmt.Errorf("--cpa-goal: %w", err)
					}
				}

				var body json.RawMessage
				if *merge {
					if *dataFile != "" {
						var overlay map[string]any
						raw, readErr := readBodyFile(*dataFile)
						if readErr != nil {
							return nil, fmt.Errorf("update: reading body: %w", readErr)
						}
						if err = json.Unmarshal(raw, &overlay); err != nil {
							return nil, fmt.Errorf("update: parsing overlay: %w", err)
						}
						for k, v := range overlay {
							currentMap[k] = v
						}
					}

					if err = ApplyFields(currentMap, Fields{
						DefaultBidAmount: resolvedBid,
						CPAGoal:          resolvedCPA,
						Status:           *status,
						Name:             *name,
						StartTime:        *startTime,
						EndTime:          *endTime,
					}, cfg, FieldLabels{}); err != nil {
						return nil, err
					}

					body, err = json.Marshal(currentMap)
					if err != nil {
						return nil, fmt.Errorf("update: marshalling merged body: %w", err)
					}
				} else if *dataFile != "" {
					var readErr error
					body, readErr = readBodyFile(*dataFile)
					if readErr != nil {
						return nil, fmt.Errorf("update: reading body: %w", readErr)
					}
				} else if hasShortcuts {
					update := make(map[string]any)
					if err = ApplyFields(update, Fields{
						DefaultBidAmount: resolvedBid,
						CPAGoal:          resolvedCPA,
						Status:           *status,
						Name:             *name,
						StartTime:        *startTime,
						EndTime:          *endTime,
					}, cfg, FieldLabels{}); err != nil {
						return nil, err
					}
					body, err = json.Marshal(update)
					if err != nil {
						return nil, fmt.Errorf("update: marshalling body: %w", err)
					}
				} else {
					return nil, shared.UsageError("--from-json, --default-bid, --status, or --merge is required")
				}

				hasCPAGoal, err := PayloadHasCPAGoal(body)
				if err != nil {
					return nil, err
				}
				var cpaCampaignAdChannelType string
				if hasCPAGoal {
					readOnlyChecks = append(readOnlyChecks, "fetched campaign to verify SEARCH channel")
					cpaCampaignAdChannelType, err = resolveCampaignAdChannelType(ctx, client, cid)
					if err != nil {
						return nil, err
					}
				}

				cpaGoalLabel := "--cpa-goal"
				if *cpaGoal == "" {
					cpaGoalLabel = "cpaGoal"
				}
				if err = ValidatePayload(ctx, client, cid, cpaCampaignAdChannelType, body, cpaGoalLabel, hasCPAGoal); err != nil {
					return nil, err
				}

				inheritedKeywordUpdates := []map[string]any(nil)
				if *updateInheritedBids {
					readOnlyChecks = append(readOnlyChecks, "listed ad group keywords")
					inheritedKeywordUpdates, err = buildInheritedKeywordBidUpdates(ctx, client, cid, agid, previousDefaultBid, resolvedBid)
					if err != nil {
						return nil, fmt.Errorf("update: finding inherited keyword bids: %w", err)
					}
					if len(inheritedKeywordUpdates) == 0 {
						warnings = append(warnings, "no keyword bids matched the ad group's previous default bid")
					}
				}

				if *check {
					return shared.NewMutationCheckSummary("update", "adgroup", shared.FormatTarget("campaign-id", cid, "adgroup-id", agid), body, shared.MutationCheckOptions{
						Count:          1 + len(inheritedKeywordUpdates),
						Safety:         []string{"bid and CPA goal limits ok"},
						ReadOnlyChecks: readOnlyChecks,
						Warnings:       warnings,
					}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, adgroups.UpdateRequest{
					CampaignID: cid,
					AdGroupID:  agid,
					RawBody:    body,
				}, &result)
				if err != nil {
					return nil, fmt.Errorf("update: %w", err)
				}
				if len(inheritedKeywordUpdates) > 0 {
					keywordBody, marshalErr := json.Marshal(inheritedKeywordUpdates)
					if marshalErr != nil {
						return nil, fmt.Errorf("update: marshalling inherited keyword updates: %w", marshalErr)
					}
					if err = shared.CheckBidLimitJSON(keywordBody); err != nil {
						return nil, err
					}
					if err = client.Do(ctx, keywordsreq.UpdateRequest{
						CampaignID: cid,
						AdGroupID:  agid,
						RawBody:    keywordBody,
					}, nil); err != nil {
						return nil, fmt.Errorf("update: applying inherited keyword bids: %w", err)
					}
				}

				return result, nil
			}

			if len(stdinFlags) > 0 {
				return shared.RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, "ADGROUPID")
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "ADGROUPID")
		},
	}
}

func buildInheritedKeywordBidUpdates(ctx context.Context, client interface {
	Do(context.Context, api.Request, any) error
}, campaignID, adgroupID string, oldBid *types.Money, newBid map[string]string) ([]map[string]any, error) {
	if oldBid == nil || newBid == nil {
		return nil, nil
	}

	var updates []map[string]any
	offset := 0

	for {
		var page types.ListResponse[map[string]any]
		if err := client.Do(ctx, keywordsreq.ListRequest{
			CampaignID: campaignID,
			AdGroupID:  adgroupID,
			Limit:      1000,
			Offset:     offset,
		}, &page); err != nil {
			return nil, err
		}

		for _, keyword := range page.Data {
			if !moneyMapsEqual(shared.ExtractMoney(keyword, "bidAmount"), oldBid) {
				continue
			}
			id, ok := keyword["id"]
			if !ok {
				continue
			}
			updates = append(updates, map[string]any{
				"id":        id,
				"bidAmount": newBid,
			})
		}

		if page.Pagination == nil || offset+len(page.Data) >= page.Pagination.TotalResults || len(page.Data) == 0 {
			break
		}
		offset += len(page.Data)
	}

	return updates, nil
}

func moneyMapsEqual(a *types.Money, b *types.Money) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return strings.EqualFold(strings.TrimSpace(a.Currency), strings.TrimSpace(b.Currency)) &&
		strings.TrimSpace(a.Amount) == strings.TrimSpace(b.Amount)
}
