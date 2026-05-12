package adgroups

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
	"github.com/imesart/apple-ads-cli/internal/output"
	"github.com/imesart/apple-ads-cli/internal/types"
)

func readBodyFile(path string) (json.RawMessage, error) {
	return shared.ReadJSONInputArg(path)
}

type Fields struct {
	DefaultBidAmount       map[string]string
	CPAGoal                map[string]string
	Status                 string
	Name                   string
	StartTime              string
	EndTime                string
	AutomatedKeywordsOptIn bool
}

type FieldLabels struct {
	StartTime string
	EndTime   string
}

func (l FieldLabels) withDefaults() FieldLabels {
	if strings.TrimSpace(l.StartTime) == "" {
		l.StartTime = "--start-time"
	}
	if strings.TrimSpace(l.EndTime) == "" {
		l.EndTime = "--end-time"
	}
	return l
}

func ApplyFields(m map[string]any, fields Fields, cfg *config.Profile, labels FieldLabels) error {
	labels = labels.withDefaults()
	if fields.DefaultBidAmount != nil {
		m["defaultBidAmount"] = fields.DefaultBidAmount
	}
	if fields.CPAGoal != nil {
		m["cpaGoal"] = fields.CPAGoal
	}
	if fields.Status != "" {
		s, err := shared.NormalizeStatus(fields.Status, "ENABLED")
		if err != nil {
			return err
		}
		m["status"] = s
	}
	if fields.Name != "" {
		m["name"] = fields.Name
	}
	if fields.StartTime != "" {
		st, err := shared.ResolveDateTimeFlag(fields.StartTime, cfg)
		if err != nil {
			return fmt.Errorf("%s: %w", labels.StartTime, err)
		}
		m["startTime"] = st
	}
	if fields.EndTime != "" {
		et, err := shared.ResolveDateTimeFlag(fields.EndTime, cfg)
		if err != nil {
			return fmt.Errorf("%s: %w", labels.EndTime, err)
		}
		m["endTime"] = et
	}
	if fields.AutomatedKeywordsOptIn {
		m["automatedKeywordsOptIn"] = true
	}
	return nil
}

func EnsureCreateStartTime(payload map[string]any, cfg *config.Profile) error {
	if raw, ok := payload["startTime"]; ok {
		if raw == nil {
			delete(payload, "startTime")
		} else {
			value, ok := raw.(string)
			if !ok {
				return shared.ValidationError("create: startTime must be a string when provided")
			}
			if strings.TrimSpace(value) != "" {
				return nil
			}
		}
	}
	startTime, err := shared.ResolveDateTimeFlag("now", cfg)
	if err != nil {
		return fmt.Errorf("create: default --start-time now: %w", err)
	}
	payload["startTime"] = startTime
	return nil
}

func normalizeCreatePayload(body json.RawMessage, cfg *config.Profile) (json.RawMessage, error) {
	var payload map[string]any
	if err := output.UnmarshalUseNumber(body, &payload); err != nil {
		return nil, fmt.Errorf("create: parsing body: %w", err)
	}
	if err := EnsureCreateStartTime(payload, cfg); err != nil {
		return nil, err
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("create: marshalling body: %w", err)
	}
	return normalized, nil
}

