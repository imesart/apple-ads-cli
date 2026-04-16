package impressionshare

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	impressionshare "github.com/imesart/apple-ads-cli/internal/api/requests/impression_share"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/types"
)

// Command returns the impression-share command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "impression-share",
		ShortUsage: "aads impression-share <subcommand>",
		ShortHelp:  "Manage impression share reports.",
		Subcommands: []*ffcli.Command{
			createCmd(),
			getCmd(),
			listCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	dataFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	name := fs.String("name", "", "Report name (required)")
	dateRange := fs.String("dateRange", "", "LAST_WEEK | LAST_2_WEEKS | LAST_4_WEEKS | CUSTOM")
	startTime := fs.String("startTime", "", "Start date (YYYY-MM-DD, now, or signed offset like -5d) (required when --dateRange=CUSTOM)")
	endTime := fs.String("endTime", "", "End date (YYYY-MM-DD, now, or signed offset like -5d) (required when --dateRange=CUSTOM)")
	granularity := fs.String("granularity", "", "DAILY (default) | WEEKLY")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aads impression-share create --name NAME [flags]",
		ShortHelp:  "Create an impression share report.",
		LongHelp: `Use shortcut flags or --from-json for the full JSON body.

Conditional shortcut flags:
  --startTime and --endTime are required when --dateRange is CUSTOM
  providing only one of --startTime/--endTime is allowed; the other default still applies
  --dateRange is not allowed with DAILY granularity

Shortcut flags:
  --name         Report name (required)
  --dateRange    LAST_WEEK | LAST_2_WEEKS | LAST_4_WEEKS | CUSTOM
  --startTime    Start date (required when --dateRange is CUSTOM; default -7d when no date flags are provided)
  --endTime      End date (required when --dateRange is CUSTOM; default now when no date flags are provided)
  --granularity  DAILY (default) | WEEKLY

JSON keys (for --from-json):
  name         string    (required) Report name
  dateRange    string    LAST_WEEK | LAST_2_WEEKS | LAST_4_WEEKS | CUSTOM
  startTime    string    Start date (required if dateRange is CUSTOM)
  endTime      string    End date (required if dateRange is CUSTOM)
  granularity  string    DAILY | WEEKLY
  selector     object    Filter conditions:
    conditions  [object]  Each: {field, operator, values}
                          Supported operator: IN
                          Supported fields: adamId, countryOrRegion

Report metrics returned: lowImpressionShare, highImpressionShare, rank, searchPopularity.
Report dimensions: adamId, appName, countryOrRegion, searchTerm.

Examples:
  aads impression-share create --name "Weekly Share Report"
  aads impression-share create --name "Weekly Share Report" --granularity WEEKLY --dateRange LAST_WEEK
  aads impression-share create --name "Custom Share Report" --dateRange CUSTOM --startTime 2026-03-01 --endTime 2026-03-07
  aads impression-share create --from-json '{"name":"Weekly Share Report","dateRange":"LAST_WEEK","granularity":"WEEKLY"}'`,
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
					body, err = shared.ReadJSONInputArg(*dataFile)
					if err != nil {
						return nil, fmt.Errorf("create: reading body: %w", err)
					}
				} else {
					if strings.TrimSpace(*name) == "" {
						return nil, shared.UsageError("--name is required")
					}

					payload := map[string]any{
						"name": strings.TrimSpace(*name),
					}

					var resolvedDateRange string
					if *dateRange != "" {
						normalizedDateRange, err := normalizeDateRange(*dateRange)
						if err != nil {
							return nil, shared.UsageErrorf("--dateRange: %v", err)
						}
						resolvedDateRange = normalizedDateRange
					}

					if *startTime != "" {
						resolvedStart, err := shared.ParseDateFlag(*startTime)
						if err != nil {
							return nil, shared.UsageErrorf("--startTime: %v", err)
						}
						payload["startTime"] = resolvedStart
					}
					if *endTime != "" {
						resolvedEnd, err := shared.ParseDateFlag(*endTime)
						if err != nil {
							return nil, shared.UsageErrorf("--endTime: %v", err)
						}
						payload["endTime"] = resolvedEnd
					}

					if resolvedDateRange == "" {
						if _, ok := payload["startTime"]; !ok {
							resolvedStart, err := shared.ParseDateFlag("-7d")
							if err != nil {
								return nil, fmt.Errorf("create: resolving default start time: %w", err)
							}
							payload["startTime"] = resolvedStart
						}
						if _, ok := payload["endTime"]; !ok {
							resolvedEnd, err := shared.ParseDateFlag("now")
							if err != nil {
								return nil, fmt.Errorf("create: resolving default end time: %w", err)
							}
							payload["endTime"] = resolvedEnd
						}
					}

					resolvedGranularity := "DAILY"
					if *granularity != "" {
						normalizedGranularity, err := normalizeGranularity(*granularity)
						if err != nil {
							return nil, shared.UsageErrorf("--granularity: %v", err)
						}
						resolvedGranularity = normalizedGranularity
					}
					payload["granularity"] = resolvedGranularity

					if resolvedDateRange != "" {
						payload["dateRange"] = resolvedDateRange
					}

					if resolvedGranularity == "DAILY" && resolvedDateRange != "" {
						return nil, shared.UsageError("--dateRange is not allowed with DAILY granularity; omit it and use start/end dates instead")
					}

					if resolvedDateRange == "CUSTOM" {
						if _, ok := payload["startTime"]; !ok {
							return nil, shared.UsageError("--startTime is required when --dateRange is CUSTOM")
						}
						if _, ok := payload["endTime"]; !ok {
							return nil, shared.UsageError("--endTime is required when --dateRange is CUSTOM")
						}
					}

					body, err = json.Marshal(payload)
					if err != nil {
						return nil, fmt.Errorf("create: marshalling body: %w", err)
					}
				}

				if *check {
					return shared.NewMutationCheckSummary("create", "impression share report", "", body, shared.MutationCheckOptions{}), nil
				}

				var result json.RawMessage
				err = client.Do(ctx, impressionshare.CreateRequest{RawBody: body}, &result)
				return result, err
			}

			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

