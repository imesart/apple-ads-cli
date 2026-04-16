package shared

import (
	"encoding/json"
	"testing"
)

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		input       string
		activeValue string
		want        string
		wantErr     bool
	}{
		// Numeric shortcuts
		{"0", "ENABLED", "PAUSED", false},
		{"1", "ENABLED", "ENABLED", false},
		{"0", "ACTIVE", "PAUSED", false},
		{"1", "ACTIVE", "ACTIVE", false},

		// Textual (lowercase)
		{"pause", "ENABLED", "PAUSED", false},
		{"paused", "ENABLED", "PAUSED", false},
		{"enable", "ENABLED", "ENABLED", false},
		{"enabled", "ENABLED", "ENABLED", false},
		{"active", "ACTIVE", "ACTIVE", false},
		{"active", "ENABLED", "ENABLED", false},

		// Mixed case
		{"PAUSED", "ENABLED", "PAUSED", false},
		{"Pause", "ENABLED", "PAUSED", false},
		{"ENABLED", "ENABLED", "ENABLED", false},
		{"Enable", "ENABLED", "ENABLED", false},
		{"ACTIVE", "ACTIVE", "ACTIVE", false},

		// Whitespace
		{"  1  ", "ENABLED", "ENABLED", false},
		{"  pause  ", "ACTIVE", "PAUSED", false},

		// Invalid
		{"2", "ENABLED", "", true},
		{"maybe", "ENABLED", "", true},
		{"", "ENABLED", "", true},
	}

	for _, tt := range tests {
		got, err := NormalizeStatus(tt.input, tt.activeValue)
		if (err != nil) != tt.wantErr {
			t.Errorf("NormalizeStatus(%q, %q) error = %v, wantErr %v", tt.input, tt.activeValue, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeStatus(%q, %q) = %q, want %q", tt.input, tt.activeValue, got, tt.want)
		}
	}
}

