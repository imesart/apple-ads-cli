package shared

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
	"github.com/imesart/apple-ads-cli/internal/fieldmeta"
	"github.com/imesart/apple-ads-cli/internal/output"
)

type localFieldKind int

const (
	localFieldString localFieldKind = iota
	localFieldNumber
	localFieldBool
	localFieldArray
)

type localFieldSpec struct {
	Path  []string
	Kind  localFieldKind
	IsID  bool
	Alias bool
}

// ApplyLocalFiltersJSON filters a standard {"data":[...]} JSON envelope locally.
// Filters use the same syntax as list-command filters and are combined with AND.
func ApplyLocalFiltersJSON(raw json.RawMessage, filters []string, entityIDName ...string) (json.RawMessage, error) {
	if len(filters) == 0 {
		return raw, nil
	}

	var envelope map[string]any
	if err := output.UnmarshalUseNumber(raw, &envelope); err != nil {
		return nil, fmt.Errorf("parsing filtered response: %w", err)
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
	parsed := make([]map[string]any, 0, len(filters))
	for _, expr := range filters {
		cond, err := parseFilter(expr)
		if err != nil {
			return nil, err
		}
		field := fmt.Sprintf("%v", cond["field"])
		spec, ok := resolveLocalField(schema, field)
		if !ok {
			return nil, fmt.Errorf("unknown field %q", field)
		}
		cond["_fieldSpec"] = spec
		if valueSpec, ok := resolveSyntheticValueSpec(schema, cond["values"]); ok {
			cond["values"] = []any{valueSpec}
		}
		parsed = append(parsed, cond)
	}

	filtered := make([]any, 0, len(rowsValue))
	for _, item := range rowsValue {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		match, err := rowMatchesAllFilters(row, parsed)
		if err != nil {
			return nil, err
		}
		if match {
			filtered = append(filtered, row)
		}
	}

	envelope["data"] = filtered
	out, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("encoding filtered response: %w", err)
	}
	return json.RawMessage(out), nil
}

func buildLocalFilterSchema(rows []any, entityIDName string) map[string]*localFieldSpec {
	schema := make(map[string]*localFieldSpec)
	for _, item := range rows {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		collectLocalFields(schema, nil, row)
	}
	registerEntityAliases(schema, fieldmeta.KindFromEntityIDName(entityIDName))
	return schema
}

func collectLocalFields(schema map[string]*localFieldSpec, prefix []string, obj map[string]any) {
	for key, value := range obj {
		path := append(append([]string{}, prefix...), key)
		if nested, ok := value.(map[string]any); ok {
			if isLocalMoneyMap(nested) {
				registerLocalAlias(schema, path, append(append([]string{}, path...), "amount"), localFieldNumber)
				registerLocalField(schema, append(path, "amount"), localFieldNumber, false, false)
				registerLocalField(schema, append(path, "currency"), localFieldString, false, false)
				continue
			}
			collectLocalFields(schema, path, nested)
			continue
		}
		registerLocalField(schema, path, inferLocalFieldKind(path, value), isIDPath(path), false)
	}
}

func registerLocalField(schema map[string]*localFieldSpec, path []string, kind localFieldKind, isID bool, alias bool) {
	key := normalizeLocalFieldPath(path)
	spec, ok := schema[key]
	if !ok {
		schema[key] = &localFieldSpec{Path: append([]string{}, path...), Kind: kind, IsID: isID, Alias: alias}
		return
	}
	if spec.Kind != kind && spec.Kind != localFieldNumber {
		spec.Kind = kind
	}
	if isID {
		spec.IsID = true
	}
}

func registerLocalAlias(schema map[string]*localFieldSpec, aliasPath []string, actualPath []string, kind localFieldKind) {
	key := normalizeLocalFieldPath(aliasPath)
	schema[key] = &localFieldSpec{
		Path:  append([]string{}, actualPath...),
		Kind:  kind,
		IsID:  false,
		Alias: true,
	}
}

func registerEntityAliases(schema map[string]*localFieldSpec, kind fieldmeta.EntityKind) {
	for _, alias := range fieldmeta.AliasInputs(kind) {
		target, ok := fieldmeta.AliasTarget(kind, alias)
		if !ok {
			continue
		}
		spec, ok := schema[normalizeLocalFilterField(target)]
		if !ok {
			continue
		}
		registerLocalAlias(schema, []string{alias}, spec.Path, spec.Kind)
	}
}

