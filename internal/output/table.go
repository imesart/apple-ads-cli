package output

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
	"github.com/olekukonko/tablewriter"
)

// PrintTable renders data as an ASCII table to stdout.
// Supports structs, slices of structs, maps, and json.RawMessage.
// For slices, each element becomes a row. For a single struct, it becomes one row.
// Nested structs like Money are flattened to "amount currency".
func PrintTable(data any) error {
	if data == nil {
		return nil
	}
	if ordered, ok := data.(OrderedData); ok {
		return printOrderedTable(ordered)
	}

	// Handle json.RawMessage / []byte containing JSON.
	// The API returns json.RawMessage which is []byte; without this,
	// reflect sees []uint8 and prints each byte as a number.
	if raw, ok := data.(json.RawMessage); ok {
		return printJSONTable(raw)
	}

	v := reflect.ValueOf(data)

	// Unwrap pointer
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return printSliceTable(v)
	case reflect.Struct:
		return printSliceTable(reflect.ValueOf([]any{data}))
	case reflect.Map:
		if v.Len() == 0 {
			fmt.Fprintln(os.Stdout, "No results.")
			return nil
		}
		if row, ok := valueToStringMap(v); ok {
			return printMapsTable([]map[string]any{row})
		}
		return printMapTable(v)
	default:
		// Scalar value: just print it
		_, err := fmt.Fprintln(os.Stdout, v.Interface())
		return err
	}
}

func printOrderedTable(data OrderedData) error {
	if len(data.Fields) == 0 {
		fmt.Fprintln(os.Stdout, "No results.")
		return nil
	}

	headers := make([]string, 0, len(data.Fields))
	for _, field := range data.Fields {
		headers = append(headers, field.Header)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	configureTable(table)

	for _, row := range data.Rows {
		values := make([]string, 0, len(row))
		for _, value := range row {
			values = append(values, formatAny(value))
		}
		table.Append(values)
	}

	table.Render()
	return nil
}

// printJSONTable unmarshals JSON and renders it as a table.
// API responses are typically {"data": [...], "pagination": {...}}.
// If the JSON has a "data" array, each element becomes a row.
func printJSONTable(raw json.RawMessage) error {
	// Try as an object with a "data" key (standard API envelope)
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Data) > 0 {
		// Check if data is an array
		var arr []map[string]any
		if err := UnmarshalUseNumber(envelope.Data, &arr); err == nil {
			return printMapsTable(arr)
		}
		// Data is a single object
		var obj map[string]any
		if err := UnmarshalUseNumber(envelope.Data, &obj); err == nil {
			return printMapsTable([]map[string]any{obj})
		}
	}

	// Try as a bare array
	var arr []map[string]any
	if err := UnmarshalUseNumber(raw, &arr); err == nil {
		return printMapsTable(arr)
	}

	// Try as a single object
	var obj map[string]any
	if err := UnmarshalUseNumber(raw, &obj); err == nil {
		return PrintTable(obj)
	}

	// Fallback: print as indented JSON
	return PrintJSON(raw, false)
}