func normalizeDateRange(value string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "LAST_WEEK", "LAST_2_WEEKS", "LAST_4_WEEKS", "CUSTOM":
		return strings.ToUpper(strings.TrimSpace(value)), nil
	default:
		return "", fmt.Errorf("invalid value %q: use LAST_WEEK, LAST_2_WEEKS, LAST_4_WEEKS, or CUSTOM", value)
	}
}

func normalizeGranularity(value string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "DAILY", "WEEKLY":
		return strings.ToUpper(strings.TrimSpace(value)), nil
	default:
		return "", fmt.Errorf("invalid value %q: use DAILY or WEEKLY", value)
	}
}

func getCmd() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	reportID := fs.String("report-id", "", "Custom Report ID")
	download := fs.String("download", "", "Write the file at downloadUri to this path; use - for stdout")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "aads impression-share get --report-id ID [--download FILE]",
		ShortHelp:  "Get a single impression share report.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*reportID) == "" {
				return shared.UsageError("--report-id is required")
			}

			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("get: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			var result json.RawMessage
			err = client.Do(ctx, impressionshare.GetRequest{ReportID: strings.TrimSpace(*reportID)}, &result)
			if err != nil {
				return fmt.Errorf("get: %w", err)
			}

			if strings.TrimSpace(*download) != "" {
				targetPath := strings.TrimSpace(*download)
				var envelope types.DataResponse[struct {
					DownloadURI string `json:"downloadUri"`
				}]
				if err := json.Unmarshal(result, &envelope); err != nil {
					return fmt.Errorf("get: decoding downloadUri: %w", err)
				}
				if strings.TrimSpace(envelope.Data.DownloadURI) == "" {
					return shared.ValidationError("downloadUri is missing from the report response")
				}
				if targetPath == "-" {
					if err := client.DownloadToWriter(ctx, envelope.Data.DownloadURI, os.Stdout); err != nil {
						return fmt.Errorf("get: downloading report: %w", err)
					}
					return nil
				}
				if err := client.DownloadToFile(ctx, envelope.Data.DownloadURI, targetPath); err != nil {
					return fmt.Errorf("get: downloading report: %w", err)
				}
				if _, err := os.Stat(targetPath); err != nil {
					return fmt.Errorf("get: verifying downloaded file: %w", err)
				}
				_, err = fmt.Fprintf(os.Stdout, "Downloaded report to %s\n", targetPath)
				return err
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

func listCmd() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	limit := fs.Int("limit", 0, "Maximum results; 0 fetches all")
	offset := fs.Int("offset", 0, "Starting offset")
	sortExpr := fs.String("sort", "", `Sort query: "field:asc" or "field:desc"`)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "aads impression-share list",
		ShortHelp:  "List all impression share reports.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			field, sortOrder, err := impressionShareSortQuery(*sortExpr)
			if err != nil {
				return err
			}
			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("list: %w", err)
			}
			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			req := impressionshare.ListRequest{
				Field:     field,
				SortOrder: sortOrder,
				Limit:     *limit,
				Offset:    *offset,
			}
			var result any
			if *limit == 0 {
				result, err = api.FetchAllRaw(ctx, client, req)
			} else {
				var raw json.RawMessage
				err = client.Do(ctx, req, &raw)
				result = raw
			}
			if err != nil {
				return fmt.Errorf("list: %w", err)
			}
			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty, "REPORTID")
		},
	}
}

func impressionShareSortQuery(expr string) (string, string, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", "", nil
	}
	field, order, ok := strings.Cut(expr, ":")
	if !ok {
		return "", "", shared.UsageErrorf("invalid --sort %q: use field:asc or field:desc", expr)
	}
	field = strings.TrimSpace(field)
	order = strings.ToLower(strings.TrimSpace(order))
	if field == "" {
		return "", "", shared.UsageError("--sort field is required")
	}
	switch order {
	case "asc":
		return field, "ASCENDING", nil
	case "desc":
		return field, "DESCENDING", nil
	default:
		return "", "", shared.UsageErrorf("invalid --sort order %q: use asc or desc", order)
	}
}
