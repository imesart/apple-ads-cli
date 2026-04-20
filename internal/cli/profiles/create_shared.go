package profiles

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

type createCommandInputs struct {
	Name                    string
	ClientID                string
	TeamID                  string
	KeyID                   string
	OrgID                   string
	PrivateKeyPath          string
	DefaultCurrency         string
	DefaultTimezone         string
	DefaultTimeOfDay        string
	MaxDailyBudget          string
	MaxBid                  string
	MaxCPAGoal              string
	MaxBudgetAmount         string
	ExplicitOrgID           bool
	ExplicitDefaultCurrency bool
	ExplicitPrivateKey      bool
}

func runCreateFlow(ctx context.Context, inputs createCommandInputs, check bool, output shared.OutputFlags) error {
	cf := config.LoadFile()
	if _, exists := cf.Profiles[inputs.Name]; exists {
		return shared.ReportError(fmt.Errorf("profile %q already exists; use 'aads profiles update' to modify it", inputs.Name))
	}

	p, discovered, err := prepareCreateProfile(ctx, inputs)
	if err != nil {
		return err
	}
	for _, warning := range discovered.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	}

	return finalizeProfileCreate(cf, inputs.Name, p, inputs.ExplicitPrivateKey, check, output)
}

func prepareCreateProfile(ctx context.Context, inputs createCommandInputs) (config.Profile, discoveredProfileDefaults, error) {
	p := config.Profile{
		ClientID:         strings.TrimSpace(inputs.ClientID),
		TeamID:           strings.TrimSpace(inputs.TeamID),
		KeyID:            strings.TrimSpace(inputs.KeyID),
		OrgID:            strings.TrimSpace(inputs.OrgID),
		PrivateKeyPath:   strings.TrimSpace(inputs.PrivateKeyPath),
		DefaultCurrency:  strings.TrimSpace(inputs.DefaultCurrency),
		DefaultTimezone:  strings.TrimSpace(inputs.DefaultTimezone),
		DefaultTimeOfDay: strings.TrimSpace(inputs.DefaultTimeOfDay),
	}
	if p.PrivateKeyPath == "" {
		p.PrivateKeyPath = defaultPrivateKeyPath(inputs.Name)
	}
	if inputs.ExplicitPrivateKey {
		if _, err := os.Stat(expandUserPath(p.PrivateKeyPath)); os.IsNotExist(err) {
			return config.Profile{}, discoveredProfileDefaults{}, shared.ReportError(fmt.Errorf("private key file %q does not exist", p.PrivateKeyPath))
		} else if err != nil {
			return config.Profile{}, discoveredProfileDefaults{}, shared.ReportError(fmt.Errorf("checking private key path %q: %w", p.PrivateKeyPath, err))
		}
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
			return config.Profile{}, discoveredProfileDefaults{}, shared.UsageError("--org-id is required unless it can be inferred from Apple Ads orgs data (Apple ACLs) using client credentials")
		} else if p.ClientID != "" || p.TeamID != "" || p.KeyID != "" || p.PrivateKeyPath != "" {
			discovered.Warnings = append(discovered.Warnings, fmt.Sprintf("could not inspect Apple Ads orgs data (ACLs) for inferred defaults: %v", err))
		}
	}
	if strings.TrimSpace(p.OrgID) == "" {
		return config.Profile{}, discoveredProfileDefaults{}, shared.UsageError("--org-id is required unless it can be inferred from Apple Ads orgs data (Apple ACLs) using client credentials")
	}
	if err := validateTimeDefaults(p.DefaultTimezone, p.DefaultTimeOfDay); err != nil {
		return config.Profile{}, discoveredProfileDefaults{}, shared.ValidationErrorf("%v", err)
	}
	var err error
	if p.MaxDailyBudget, err = parseProfileLimitFlag(inputs.MaxDailyBudget, p.DefaultCurrency); err != nil {
		return config.Profile{}, discoveredProfileDefaults{}, shared.ValidationErrorf("--max-daily-budget: %v", err)
	}
	if p.MaxBid, err = parseProfileLimitFlag(inputs.MaxBid, p.DefaultCurrency); err != nil {
		return config.Profile{}, discoveredProfileDefaults{}, shared.ValidationErrorf("--max-bid: %v", err)
	}
	if p.MaxCPAGoal, err = parseProfileLimitFlag(inputs.MaxCPAGoal, p.DefaultCurrency); err != nil {
		return config.Profile{}, discoveredProfileDefaults{}, shared.ValidationErrorf("--max-cpa-goal: %v", err)
	}
	if p.MaxBudgetAmount, err = parseProfileLimitFlag(inputs.MaxBudgetAmount, p.DefaultCurrency); err != nil {
		return config.Profile{}, discoveredProfileDefaults{}, shared.ValidationErrorf("--max-budget: %v", err)
	}

	return p, discovered, nil
}

func finalizeProfileCreate(cf *config.ConfigFile, profileName string, p config.Profile, explicitPrivateKeyPath bool, check bool, output shared.OutputFlags) error {
	if p.PrivateKeyPath != "" && !explicitPrivateKeyPath {
		if _, err := os.Stat(expandUserPath(p.PrivateKeyPath)); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, createMissingPrivateKeyWarning(profileName, p.PrivateKeyPath, false))
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
	if check {
		return shared.PrintOutput(shared.NewMutationCheckSummary("create", "profile", shared.FormatTarget("name", profileName), body, shared.MutationCheckOptions{}), *output.Output, *output.Fields, *output.Pretty)
	}

	cf.Profiles[profileName] = p
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
}
