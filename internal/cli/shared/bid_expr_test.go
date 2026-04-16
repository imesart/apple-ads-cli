package shared

import (
	"testing"

	"github.com/imesart/apple-ads-cli/internal/types"
)

func TestParseBidExpr_Absolute(t *testing.T) {
	expr, err := ParseBidExpr("1.50 USD")
	if err != nil {
		t.Fatalf("ParseBidExpr(\"1.50 USD\") error: %v", err)
	}
	if expr.Kind != BidAbsolute {
		t.Errorf("kind = %d, want BidAbsolute", expr.Kind)
	}
	if expr.Money["amount"] != "1.50" || expr.Money["currency"] != "USD" {
		t.Errorf("money = %v, want amount=1.50 currency=USD", expr.Money)
	}
}

func TestParseBidExpr_DeltaPositive(t *testing.T) {
	expr, err := ParseBidExpr("+1.00")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if expr.Kind != BidDelta {
		t.Errorf("kind = %d, want BidDelta", expr.Kind)
	}
	if expr.Delta != 1.0 {
		t.Errorf("delta = %f, want 1.0", expr.Delta)
	}
	if expr.DeltaCur != "" {
		t.Errorf("deltaCur = %q, want empty", expr.DeltaCur)
	}
}

func TestParseBidExpr_DeltaNegative(t *testing.T) {
	expr, err := ParseBidExpr("-0.50")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if expr.Kind != BidDelta {
		t.Errorf("kind = %d, want BidDelta", expr.Kind)
	}
	if expr.Delta != -0.50 {
		t.Errorf("delta = %f, want -0.50", expr.Delta)
	}
}

func TestParseBidExpr_DeltaWithCurrency(t *testing.T) {
	expr, err := ParseBidExpr("+1.00 EUR")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if expr.Kind != BidDelta {
		t.Errorf("kind = %d, want BidDelta", expr.Kind)
	}
	if expr.Delta != 1.0 {
		t.Errorf("delta = %f, want 1.0", expr.Delta)
	}
	if expr.DeltaCur != "EUR" {
		t.Errorf("deltaCur = %q, want EUR", expr.DeltaCur)
	}
}

func TestParseBidExpr_MultiplierX(t *testing.T) {
	for _, prefix := range []string{"x", "X", "*"} {
		expr, err := ParseBidExpr(prefix + "1.10")
		if err != nil {
			t.Fatalf("error with prefix %q: %v", prefix, err)
		}
		if expr.Kind != BidMultiply {
			t.Errorf("prefix %q: kind = %d, want BidMultiply", prefix, expr.Kind)
		}
		if expr.Multiplier != 1.10 {
			t.Errorf("prefix %q: multiplier = %f, want 1.10", prefix, expr.Multiplier)
		}
	}
}

func TestParseBidExpr_MultiplierZero(t *testing.T) {
	expr, err := ParseBidExpr("x0")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if expr.Kind != BidMultiply || expr.Multiplier != 0 {
		t.Errorf("got kind=%d multiplier=%f, want BidMultiply 0", expr.Kind, expr.Multiplier)
	}
}

func TestParseBidExpr_MultiplierNegative(t *testing.T) {
	_, err := ParseBidExpr("x-1")
	if err == nil {
		t.Fatal("expected error for negative multiplier")
	}
}

func TestParseBidExpr_PercentPositive(t *testing.T) {
	expr, err := ParseBidExpr("+10%")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if expr.Kind != BidMultiply {
		t.Errorf("kind = %d, want BidMultiply", expr.Kind)
	}
	if expr.Multiplier != 1.10 {
		t.Errorf("multiplier = %f, want 1.10", expr.Multiplier)
	}
}

func TestParseBidExpr_PercentNegative(t *testing.T) {
	expr, err := ParseBidExpr("-15%")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if expr.Kind != BidMultiply {
		t.Errorf("kind = %d, want BidMultiply", expr.Kind)
	}
	if expr.Multiplier != 0.85 {
		t.Errorf("multiplier = %f, want 0.85", expr.Multiplier)
	}
}

