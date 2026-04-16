package shared

import (
	"encoding/json"
	"testing"
)

func TestApplyLocalFiltersJSON_IDEqualityIsExact(t *testing.T) {
	raw := json.RawMessage(`{"data":[{"campaignId":123},{"campaignId":1234}]}`)

	got, err := ApplyLocalFiltersJSON(raw, []string{"campaignId=123"})
	if err != nil {
		t.Fatalf("ApplyLocalFiltersJSON error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got, &envelope); err != nil {
		t.Fatalf("unmarshal filtered envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
	row := data[0].(map[string]any)
	if row["campaignId"] != float64(123) {
		t.Fatalf("campaignId = %v, want 123", row["campaignId"])
	}
}

func TestApplyLocalFiltersJSON_EntityAliasMatchesIDField(t *testing.T) {
	raw := json.RawMessage(`{"data":[{"id":900601},{"id":900602}]}`)

	got, err := ApplyLocalFiltersJSON(raw, []string{"CREATIVE_ID=900601"}, "CREATIVEID")
	if err != nil {
		t.Fatalf("ApplyLocalFiltersJSON error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got, &envelope); err != nil {
		t.Fatalf("unmarshal filtered envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
	row := data[0].(map[string]any)
	if row["id"] != float64(900601) {
		t.Fatalf("id = %v, want 900601", row["id"])
	}
}

func TestApplyLocalFiltersJSON_MoneyAliasAndNumericComparison(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"localSpend":{"amount":"12.34","currency":"USD"},"impressions":100},
		{"localSpend":{"amount":"8.10","currency":"EUR"},"impressions":200}
	]}`)

	got, err := ApplyLocalFiltersJSON(raw, []string{"LOCAL_SPEND GREATER_THAN 10"})
	if err != nil {
		t.Fatalf("ApplyLocalFiltersJSON error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got, &envelope); err != nil {
		t.Fatalf("unmarshal filtered envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
	row := data[0].(map[string]any)
	localSpend := row["localSpend"].(map[string]any)
	if localSpend["amount"] != "12.34" {
		t.Fatalf("localSpend.amount = %v, want 12.34", localSpend["amount"])
	}
}

func TestApplyLocalFiltersJSON_MissingNumericDefaultsToZero(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"campaignId":111,"impressions":10},
		{"campaignId":222}
	]}`)

	got, err := ApplyLocalFiltersJSON(raw, []string{"impressions=0"})
	if err != nil {
		t.Fatalf("ApplyLocalFiltersJSON error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got, &envelope); err != nil {
		t.Fatalf("unmarshal filtered envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
	row := data[0].(map[string]any)
	if row["campaignId"] != float64(222) {
		t.Fatalf("campaignId = %v, want 222", row["campaignId"])
	}
}

func TestApplyLocalFiltersJSON_InvalidFieldErrors(t *testing.T) {
	raw := json.RawMessage(`{"data":[{"campaignId":123}]}`)

	_, err := ApplyLocalFiltersJSON(raw, []string{"bogus=1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApplyLocalFiltersJSON_FieldToSyntheticMoneyComparison(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"localSpend":{"amount":"12.34","currency":"USD"},"dailyBudgetAmount":{"amount":"10.00","currency":"USD"}},
		{"localSpend":{"amount":"8.10","currency":"USD"},"dailyBudgetAmount":{"amount":"10.00","currency":"USD"}}
	]}`)

	got, err := ApplyLocalFiltersJSON(raw, []string{"localSpend > dailyBudgetAmount"})
	if err != nil {
		t.Fatalf("ApplyLocalFiltersJSON error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got, &envelope); err != nil {
		t.Fatalf("unmarshal filtered envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
}

func TestApplyLocalFiltersJSON_FieldToSyntheticMoneyComparisonCurrencyMismatch(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"localSpend":{"amount":"12.34","currency":"USD"},"dailyBudgetAmount":{"amount":"10.00","currency":"EUR"}}
	]}`)

	_, err := ApplyLocalFiltersJSON(raw, []string{"localSpend > dailyBudgetAmount"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err != nil && err.Error() != "currency mismatch: USD vs EUR" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyLocalFiltersJSON_NotEqualsMixedPresentAndMissing(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"searchTermText":"brand"},
		{"searchTermText":""},
		{}
	]}`)

	got, err := ApplyLocalFiltersJSON(raw, []string{`searchTermText != ''`})
	if err != nil {
		t.Fatalf("ApplyLocalFiltersJSON error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got, &envelope); err != nil {
		t.Fatalf("unmarshal filtered envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
	row := data[0].(map[string]any)
	if row["searchTermText"] != "brand" {
		t.Fatalf("searchTermText = %v, want brand", row["searchTermText"])
	}
}

func TestApplyLocalFiltersJSON_NotEqualsMissingComparedToNullIsFalse(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"searchTermText":"brand"},
		{"searchTermText":null},
		{}
	]}`)

	got, err := ApplyLocalFiltersJSON(raw, []string{`searchTermText != null`})
	if err != nil {
		t.Fatalf("ApplyLocalFiltersJSON error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got, &envelope); err != nil {
		t.Fatalf("unmarshal filtered envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
	row := data[0].(map[string]any)
	if row["searchTermText"] != "brand" {
		t.Fatalf("searchTermText = %v, want brand", row["searchTermText"])
	}
}
