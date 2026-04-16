package productpages

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	productpages "github.com/imesart/apple-ads-cli/internal/api/requests/product_pages"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

// Command returns the product-pages command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "product-pages",
		ShortUsage: "aads product-pages <subcommand>",
		ShortHelp:  "Manage custom product pages.",
		Subcommands: []*ffcli.Command{
			listCmd(),
			getCmd(),
			localesCmd(),
			countriesCmd(),
			devicesCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func listCmd() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	adamID := fs.String("adam-id", "", "App Adam ID (or - to read IDs from stdin) (required)")
	limit := fs.Int("limit", 0, "Maximum results; 0 fetches all")
	offset := fs.Int("offset", 0, "Starting offset")
	var filters repeatedStrings
	fs.Var(&filters, "filter", `Filter query: "name=value" or "state=value" (repeatable)`)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "aads product-pages list --adam-id ID",
		ShortHelp:  "List custom product pages.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			stdinFlags := shared.CollectStdinFlags(shared.StdinFlag{Name: "adam-id", Ptr: adamID})
			execOnce := func() (any, error) {
				aid := strings.TrimSpace(*adamID)
				if aid == "" {
					return nil, shared.UsageError("--adam-id is required")
				}
				filterValues, err := productPageFilterQuery(filters)
				if err != nil {
					return nil, err
				}
				client, err := shared.GetClient()
				if err != nil {
					return nil, fmt.Errorf("list: %w", err)
				}
				ctx, cancel := shared.ContextWithTimeout(ctx)
				defer cancel()

				req := productpages.ListRequest{
					AdamID: aid,
					Name:   filterValues["name"],
					State:  filterValues["state"],
					Limit:  *limit,
					Offset: *offset,
				}
				if *limit == 0 {
					return api.FetchAllRaw(ctx, client, req)
				}
				var result json.RawMessage
				err = client.Do(ctx, req, &result)
				return result, err
			}
			if len(stdinFlags) > 0 {
				return shared.RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, "PRODUCTPAGEID")
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return shared.PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, "PRODUCTPAGEID")
		},
	}
}

func getCmd() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	adamID := fs.String("adam-id", "", "App Adam ID")
	productPageID := fs.String("product-page-id", "", "Product Page ID")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "aads product-pages get --adam-id ID --product-page-id PPID",
		ShortHelp:  "Get a product page.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			aid := strings.TrimSpace(*adamID)
			if aid == "" {
				return shared.UsageErrorf("--adam-id is required")
			}
			ppid := strings.TrimSpace(*productPageID)
			if ppid == "" {
				return shared.UsageErrorf("--product-page-id is required")
			}

			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("get: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			var result json.RawMessage
			err = client.Do(ctx, productpages.GetRequest{
				AdamID:        aid,
				ProductPageID: ppid,
			}, &result)
			if err != nil {
				return fmt.Errorf("get: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

func localesCmd() *ffcli.Command {
	fs := flag.NewFlagSet("locales", flag.ContinueOnError)
	adamID := fs.String("adam-id", "", "App Adam ID")
	productPageID := fs.String("product-page-id", "", "Product Page ID")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "locales",
		ShortUsage: "aads product-pages locales --adam-id ID --product-page-id PPID",
		ShortHelp:  "Get product page locales.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			aid := strings.TrimSpace(*adamID)
			if aid == "" {
				return shared.UsageErrorf("--adam-id is required")
			}
			ppid := strings.TrimSpace(*productPageID)
			if ppid == "" {
				return shared.UsageErrorf("--product-page-id is required")
			}

			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("locales: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			var result json.RawMessage
			err = client.Do(ctx, productpages.LocalesRequest{
				AdamID:        aid,
				ProductPageID: ppid,
			}, &result)
			if err != nil {
				return fmt.Errorf("locales: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

func countriesCmd() *ffcli.Command {
	fs := flag.NewFlagSet("countries", flag.ContinueOnError)
	countriesOrRegions := fs.String("countries-or-regions", "", "Comma-separated ISO alpha-2 country or region codes")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "countries",
		ShortUsage: "aads product-pages countries",
		ShortHelp:  "Get supported countries/regions.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("countries: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			var result json.RawMessage
			err = client.Do(ctx, productpages.CountriesRequest{
				CountriesOrRegions: strings.TrimSpace(*countriesOrRegions),
			}, &result)
			if err != nil {
				return fmt.Errorf("countries: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

type repeatedStrings []string

func (s *repeatedStrings) String() string {
	return strings.Join(*s, ",")
}

func (s *repeatedStrings) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func productPageFilterQuery(filters []string) (map[string]string, error) {
	out := map[string]string{}
	for _, filter := range filters {
		key, value, ok := strings.Cut(filter, "=")
		if !ok {
			fields := strings.Fields(filter)
			if len(fields) == 3 && strings.EqualFold(fields[1], "EQUALS") {
				key, value, ok = fields[0], fields[2], true
			}
		}
		if !ok {
			return nil, shared.UsageErrorf("invalid --filter %q: use name=value or state=value", filter)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "name" && key != "state" {
			return nil, shared.UsageErrorf("unsupported --filter field %q: supported fields are name and state", key)
		}
		if value == "" {
			return nil, shared.UsageErrorf("--filter %q has an empty value", filter)
		}
		out[key] = value
	}
	return out, nil
}

func devicesCmd() *ffcli.Command {
	fs := flag.NewFlagSet("devices", flag.ContinueOnError)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "devices",
		ShortUsage: "aads product-pages devices",
		ShortHelp:  "Get app preview device sizes.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("devices: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			var result json.RawMessage
			err = client.Do(ctx, productpages.DevicesRequest{}, &result)
			if err != nil {
				return fmt.Errorf("devices: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}
