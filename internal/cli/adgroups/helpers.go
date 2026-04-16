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
)

func readBodyFile(path string) (json.RawMessage, error) {
	return shared.ReadJSONInputArg(path)
}

// checkSearchCampaign fetches the campaign and verifies adChannelType is SEARCH.
func checkSearchCampaign(ctx context.Context, client *api.Client, campaignID string) error {
	var raw json.RawMessage
	if err := client.Do(ctx, campaigns.GetRequest{CampaignID: campaignID}, &raw); err != nil {
		return fmt.Errorf("update: fetching campaign: %w", err)
	}
	var resp struct {
		Data struct {
			AdChannelType string `json:"adChannelType"`
		} `json:"data"`
		AdChannelType string `json:"adChannelType"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("update: parsing campaign: %w", err)
	}
	acht := resp.Data.AdChannelType
	if acht == "" {
		acht = resp.AdChannelType
	}
	if !strings.EqualFold(acht, "SEARCH") {
		return fmt.Errorf("--cpa-goal requires a SEARCH campaign (this campaign's adChannelType is %s)", acht)
	}
	return nil
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
