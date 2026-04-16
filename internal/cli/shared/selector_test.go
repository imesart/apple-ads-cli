package shared

import (
	"encoding/json"
	"testing"
)

func TestParseFilter_Equality(t *testing.T) {
	tests := []string{
		"status=ENABLED",
		"status = ENABLED",
		"status= ENABLED",
		"status =ENABLED",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			cond, err := parseFilter(input)
			if err != nil {
				t.Fatal(err)
			}
			if cond["field"] != "status" {
				t.Errorf("field = %q, want status", cond["field"])
			}
			if cond["operator"] != "EQUALS" {
				t.Errorf("operator = %q, want EQUALS", cond["operator"])
			}
			vals := cond["values"].([]any)
			if len(vals) != 1 || vals[0] != "ENABLED" {
				t.Errorf("values = %v, want [ENABLED]", vals)
			}
		})
	}
}

func TestParseFilter_OperatorSpaceSeparated(t *testing.T) {
	tests := []struct {
		input    string
		field    string
		operator string
		values   []any
	}{
		{"name STARTSWITH MyApp", "name", "STARTSWITH", []any{"MyApp"}},
		{"name startswith MyApp", "name", "STARTSWITH", []any{"MyApp"}},
		{"name contains hello world", "name", "CONTAINS", []any{"hello world"}},
		{"field < 100", "field", "LESS_THAN", []any{"100"}},
		{"field > 50", "field", "GREATER_THAN", []any{"50"}},
		{"field != 50", "field", "NOT_EQUALS", []any{"50"}},
		{"field LESS_THAN 100", "field", "LESS_THAN", []any{"100"}},
		{"field lessthan 100", "field", "LESS_THAN", []any{"100"}},
		{"field greaterthan 50", "field", "GREATER_THAN", []any{"50"}},
		{"field containsall [a, b]", "field", "CONTAINS_ALL", []any{"a", "b"}},
		{"field containsany [x, y]", "field", "CONTAINS_ANY", []any{"x", "y"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cond, err := parseFilter(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if cond["field"] != tt.field {
				t.Errorf("field = %q, want %q", cond["field"], tt.field)
			}
			if cond["operator"] != tt.operator {
				t.Errorf("operator = %q, want %q", cond["operator"], tt.operator)
			}
			vals := cond["values"].([]any)
			if len(vals) != len(tt.values) {
				t.Fatalf("values len = %d, want %d", len(vals), len(tt.values))
			}
			for i, v := range vals {
				if v != tt.values[i] {
					t.Errorf("values[%d] = %q, want %q", i, v, tt.values[i])
				}
			}
		})
	}
}

func TestParseFilter_ArrayValues(t *testing.T) {
	cond, err := parseFilter("adamId IN [123, 456, 789]")
	if err != nil {
		t.Fatal(err)
	}
	if cond["field"] != "adamId" {
		t.Errorf("field = %q, want adamId", cond["field"])
	}
	if cond["operator"] != "IN" {
		t.Errorf("operator = %q, want IN", cond["operator"])
	}
	vals := cond["values"].([]any)
	if len(vals) != 3 {
		t.Fatalf("values len = %d, want 3", len(vals))
	}
	expected := []string{"123", "456", "789"}
	for i, v := range vals {
		if v != expected[i] {
			t.Errorf("values[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestParseFilter_Between(t *testing.T) {
	cond, err := parseFilter("budget BETWEEN [0, 5]")
	if err != nil {
		t.Fatal(err)
	}
	if cond["operator"] != "BETWEEN" {
		t.Errorf("operator = %q, want BETWEEN", cond["operator"])
	}
	vals := cond["values"].([]any)
	if len(vals) != 2 || vals[0] != "0" || vals[1] != "5" {
		t.Errorf("values = %v, want [0 5]", vals)
	}
}

func TestParseFilter_InvalidOperator(t *testing.T) {
	_, err := parseFilter("field BOGUS value")
	if err == nil {
		t.Error("expected error for unknown operator")
	}
}

func TestParseFilter_CompactComparisonAliases(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    any
	}{
		{"impressions<100", "LESS_THAN", "100"},
		{"localSpend.amount>10", "GREATER_THAN", "10"},
		{"name!=Test", "NOT_EQUALS", "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cond, err := parseFilter(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if cond["operator"] != tt.operator {
				t.Fatalf("operator = %q, want %q", cond["operator"], tt.operator)
			}
			vals := cond["values"].([]any)
			if len(vals) != 1 || vals[0] != tt.value {
				t.Fatalf("values = %v, want [%s]", vals, tt.value)
			}
		})
	}
}

func TestParseFilter_NullAndQuotedValues(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{`field != null`, nil},
		{`field != ''`, ""},
		{`field != ""`, ""},
		{`field = 'null'`, "null"},
		{`field = "null"`, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cond, err := parseFilter(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			vals := cond["values"].([]any)
			if len(vals) != 1 || vals[0] != tt.want {
				t.Fatalf("values = %#v, want [%#v]", vals, tt.want)
			}
		})
	}
}

