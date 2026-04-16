package reports

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func flattenReportResponse(raw json.RawMessage) (any, error) {
	var envelope map[string]any
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("parsing report response: %w", err)
	}

	if data, ok := envelope["data"]; ok && data == nil {
		normalized, err := json.Marshal(map[string]any{"data": []any{}})
		if err != nil {
			return nil, fmt.Errorf("encoding empty report response: %w", err)
		}
		return json.RawMessage(normalized), nil
	}

	reporting := extractReportingEnvelope(envelope)
	if reporting == nil {
		return raw, nil
	}

	rowsValue, _ := reporting["row"].([]any)
	rows := make([]map[string]any, 0, len(rowsValue))
	for _, item := range rowsValue {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rows = append(rows, flattenReportingRow(row)...)
	}

	result := map[string]any{
		"data": rows,
	}

	if grandTotals, ok := reporting["grandTotals"].(map[string]any); ok {
		result["grandTotals"] = flattenLeafMap(grandTotals)
	}

	normalized, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("encoding flattened report response: %w", err)
	}
	return json.RawMessage(normalized), nil
}

func extractReportingEnvelope(envelope map[string]any) map[string]any {
	if reporting, ok := envelope["reportingDataResponse"].(map[string]any); ok {
		return reporting
	}
	if data, ok := envelope["data"].(map[string]any); ok {
		if reporting, ok := data["reportingDataResponse"].(map[string]any); ok {
			return reporting
		}
	}
	return nil
}

func flattenReportingRow(row map[string]any) []map[string]any {
	base := make(map[string]any)

	if metadata, ok := row["metadata"].(map[string]any); ok {
		flattenObjectInto(base, metadata, "")
	}
	if insights, ok := row["insights"].(map[string]any); ok {
		flattenObjectInto(base, insights, "insights")
	}
	if other, ok := row["other"]; ok {
		base["other"] = other
	}

	if granularity, ok := row["granularity"].([]any); ok && len(granularity) > 0 {
		rows := make([]map[string]any, 0, len(granularity))
		for _, bucketValue := range granularity {
			bucket, ok := bucketValue.(map[string]any)
			if !ok {
				continue
			}
			flattened := cloneMap(base)
			flattenObjectInto(flattened, bucket, "")
			rows = append(rows, flattened)
		}
		return rows
	}

	if total, ok := row["total"].(map[string]any); ok {
		flattenObjectInto(base, total, "")
	}

	return []map[string]any{base}
}

func flattenLeafMap(obj map[string]any) map[string]any {
	out := make(map[string]any)
	flattenObjectInto(out, obj, "")
	return out
}

func flattenObjectInto(dst map[string]any, obj map[string]any, prefix string) {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		value := obj[key]
		flatKey := key
		if prefix != "" {
			flatKey = prefix + upperFirst(key)
		}

		nested, ok := value.(map[string]any)
		if ok && !isMoneyMap(nested) {
			flattenObjectInto(dst, nested, flatKey)
			continue
		}

		dst[flatKey] = value
	}
}

func isMoneyMap(m map[string]any) bool {
	_, hasAmount := m["amount"]
	_, hasCurrency := m["currency"]
	return hasAmount && hasCurrency
}

func cloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
