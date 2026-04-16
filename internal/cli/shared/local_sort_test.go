package shared

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestApplyLocalSortsJSON_SortsMoneyAmounts(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"name":"high","dailyBudgetAmount":{"amount":"20","currency":"EUR"}},
		{"name":"low","dailyBudgetAmount":{"amount":"10","currency":"EUR"}}
	]}`)

	out, err := ApplyLocalSortsJSON(raw, stringSlice{"dailyBudgetAmount:asc"})
	if err != nil {
		t.Fatalf("ApplyLocalSortsJSON error: %v", err)
	}
	text := string(out)
	low := strings.Index(text, `"name":"low"`)
	high := strings.Index(text, `"name":"high"`)
	if low < 0 || high < 0 || low > high {
		t.Fatalf("expected rows sorted by money amount asc, got %s", text)
	}
}

func TestApplyLocalSortsJSON_RejectsMixedMoneyCurrencies(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"name":"euro","dailyBudgetAmount":{"amount":"10","currency":"EUR"}},
		{"name":"dollar","dailyBudgetAmount":{"amount":"10","currency":"USD"}}
	]}`)

	_, err := ApplyLocalSortsJSON(raw, stringSlice{"dailyBudgetAmount:asc"})
	if err == nil {
		t.Fatal("expected mixed-currency error")
	}
	if !strings.Contains(err.Error(), "mixed currencies") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyLocalSortsJSON_UsesMissingDefaults(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"name":"present","impressions":5},
		{"name":"missing"}
	]}`)

	out, err := ApplyLocalSortsJSON(raw, stringSlice{"impressions:asc"})
	if err != nil {
		t.Fatalf("ApplyLocalSortsJSON error: %v", err)
	}
	text := string(out)
	missing := strings.Index(text, `"name":"missing"`)
	present := strings.Index(text, `"name":"present"`)
	if missing < 0 || present < 0 || missing > present {
		t.Fatalf("expected missing numeric field to sort as 0, got %s", text)
	}
}