func TestParseBidExpr_PercentZero(t *testing.T) {
	expr, err := ParseBidExpr("+0%")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if expr.Multiplier != 1.0 {
		t.Errorf("multiplier = %f, want 1.0", expr.Multiplier)
	}
}

func TestParseBidExpr_Empty(t *testing.T) {
	_, err := ParseBidExpr("")
	if err == nil {
		t.Fatal("expected error for empty expression")
	}
}

func TestParseBidExpr_InvalidPercent(t *testing.T) {
	_, err := ParseBidExpr("+abc%")
	if err == nil {
		t.Fatal("expected error for invalid percent")
	}
}

func TestParseBidExpr_InvalidMultiplier(t *testing.T) {
	_, err := ParseBidExpr("xabc")
	if err == nil {
		t.Fatal("expected error for invalid multiplier")
	}
}

func TestParseBidExpr_InvalidDelta(t *testing.T) {
	_, err := ParseBidExpr("+abc")
	if err == nil {
		t.Fatal("expected error for invalid delta")
	}
}

func TestBidExpr_IsRelative(t *testing.T) {
	if (&BidExpr{Kind: BidAbsolute}).IsRelative() {
		t.Error("absolute should not be relative")
	}
	if !(&BidExpr{Kind: BidDelta}).IsRelative() {
		t.Error("delta should be relative")
	}
	if !(&BidExpr{Kind: BidMultiply}).IsRelative() {
		t.Error("multiply should be relative")
	}
}