// PayloadHasCPAGoal reports whether the JSON body sets a non-null cpaGoal.
// Callers use this to decide whether a campaign GET will be needed (so they
// can pre-fetch and announce the read-only check up front).
func PayloadHasCPAGoal(body json.RawMessage) (bool, error) {
	var payload struct {
		CPAGoal *types.Money `json:"cpaGoal"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, fmt.Errorf("parsing payload for cpaGoal check: %w", err)
	}
	return payload.CPAGoal != nil, nil
}

// ValidatePayload validates an ad group payload's CPA goal (if present)
// against the campaign's adChannelType, then runs bid/CPA limit checks.
//
// hasCPAGoal must be supplied by the caller (typically via PayloadHasCPAGoal)
// so callers can pre-fetch and announce the campaign GET in --check output
// without this helper re-parsing the body. adChannelType, when non-empty, is
// trusted; otherwise the campaign is fetched via client. cpaGoalLabel
// customizes the error label (e.g. "cpaGoal" or "--cpa-goal"); if empty it
// defaults to "cpaGoal".
func ValidatePayload(ctx context.Context, client *api.Client, campaignID string, adChannelType string, body json.RawMessage, cpaGoalLabel string, hasCPAGoal bool) error {
	if cpaGoalLabel == "" {
		cpaGoalLabel = "cpaGoal"
	}
	if err := validateCPAGoal(ctx, client, campaignID, adChannelType, cpaGoalLabel, hasCPAGoal); err != nil {
		return err
	}
	return shared.CheckBidLimitJSON(body)
}

func validateCPAGoal(ctx context.Context, client *api.Client, campaignID string, campaignAdChannelType string, cpaGoalLabel string, hasCPAGoal bool) error {
	if !hasCPAGoal {
		return nil
	}
	acht := strings.TrimSpace(campaignAdChannelType)
	var err error
	if acht == "" {
		acht, err = resolveCampaignAdChannelType(ctx, client, campaignID)
		if err != nil {
			return err
		}
	}
	if !strings.EqualFold(acht, "SEARCH") {
		return fmt.Errorf("%s requires a SEARCH campaign (this campaign's adChannelType is %s)", cpaGoalLabel, acht)
	}
	return nil
}

func resolveCampaignAdChannelType(ctx context.Context, client *api.Client, campaignID string) (string, error) {
	var raw json.RawMessage
	if err := client.Do(ctx, campaigns.GetRequest{CampaignID: campaignID}, &raw); err != nil {
		return "", fmt.Errorf("fetching campaign: %w", err)
	}
	var resp struct {
		Data struct {
			AdChannelType string `json:"adChannelType"`
		} `json:"data"`
		AdChannelType string `json:"adChannelType"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("parsing campaign: %w", err)
	}
	acht := resp.Data.AdChannelType
	if acht == "" {
		acht = resp.AdChannelType
	}
	return acht, nil
}

func buildTargetingDimensions(age, gender, deviceClass, countryCode, adminArea, locality string) (map[string]any, error) {
	td := make(map[string]any)

	if age != "" {
		parts := strings.SplitN(age, "-", 2)
		if len(parts) == 2 {
			minAge, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			maxAge, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			td["age"] = map[string]any{
				"included": []map[string]any{
					{"minAge": minAge, "maxAge": maxAge},
				},
			}
		}
	}
	if gender != "" {
		genders, err := normalizeList(gender, shared.NormalizeGender)
		if err != nil {
			return nil, err
		}
		td["gender"] = map[string]any{"included": genders}
	}
	if deviceClass != "" {
		devices, err := normalizeList(deviceClass, shared.NormalizeDeviceClass)
		if err != nil {
			return nil, err
		}
		td["deviceClass"] = map[string]any{"included": devices}
	}
	if countryCode != "" {
		td["country"] = map[string]any{"included": splitTrimUpper(countryCode)}
	}
	if adminArea != "" {
		td["adminArea"] = map[string]any{"included": splitTrim(adminArea)}
	}
	if locality != "" {
		td["locality"] = map[string]any{"included": splitTrim(locality)}
	}

	if len(td) == 0 {
		return nil, nil
	}
	return td, nil
}

// normalizeList splits a comma-separated string and normalizes each item.
func normalizeList(s string, normalize func(string) (string, error)) ([]string, error) {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		normalized, err := normalize(p)
		if err != nil {
			return nil, err
		}
		result = append(result, normalized)
	}
	return result, nil
}

func splitTrimUpper(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, strings.ToUpper(p))
		}
	}
	return result
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