func resolveLocalField(schema map[string]*localFieldSpec, field string) (*localFieldSpec, bool) {
	spec, ok := schema[normalizeLocalFilterField(field)]
	return spec, ok
}

func rowMatchesAllFilters(row map[string]any, filters []map[string]any) (bool, error) {
	for _, cond := range filters {
		spec := cond["_fieldSpec"].(*localFieldSpec)
		values := cond["values"].([]any)
		ok, err := evaluateLocalCondition(row, spec, fmt.Sprintf("%v", cond["operator"]), values)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func evaluateLocalCondition(row map[string]any, spec *localFieldSpec, operator string, values []any) (bool, error) {
	value, exists := extractLocalPath(row, spec.Path)
	if !exists {
		value = defaultLocalValue(spec.Kind)
	}
	if spec.Alias && len(spec.Path) > 0 && spec.Path[len(spec.Path)-1] == "amount" {
		if parentValue, ok := extractLocalPath(row, spec.Path[:len(spec.Path)-1]); ok {
			value = parentValue
		}
	}
	if len(values) > 0 {
		if valueSpec, ok := values[0].(*localFieldSpec); ok {
			resolved, err := extractComparableLocalValue(row, valueSpec)
			if err != nil {
				return false, err
			}
			values = []any{resolved}
		}
	}

	switch operator {
	case "EQUALS":
		if spec.IsID {
			return compareExactID(value, values), nil
		}
		return compareLocalEquals(value, spec.Kind, values), nil
	case "NOT_EQUALS":
		if !exists {
			return compareMissingNotEquals(values), nil
		}
		if spec.IsID {
			return !compareExactID(value, values), nil
		}
		return !compareLocalEquals(value, spec.Kind, values), nil
	case "CONTAINS":
		return compareLocalContains(value, values), nil
	case "STARTSWITH":
		return strings.HasPrefix(toLocalString(value), fmt.Sprintf("%v", values[0])), nil
	case "ENDSWITH":
		return strings.HasSuffix(toLocalString(value), fmt.Sprintf("%v", values[0])), nil
	case "IN":
		if spec.IsID {
			return compareExactID(value, values), nil
		}
		for _, candidate := range values {
			if compareLocalEquals(value, spec.Kind, []any{candidate}) {
				return true, nil
			}
		}
		return false, nil
	case "LESS_THAN":
		return compareLocalNumeric(value, values, func(a, b float64) bool { return a < b })
	case "GREATER_THAN":
		return compareLocalNumeric(value, values, func(a, b float64) bool { return a > b })
	case "BETWEEN":
		if len(values) != 2 {
			return false, fmt.Errorf("BETWEEN requires two values")
		}
		actual := toLocalNumber(value)
		low := toLocalNumber(values[0])
		high := toLocalNumber(values[1])
		return actual >= low && actual <= high, nil
	case "CONTAINS_ALL":
		actualItems := toLocalStringSlice(value)
		for _, candidate := range values {
			if !containsLocalString(actualItems, fmt.Sprintf("%v", candidate)) {
				return false, nil
			}
		}
		return true, nil
	case "CONTAINS_ANY":
		actualItems := toLocalStringSlice(value)
		for _, candidate := range values {
			if containsLocalString(actualItems, fmt.Sprintf("%v", candidate)) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported operator %q", operator)
	}
}

func compareExactID(value any, values []any) bool {
	actual := localIDString(value)
	for _, candidate := range values {
		if actual == fmt.Sprintf("%v", candidate) {
			return true
		}
	}
	return false
}

func compareLocalEquals(value any, kind localFieldKind, values []any) bool {
	if len(values) == 0 {
		return false
	}
	target := values[0]
	if target == nil {
		return value == nil
	}
	switch kind {
	case localFieldNumber:
		if money, ok := value.(map[string]any); ok {
			amount, _ := extractMoneyValue(money)
			return toLocalNumber(amount) == toLocalNumber(target)
		}
		return toLocalNumber(value) == toLocalNumber(target)
	case localFieldBool:
		return strings.EqualFold(toLocalString(value), fmt.Sprintf("%v", target))
	case localFieldArray:
		actualItems := toLocalStringSlice(value)
		return len(actualItems) == 1 && actualItems[0] == fmt.Sprintf("%v", target)
	default:
		return toLocalString(value) == fmt.Sprintf("%v", target)
	}
}

func compareMissingNotEquals(values []any) bool {
	if len(values) == 0 {
		return true
	}
	target := values[0]
	if target == nil {
		return false
	}
	return fmt.Sprintf("%v", target) != ""
}

func compareLocalContains(value any, values []any) bool {
	if len(values) == 0 {
		return false
	}
	needle := fmt.Sprintf("%v", values[0])
	if items := toLocalStringSlice(value); len(items) > 1 {
		for _, item := range items {
			if strings.Contains(item, needle) {
				return true
			}
		}
		return false
	}
	return strings.Contains(toLocalString(value), needle)
}

func compareLocalNumeric(value any, values []any, fn func(a, b float64) bool) (bool, error) {
	if len(values) == 0 {
		return false, nil
	}
	if valueMoney, ok := value.(map[string]any); ok {
		leftAmount, leftCurrency := extractMoneyValue(valueMoney)
		rightAmount, rightCurrency, rightHasCurrency := extractComparableMoneyOperand(values[0])
		if rightHasCurrency && leftCurrency != "" && rightCurrency != "" && !strings.EqualFold(leftCurrency, rightCurrency) {
			return false, fmt.Errorf("currency mismatch: %s vs %s", leftCurrency, rightCurrency)
		}
		return fn(toLocalNumber(leftAmount), toLocalNumber(rightAmount)), nil
	}
	return fn(toLocalNumber(value), toLocalNumber(values[0])), nil
}

func resolveSyntheticValueSpec(schema map[string]*localFieldSpec, values any) (*localFieldSpec, bool) {
	parsed, ok := values.([]any)
	if !ok || len(parsed) != 1 {
		return nil, false
	}
	name := strings.TrimSpace(fmt.Sprintf("%v", parsed[0]))
	if !fieldmeta.IsCarriedSynthetic(name) {
		return nil, false
	}
	spec, ok := resolveLocalField(schema, name)
	return spec, ok
}

func extractComparableLocalValue(row map[string]any, spec *localFieldSpec) (any, error) {
	value, exists := extractLocalPath(row, spec.Path)
	if !exists {
		return nil, ValidationErrorf("field %q is not available from stdin for this request scope", strings.Join(spec.Path, "."))
	}
	if spec.Alias && len(spec.Path) > 0 && spec.Path[len(spec.Path)-1] == "amount" {
		parentValue, _ := extractLocalPath(row, spec.Path[:len(spec.Path)-1])
		return parentValue, nil
	}
	return value, nil
}

func extractComparableMoneyOperand(value any) (amount string, currency string, hasCurrency bool) {
	if money, ok := value.(map[string]any); ok {
		amount, currency = extractMoneyValue(money)
		return amount, currency, currency != ""
	}
	return toLocalString(value), "", false
}

func extractLocalPath(obj map[string]any, path []string) (any, bool) {
	var current any = obj
	for _, segment := range path {
		next, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = next[segment]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func defaultLocalValue(kind localFieldKind) any {
	switch kind {
	case localFieldNumber:
		return float64(0)
	case localFieldArray:
		return []any{}
	case localFieldBool:
		return false
	default:
		return ""
	}
}

func inferLocalFieldKind(path []string, value any) localFieldKind {
	if isIDPath(path) {
		return localFieldString
	}
	switch value.(type) {
	case float64, float32, int, int64, int32, json.Number:
		return localFieldNumber
	case bool:
		return localFieldBool
	case []any, []string:
		return localFieldArray
	default:
		return localFieldString
	}
}

func isIDPath(path []string) bool {
	if len(path) == 0 {
		return false
	}
	last := strings.ToLower(path[len(path)-1])
	return strings.HasSuffix(last, "id")
}

func isLocalMoneyMap(m map[string]any) bool {
	_, hasAmount := m["amount"]
	_, hasCurrency := m["currency"]
	return hasAmount && hasCurrency
}

func normalizeLocalFilterField(field string) string {
	parts := strings.Split(strings.TrimSpace(field), ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, columnname.Compact(columnname.FromField(part)))
	}
	return strings.Join(out, ".")
}

func normalizeLocalFieldPath(path []string) string {
	return normalizeLocalFilterField(strings.Join(path, "."))
}

func toLocalString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	case []any:
		if len(v) == 0 {
			return ""
		}
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, toLocalString(item))
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toLocalStringSlice(value any) []string {
	switch v := value.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, toLocalString(item))
		}
		return out
	case []string:
		return append([]string{}, v...)
	case nil:
		return nil
	default:
		return []string{toLocalString(v)}
	}
}

func containsLocalString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func toLocalNumber(value any) float64 {
	switch v := value.(type) {
	case nil:
		return 0
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f
	default:
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		return f
	}
}

func localIDString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case json.Number:
		return v.String()
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
