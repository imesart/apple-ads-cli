package shared

import (
	"fmt"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/fieldmeta"
)

func normalizeSelectorFieldForEntity(entityIDName string, field string) string {
	canonical := fieldmeta.CanonicalField(field)
	if target, ok := fieldmeta.AliasTarget(fieldmeta.KindFromEntityIDName(entityIDName), canonical); ok {
		return target
	}
	return canonical
}

func resolveSyntheticFilterValue(entityIDName string, field string, value any) (any, bool, error) {
	raw := strings.TrimSpace(fmt.Sprintf("%v", value))
	if !fieldmeta.IsCarriedSynthetic(raw) {
		return value, false, nil
	}

	ctx := currentSyntheticContext()
	if len(ctx) == 0 {
		return nil, false, ValidationErrorf("field %q is not available from stdin for this request scope", raw)
	}
	resolved, ok := ctx[fieldmeta.CanonicalSyntheticName(raw)]
	if !ok {
		return nil, false, ValidationErrorf("field %q is not available from stdin for this request scope", raw)
	}

	if isMoneyMapValue(resolved) {
		amount, currency := extractMoneyValue(resolved)
		if amount == "" {
			return nil, false, ValidationErrorf("field %q is not available from stdin for this request scope", raw)
		}
		_ = currency
		return amount, true, nil
	}
	return resolved, true, nil
}

func normalizeSelectorCondition(entityIDName string, cond map[string]any) (map[string]any, error) {
	field := strings.TrimSpace(fmt.Sprintf("%v", cond["field"]))
	cond["field"] = normalizeSelectorFieldForEntity(entityIDName, field)

	values, ok := cond["values"].([]any)
	if !ok || len(values) == 0 {
		return cond, nil
	}
	for i, value := range values {
		resolved, changed, err := resolveSyntheticFilterValue(entityIDName, field, value)
		if err != nil {
			return nil, err
		}
		if changed {
			values[i] = resolved
		}
	}
	return cond, nil
}

func normalizeSelectorSort(entityIDName string, order map[string]any) map[string]any {
	field := strings.TrimSpace(fmt.Sprintf("%v", order["field"]))
	order["field"] = normalizeSelectorFieldForEntity(entityIDName, field)
	return order
}

func isMoneyMapValue(v any) bool {
	_, ok := v.(map[string]any)
	if !ok {
		return false
	}
	amount, currency := extractMoneyValue(v)
	return amount != "" || currency != ""
}

func extractMoneyValue(v any) (string, string) {
	m, ok := v.(map[string]any)
	if !ok {
		return "", ""
	}
	var amount, currency string
	if raw, ok := m["amount"]; ok {
		amount = strings.TrimSpace(fmt.Sprintf("%v", raw))
	}
	if raw, ok := m["currency"]; ok {
		currency = strings.TrimSpace(fmt.Sprintf("%v", raw))
	}
	return amount, currency
}
