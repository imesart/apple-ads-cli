package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
)

// idHierarchy defines the column order for ids output.
// Parent IDs come first, resource ID last.
var idHierarchy = []string{"campaignId", "adGroupId", "id"}

// PrintIDs extracts ID fields from API responses and prints them
// tab-separated with a header row. Column order follows the
// resource hierarchy: campaignId, adGroupId, id.
// entityIDName specifies the column name for the "id" field
// (e.g., "KEYWORD_ID"). If empty, it is inferred from the hierarchy depth.
func PrintIDs(data any, entityIDName string) error {
	records, err := recordsFromData(data)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	// Determine which ID columns exist from the first record.
	cols := idColumns(records[0])
	if len(cols) == 0 {
		return fmt.Errorf("no id fields found in response")
	}

	// Build header names.
	headers := make([]string, len(cols))
	for i, col := range cols {
		if col == "id" {
			if entityIDName != "" {
				headers[i] = columnname.FromField(entityIDName)
			} else {
				headers[i] = inferIDColumnName(cols)
			}
		} else {
			headers[i] = columnname.FromField(col)
		}
	}

	// Print header row.
	fmt.Fprintln(os.Stdout, strings.Join(headers, "\t"))

	// Print data rows.
	for _, rec := range records {
		vals := make([]string, len(cols))
		for i, col := range cols {
			vals[i] = formatIDValue(rec[col])
		}
		fmt.Fprintln(os.Stdout, strings.Join(vals, "\t"))
	}
	return nil
}

func recordsFromData(data any) ([]map[string]any, error) {
	if raw, ok := data.(json.RawMessage); ok {
		return unwrapRecords(raw)
	}

	normalized := unwrapOutputData(data)
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("cannot encode records: %w", err)
	}
	return unwrapRecords(raw)
}

// inferIDColumnName determines the column name for the "id" field
// based on what parent ID fields are present in the record.
func inferIDColumnName(cols []string) string {
	hasCampaignId := false
	hasAdGroupId := false
	for _, col := range cols {
		switch col {
		case "campaignId":
			hasCampaignId = true
		case "adGroupId":
			hasAdGroupId = true
		}
	}
	if hasAdGroupId {
		return columnname.FromField("keywordId")
	}
	if hasCampaignId {
		return columnname.FromField("adGroupId")
	}
	return columnname.FromField("campaignId")
}

// unwrapRecords extracts []map[string]any from a JSON response.
// Handles: {"data": [...]}, {"data": {...}}, bare [...], bare {...}.
func unwrapRecords(raw json.RawMessage) ([]map[string]any, error) {
	// Try envelope with "data" key
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Data) > 0 {
		// Try as array
		var arr []map[string]any
		if err := UnmarshalUseNumber(envelope.Data, &arr); err == nil {
			return arr, nil
		}
		// Try as single object
		var obj map[string]any
		if err := UnmarshalUseNumber(envelope.Data, &obj); err == nil {
			return []map[string]any{obj}, nil
		}
	}

	// Try as bare array
	var arr []map[string]any
	if err := UnmarshalUseNumber(raw, &arr); err == nil {
		return arr, nil
	}

	// Try as bare single object
	var obj map[string]any
	if err := UnmarshalUseNumber(raw, &obj); err == nil {
		return []map[string]any{obj}, nil
	}

	return nil, fmt.Errorf("cannot extract records from response")
}

// idColumns returns the ID field names present in a record, in hierarchy order.
func idColumns(rec map[string]any) []string {
	var cols []string
	for _, key := range idHierarchy {
		if _, ok := rec[key]; ok {
			cols = append(cols, key)
		}
	}
	return cols
}

// formatIDValue formats an ID value as a string.
// JSON numbers are float64; display as integers.
func formatIDValue(v any) string {
	switch val := v.(type) {
	case float64:
		return fmt.Sprintf("%d", int64(val))
	case json.Number:
		return val.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}
