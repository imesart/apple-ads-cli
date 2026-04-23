package campaigns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/api"
	reqcampaigns "github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func readBodyFile(path string) (json.RawMessage, error) {
	return shared.ReadJSONInputArg(path)
}

func readJSONObjectArg(arg string) (map[string]any, error) {
	if arg == "" {
		return nil, nil
	}
	raw, err := shared.ReadJSONInputArg(arg)
	if err != nil {
		return nil, err
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("expected JSON object: %w", err)
	}
	return obj, nil
}

func applyCampaignShortcuts(m map[string]any, status, name, budgetAmt, dailyBudgetAmt, targetCpa, locInvoiceDetails, countries string) error {
	if status != "" {
		s, err := shared.NormalizeStatus(status, "ENABLED")
		if err != nil {
			return err
		}
		m["status"] = s
	}
	if name != "" {
		m["name"] = name
	}
	if budgetAmt != "" {
		money, err := shared.ParseMoneyFlag(budgetAmt)
		if err != nil {
			return err
		}
		m["budgetAmount"] = money
	}
	if dailyBudgetAmt != "" {
		money, err := shared.ParseMoneyFlag(dailyBudgetAmt)
		if err != nil {
			return err
		}
		m["dailyBudgetAmount"] = money
	}
	if targetCpa != "" {
		money, err := shared.ParseMoneyFlag(targetCpa)
		if err != nil {
			return err
		}
		m["targetCpa"] = money
	}
	if locInvoiceDetails != "" {
		loc, err := readJSONObjectArg(locInvoiceDetails)
		if err != nil {
			return fmt.Errorf("--loc-invoice-details: %w", err)
		}
		m["locInvoiceDetails"] = loc
	}
	if countries != "" {
		parts := strings.Split(countries, ",")
		codes := make([]string, 0, len(parts))
		for _, p := range parts {
			c := strings.TrimSpace(p)
			if c != "" {
				codes = append(codes, strings.ToUpper(c))
			}
		}
		m["countriesOrRegions"] = codes
	}
	return nil
}

func ValidatePayload(ctx context.Context, client *api.Client, body json.RawMessage, campaignID string) error {
	if err := validateCampaignTargetCPA(ctx, client, body, campaignID); err != nil {
		return err
	}
	if err := shared.CheckBudgetLimitJSON(body); err != nil {
		return err
	}
	return shared.CheckBidLimitJSON(body)
}

func validateCampaignTargetCPA(ctx context.Context, client *api.Client, body json.RawMessage, campaignID string) error {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if _, ok := payload["targetCpa"]; !ok {
		return nil
	}

	adChannelType := strings.TrimSpace(stringValue(payload["adChannelType"]))
	if adChannelType == "" && strings.TrimSpace(campaignID) != "" {
		current, err := fetchCampaignForValidation(ctx, client, campaignID)
		if err != nil {
			return err
		}
		adChannelType = strings.TrimSpace(stringValue(current["adChannelType"]))
	}
	if adChannelType == "" {
		adChannelType = "SEARCH"
	}
	if !strings.EqualFold(adChannelType, "SEARCH") {
		return shared.ValidationError("targetCpa is supported only for SEARCH campaigns")
	}
	return nil
}

func fetchCampaignForValidation(ctx context.Context, client *api.Client, campaignID string) (map[string]any, error) {
	var raw json.RawMessage
	if err := client.Do(ctx, reqcampaigns.GetRequest{CampaignID: campaignID}, &raw); err != nil {
		return nil, err
	}
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("parsing campaign response: %w", err)
	}
	if envelope.Data == nil {
		return nil, fmt.Errorf("missing campaign data")
	}
	return envelope.Data, nil
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}
