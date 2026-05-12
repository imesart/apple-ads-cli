package adgroups

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	shared.BindForceFlag(fs)
	nameFlag := fs.String("name", "", "Ad group name (accepts template variables like %(fieldName) or %(FIELD_NAME)) (required)")
	defaultBid := fs.String("default-bid", "", "Default CPC bid amount (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency) (required)")
	statusFlag := fs.String("status", "", "ENABLED (default) | PAUSED (also 1/0, enable/pause)")
	startTime := fs.String("start-time", "", "Start time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)")
	endTime := fs.String("end-time", "", "End time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)")
	autoKeywords := fs.Bool("automated-keywords-opt-in", false, "Enable automated keywords")
	cpaGoal := fs.String("cpa-goal", "", "CPA goal amount (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency); Search campaigns only")
	age := fs.String("age", "", "Age range (e.g. \"18-65\")")
	gender := fs.String("gender", "", "Gender targeting (comma-separated: M,F)")
	deviceClass := fs.String("device-class", "", "Device class (comma-separated: IPHONE,IPAD)")
	countryCode := fs.String("country-code", "", "Country targeting (comma-separated: US,GB)")
	adminArea := fs.String("admin-area", "", "Admin area targeting (comma-separated: US|CA,US|NY)")
	locality := fs.String("locality", "", "Locality targeting (comma-separated: US|CA|Los Angeles)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aads adgroups create --campaign-id CID --name NAME --default-bid BID [flags]",
		ShortHelp:  "Create an ad group.",
		LongHelp: `Use shortcut flags or --from-json for the full JSON body.

JSON keys (for --from-json):
  name                    string  (required) Ad group name
  pricingModel            string  CPC (default)
  defaultBidAmount        Money   (required) Default bid
  status                  string  ENABLED (default) | PAUSED
  cpaGoal                 Money   Target cost per acquisition (Search only)
  automatedKeywordsOptIn  bool    Enable automated keywords
  startTime               string  ISO 8601 datetime
  endTime                 string  ISO 8601 datetime
  targetingDimensions     object  Audience targeting

Money object: {"amount": "1.50", "currency": "USD"}

Examples:
  aads adgroups create --campaign-id 123 --name "Search Group" --default-bid 1.50
  aads adgroups create --campaign-id 123 --name "Search Group" --default-bid "1.50 USD" --status PAUSED
  aads adgroups create --campaign-id 123 --name "Search Group %(PRICING_MODEL)" --default-bid 1.50 --cpa-goal 2.00 --age 18-65 --gender M,F
  aads adgroups create --campaign-id 123 --from-json adgroup.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			if *dataFile != "" {
				conflicts := shared.VisitedFlagNames(fs,
					"name", "default-bid", "status", "start-time", "end-time",
					"automated-keywords-opt-in", "cpa-goal",
					"age", "gender", "device-class", "country-code", "admin-area", "locality",
				)
				if len(conflicts) > 0 {
					return shared.UsageErrorf("--from-json cannot be combined with --%s (shortcut flags are ignored under --from-json)", conflicts[0])
				}
			}

			execOnce := func() (any, error) {
				cid := strings.TrimSpace(*campaignID)
				if cid == "" {
					return nil, shared.UsageError("--campaign-id is required")
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

				readOnlyChecks := []string{}

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
					if *defaultBid == "" {
						return nil, shared.UsageError("--default-bid is required")
					}

					bid, err := shared.ParseMoneyFlag(*defaultBid)
					if err != nil {
						return nil, err
					}

					m := map[string]any{
						"pricingModel":     "CPC",
						"defaultBidAmount": bid,
					}

					fields := Fields{
						DefaultBidAmount:       bid,
						Status:                 *statusFlag,
						StartTime:              *startTime,
						EndTime:                *endTime,
						AutomatedKeywordsOptIn: *autoKeywords,
					}
					if *cpaGoal != "" {
						money, err := shared.ParseMoneyFlag(*cpaGoal)
						if err != nil {
							return nil, err
						}
						fields.CPAGoal = money
					}
					if err := ApplyFields(m, fields, cfg, FieldLabels{}); err != nil {
						return nil, err
					}

					td, err := buildTargetingDimensions(*age, *gender, *deviceClass, *countryCode, *adminArea, *locality)
					if err != nil {
						return nil, err
					}
					if td != nil {
						m["targetingDimensions"] = td
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
				body, err = normalizeCreatePayload(body, cfg)
				if err != nil {
					return nil, err
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
				if err := ValidatePayload(ctx, client, cid, cpaCampaignAdChannelType, body, cpaGoalLabel, hasCPAGoal); err != nil {
					return nil, err
				}
				if *check {
					return shared.NewMutationCheckSummary("create", "adgroup", shared.FormatTarget("campaign-id", cid), body, shared.MutationCheckOptions{
						Safety:         []string{"bid and CPA goal limits ok"},
						ReadOnlyChecks: readOnlyChecks,
					}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, adgroups.CreateRequest{
					CampaignID: cid,
					RawBody:    body,
				}, &result)
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
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