// printMapsTable renders a slice of maps as a table, using consistent column
// ordering derived from the first element's keys.
func printMapsTable(rows []map[string]any) error {
	if len(rows) == 0 {
		fmt.Fprintln(os.Stdout, "No results.")
		return nil
	}

	// Collect column order from the keys of the first row
	keys := mapKeys(rows[0])

	headers := make([]string, len(keys))
	for i, k := range keys {
		headers[i] = columnname.FromField(k)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	configureTable(table)

	for _, row := range rows {
		vals := make([]string, len(keys))
		for i, k := range keys {
			vals[i] = formatAny(row[k])
		}
		table.Append(vals)
	}

	table.Render()
	return nil
}

// mapKeys returns the keys of a map in a stable order.
// Puts common ID/name fields first, then remaining keys alphabetically.
func mapKeys(m map[string]any) []string {
	// Priority fields shown first
	priority := []string{"id", "name", "status", "text", "matchType"}
	var keys []string
	seen := make(map[string]bool)

	for _, p := range priority {
		if _, ok := m[p]; ok {
			keys = append(keys, p)
			seen[p] = true
		}
	}

	// Remaining keys in sorted order
	var rest []string
	for k := range m {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	slices.Sort(rest)
	keys = append(keys, rest...)
	return keys
}

// formatAny converts an arbitrary JSON-decoded value to a display string.
func formatAny(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case map[string]any:
		// Money-like: {amount, currency}
		if amt, ok := val["amount"]; ok {
			if cur, ok := val["currency"]; ok {
				return fmt.Sprintf("%v %v", amt, cur)
			}
		}
		var parts []string
		for k, v := range val {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		return strings.Join(parts, ", ")
	case []any:
		var parts []string
		for _, item := range val {
			parts = append(parts, formatAny(item))
		}
		return strings.Join(parts, ", ")
	case json.Number:
		return val.String()
	case float64:
		// Fallback for non-UseNumber paths; display integers without decimal
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// printSliceTable renders a slice of structs as a table.
func printSliceTable(v reflect.Value) error {
	if v.Len() == 0 {
		fmt.Fprintln(os.Stdout, "No results.")
		return nil
	}

	if rows, ok := sliceToMaps(v); ok {
		return printMapsTable(rows)
	}

	// Get the element type, dereferencing pointers
	elemType := v.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	// For interface{} slices with struct elements, inspect the first element
	if elemType.Kind() == reflect.Interface {
		first := v.Index(0)
		for first.Kind() == reflect.Interface || first.Kind() == reflect.Ptr {
			first = first.Elem()
		}
		if first.Kind() != reflect.Struct {
			return printScalarSlice(v)
		}
		elemType = first.Type()
	}

	if elemType.Kind() != reflect.Struct {
		return printScalarSlice(v)
	}

	headers := structHeaders(elemType)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	configureTable(table)

	for i := range v.Len() {
		elem := v.Index(i)
		for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		row := structRow(elem, elemType)
		table.Append(row)
	}

	table.Render()
	return nil
}

func sliceToMaps(v reflect.Value) ([]map[string]any, bool) {
	rows := make([]map[string]any, 0, v.Len())
	for i := range v.Len() {
		elem := v.Index(i)
		for elem.Kind() == reflect.Interface || elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				return nil, false
			}
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Map {
			return nil, false
		}

		row := make(map[string]any, elem.Len())
		iter := elem.MapRange()
		for iter.Next() {
			key, ok := iter.Key().Interface().(string)
			if !ok {
				return nil, false
			}
			row[key] = iter.Value().Interface()
		}
		rows = append(rows, row)
	}
	return rows, true
}

func valueToStringMap(v reflect.Value) (map[string]any, bool) {
	row := make(map[string]any, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		key, ok := iter.Key().Interface().(string)
		if !ok {
			return nil, false
		}
		row[key] = iter.Value().Interface()
	}
	return row, true
}

// printMapTable renders a map as a two-column key/value table.
func printMapTable(v reflect.Value) error {
	if v.Len() == 0 {
		fmt.Fprintln(os.Stdout, "No results.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"KEY", "VALUE"})
	configureTable(table)

	iter := v.MapRange()
	for iter.Next() {
		key := fmt.Sprintf("%v", iter.Key().Interface())
		val := formatValue(iter.Value())
		table.Append([]string{key, val})
	}

	table.Render()
	return nil
}

// printScalarSlice prints a slice of non-struct values as a single-column table.
func printScalarSlice(v reflect.Value) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"VALUE"})
	configureTable(table)

	for i := range v.Len() {
		table.Append([]string{formatValue(v.Index(i))})
	}

	table.Render()
	return nil
}

// structHeaders returns column headers from struct field names or json tags.
func structHeaders(t reflect.Type) []string {
	var headers []string
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name := fieldDisplayName(f)
		if name == "-" {
			continue
		}

		headers = append(headers, columnname.FromField(name))
	}
	return headers
}

// structRow extracts display values from a struct value.
func structRow(v reflect.Value, t reflect.Type) []string {
	var row []string
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name := fieldDisplayName(f)
		if name == "-" {
			continue
		}

		fv := v.Field(i)
		row = append(row, formatValue(fv))
	}
	return row
}

// fieldDisplayName returns the display name for a struct field.
// Uses the json tag name if available, otherwise the field name.
func fieldDisplayName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" || tag == "-" {
		if tag == "-" {
			return "-"
		}
		return f.Name
	}
	parts := strings.SplitN(tag, ",", 2)
	if parts[0] == "" || parts[0] == "-" {
		if parts[0] == "-" {
			return "-"
		}
		return f.Name
	}
	return parts[0]
}

// formatValue converts a reflect.Value to a display string.
// Handles pointers, nested structs (like Money), and basic types.
func formatValue(v reflect.Value) string {
	// Unwrap interfaces
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	// Unwrap pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return formatStruct(v)
	case reflect.Slice, reflect.Array:
		return formatSlice(v)
	case reflect.Map:
		return formatMap(v)
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// formatStruct flattens a struct to a string.
// For Money-like structs (Amount + Currency), produces "amount currency".
// For other structs, produces "field1=val1, field2=val2".
func formatStruct(v reflect.Value) string {
	t := v.Type()

	// Special case: Money-like struct with Amount and Currency fields
	amountField := v.FieldByName("Amount")
	currencyField := v.FieldByName("Currency")
	if amountField.IsValid() && currencyField.IsValid() {
		return fmt.Sprintf("%s %s", amountField.Interface(), currencyField.Interface())
	}

	var parts []string
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		fv := v.Field(i)
		val := formatValue(fv)
		if val != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", fieldDisplayName(f), val))
		}
	}
	return strings.Join(parts, ", ")
}

// formatSlice formats a slice as a comma-separated string.
func formatSlice(v reflect.Value) string {
	if v.Len() == 0 {
		return ""
	}
	var parts []string
	for i := range v.Len() {
		parts = append(parts, formatValue(v.Index(i)))
	}
	return strings.Join(parts, ", ")
}

// formatMap formats a map as a comma-separated key=value string.
func formatMap(v reflect.Value) string {
	if v.Len() == 0 {
		return ""
	}
	var parts []string
	iter := v.MapRange()
	for iter.Next() {
		parts = append(parts, fmt.Sprintf("%v=%s", iter.Key().Interface(), formatValue(iter.Value())))
	}
	return strings.Join(parts, ", ")
}

// configureTable sets standard table formatting options.
func configureTable(table *tablewriter.Table) {
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
}
