package negatives

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	negadgroup "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_adgroup"
	negcampaign "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_campaign"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	campaignID := fs.String("campaign-id", "", "Campaign ID (or - to read IDs from stdin) (required)")
	adgroupID := fs.String("adgroup-id", "", "Ad Group ID; omit for campaign-level (or - to read IDs from stdin)")
	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	text := fs.String("text", "", "Keyword text (comma-separated for multiple; quote individual keywords)")
	matchType := fs.String("match-type", "EXACT", "BROAD | EXACT (default)")
	statusFlag := fs.String("status", "", "ACTIVE (default) | PAUSED (also 1/0, enable/pause)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aads negatives create --campaign-id CID [--adgroup-id AGID] --text TEXT [flags]",
		ShortHelp:  "Create negative keywords.",
		LongHelp: `Create negative keywords using shortcut flags or --from-json for full JSON.

Without --adgroup-id, creates campaign-level negative keywords.
With --adgroup-id, creates ad group-level negative keywords.

With --text, creates one negative keyword per comma-separated item.
All share the same --match-type and --status. Quote items with commas.

JSON keys per negative keyword (for --from-json):
  text       string  (required) Keyword text to exclude
  matchType  string  BROAD | EXACT

Examples:
  aads negatives create --campaign-id 101 --text "free workout,fitness wallpaper"
  aads negatives create --campaign-id 101 --adgroup-id 201 --text "yoga mat,protein powder" --match-type BROAD
  aads negatives create --campaign-id 1 --from-json negatives.json`,
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
				target := shared.FormatTarget("campaign-id", cid)
				if isAdGroupLevel(*adgroupID) {
					target = shared.FormatTarget("campaign-id", cid, "adgroup-id", strings.TrimSpace(*adgroupID))
				}
				if *check {
					return shared.NewMutationCheckSummary("create", "negative keyword", target, body, shared.MutationCheckOptions{}), nil
				}

				if isAdGroupLevel(*adgroupID) {
					var result json.RawMessage
					err := client.Do(ctx, negadgroup.CreateRequest{
						CampaignID: cid,
						AdGroupID:  strings.TrimSpace(*adgroupID),
						RawBody:    body,
					}, &result)
					return result, err
				}
				var result json.RawMessage
				err = client.Do(ctx, negcampaign.CreateRequest{
					CampaignID: cid,
					RawBody:    body,
				}, &result)
				return result, err
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
