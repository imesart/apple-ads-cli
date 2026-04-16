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

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID (or - to read IDs from stdin) (required)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	shared.BindForceFlag(fs)
	text := fs.String("text", "", "Keyword text (comma-separated for multiple; quote individual keywords)")
	matchType := fs.String("match-type", "EXACT", "BROAD | EXACT (default)")
	bid := fs.String("bid", "", "Bid amount (AMOUNT or \"AMOUNT CURRENCY\"; bare amount uses default currency)")
	statusFlag := fs.String("status", "", "ACTIVE (default) | PAUSED (also 1/0, enable/pause)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aads keywords create --campaign-id CID --adgroup-id AGID --text TEXT [flags]",
		ShortHelp:  "Create targeting keywords.",
		LongHelp: `Create targeting keywords using shortcut flags or --from-json for full JSON.

With --text, creates one keyword per comma-separated item. All share
the same --match-type, --bid, and --status. Quote items with commas.

JSON keys per keyword (for --from-json):
  text       string  (required) Keyword text
  matchType  string  BROAD | EXACT
  status     string  ACTIVE (default) | PAUSED
  bidAmount  Money   Keyword-level bid override

Examples:
  aads keywords create --campaign-id 101 --adgroup-id 201 --text "fitness coach"
  aads keywords create --campaign-id 101 --adgroup-id 201 --text "fitness coach,home workout planner" --match-type BROAD
  aads keywords create --campaign-id 101 --adgroup-id 201 --text "fitness coach" --bid 1.50
  aads keywords create --campaign-id 1 --adgroup-id 2 --from-json keywords.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(
				shared.StdinFlag{Name: "campaign-id", Ptr: campaignID},
				shared.StdinFlag{Name: "adgroup-id", Ptr: adgroupID},
			)

			if len(stdinFlags) > 0 && shared.IsStdinJSONInputArg(*dataFile) {
				return shared.UsageError("cannot use --from-json @- with stdin-piped ID flags")
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
				} else if *text != "" {
					items := shared.ParseTextList(*text)
					mt, err := shared.NormalizeMatchType(*matchType)
					if err != nil {
						return nil, err
					}

					var bidMoney map[string]string
					if *bid != "" {
						bidMoney, err = shared.ParseMoneyFlag(*bid)
						if err != nil {
							return nil, err
						}
					}

					var s string
					if *statusFlag != "" {
						s, err = shared.NormalizeStatus(*statusFlag, "ACTIVE")
						if err != nil {
							return nil, err
						}
					}

					kws := make([]map[string]any, 0, len(items))
					for _, t := range items {
						kw := map[string]any{
							"text":      t,
							"matchType": mt,
						}
						if bidMoney != nil {
							kw["bidAmount"] = bidMoney
						}
						if s != "" {
							kw["status"] = s
						}
						kws = append(kws, kw)
					}

					body, err = json.Marshal(kws)
					if err != nil {
						return nil, fmt.Errorf("create: marshalling body: %w", err)
					}
				} else {
					return nil, shared.UsageError("--from-json or --text is required")
				}

				if err := shared.CheckBidLimitJSON(body); err != nil {
					return nil, err
				}
				if *check {
					return shared.NewMutationCheckSummary("create", "keyword", shared.FormatTarget("campaign-id", cid, "adgroup-id", agid), body, shared.MutationCheckOptions{
						Safety: []string{"bid limits ok"},
					}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, keywords.CreateRequest{
					CampaignID: cid,
					AdGroupID:  agid,
					RawBody:    body,
				}, &result)
				if err != nil {
					return nil, fmt.Errorf("create: %w", err)
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
