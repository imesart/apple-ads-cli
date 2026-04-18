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

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("profiles create", flag.ContinueOnError)

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
		Name:       "create",
		ShortUsage: "aads profiles create --name NAME [--org-id ID] [flags]",
		ShortHelp:  "Create a new profile.",
		LongHelp: `Create a new configuration profile. If --org-id is omitted, the CLI tries to infer it
from Apple Ads by calling orgs user and using parentOrgId. It then looks up the
matching orgs list row to infer default_currency and default_timezone unless
you already provided those flags. Apple calls this ACL data; the CLI exposes it
under the orgs command group. If the lookup fails, the CLI warns and still
creates the profile as long as org_id was resolved. If this is the first
profile, it becomes the default automatically.

Example:
  aads profiles create --name default --client-id SEARCHADS.abc --team-id SEARCHADS.abc --key-id abc --org-id 123
  aads profiles create --name work --client-id SEARCHADS.def --team-id SEARCHADS.def --key-id def --org-id 456 --private-key-path ~/.aads/keys/work-private-key.pem

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
			explicitPrivateKeyPath := strings.TrimSpace(*privateKeyPath) != ""

			cf := config.LoadFile()

			if _, exists := cf.Profiles[profileName]; exists {
				return shared.ReportError(fmt.Errorf("profile %q already exists; use 'aads profiles update' to modify it", profileName))
			}

			p := config.Profile{
				ClientID:         strings.TrimSpace(*clientID),
				TeamID:           strings.TrimSpace(*teamID),
				KeyID:            strings.TrimSpace(*keyID),
				OrgID:            strings.TrimSpace(*orgID),
				PrivateKeyPath:   strings.TrimSpace(*privateKeyPath),
				DefaultCurrency:  strings.TrimSpace(*defaultCurrency),
				DefaultTimezone:  strings.TrimSpace(*defaultTimezone),
				DefaultTimeOfDay: strings.TrimSpace(*defaultTimeOfDay),
			}
			if p.PrivateKeyPath == "" {
				p.PrivateKeyPath = defaultPrivateKeyPath(profileName)
			}
			discovered := discoveredProfileDefaults{}
			explicitOrgID := p.OrgID
			if p.OrgID == "" || p.DefaultCurrency == "" || p.DefaultTimezone == "" {
				discoveredCfg := config.Profile{
					ClientID:       p.ClientID,
					TeamID:         p.TeamID,
					KeyID:          p.KeyID,
					PrivateKeyPath: p.PrivateKeyPath,
				}
				if inferred, err := discoverProfileDefaults(ctx, discoveredCfg, explicitOrgID); err == nil {
					discovered = inferred
					if p.OrgID == "" {
						p.OrgID = inferred.OrgID
					}
					if p.DefaultCurrency == "" {
						p.DefaultCurrency = inferred.DefaultCurrency
					}
					if p.DefaultTimezone == "" {
						p.DefaultTimezone = inferred.DefaultTimezone
					}
				} else if explicitOrgID == "" {
					return shared.UsageError("--org-id is required unless it can be inferred from Apple Ads orgs data (Apple ACLs) using client credentials")
				} else if p.ClientID != "" || p.TeamID != "" || p.KeyID != "" || p.PrivateKeyPath != "" {
					fmt.Fprintf(os.Stderr, "Warning: could not inspect Apple Ads orgs data (ACLs) for inferred defaults: %v\n", err)
				}
			}
			if strings.TrimSpace(p.OrgID) == "" {
				return shared.UsageError("--org-id is required unless it can be inferred from Apple Ads orgs data (ACLs) using client credentials")
			}
			for _, warning := range discovered.Warnings {
				fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
			}
			if err := validateTimeDefaults(p.DefaultTimezone, p.DefaultTimeOfDay); err != nil {
				return shared.ValidationErrorf("%v", err)
			}
			var err error
			if p.MaxDailyBudget, err = parseProfileLimitFlag(*maxDailyBudget, p.DefaultCurrency); err != nil {
				return shared.ValidationErrorf("--max-daily-budget: %v", err)
			}
			if p.MaxBid, err = parseProfileLimitFlag(*maxBid, p.DefaultCurrency); err != nil {
				return shared.ValidationErrorf("--max-bid: %v", err)
			}
			if p.MaxCPAGoal, err = parseProfileLimitFlag(*maxCPAGoal, p.DefaultCurrency); err != nil {
				return shared.ValidationErrorf("--max-cpa-goal: %v", err)
			}
			if p.MaxBudgetAmount, err = parseProfileLimitFlag(*maxBudgetAmount, p.DefaultCurrency); err != nil {
				return shared.ValidationErrorf("--max-budget: %v", err)
			}
			if p.PrivateKeyPath != "" {
				if _, err := os.Stat(expandUserPath(p.PrivateKeyPath)); os.IsNotExist(err) {
					fmt.Fprintln(os.Stderr, createMissingPrivateKeyWarning(profileName, p.PrivateKeyPath, explicitPrivateKeyPath))
				}
			}

			summaryFields := map[string]any{"name": profileName}
			if p.ClientID != "" {
				summaryFields["clientId"] = p.ClientID
			}
			if p.TeamID != "" {
				summaryFields["teamId"] = p.TeamID
			}
			if p.KeyID != "" {
				summaryFields["keyId"] = p.KeyID
			}
			if p.OrgID != "" {
				summaryFields["orgId"] = p.OrgID
			}
			if p.PrivateKeyPath != "" {
				summaryFields["privateKeyPath"] = p.PrivateKeyPath
			}
			if p.DefaultCurrency != "" {
				summaryFields["defaultCurrency"] = p.DefaultCurrency
			}
			if p.DefaultTimezone != "" {
				summaryFields["defaultTimezone"] = p.DefaultTimezone
			}
			if p.DefaultTimeOfDay != "" {
				summaryFields["defaultTimeOfDay"] = p.DefaultTimeOfDay
			}
			if p.MaxDailyBudget.Enabled() {
				summaryFields["maxDailyBudget"] = p.MaxDailyBudget.String()
			}
			if p.MaxBid.Enabled() {
				summaryFields["maxBid"] = p.MaxBid.String()
			}
			if p.MaxCPAGoal.Enabled() {
				summaryFields["maxCpaGoal"] = p.MaxCPAGoal.String()
			}
			if p.MaxBudgetAmount.Enabled() {
				summaryFields["maxBudgetAmount"] = p.MaxBudgetAmount.String()
			}

			body, err := json.Marshal(summaryFields)
			if err != nil {
				return err
			}
			if *check {
				return shared.PrintOutput(shared.NewMutationCheckSummary("create", "profile", shared.FormatTarget("name", profileName), body, shared.MutationCheckOptions{}), *output.Output, *output.Fields, *output.Pretty)
			}

			cf.Profiles[profileName] = p

			// First profile becomes the default
			if cf.DefaultProfile == "" || len(cf.Profiles) == 1 {
				cf.DefaultProfile = profileName
			}

			if err := config.SaveFile(cf); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Profile %q created in %s\n", profileName, config.DefaultConfigPath())
			if cf.DefaultProfile == profileName {
				fmt.Fprintf(os.Stderr, "Set as default profile.\n")
			}
			return nil
		},
	}
}
