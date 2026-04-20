package profiles

import (
	"context"
	"flag"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func createCmd() *ffcli.Command {
	fs := flag.NewFlagSet("profiles create", flag.ContinueOnError)

	name := fs.String("name", "", "Profile name")
	interactive := fs.Bool("interactive", false, "Prompt for missing profile fields in a terminal wizard")
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
		ShortUsage: "aads profiles create [--name NAME] [--org-id ID] [flags]",
		ShortHelp:  "Create a new profile.",
		LongHelp: `Create a new configuration profile. Use --interactive to launch a
terminal wizard that prompts for missing values, guides key setup, and
gathers Apple Ads credentials before writing the profile.

If --org-id is omitted, the CLI tries to infer it from Apple Ads by calling
orgs user and using parentOrgId. It then looks up the matching orgs list row
to infer default_currency and default_timezone unless you already provided
those flags. Apple calls this ACL data; the CLI exposes it under the orgs
command group. If the lookup fails, the CLI warns and still creates the
profile as long as org_id was resolved. If this is the first profile, it
becomes the default automatically.

Example:
  aads profiles create --interactive
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
			inputs := createCommandInputs{
				Name:                    strings.TrimSpace(*name),
				ClientID:                strings.TrimSpace(*clientID),
				TeamID:                  strings.TrimSpace(*teamID),
				KeyID:                   strings.TrimSpace(*keyID),
				OrgID:                   strings.TrimSpace(*orgID),
				PrivateKeyPath:          strings.TrimSpace(*privateKeyPath),
				DefaultCurrency:         strings.TrimSpace(*defaultCurrency),
				DefaultTimezone:         strings.TrimSpace(*defaultTimezone),
				DefaultTimeOfDay:        strings.TrimSpace(*defaultTimeOfDay),
				MaxDailyBudget:          strings.TrimSpace(*maxDailyBudget),
				MaxBid:                  strings.TrimSpace(*maxBid),
				MaxCPAGoal:              strings.TrimSpace(*maxCPAGoal),
				MaxBudgetAmount:         strings.TrimSpace(*maxBudgetAmount),
				ExplicitOrgID:           strings.TrimSpace(*orgID) != "",
				ExplicitDefaultCurrency: strings.TrimSpace(*defaultCurrency) != "",
				ExplicitPrivateKey:      strings.TrimSpace(*privateKeyPath) != "",
			}
			if *interactive {
				return runInteractiveCreate(ctx, inputs, *check, output)
			}
			if inputs.Name == "" {
				return shared.UsageError("--name is required")
			}
			return runCreateFlow(ctx, inputs, *check, output)
		},
	}
}
