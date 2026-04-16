package output

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
)

// PrintPipe renders records as tab-separated rows with a header.
// It is intended for downstream aads commands consuming stdin.
func PrintPipe(data any) error {
	records, err := recordsFromData(data)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	cols := pipeColumns(records)
	if len(cols) == 0 {
		return fmt.Errorf("pipe output requires object or collection output")
	}

	headers := make([]string, 0, len(cols))
	for _, col := range cols {
		headers = append(headers, columnname.FromField(col))
	}
	fmt.Fprintln(os.Stdout, strings.Join(headers, "\t"))

	for _, rec := range records {
		row := make([]string, 0, len(cols))
		for _, col := range cols {
			row = append(row, formatPipeValue(rec[col]))
		}
		fmt.Fprintln(os.Stdout, strings.Join(row, "\t"))
	}
	return nil
}

func pipeColumns(records []map[string]any) []string {
	seen := make(map[string]struct{})
	for _, rec := range records {
		for key := range rec {
			seen[key] = struct{}{}
		}
	}

	var cols []string
	for _, key := range idHierarchy {
		if _, ok := seen[key]; ok {
			cols = append(cols, key)
			delete(seen, key)
		}
	}
	for _, key := range []string{"adamId", "appName", "appNameShort", "creativeId", "campaignName", "budgetAmount", "dailyBudgetAmount", "productPageId", "adGroupName", "defaultBidAmount", "cpaGoal"} {
		if _, ok := seen[key]; ok {
			cols = append(cols, key)
			delete(seen, key)
		}
	}

	rest := make([]string, 0, len(seen))
	for key := range seen {
		rest = append(rest, key)
	}
	slices.Sort(rest)
	return append(cols, rest...)
}

func formatPipeValue(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case json.Number:
		return val.String()
	case float64:
		if float64(int64(val)) == val {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%v", val)
	case string:
		return strings.ReplaceAll(val, "\t", " ")
	default:
		encoded, err := json.Marshal(val)
		if err != nil {
			return strings.ReplaceAll(fmt.Sprintf("%v", val), "\t", " ")
		}
		return strings.ReplaceAll(string(encoded), "\t", " ")
	}
}
