package shared

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/fieldmeta"
)

// ParseMoneyFlag parses a money flag value like "100" or "100 USD".
// If no currency is specified, uses --currency flag or config default_currency.
// Returns {"amount": ..., "currency": ...}.
func ParseMoneyFlag(value string) (map[string]string, error) {
	parts := strings.Fields(value)
	var amount, currency string
	switch len(parts) {
	case 1:
		amount = parts[0]
		currency = Currency()
		if currency == "" {
			cfg, err := GetConfig()
			if err == nil && cfg.DefaultCurrency != "" {
				currency = cfg.DefaultCurrency
			}
		}
		if currency == "" {
			return nil, fmt.Errorf("no currency specified: use \"AMOUNT CURRENCY\" or set --currency / default_currency in config")
		}
	case 2:
		amount = parts[0]
		currency = parts[1]
	default:
		return nil, fmt.Errorf("invalid money format %q: use \"AMOUNT\" or \"AMOUNT CURRENCY\"", value)
	}
	return map[string]string{
		"amount":   amount,
		"currency": strings.ToUpper(currency),
	}, nil
}

// ParseTextList splits a comma-separated string into trimmed items.
// Individual items can be quoted to include commas: "hello, world","other"
func ParseTextList(value string) []string {
	var items []string
	var current strings.Builder
	inQuote := false
	for i := 0; i < len(value); i++ {
		ch := value[i]
		switch {
		case ch == '"':
			inQuote = !inQuote
		case ch == ',' && !inQuote:
			s := strings.TrimSpace(current.String())
			if s != "" {
				items = append(items, s)
			}
			current.Reset()
		default:
			current.WriteByte(ch)
		}
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		items = append(items, s)
	}
	return items
}

// ParseIDList splits a comma-separated ID flag value into trimmed IDs.
func ParseIDList(value string) ([]string, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, fmt.Errorf("no IDs provided")
	}

	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id == "" {
			return nil, fmt.Errorf("invalid empty ID in %q", value)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// NormalizeStatus normalizes a status flag value.
// Accepts: 0/1, pause/paused, enable/enabled, active, PAUSED/ENABLED/ACTIVE.
// activeValue should be "ENABLED" (campaigns, adgroups, ads) or "ACTIVE" (keywords, negatives).
func NormalizeStatus(input string, activeValue string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "0", "pause", "paused":
		return "PAUSED", nil
	case "1", "enable", "enabled", "active":
		return activeValue, nil
	default:
		return "", fmt.Errorf("invalid status %q: use 0/1, pause/enable, or PAUSED/%s", input, activeValue)
	}
}

// NormalizeAdChannelType normalizes an ad channel type flag value.
// Accepts: search, display (case-insensitive).
func NormalizeAdChannelType(input string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "search":
		return "SEARCH", nil
	case "display":
		return "DISPLAY", nil
	default:
		return "", fmt.Errorf("invalid ad channel type %q: use SEARCH or DISPLAY", input)
	}
}

// NormalizeBillingEvent normalizes a billing event flag value.
// Accepts: taps, impressions (case-insensitive).
func NormalizeBillingEvent(input string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "taps", "tap", "cpc":
		return "TAPS", nil
	case "impressions", "impression", "cpm":
		return "IMPRESSIONS", nil
	default:
		return "", fmt.Errorf("invalid billing event %q: use TAPS or IMPRESSIONS", input)
	}
}

// NormalizeMatchType normalizes a match type flag value.
// Accepts: broad, exact (case-insensitive).
func NormalizeMatchType(input string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "broad":
		return "BROAD", nil
	case "exact":
		return "EXACT", nil
	default:
		return "", fmt.Errorf("invalid match type %q: use BROAD or EXACT", input)
	}
}

// NormalizeGender normalizes a gender targeting value.
// Accepts: m, male, f, female (case-insensitive).
func NormalizeGender(input string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "m", "male":
		return "M", nil
	case "f", "female":
		return "F", nil
	default:
		return "", fmt.Errorf("invalid gender %q: use M or F", input)
	}
}