func TestNormalizeAdChannelType(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"SEARCH", "SEARCH", false},
		{"search", "SEARCH", false},
		{"Search", "SEARCH", false},
		{"DISPLAY", "DISPLAY", false},
		{"display", "DISPLAY", false},
		{"  search  ", "SEARCH", false},
		{"", "", true},
		{"VIDEO", "", true},
	}

	for _, tt := range tests {
		got, err := NormalizeAdChannelType(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("NormalizeAdChannelType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeAdChannelType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeBillingEvent(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"TAPS", "TAPS", false},
		{"taps", "TAPS", false},
		{"tap", "TAPS", false},
		{"cpc", "TAPS", false},
		{"IMPRESSIONS", "IMPRESSIONS", false},
		{"impressions", "IMPRESSIONS", false},
		{"impression", "IMPRESSIONS", false},
		{"cpm", "IMPRESSIONS", false},
		{"  taps  ", "TAPS", false},
		{"", "", true},
		{"CLICKS", "", true},
	}

	for _, tt := range tests {
		got, err := NormalizeBillingEvent(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("NormalizeBillingEvent(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeBillingEvent(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeMatchType(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"BROAD", "BROAD", false},
		{"broad", "BROAD", false},
		{"Broad", "BROAD", false},
		{"EXACT", "EXACT", false},
		{"exact", "EXACT", false},
		{"  broad  ", "BROAD", false},
		{"", "", true},
		{"PHRASE", "", true},
	}

	for _, tt := range tests {
		got, err := NormalizeMatchType(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("NormalizeMatchType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeMatchType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeGender(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"M", "M", false},
		{"m", "M", false},
		{"male", "M", false},
		{"Male", "M", false},
		{"F", "F", false},
		{"f", "F", false},
		{"female", "F", false},
		{"Female", "F", false},
		{"  m  ", "M", false},
		{"", "", true},
		{"X", "", true},
		{"other", "", true},
	}

	for _, tt := range tests {
		got, err := NormalizeGender(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("NormalizeGender(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeGender(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeDeviceClass(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"IPHONE", "IPHONE", false},
		{"iphone", "IPHONE", false},
		{"iPhone", "IPHONE", false},
		{"IPAD", "IPAD", false},
		{"ipad", "IPAD", false},
		{"iPad", "IPAD", false},
		{"  iphone  ", "IPHONE", false},
		{"", "", true},
		{"ANDROID", "", true},
		{"mac", "", true},
	}

	for _, tt := range tests {
		got, err := NormalizeDeviceClass(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("NormalizeDeviceClass(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeDeviceClass(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseMoneyFlag(t *testing.T) {
	// Set global currency for tests that rely on default
	oldCurrency := globalCurrency
	defer func() { globalCurrency = oldCurrency }()

	// With default currency set
	globalCurrency = "USD"
	m, err := ParseMoneyFlag("1.50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["amount"] != "1.50" || m["currency"] != "USD" {
		t.Errorf("got %v, want amount=1.50 currency=USD", m)
	}

	// With explicit currency
	m, err = ParseMoneyFlag("2.00 EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["amount"] != "2.00" || m["currency"] != "EUR" {
		t.Errorf("got %v, want amount=2.00 currency=EUR", m)
	}

	// Currency uppercased
	m, err = ParseMoneyFlag("3.00 gbp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["currency"] != "GBP" {
		t.Errorf("got currency=%s, want GBP", m["currency"])
	}

	// No default currency → error
	globalCurrency = ""
	_, err = ParseMoneyFlag("1.50")
	if err == nil {
		t.Error("expected error when no default currency")
	}

	// Too many parts → error
	globalCurrency = "USD"
	_, err = ParseMoneyFlag("1.50 USD extra")
	if err == nil {
		t.Error("expected error for too many parts")
	}
}

func TestAddSelectorEqualsCondition_AppendsCondition(t *testing.T) {
	selector := json.RawMessage(`{"pagination":{"offset":0,"limit":1000},"conditions":[{"field":"status","operator":"EQUALS","values":["ACTIVE"]}]}`)

	got, err := AddSelectorEqualsCondition(selector, "adGroupId", "456")
	if err != nil {
		t.Fatalf("AddSelectorEqualsCondition() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("unmarshal selector: %v", err)
	}
	conditions := payload["conditions"].([]any)
	if len(conditions) != 2 {
		t.Fatalf("conditions len = %d, want 2", len(conditions))
	}
}

func TestAddSelectorEqualsCondition_SameValueIsNoOp(t *testing.T) {
	selector := json.RawMessage(`{"conditions":[{"field":"adGroupId","operator":"EQUALS","values":["456"]}]}`)

	got, err := AddSelectorEqualsCondition(selector, "adGroupId", "456")
	if err != nil {
		t.Fatalf("AddSelectorEqualsCondition() error = %v", err)
	}
	if string(got) != string(selector) {
		t.Fatalf("selector changed: got %s want %s", got, selector)
	}
}

func TestAddSelectorEqualsCondition_ConflictingValueFails(t *testing.T) {
	selector := json.RawMessage(`{"conditions":[{"field":"adGroupId","operator":"EQUALS","values":["123"]}]}`)

	_, err := AddSelectorEqualsCondition(selector, "adGroupId", "456")
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestNormalizeStatusSelector_NormalizesCanonicalStatusFieldAndNumericAlias(t *testing.T) {
	selector := json.RawMessage(`{"conditions":[{"field":"campaignStatus","operator":"EQUALS","values":[1]}]}`)

	got, err := NormalizeStatusSelector(selector, "ENABLED")
	if err != nil {
		t.Fatalf("NormalizeStatusSelector() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("unmarshal selector: %v", err)
	}
	conditions := payload["conditions"].([]any)
	first := conditions[0].(map[string]any)
	values := first["values"].([]any)
	if len(values) != 1 || values[0] != "ENABLED" {
		t.Fatalf("values = %v, want [ENABLED]", values)
	}
}