func TestParseFilter_Empty(t *testing.T) {
	_, err := parseFilter("")
	if err == nil {
		t.Error("expected error for empty filter")
	}
}

func TestParseSort(t *testing.T) {
	tests := []struct {
		input string
		field string
		order string
	}{
		{"name:asc", "name", "ASCENDING"},
		{"name:ASC", "name", "ASCENDING"},
		{"name:ascending", "name", "ASCENDING"},
		{"id:desc", "id", "DESCENDING"},
		{"id:DESC", "id", "DESCENDING"},
		{"id:descending", "id", "DESCENDING"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			order, err := parseSort(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if order["field"] != tt.field {
				t.Errorf("field = %q, want %q", order["field"], tt.field)
			}
			if order["sortOrder"] != tt.order {
				t.Errorf("sortOrder = %q, want %q", order["sortOrder"], tt.order)
			}
		})
	}
}

func TestParseSort_Invalid(t *testing.T) {
	_, err := parseSort("name")
	if err == nil {
		t.Error("expected error for missing colon")
	}
	_, err = parseSort("name:sideways")
	if err == nil {
		t.Error("expected error for invalid sort order")
	}
}

func TestBuildReportRequest_UsesSelectorSortOrder(t *testing.T) {
	req := buildReportRequest("2026-03-18", "2026-03-25", "", "", "UTC", true, true, false, "", "impressions:desc")

	selector, ok := req["selector"].(map[string]any)
	if !ok {
		t.Fatalf("selector type = %T, want map[string]any", req["selector"])
	}

	orderBy, ok := selector["orderBy"].([]map[string]any)
	if ok {
		if len(orderBy) != 1 {
			t.Fatalf("orderBy len = %d, want 1", len(orderBy))
		}
		if orderBy[0]["field"] != "impressions" {
			t.Fatalf("field = %v, want impressions", orderBy[0]["field"])
		}
		if orderBy[0]["sortOrder"] != "DESCENDING" {
			t.Fatalf("sortOrder = %v, want DESCENDING", orderBy[0]["sortOrder"])
		}
		return
	}

	raw, err := json.Marshal(selector["orderBy"])
	if err != nil {
		t.Fatalf("marshal orderBy: %v", err)
	}
	t.Fatalf("orderBy type = %T, value = %s", selector["orderBy"], raw)
}

func TestParseFilterValues_Bracket(t *testing.T) {
	vals := parseFilterValues("[a, b, c]")
	if len(vals) != 3 || vals[0] != "a" || vals[1] != "b" || vals[2] != "c" {
		t.Errorf("parseFilterValues([a, b, c]) = %v, want [a b c]", vals)
	}
}

func TestParseFilterValues_Single(t *testing.T) {
	vals := parseFilterValues("hello")
	if len(vals) != 1 || vals[0] != "hello" {
		t.Errorf("parseFilterValues(hello) = %v, want [hello]", vals)
	}
}

func TestParseFilterValues_QuotedEmptyString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`''`, ""},
		{`""`, ""},
		{`'null'`, "null"},
		{`"null"`, "null"},
	}
	for _, tt := range tests {
		vals := parseFilterValues(tt.input)
		if len(vals) != 1 || vals[0] != tt.want {
			t.Errorf("parseFilterValues(%q) = %v, want [%q]", tt.input, vals, tt.want)
		}
	}
}

func TestNormalizeOperator(t *testing.T) {
	tests := map[string]string{
		"equals":      "EQUALS",
		"!=":          "NOT_EQUALS",
		"EQUALS":      "EQUALS",
		"<":           "LESS_THAN",
		">":           "GREATER_THAN",
		"lessthan":    "LESS_THAN",
		"LESSTHAN":    "LESS_THAN",
		"LESS_THAN":   "LESS_THAN",
		"greaterthan": "GREATER_THAN",
		"containsall": "CONTAINS_ALL",
		"containsany": "CONTAINS_ANY",
		"startswith":  "STARTSWITH",
	}
	for input, want := range tests {
		got := normalizeOperator(input)
		if got != want {
			t.Errorf("normalizeOperator(%q) = %q, want %q", input, got, want)
		}
	}
}