// NormalizeDeviceClass normalizes a device class targeting value.
// Accepts: iphone, ipad (case-insensitive).
func NormalizeDeviceClass(input string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "iphone":
		return "IPHONE", nil
	case "ipad":
		return "IPAD", nil
	default:
		return "", fmt.Errorf("invalid device class %q: use IPHONE or IPAD", input)
	}
}

// AddSelectorEqualsCondition appends an equality condition to a selector body.
// If the same field already has an EQUALS condition with the same single value,
// the selector is returned unchanged. Conflicting values are rejected.
func AddSelectorEqualsCondition(selector json.RawMessage, field, value string) (json.RawMessage, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return selector, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(selector, &payload); err != nil {
		return nil, fmt.Errorf("parsing selector: %w", err)
	}

	var conditions []any
	if rawConditions, ok := payload["conditions"]; ok {
		switch c := rawConditions.(type) {
		case []any:
			conditions = c
		case []map[string]any:
			for _, cond := range c {
				conditions = append(conditions, cond)
			}
		default:
			return nil, fmt.Errorf("invalid selector conditions")
		}
	}

	for _, raw := range conditions {
		cond, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		condField, _ := cond["field"].(string)
		if !strings.EqualFold(strings.TrimSpace(condField), field) {
			continue
		}
		operator, _ := cond["operator"].(string)
		if !strings.EqualFold(strings.TrimSpace(operator), "EQUALS") {
			continue
		}
		values, ok := cond["values"].([]any)
		if !ok || len(values) != 1 {
			continue
		}
		existing, ok := values[0].(string)
		if !ok {
			continue
		}
		if existing == value {
			return selector, nil
		}
		return nil, ValidationErrorf("conflicting selector condition for %s: %q does not match --%s %q", field, existing, kebabCase(field), value)
	}

	conditions = append(conditions, map[string]any{
		"field":    field,
		"operator": "EQUALS",
		"values":   []any{value},
	})
	payload["conditions"] = conditions

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("building selector: %w", err)
	}
	return json.RawMessage(data), nil
}

func kebabCase(field string) string {
	var b strings.Builder
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('-')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

// NormalizeStatusSelector rewrites selector conditions on the "status" field to
// use the canonical API values for the entity type. This allows friendly inputs
// like ACTIVE for ENABLED-based entities and returns a usage error for invalid
// status values before the API call.
func NormalizeStatusSelector(selector json.RawMessage, activeValue string) (json.RawMessage, error) {
	var payload map[string]any
	if err := json.Unmarshal(selector, &payload); err != nil {
		return selector, nil
	}

	rawConditions, ok := payload["conditions"].([]any)
	if !ok || len(rawConditions) == 0 {
		return selector, nil
	}

	changed := false
	for _, raw := range rawConditions {
		cond, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		field, _ := cond["field"].(string)
		if !isStatusSelectorField(field) {
			continue
		}
		values, ok := cond["values"].([]any)
		if !ok {
			continue
		}
		for i, rawValue := range values {
			s := strings.TrimSpace(fmt.Sprintf("%v", rawValue))
			if s == "" || s == "<nil>" {
				continue
			}
			normalized, err := NormalizeStatus(s, activeValue)
			if err != nil {
				return nil, ValidationErrorf("invalid status filter value %q: valid values are PAUSED and %s", s, activeValue)
			}
			if normalized != s {
				values[i] = normalized
				changed = true
			}
		}
	}

	if !changed {
		return selector, nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("normalizing status selector: %w", err)
	}
	return json.RawMessage(data), nil
}

func isStatusSelectorField(field string) bool {
	switch fieldmeta.CanonicalField(field) {
	case "status", "campaignStatus", "adGroupStatus", "keywordStatus":
		return true
	default:
		return false
	}
}