func TestBidExpr_Resolve_Absolute(t *testing.T) {
	expr := &BidExpr{
		Kind:  BidAbsolute,
		Money: map[string]string{"amount": "1.50", "currency": "USD"},
	}
	result, err := expr.Resolve(nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "1.50" || result["currency"] != "USD" {
		t.Errorf("result = %v", result)
	}
}

func TestBidExpr_Resolve_DeltaPositive(t *testing.T) {
	expr := &BidExpr{Kind: BidDelta, Delta: 1.0}
	current := &types.Money{Amount: "2.00", Currency: "USD"}
	result, err := expr.Resolve(current)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "3.00" {
		t.Errorf("amount = %q, want 3.00", result["amount"])
	}
	if result["currency"] != "USD" {
		t.Errorf("currency = %q, want USD", result["currency"])
	}
}

func TestBidExpr_Resolve_DeltaNegative(t *testing.T) {
	expr := &BidExpr{Kind: BidDelta, Delta: -0.25}
	current := &types.Money{Amount: "1.00", Currency: "EUR"}
	result, err := expr.Resolve(current)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "0.75" {
		t.Errorf("amount = %q, want 0.75", result["amount"])
	}
}

func TestBidExpr_Resolve_DeltaFloorsAtZero(t *testing.T) {
	expr := &BidExpr{Kind: BidDelta, Delta: -5.0}
	current := &types.Money{Amount: "2.00", Currency: "USD"}
	result, err := expr.Resolve(current)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "0.00" {
		t.Errorf("amount = %q, want 0.00", result["amount"])
	}
}

func TestBidExpr_Resolve_Multiply(t *testing.T) {
	expr := &BidExpr{Kind: BidMultiply, Multiplier: 1.5}
	current := &types.Money{Amount: "2.00", Currency: "EUR"}
	result, err := expr.Resolve(current)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "3.00" {
		t.Errorf("amount = %q, want 3.00", result["amount"])
	}
	if result["currency"] != "EUR" {
		t.Errorf("currency = %q, want EUR", result["currency"])
	}
}

func TestBidExpr_Resolve_MultiplyRounding(t *testing.T) {
	expr := &BidExpr{Kind: BidMultiply, Multiplier: 1.1}
	current := &types.Money{Amount: "1.05", Currency: "USD"}
	result, err := expr.Resolve(current)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "1.16" {
		t.Errorf("amount = %q, want 1.16 (1.05*1.1=1.155 rounded)", result["amount"])
	}
}

func TestBidExpr_Resolve_MultiplyFloorsAtZero(t *testing.T) {
	// Extreme negative percent (-200%) gives multiplier -1.0
	expr := &BidExpr{Kind: BidMultiply, Multiplier: -1.0}
	current := &types.Money{Amount: "2.00", Currency: "USD"}
	result, err := expr.Resolve(current)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "0.00" {
		t.Errorf("amount = %q, want 0.00", result["amount"])
	}
}

func TestBidExpr_Resolve_DeltaCurrencyMismatch(t *testing.T) {
	expr := &BidExpr{Kind: BidDelta, Delta: 1.0, DeltaCur: "EUR"}
	current := &types.Money{Amount: "2.00", Currency: "USD"}
	_, err := expr.Resolve(current)
	if err == nil {
		t.Fatal("expected currency mismatch error")
	}
}

func TestBidExpr_Resolve_DeltaCurrencyMatch(t *testing.T) {
	expr := &BidExpr{Kind: BidDelta, Delta: 1.0, DeltaCur: "USD"}
	current := &types.Money{Amount: "2.00", Currency: "USD"}
	result, err := expr.Resolve(current)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "3.00" {
		t.Errorf("amount = %q, want 3.00", result["amount"])
	}
}

func TestBidExpr_Resolve_DeltaCurrencyCaseInsensitive(t *testing.T) {
	expr := &BidExpr{Kind: BidDelta, Delta: 1.0, DeltaCur: "usd"}
	current := &types.Money{Amount: "2.00", Currency: "USD"}
	result, err := expr.Resolve(current)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["amount"] != "3.00" {
		t.Errorf("amount = %q, want 3.00", result["amount"])
	}
}

func TestBidExpr_Resolve_RelativeNoCurrent(t *testing.T) {
	expr := &BidExpr{Kind: BidDelta, Delta: 1.0}
	_, err := expr.Resolve(nil)
	if err == nil {
		t.Fatal("expected error for nil current with delta")
	}

	expr2 := &BidExpr{Kind: BidMultiply, Multiplier: 1.1}
	_, err = expr2.Resolve(nil)
	if err == nil {
		t.Fatal("expected error for nil current with multiplier")
	}
}

func TestExtractMoney(t *testing.T) {
	m := map[string]any{
		"bidAmount": map[string]any{
			"amount":   "1.50",
			"currency": "USD",
		},
	}
	money := ExtractMoney(m, "bidAmount")
	if money == nil {
		t.Fatal("expected non-nil money")
	}
	if money.Amount != "1.50" || money.Currency != "USD" {
		t.Errorf("money = %+v", money)
	}
}

func TestExtractMoney_Missing(t *testing.T) {
	m := map[string]any{}
	if ExtractMoney(m, "bidAmount") != nil {
		t.Error("expected nil for missing key")
	}
}

func TestExtractMoney_NilMap(t *testing.T) {
	if ExtractMoney(nil, "bidAmount") != nil {
		t.Error("expected nil for nil map")
	}
}

func TestExtractMoney_NotObject(t *testing.T) {
	m := map[string]any{"bidAmount": "not an object"}
	if ExtractMoney(m, "bidAmount") != nil {
		t.Error("expected nil for non-object value")
	}
}

func TestExtractMoney_EmptyAmount(t *testing.T) {
	m := map[string]any{
		"bidAmount": map[string]any{
			"amount":   "",
			"currency": "USD",
		},
	}
	if ExtractMoney(m, "bidAmount") != nil {
		t.Error("expected nil for empty amount")
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{1.0, "1.00"},
		{1.5, "1.50"},
		{1.155, "1.16"},
		{1.154, "1.15"},
		{0, "0.00"},
		{100.999, "101.00"},
	}
	for _, tc := range tests {
		got := formatAmount(tc.in)
		if got != tc.want {
			t.Errorf("formatAmount(%f) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
