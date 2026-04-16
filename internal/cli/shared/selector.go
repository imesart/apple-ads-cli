package shared

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
)

// stringSlice implements flag.Value for repeatable string flags.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

// SelectorFlags holds filter, sort, and selector flags for list commands.
type SelectorFlags struct {
	filters  stringSlice
	sorts    stringSlice
	selector *string
}

// BindSelectorFlags registers --filter, --sort, and --selector on the given flag set.
func BindSelectorFlags(fs *flag.FlagSet) *SelectorFlags {
	return BindNamedSelectorFlags(fs, "filter", "sort", "selector")
}

// BindNamedSelectorFlags registers selector-related flags with custom names.
func BindNamedSelectorFlags(fs *flag.FlagSet, filterName, sortName, selectorName string) *SelectorFlags {
	sf := &SelectorFlags{}
	fs.Var(&sf.filters, filterName, `Filter: "field=value" or "field OPERATOR value" (repeatable)`)
	fs.Var(&sf.sorts, sortName, `Sort: "field:asc" or "field:desc" (repeatable)`)
	sf.selector = fs.String(selectorName, "", `Selector input: inline JSON, @file.json, or @- for stdin`)
	return sf
}

// HasFlags returns true if any selector flag was provided.
func (sf *SelectorFlags) HasFlags() bool {
	return len(sf.filters) > 0 || len(sf.sorts) > 0 || *sf.selector != ""
}

// Build constructs a JSON selector from the provided flags.
func (sf *SelectorFlags) Build(limit, offset int, entityIDName string) (json.RawMessage, error) {
	return buildSelector(sf.filters, sf.sorts, *sf.selector, limit, offset, entityIDName)
}

// validOperators is the set of selector operators supported by the API.
var validOperators = map[string]bool{
	"EQUALS": true, "CONTAINS": true, "STARTSWITH": true, "ENDSWITH": true,
	"IN": true, "LESS_THAN": true, "GREATER_THAN": true, "BETWEEN": true,
	"CONTAINS_ALL": true, "CONTAINS_ANY": true,
}

var localOnlyOperators = map[string]bool{
	"NOT_EQUALS": true,
}

// normalizeOperator uppercases and canonicalizes an operator name.
func normalizeOperator(op string) string {
	op = strings.ToUpper(op)
	switch op {
	case "!=":
		return "NOT_EQUALS"
	case "<":
		return "LESS_THAN"
	case ">":
		return "GREATER_THAN"
	case "LESSTHAN":
		return "LESS_THAN"
	case "GREATERTHAN":
		return "GREATER_THAN"
	case "CONTAINSALL":
		return "CONTAINS_ALL"
	case "CONTAINSANY":
		return "CONTAINS_ANY"
	}
	return op
}

func isKnownOperator(op string) bool {
	return validOperators[op] || localOnlyOperators[op]
}

func isLocalOnlyOperator(op string) bool {
	return localOnlyOperators[op]
}

// parseFilterValues parses a value string into a slice of values.
// Bracket notation [a, b, c] is parsed as multiple values.
func parseFilterValues(s string) []any {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := s[1 : len(s)-1]
		parts := strings.Split(inner, ",")
		var values []any
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				values = append(values, parseFilterValue(p))
			}
		}
		return values
	}
	if s == "" {
		return nil
	}
	return []any{parseFilterValue(s)}
}

func parseFilterValue(s string) any {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	if strings.EqualFold(strings.TrimSpace(s), "null") {
		return nil
	}
	return s
}

