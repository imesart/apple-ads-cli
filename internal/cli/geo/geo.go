package geo

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/geo"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/types"
)

// Command returns the geo command group.
func Command() *ffcli.Command {
	return &ffcli.Command{
		Name:       "geo",
		ShortUsage: "aads geo <subcommand>",
		ShortHelp:  "Search geolocations.",
		Subcommands: []*ffcli.Command{
			searchCmd(),
			getCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func searchCmd() *ffcli.Command {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	query := fs.String("query", "", "Search query (required)")
	entity := fs.String("entity", "", "Entity type: Country, AdminArea, Locality")
	countryCode := fs.String("country-code", "", "Country code, ISO 3166-1 alpha-2")
	limit := fs.Int("limit", 0, "Maximum results; 0 fetches all pages")
	offset := fs.Int("offset", 0, "Starting offset")
	sorts := shared.BindLocalSortFlags(fs)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "search",
		ShortUsage: "aads geo search --query TEXT [--entity TYPE] [--country-code CC] [flags]",
		ShortHelp:  "Search for geolocations.",
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

			req := geo.SearchRequest{
				SearchQuery: q,
				Entity:      strings.TrimSpace(*entity),
				CountryCode: strings.TrimSpace(*countryCode),
				Limit:       *limit,
				Offset:      *offset,
			}
			var result json.RawMessage
			if *limit == 0 {
				rows, err := api.FetchAll[json.RawMessage](ctx, client, req)
				if err != nil {
					return fmt.Errorf("search: %w", err)
				}
				data, err := json.Marshal(map[string]any{"data": rows})
				if err != nil {
					return fmt.Errorf("search: marshalling fetched rows: %w", err)
				}
				result = json.RawMessage(data)
			} else {
				if err := client.Do(ctx, req, &result); err != nil {
					return fmt.Errorf("search: %w", err)
				}
			}
			result, err = shared.MaybeApplyLocalSorts(result, sorts.Values(), "search")
			if err != nil {
				return err
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

func getCmd() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	entity := fs.String("entity", "", "Entity type: Country, AdminArea, Locality (required)")
	geoID := fs.String("geo-id", "", "Geo identifier (required)")
	limit := fs.Int("limit", 0, "Maximum results")
	offset := fs.Int("offset", 0, "Starting offset")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "aads geo get --entity TYPE --geo-id ID [flags]",
		ShortHelp:  "Get geolocation details.",
		LongHelp: `Get geolocation details for a specific geo identifier.

Required flags:
  --entity  Country | AdminArea | Locality
  --geo-id  Geo identifier

Examples:
  aads geo get --entity Country --geo-id US
  aads geo get --entity AdminArea --geo-id US|CA
  aads geo get --entity Locality --geo-id US|CA|San Francisco`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			ent, err := normalizeGeoEntity(*entity)
			if err != nil {
				return err
			}
			id := strings.TrimSpace(*geoID)
			if id == "" {
				return shared.UsageError("--geo-id is required")
			}

			client, err := shared.GetClient()
			if err != nil {
				return fmt.Errorf("get: %w", err)
			}

			ctx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			var result json.RawMessage
			err = client.Do(ctx, geo.GetRequest{
				ID:     id,
				Entity: ent,
				Limit:  *limit,
				Offset: *offset,
			}, &result)
			if err != nil {
				return fmt.Errorf("get: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

func normalizeGeoEntity(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return "", shared.UsageError("--entity is required")
	case strings.ToLower(string(types.GeolocationEntityCountry)):
		return string(types.GeolocationEntityCountry), nil
	case strings.ToLower(string(types.GeolocationEntityAdminArea)):
		return string(types.GeolocationEntityAdminArea), nil
	case strings.ToLower(string(types.GeolocationEntityLocality)):
		return string(types.GeolocationEntityLocality), nil
	default:
		return "", shared.ValidationError(`--entity must be one of: Country, AdminArea, Locality`)
	}
}
