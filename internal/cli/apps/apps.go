package apps

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/apps"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

// Command returns the apps command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "apps",
		ShortUsage: "aads apps <subcommand>",
		ShortHelp:  "Search and inspect apps.",
		Subcommands: []*ffcli.Command{
			searchCmd(),
			eligibilityCmd(),
			detailsCmd(),
			localizedCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func searchCmd() *ffcli.Command {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	query := fs.String("query", "", "Search query string (required)")
	onlyOwnedApps := fs.Bool("only-owned-apps", false, "Only return apps owned by the current organization")
	limit := fs.Int("limit", 0, "Maximum results; 0 fetches all pages")
	offset := fs.Int("offset", 0, "Starting offset")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "search",
		ShortUsage: "aads apps search --query TEXT [flags]",
		ShortHelp:  "Search for iOS apps.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			q := strings.TrimSpace(*query)
			if q == "" {
				return shared.UsageErrorf("--query is required")
			}

			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("search: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			req := apps.SearchRequest{
				SearchQuery:     q,
				ReturnOwnedApps: *onlyOwnedApps,
				Limit:           *limit,
				Offset:          *offset,
			}
			if *limit == 0 {
				rows, err := api.FetchAll[json.RawMessage](ctx, client, req)
				if err != nil {
					return fmt.Errorf("search: %w", err)
				}
				data, err := json.Marshal(map[string]any{"data": rows})
				if err != nil {
					return fmt.Errorf("search: marshalling fetched rows: %w", err)
				}
				return shared.PrintOutput(json.RawMessage(data), *output.Output, *output.Fields, *output.Pretty)
			}

			var result json.RawMessage
			err = client.Do(ctx, req, &result)
			if err != nil {
				return fmt.Errorf("search: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

// validSupplySources lists accepted supply source enum values.
var validSupplySources = map[string]bool{
	"APPSTORE_SEARCH_RESULTS":       true,
	"APPSTORE_SEARCH_TAB":           true,
	"APPSTORE_PRODUCT_PAGES_BROWSE": true,
	"APPSTORE_TODAY_TAB":            true,
}

func eligibilityCmd() *ffcli.Command {
	fs := flag.NewFlagSet("eligibility", flag.ContinueOnError)

	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	adamID := fs.String("adam-id", "", "App Store app ID (required)")
	country := fs.String("country-or-region", "", "ISO alpha-2 country code, e.g. US (required)")
	deviceClass := fs.String("device-class", "", "IPHONE | IPAD")
	supplySource := fs.String("supply-source", "", "APPSTORE_SEARCH_RESULTS | APPSTORE_SEARCH_TAB | APPSTORE_PRODUCT_PAGES_BROWSE | APPSTORE_TODAY_TAB")
	minAge := fs.Int("min-age", 0, "Minimum age restriction")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "eligibility",
		ShortUsage: "aads apps eligibility --adam-id ID --country-or-region CC [flags]",
		ShortHelp:  "Check app eligibility.",
		LongHelp: `Use shortcut flags or --from-json for the selector body.

Selector JSON keys (for --from-json):
  conditions       [object] Each: {field, operator, values}
  fields           [string] Fields to return
  orderBy          [object] Each: {field, sortOrder}
  pagination       object   {offset, limit}

Shortcut flags map to selector conditions:
  --country-or-region -> countryOrRegion EQUALS <value>
  --device-class      -> deviceClass EQUALS <value>
  --supply-source     -> supplySource EQUALS <value>
  --min-age           -> minAge EQUALS <value>

Alternate input:
  --from-json also accepts an alternate body shape with keys:
  adamId, countryOrRegion, deviceClass, supplySource, minAge

Response includes eligibility state: ELIGIBLE or INELIGIBLE.

Examples:
  aads apps eligibility --adam-id 900001 --country-or-region US
  aads apps eligibility --adam-id 900001 --country-or-region US --device-class IPHONE --supply-source APPSTORE_SEARCH_RESULTS
  aads apps eligibility --adam-id 900001 --from-json selector.json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			execOnce := func() (any, error) {
				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("eligibility: %w", err)
				}

				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				var body json.RawMessage
				var adamIDVal string
				if *dataFile != "" {
					body, err = shared.ReadJSONInputArg(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("eligibility: reading body: %w", err)
					}
					var parsed map[string]json.RawMessage
					if err := json.Unmarshal(body, &parsed); err != nil {
						return nil, fmt.Errorf("eligibility: parsing body: %w", err)
					}

					if rawConditions, ok := parsed["conditions"]; ok && len(rawConditions) > 0 {
						if strings.TrimSpace(*adamID) == "" {
							return nil, shared.UsageError("--adam-id is required when --from-json provides a selector body")
						}
						adamIDVal = strings.TrimSpace(*adamID)
					} else {
						body, adamIDVal, err = buildLegacyEligibilitySelector(body)
						if err != nil {
							return nil, err
						}
					}
					if strings.TrimSpace(*adamID) != "" {
						adamIDVal = strings.TrimSpace(*adamID)
					}
				} else {
					if *adamID == "" {
						return nil, shared.UsageError("--adam-id is required")
					}
					if *country == "" {
						return nil, shared.UsageError("--country-or-region is required")
					}
					adamIDVal = strings.TrimSpace(*adamID)

					conditions := []map[string]any{
						{
							"field":    "countryOrRegion",
							"operator": "EQUALS",
							"values":   []any{strings.ToUpper(strings.TrimSpace(*country))},
						},
					}

					if *deviceClass != "" {
						dc, err := shared.NormalizeDeviceClass(*deviceClass)
						if err != nil {
							return nil, err
						}
						conditions = append(conditions, map[string]any{
							"field":    "deviceClass",
							"operator": "EQUALS",
							"values":   []any{dc},
						})
					}

					if *supplySource != "" {
						ss := strings.ToUpper(strings.TrimSpace(*supplySource))
						if !validSupplySources[ss] {
							return nil, fmt.Errorf("invalid supply source %q: use APPSTORE_SEARCH_RESULTS, APPSTORE_SEARCH_TAB, APPSTORE_PRODUCT_PAGES_BROWSE, or APPSTORE_TODAY_TAB", *supplySource)
						}
						conditions = append(conditions, map[string]any{
							"field":    "supplySource",
							"operator": "EQUALS",
							"values":   []any{ss},
						})
					}

					if *minAge > 0 {
						conditions = append(conditions, map[string]any{
							"field":    "minAge",
							"operator": "EQUALS",
							"values":   []any{*minAge},
						})
					}

					body, err = json.Marshal(map[string]any{
						"conditions": conditions,
					})
					if err != nil {
						return nil, fmt.Errorf("eligibility: marshalling body: %w", err)
					}
				}

				if *check {
					return shared.NewMutationCheckSummary("check", "app eligibility", "", body, shared.MutationCheckOptions{}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, apps.EligibilityRequest{AdamID: adamIDVal, RawBody: body}, &result)
				if err != nil {
					return nil, fmt.Errorf("eligibility: %w", err)
				}
				return result, nil
			}

			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

func buildLegacyEligibilitySelector(body json.RawMessage) (json.RawMessage, string, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, "", fmt.Errorf("eligibility: parsing body: %w", err)
	}

	rawAdamID, ok := payload["adamId"]
	if !ok {
		return nil, "", shared.UsageError("adamId is required in legacy JSON body")
	}

	adamIDVal := strings.TrimSpace(fmt.Sprint(rawAdamID))
	if adamIDVal == "" || adamIDVal == "<nil>" {
		return nil, "", shared.UsageError("adamId is required in legacy JSON body")
	}

	rawCountry, ok := payload["countryOrRegion"]
	if !ok {
		return nil, "", shared.UsageError("countryOrRegion is required in legacy JSON body")
	}
	country := strings.ToUpper(strings.TrimSpace(fmt.Sprint(rawCountry)))
	if country == "" || country == "<nil>" {
		return nil, "", shared.UsageError("countryOrRegion is required in legacy JSON body")
	}

	conditions := []map[string]any{
		{
			"field":    "countryOrRegion",
			"operator": "EQUALS",
			"values":   []any{country},
		},
	}

	if rawDeviceClass, ok := payload["deviceClass"]; ok {
		dc, err := shared.NormalizeDeviceClass(fmt.Sprint(rawDeviceClass))
		if err != nil {
			return nil, "", err
		}
		conditions = append(conditions, map[string]any{
			"field":    "deviceClass",
			"operator": "EQUALS",
			"values":   []any{dc},
		})
	}

	if rawSupplySource, ok := payload["supplySource"]; ok {
		ss := strings.ToUpper(strings.TrimSpace(fmt.Sprint(rawSupplySource)))
		if !validSupplySources[ss] {
			return nil, "", fmt.Errorf("invalid supply source %q: use APPSTORE_SEARCH_RESULTS, APPSTORE_SEARCH_TAB, APPSTORE_PRODUCT_PAGES_BROWSE, or APPSTORE_TODAY_TAB", rawSupplySource)
		}
		conditions = append(conditions, map[string]any{
			"field":    "supplySource",
			"operator": "EQUALS",
			"values":   []any{ss},
		})
	}

	if rawMinAge, ok := payload["minAge"]; ok {
		conditions = append(conditions, map[string]any{
			"field":    "minAge",
			"operator": "EQUALS",
			"values":   []any{rawMinAge},
		})
	}

	selectorBody, err := json.Marshal(map[string]any{
		"conditions": conditions,
	})
	if err != nil {
		return nil, "", fmt.Errorf("eligibility: marshalling selector body: %w", err)
	}
	return selectorBody, adamIDVal, nil
}

func detailsCmd() *ffcli.Command {
	return shared.BuildIDGetCommand(shared.IDGetCommandConfig{
		Name:       "details",
		ShortUsage: "aads apps details --adam-id ID",
		ShortHelp:  "Get app details.",
		IDFlag:     "adam-id",
		IDUsage:    "App Adam ID",
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, apps.DetailsRequest{AdamID: id}, &result)
			return result, err
		},
	})
}

func localizedCmd() *ffcli.Command {
	return shared.BuildIDGetCommand(shared.IDGetCommandConfig{
		Name:       "localized",
		ShortUsage: "aads apps localized --adam-id ID",
		ShortHelp:  "Get localized app details.",
		IDFlag:     "adam-id",
		IDUsage:    "App Adam ID",
		Exec: func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) (any, error) {
			var result json.RawMessage
			err := client.Do(ctx, apps.LocalizedRequest{AdamID: id}, &result)
			return result, err
		},
	})
}