// parseFilter parses a filter expression into a selector condition.
// Formats:
//
//	"field=value"               → EQUALS
//	"field OPERATOR value"      → named operator (case-insensitive)
//	"field IN [a, b, c]"       → set membership
//	"field BETWEEN [0, 5]"     → range
func parseFilter(expr string) (map[string]any, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty filter expression")
	}

	// Compact comparison aliases: field!=value, field<value, field>value
	for _, symbol := range []string{"!=", "<", ">"} {
		if idx := strings.Index(expr, symbol); idx > 0 && !strings.Contains(expr[:idx], " ") {
			field := strings.TrimSpace(expr[:idx])
			value := strings.TrimSpace(expr[idx+1:])
			if symbol == "!=" {
				value = strings.TrimSpace(expr[idx+len(symbol):])
			}
			if field != "" && value != "" {
				return map[string]any{
					"field":    field,
					"operator": normalizeOperator(symbol),
					"values":   parseFilterValues(value),
				}, nil
			}
		}
	}

	// Not equals: field != value
	if neqIdx := strings.Index(expr, "!="); neqIdx > 0 {
		field := strings.TrimSpace(expr[:neqIdx])
		value := strings.TrimSpace(expr[neqIdx+2:])
		if field != "" && value != "" {
			return map[string]any{
				"field":    field,
				"operator": "NOT_EQUALS",
				"values":   parseFilterValues(value),
			}, nil
		}
	}

	// Equality: field=value or field = value
	if eqIdx := strings.Index(expr, "="); eqIdx > 0 {
		field := strings.TrimSpace(expr[:eqIdx])
		value := strings.TrimSpace(expr[eqIdx+1:])
		if field != "" && value != "" {
			return map[string]any{
				"field":    field,
				"operator": "EQUALS",
				"values":   parseFilterValues(value),
			}, nil
		}
	}

	// Space-separated: field OPERATOR value(s)
	// Find the operator by scanning words after the first
	words := strings.Fields(expr)
	if len(words) < 2 {
		return nil, fmt.Errorf("invalid filter %q: use \"field=value\" or \"field OPERATOR value\"", expr)
	}

	field := words[0]
	op := normalizeOperator(words[1])

	if !isKnownOperator(op) {
		return nil, fmt.Errorf("unknown operator %q in filter %q\nValid: EQUALS, !=, CONTAINS, STARTSWITH, ENDSWITH, IN, LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY", words[1], expr)
	}

	// Everything after "field OPERATOR" is the value part.
	// Find position after the operator word in the original string.
	afterField := strings.TrimSpace(expr[len(field):])
	afterOp := strings.TrimSpace(afterField[len(words[1]):])

	values := parseFilterValues(afterOp)

	return map[string]any{
		"field":    field,
		"operator": op,
		"values":   values,
	}, nil
}

// parseSort parses "field:asc" or "field:desc" into a selector orderBy entry.
func parseSort(expr string) (map[string]any, error) {
	expr = strings.TrimSpace(expr)
	parts := strings.SplitN(expr, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid sort %q: use \"field:asc\" or \"field:desc\"", expr)
	}
	field := strings.TrimSpace(parts[0])
	order := strings.ToUpper(strings.TrimSpace(parts[1]))
	switch order {
	case "ASC", "ASCENDING":
		order = "ASCENDING"
	case "DESC", "DESCENDING":
		order = "DESCENDING"
	default:
		return nil, fmt.Errorf("invalid sort order %q: use asc or desc", parts[1])
	}
	return map[string]any{
		"field":     field,
		"sortOrder": order,
	}, nil
}

// buildSelector builds a JSON selector from flags or raw JSON.
func buildSelector(filters stringSlice, sorts stringSlice, selector string, limit int, offset int, entityIDName string) (json.RawMessage, error) {
	if selector != "" {
		if len(filters) > 0 {
			return nil, UsageError("--filter and --selector are mutually exclusive")
		}
		if len(sorts) > 0 {
			return nil, UsageError("--sort and --selector are mutually exclusive; include orderBy in the selector JSON instead")
		}
		data, err := ReadJSONInputArg(selector)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(data), nil
	}
	sel := map[string]any{
		"pagination": map[string]any{
			"offset": offset,
			"limit":  limit,
		},
	}
	if len(filters) > 0 {
		conditions := make([]map[string]any, 0, len(filters))
		for _, f := range filters {
			cond, err := parseFilter(f)
			if err != nil {
				return nil, err
			}
			if isLocalOnlyOperator(fmt.Sprintf("%v", cond["operator"])) {
				return nil, UsageErrorf("operator %q is only supported for local filtering", "!=")
			}
			cond, err = normalizeSelectorCondition(entityIDName, cond)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, cond)
		}
		sel["conditions"] = conditions
	}
	if len(sorts) > 0 {
		orderBy := make([]map[string]any, 0, len(sorts))
		for _, s := range sorts {
			order, err := parseSort(s)
			if err != nil {
				return nil, err
			}
			order = normalizeSelectorSort(entityIDName, order)
			orderBy = append(orderBy, order)
		}
		sel["orderBy"] = orderBy
	}
	data, err := json.Marshal(sel)
	if err != nil {
		return nil, fmt.Errorf("building selector: %w", err)
	}
	return data, nil
}

func splitFiltersByExecution(filters stringSlice) (stringSlice, stringSlice, error) {
	var remote stringSlice
	var local stringSlice
	for _, f := range filters {
		cond, err := parseFilter(f)
		if err != nil {
			return nil, nil, err
		}
		if isLocalOnlyOperator(fmt.Sprintf("%v", cond["operator"])) {
			local = append(local, f)
			continue
		}
		remote = append(remote, f)
	}
	return remote, local, nil
}
