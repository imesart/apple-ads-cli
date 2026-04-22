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

func TestApplyLocalSortsJSON_SortsStringsCaseInsensitively(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"name":"beta"},
		{"name":"Alpha"},
		{"name":"zeta"}
	]}`)

	out, err := ApplyLocalSortsJSON(raw, stringSlice{"name:asc"})
	if err != nil {
		t.Fatalf("ApplyLocalSortsJSON error: %v", err)
	}
	text := string(out)
	alpha := strings.Index(text, `"name":"Alpha"`)
	beta := strings.Index(text, `"name":"beta"`)
	zeta := strings.Index(text, `"name":"zeta"`)
	if alpha < 0 || beta < 0 || zeta < 0 || alpha > beta || beta > zeta {
		t.Fatalf("expected case-insensitive string sort, got %s", text)
	}
}

func TestApplyLocalSortsJSON_SortsDescending(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"name":"alpha"},
		{"name":"zeta"},
		{"name":"mu"}
	]}`)

	out, err := ApplyLocalSortsJSON(raw, stringSlice{"name:desc"})
	if err != nil {
		t.Fatalf("ApplyLocalSortsJSON error: %v", err)
	}
	text := string(out)
	zeta := strings.Index(text, `"name":"zeta"`)
	mu := strings.Index(text, `"name":"mu"`)
	alpha := strings.Index(text, `"name":"alpha"`)
	if zeta < 0 || mu < 0 || alpha < 0 || zeta > mu || mu > alpha {
		t.Fatalf("expected descending name sort, got %s", text)
	}
}

func TestApplyLocalSortsJSON_MultipleSortsAppliedInOrder(t *testing.T) {
	raw := json.RawMessage(`{"data":[
		{"team":"beta","score":5},
		{"team":"alpha","score":10},
		{"team":"alpha","score":1},
		{"team":"beta","score":2}
	]}`)

	out, err := ApplyLocalSortsJSON(raw, stringSlice{"team:asc", "score:desc"})
	if err != nil {
		t.Fatalf("ApplyLocalSortsJSON error: %v", err)
	}
	text := string(out)
	wantOrder := []string{
		`"score":10,"team":"alpha"`,
		`"score":1,"team":"alpha"`,
		`"score":5,"team":"beta"`,
		`"score":2,"team":"beta"`,
	}
	prev := -1
	for _, fragment := range wantOrder {
		idx := strings.Index(text, fragment)
		if idx < 0 {
			t.Fatalf("missing row %q in %s", fragment, text)
		}
		if idx < prev {
			t.Fatalf("rows out of order at %q: got %s", fragment, text)
		}
		prev = idx
	}
}
