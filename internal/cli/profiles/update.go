package profiles

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

func updateCmd() *ffcli.Command {
	fs := flag.NewFlagSet("profiles update", flag.ContinueOnError)

	name := fs.String("name", "", "Profile name (required)")
	clientID := fs.String("client-id", "", "Apple Ads client ID")
	teamID := fs.String("team-id", "", "Apple Ads team ID")
	keyID := fs.String("key-id", "", "Apple Ads key ID")
	orgID := fs.String("org-id", "", "Apple Ads organization ID")
	privateKeyPath := fs.String("private-key-path", "", "Path to ES256 private key PEM file")
	defaultCurrency := fs.String("default-currency", "", "Default currency, e.g. USD")
	defaultTimezone := fs.String("default-timezone", "", "Default timezone for time flags, e.g. Europe/Luxembourg (default local machine timezone)")
	defaultTimeOfDay := fs.String("default-time-of-day", "", "Default time-of-day for date-only time flags: HH:MM or HH:MM:SS (default current time in the selected timezone)")
	maxDailyBudget := fs.String("max-daily-budget", "", "Safety limit in default currency: \"AMOUNT\" or \"AMOUNT CURRENCY\" (0 or empty = disabled)")
	maxBid := fs.String("max-bid", "", "Safety limit in default currency: \"AMOUNT\" or \"AMOUNT CURRENCY\" (0 or empty = disabled)")
	maxCPAGoal := fs.String("max-cpa-goal", "", "Safety limit in default currency: \"AMOUNT\" or \"AMOUNT CURRENCY\" (0 or empty = disabled)")
	maxBudgetAmount := fs.String("max-budget", "", "Safety limit in default currency: \"AMOUNT\" or \"AMOUNT CURRENCY\" (0 or empty = disabled)")
	check := fs.Bool("check", false, "Validate and summarize without writing config")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "aads profiles update --name NAME [flags]",
		ShortHelp:  "Update an existing profile.",
		LongHelp: `Update fields on an existing profile. Only provided flags are changed.

Example:
  aads profiles update --name default --org-id 456
  aads profiles update --name work --max-daily-budget 500

Time defaults apply to mutation time flags like --start-time and --end-time.
If default_timezone is empty, the local machine timezone is used. If
default_time_of_day is empty, the current time in the selected timezone is
used. Report day flags do not use default_timezone unless the report timezone
is UTC.

Limit flags are stored in the config file as decimal text in default_currency.
They accept "AMOUNT" or "AMOUNT CURRENCY". If a currency is provided, it must
match default_currency. Use 0 or an empty value to disable a limit.`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			profileName := strings.TrimSpace(*name)
			if profileName == "" {
				return shared.UsageError("--name is required")
			}
			visited := map[string]bool{}
			fs.Visit(func(f *flag.Flag) {
				visited[f.Name] = true
			})

			cf := config.LoadFile()

			p, exists := cf.Profiles[profileName]
			if !exists {
				return shared.ReportError(fmt.Errorf("profile %q not found", profileName))
			}

			// Track whether any flag was provided
			updated := false
			summaryFields := map[string]any{}

			if v := strings.TrimSpace(*clientID); v != "" {
				p.ClientID = v
				summaryFields["clientId"] = v
				updated = true
			}
			if v := strings.TrimSpace(*teamID); v != "" {
				p.TeamID = v
				summaryFields["teamId"] = v
				updated = true
			}
			if v := strings.TrimSpace(*keyID); v != "" {
				p.KeyID = v
				summaryFields["keyId"] = v
				updated = true
			}
			if v := strings.TrimSpace(*orgID); v != "" {
				p.OrgID = v
				summaryFields["orgId"] = v
				updated = true
			}
			if v := strings.TrimSpace(*privateKeyPath); v != "" {
				p.PrivateKeyPath = v
				summaryFields["privateKeyPath"] = v
				updated = true
			}
			if v := strings.TrimSpace(*defaultCurrency); v != "" {
				p.DefaultCurrency = v
				summaryFields["defaultCurrency"] = v
				updated = true
			}
			if visited["default-timezone"] {
				p.DefaultTimezone = strings.TrimSpace(*defaultTimezone)
				summaryFields["defaultTimezone"] = p.DefaultTimezone
				updated = true
			}
			if visited["default-time-of-day"] {
				p.DefaultTimeOfDay = strings.TrimSpace(*defaultTimeOfDay)
				summaryFields["defaultTimeOfDay"] = p.DefaultTimeOfDay
				updated = true
			}
			if err := validateTimeDefaults(p.DefaultTimezone, p.DefaultTimeOfDay); err != nil {
				return shared.ValidationErrorf("%v", err)
			}
			if visited["max-daily-budget"] {
				parsed, err := parseProfileLimitFlag(*maxDailyBudget, p.DefaultCurrency)
				if err != nil {
					return shared.ValidationErrorf("--max-daily-budget: %v", err)
				}
				p.MaxDailyBudget = parsed
				summaryFields["maxDailyBudget"] = parsed.String()
				updated = true
			}
			if visited["max-bid"] {
				parsed, err := parseProfileLimitFlag(*maxBid, p.DefaultCurrency)
				if err != nil {
					return shared.ValidationErrorf("--max-bid: %v", err)
				}
				p.MaxBid = parsed
				summaryFields["maxBid"] = parsed.String()
				updated = true
			}
			if visited["max-cpa-goal"] {
				parsed, err := parseProfileLimitFlag(*maxCPAGoal, p.DefaultCurrency)
				if err != nil {
					return shared.ValidationErrorf("--max-cpa-goal: %v", err)
				}
				p.MaxCPAGoal = parsed
				summaryFields["maxCpaGoal"] = parsed.String()
				updated = true
			}
			if visited["max-budget"] {
				parsed, err := parseProfileLimitFlag(*maxBudgetAmount, p.DefaultCurrency)
				if err != nil {
					return shared.ValidationErrorf("--max-budget: %v", err)
				}
				p.MaxBudgetAmount = parsed
				summaryFields["maxBudgetAmount"] = parsed.String()
				updated = true
			}

			if p.PrivateKeyPath != "" {
				if _, err := os.Stat(p.PrivateKeyPath); os.IsNotExist(err) {
					fmt.Fprintf(os.Stderr, "Warning: private key file %q does not exist\n", p.PrivateKeyPath)
				}
			}

			if !updated {
				return shared.UsageError("at least one configuration flag is required")
			}
			body, err := json.Marshal(summaryFields)
			if err != nil {
				return err
			}
			if *check {
				return shared.PrintOutput(shared.NewMutationCheckSummary("update", "profile", shared.FormatTarget("name", profileName), body, shared.MutationCheckOptions{}), *output.Output, *output.Fields, *output.Pretty)
			}

			cf.Profiles[profileName] = p

			if err := config.SaveFile(cf); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Profile %q updated in %s\n", profileName, config.DefaultConfigPath())
			return nil
		},
	}
}
