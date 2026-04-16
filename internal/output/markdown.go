package output

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
)

// PrintMarkdown renders data as a GitHub-flavored markdown table.
func PrintMarkdown(data any) error {
	if data == nil {
		return nil
	}
	if ordered, ok := data.(OrderedData); ok {
		return printMarkdownRows(ordered.Fields, ordered.Rows)
	}
	if raw, ok := data.(json.RawMessage); ok {
		return printJSONMarkdown(raw)
	}

	v := reflect.ValueOf(data)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return printMarkdownValueSlice(v)
	case reflect.Struct:
		return printMarkdownValueSlice(reflect.ValueOf([]any{data}))
	case reflect.Map:
		if v.Len() == 0 {
			fmt.Fprintln(os.Stdout, "No results.")
			return nil
		}
		if row, ok := valueToStringMap(v); ok {
			return printMarkdownMaps([]map[string]any{row})
		}
		return printMarkdownMap(v)
	default:
		_, err := fmt.Fprintln(os.Stdout, v.Interface())
		return err
	}
}

func printJSONMarkdown(raw json.RawMessage) error {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Data) > 0 {
		var arr []map[string]any
		if err := UnmarshalUseNumber(envelope.Data, &arr); err == nil {
			return printMarkdownMaps(arr)
		}
		var obj map[string]any
		if err := UnmarshalUseNumber(envelope.Data, &obj); err == nil {
			return printMarkdownMaps([]map[string]any{obj})
		}
	}

	var arr []map[string]any
	if err := UnmarshalUseNumber(raw, &arr); err == nil {
		return printMarkdownMaps(arr)
	}

	var obj map[string]any
	if err := UnmarshalUseNumber(raw, &obj); err == nil {
		return printMarkdownMaps([]map[string]any{obj})
	}

	return PrintJSON(raw, false)
}

func printMarkdownMaps(rows []map[string]any) error {
	if len(rows) == 0 {
		fmt.Fprintln(os.Stdout, "No results.")
		return nil
	}
	keys := mapKeys(rows[0])
	fields := make([]OrderedField, 0, len(keys))
	values := make([][]any, 0, len(rows))
	for _, key := range keys {
		fields = append(fields, OrderedField{Key: key, Header: columnname.FromField(key)})
	}
	for _, row := range rows {
		current := make([]any, 0, len(keys))
		for _, key := range keys {
			current = append(current, row[key])
		}
		values = append(values, current)
	}
	return printMarkdownRows(fields, values)
}

func printMarkdownMap(v reflect.Value) error {
	if v.Len() == 0 {
		fmt.Fprintln(os.Stdout, "No results.")
		return nil
	}
	fields := []OrderedField{
		{Key: "key", Header: "KEY"},
		{Key: "value", Header: "VALUE"},
	}
	rows := make([][]any, 0, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		rows = append(rows, []any{iter.Key().Interface(), formatValue(iter.Value())})
	}
	return printMarkdownRows(fields, rows)
}

func printMarkdownValueSlice(v reflect.Value) error {
	if v.Len() == 0 {
		fmt.Fprintln(os.Stdout, "No results.")
		return nil
	}

	if rows, ok := sliceToMaps(v); ok {
		return printMarkdownMaps(rows)
	}

	elemType := v.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() == reflect.Interface {
		first := v.Index(0)
		for first.Kind() == reflect.Interface || first.Kind() == reflect.Ptr {
			first = first.Elem()
		}
		if first.Kind() != reflect.Struct {
			return printMarkdownScalars(v)
		}
		elemType = first.Type()
	}
	if elemType.Kind() != reflect.Struct {
		return printMarkdownScalars(v)
	}

	headers := structHeaders(elemType)
	fields := make([]OrderedField, 0, len(headers))
	for _, header := range headers {
		fields = append(fields, OrderedField{Header: header, Key: strings.ToLower(header)})
	}
	rows := make([][]any, 0, v.Len())
	for i := range v.Len() {
		elem := v.Index(i)
		for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		row := make([]any, 0, elemType.NumField())
		for _, value := range structRow(elem, elemType) {
			row = append(row, value)
		}
		rows = append(rows, row)
	}
	return printMarkdownRows(fields, rows)
}

func printMarkdownScalars(v reflect.Value) error {
	fields := []OrderedField{{Key: "value", Header: "VALUE"}}
	rows := make([][]any, 0, v.Len())
	for i := range v.Len() {
		rows = append(rows, []any{formatValue(v.Index(i))})
	}
	return printMarkdownRows(fields, rows)
}

func printMarkdownRows(fields []OrderedField, rows [][]any) error {
	if len(fields) == 0 {
		fmt.Fprintln(os.Stdout, "No results.")
		return nil
	}

	headers := make([]string, 0, len(fields))
	separators := make([]string, 0, len(fields))
	for _, field := range fields {
		headers = append(headers, escapeMarkdown(field.Header))
		separators = append(separators, "---")
	}
	fmt.Fprintf(os.Stdout, "| %s |\n", strings.Join(headers, " | "))
	fmt.Fprintf(os.Stdout, "| %s |\n", strings.Join(separators, " | "))
	for _, row := range rows {
		values := make([]string, 0, len(row))
		for _, value := range row {
			values = append(values, escapeMarkdown(formatAny(value)))
		}
		fmt.Fprintf(os.Stdout, "| %s |\n", strings.Join(values, " | "))
	}
	return nil
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "|", "\\|")
}
