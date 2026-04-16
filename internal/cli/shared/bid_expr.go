package shared

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/types"
)

// BidExprKind describes the type of bid expression.
type BidExprKind int

const (
	BidAbsolute BidExprKind = iota
	BidDelta
	BidMultiply
)

// BidExpr represents a parsed bid expression.
type BidExpr struct {
	Kind       BidExprKind
	Money      map[string]string // for BidAbsolute
	Delta      float64           // for BidDelta
	DeltaCur   string            // explicit currency on delta (may be empty)
	Multiplier float64           // for BidMultiply
}

// ParseBidExpr parses a bid flag value into a BidExpr.
//
// Formats:
//   - Absolute: "1.50", "1.50 USD"
//   - Delta: "+1", "-0.50", "+1 USD", "-0.50 EUR"
//   - Multiplier: "x1.1", "x0.9", "*1.1", "*0.9"
//   - Percent: "+10%", "-15%"
func ParseBidExpr(value string) (*BidExpr, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty bid expression")
	}

	// Percent: +10%, -15%
	if strings.HasSuffix(value, "%") {
		body := strings.TrimSuffix(value, "%")
		pct, err := strconv.ParseFloat(body, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid percent expression %q", value)
		}
		return &BidExpr{
			Kind:       BidMultiply,
			Multiplier: 1 + pct/100,
		}, nil
	}

	// Multiplier: x1.1, X1.1, *1.1
	if len(value) > 1 && (value[0] == 'x' || value[0] == 'X' || value[0] == '*') {
		m, err := strconv.ParseFloat(value[1:], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid multiplier expression %q", value)
		}
		if m < 0 {
			return nil, fmt.Errorf("multiplier cannot be negative: %q", value)
		}
		return &BidExpr{
			Kind:       BidMultiply,
			Multiplier: m,
		}, nil
	}

	// Delta: +1, -0.50, "+1 USD"
	if value[0] == '+' || value[0] == '-' {
		parts := strings.Fields(value)
		delta, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid delta expression %q", value)
		}
		var cur string
		if len(parts) == 2 {
			cur = strings.ToUpper(parts[1])
		} else if len(parts) > 2 {
			return nil, fmt.Errorf("invalid delta expression %q", value)
		}
		return &BidExpr{
			Kind:     BidDelta,
			Delta:    delta,
			DeltaCur: cur,
		}, nil
	}

	// Absolute: delegate to ParseMoneyFlag
	money, err := ParseMoneyFlag(value)
	if err != nil {
		return nil, err
	}
	return &BidExpr{
		Kind:  BidAbsolute,
		Money: money,
	}, nil
}

// IsRelative returns true if the expression needs a current value to resolve.
func (e *BidExpr) IsRelative() bool {
	return e.Kind == BidDelta || e.Kind == BidMultiply
}

// Resolve applies the expression to produce a money map {"amount": ..., "currency": ...}.
// For absolute expressions, current may be nil.
// For relative expressions, current must be non-nil.
func (e *BidExpr) Resolve(current *types.Money) (map[string]string, error) {
	switch e.Kind {
	case BidAbsolute:
		return e.Money, nil

	case BidDelta:
		if current == nil {
			return nil, fmt.Errorf("no current value to apply delta to")
		}
		cur, err := strconv.ParseFloat(current.Amount, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse current amount %q: %w", current.Amount, err)
		}
		if e.DeltaCur != "" && current.Currency != "" &&
			!strings.EqualFold(e.DeltaCur, current.Currency) {
			return nil, fmt.Errorf("delta currency %s does not match current currency %s", e.DeltaCur, current.Currency)
		}
		result := cur + e.Delta
		if result < 0 {
			result = 0
		}
		return map[string]string{
			"amount":   formatAmount(result),
			"currency": current.Currency,
		}, nil

	case BidMultiply:
		if current == nil {
			return nil, fmt.Errorf("no current value to apply multiplier to")
		}
		cur, err := strconv.ParseFloat(current.Amount, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse current amount %q: %w", current.Amount, err)
		}
		result := cur * e.Multiplier
		if result < 0 {
			result = 0
		}
		return map[string]string{
			"amount":   formatAmount(result),
			"currency": current.Currency,
		}, nil
	}

	return nil, fmt.Errorf("unknown bid expression kind: %d", e.Kind)
}

// formatAmount formats a float64 as a decimal string with 2 decimal places.
func formatAmount(v float64) string {
	v = math.Round(v*100) / 100
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// ExtractMoney extracts a Money value from a JSON-unmarshalled map.
// Returns nil if the key is missing or not a valid money object.
func ExtractMoney(m map[string]any, key string) *types.Money {
	v, ok := m[key]
	if !ok {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	amount, _ := obj["amount"].(string)
	currency, _ := obj["currency"].(string)
	if amount == "" {
		return nil
	}
	return &types.Money{Amount: amount, Currency: currency}
}
