package shared

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/fieldmeta"
	"github.com/imesart/apple-ads-cli/internal/output"
)

type localSortKey struct {
	Field string
	Desc  bool
	Spec  *localFieldSpec
}

func splitSortsByExecution(sorts stringSlice, entityIDName string) (remote stringSlice, localRequired bool, err error) {
	kind := fieldmeta.KindFromEntityIDName(entityIDName)
	for _, expr := range sorts {
		order, err := parseSort(expr)
		if err != nil {
			return nil, false, err
		}
		field := strings.TrimSpace(fmt.Sprintf("%v", order["field"]))
		if _, ok := fieldmeta.AliasTarget(kind, field); ok {
			remote = append(remote, expr)
			continue
		}
		if isNativeCarriedSortField(kind, field) {
			remote = append(remote, expr)
			continue
		}
		if fieldmeta.IsCarriedSynthetic(field) {
			localRequired = true
			continue
		}
		remote = append(remote, expr)
	}
	return remote, localRequired, nil
}

func isNativeCarriedSortField(kind fieldmeta.EntityKind, field string) bool {
	canonical := fieldmeta.CanonicalSyntheticName(field)
	switch kind {
	case fieldmeta.EntityCampaign:
		switch canonical {
		case "adamId", "budgetAmount", "dailyBudgetAmount":
			return true
		}
	case fieldmeta.EntityAdGroup:
		switch canonical {
		case "defaultBidAmount", "cpaGoal":
			return true
		}
	case fieldmeta.EntityCreative:
		return canonical == "adamId"
	case fieldmeta.EntityProductPage:
		return canonical == "adamId"
	}
	return false
}

func applyLocalListSorts(resp any, sorts stringSlice, entityIDName string) (any, error) {
	raw, ok := resp.(json.RawMessage)
	if !ok {
		return nil, fmt.Errorf("local sorting requires JSON response data")
	}
	if ctx := currentSyntheticContext(); len(ctx) > 0 {
		raw = augmentWithContext(raw, ctx).(json.RawMessage)
	}
	sorted, err := ApplyLocalSortsJSON(raw, sorts, entityIDName)
	if err != nil {
		return nil, err
	}
	return sorted, nil
}

// ApplyLocalSortsJSON sorts a standard {"data":[...]} JSON envelope locally.
// Sort keys use the same field normalization as local filters and are applied
// in the order the user provided them.
func ApplyLocalSortsJSON(raw json.RawMessage, sorts stringSlice, entityIDName ...string) (json.RawMessage, error) {
	if len(sorts) == 0 {
		return raw, nil
	}

	var envelope map[string]any
	if err := output.UnmarshalUseNumber(raw, &envelope); err != nil {
		return nil, fmt.Errorf("parsing sorted response: %w", err)
	}

	rowsValue, ok := envelope["data"].([]any)
	if !ok {
		return raw, nil
	}
	if len(rowsValue) == 0 {
		return raw, nil
	}

	name := ""
	if len(entityIDName) > 0 {
		name = entityIDName[0]
	}
	schema := buildLocalFilterSchema(rowsValue, name)
	keys := make([]localSortKey, 0, len(sorts))
	for _, expr := range sorts {
		order, err := parseSort(expr)
		if err != nil {
			return nil, err
		}
		field := strings.TrimSpace(fmt.Sprintf("%v", order["field"]))
		spec, ok := resolveLocalField(schema, field)
		if !ok {
			return nil, fmt.Errorf("unknown field %q", field)
		}
		keys = append(keys, localSortKey{
			Field: field,
			Desc:  fmt.Sprintf("%v", order["sortOrder"]) == "DESCENDING",
			Spec:  spec,
		})
	}

	if err := validateSortCurrencies(rowsValue, keys); err != nil {
		return nil, err
	}

	sort.SliceStable(rowsValue, func(i, j int) bool {
		left, leftOK := rowsValue[i].(map[string]any)
		right, rightOK := rowsValue[j].(map[string]any)
		if !leftOK || !rightOK {
			return false
		}
		for _, key := range keys {
			cmp := compareLocalSortValues(left, right, key.Spec)
			if cmp == 0 {
				continue
			}
			if key.Desc {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})

	envelope["data"] = rowsValue
	out, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("encoding sorted response: %w", err)
	}
	return json.RawMessage(out), nil
}

func validateSortCurrencies(rows []any, keys []localSortKey) error {
	for _, key := range keys {
		if len(key.Spec.Path) == 0 || key.Spec.Path[len(key.Spec.Path)-1] == "currency" {
			continue
		}
		currency := ""
		for _, item := range rows {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			value, exists := extractLocalPath(row, key.Spec.Path)
			if key.Spec.Alias && len(key.Spec.Path) > 0 && key.Spec.Path[len(key.Spec.Path)-1] == "amount" {
				if parentValue, parentExists := extractLocalPath(row, key.Spec.Path[:len(key.Spec.Path)-1]); parentExists {
					value = parentValue
					exists = true
				}
			}
			if !exists {
				continue
			}
			if money, ok := value.(map[string]any); ok {
				_, rowCurrency := extractMoneyValue(money)
				if rowCurrency == "" {
					continue
				}
				if currency == "" {
					currency = rowCurrency
					continue
				}
				if !strings.EqualFold(currency, rowCurrency) {
					return ValidationErrorf("cannot sort field %q with mixed currencies: %s vs %s", key.Field, currency, rowCurrency)
				}
			}
		}
	}
	return nil
}

func compareLocalSortValues(left, right map[string]any, spec *localFieldSpec) int {
	leftValue, leftExists := extractLocalPath(left, spec.Path)
	rightValue, rightExists := extractLocalPath(right, spec.Path)
	if !leftExists {
		leftValue = defaultLocalValue(spec.Kind)
	}
	if !rightExists {
		rightValue = defaultLocalValue(spec.Kind)
	}
	if spec.Alias && len(spec.Path) > 0 && spec.Path[len(spec.Path)-1] == "amount" {
		if parentValue, ok := extractLocalPath(left, spec.Path[:len(spec.Path)-1]); ok {
			leftValue = parentValue
		}
		if parentValue, ok := extractLocalPath(right, spec.Path[:len(spec.Path)-1]); ok {
			rightValue = parentValue
		}
	}

	switch spec.Kind {
	case localFieldNumber:
		leftNumber := comparableLocalNumber(leftValue)
		rightNumber := comparableLocalNumber(rightValue)
		switch {
		case leftNumber < rightNumber:
			return -1
		case leftNumber > rightNumber:
			return 1
		default:
			return 0
		}
	case localFieldBool:
		leftBool := strings.EqualFold(toLocalString(leftValue), "true")
		rightBool := strings.EqualFold(toLocalString(rightValue), "true")
		switch {
		case !leftBool && rightBool:
			return -1
		case leftBool && !rightBool:
			return 1
		default:
			return 0
		}
	default:
		return strings.Compare(toLocalString(leftValue), toLocalString(rightValue))
	}
}

func comparableLocalNumber(value any) float64 {
	if money, ok := value.(map[string]any); ok {
		amount, _ := extractMoneyValue(money)
		return toLocalNumber(amount)
	}
	return toLocalNumber(value)
}
