package shared

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
	"github.com/imesart/apple-ads-cli/internal/fieldmeta"
	"github.com/imesart/apple-ads-cli/internal/output"
)

// StdinFlag represents a flag whose value is "-" and should be read from stdin.
type StdinFlag struct {
	Name string  // flag name, e.g. "campaign-id"
	Ptr  *string // pointer to the flag's string value
}

// currentStdinContext holds extra key-value pairs injected from a piped stdin
// row (e.g. campaignId, adGroupId) so downstream commands can reference parent
// IDs without the user passing them explicitly. It is package-level state set
// before each command invocation in RunWithStdin and cleared afterward.
var currentStdinContext map[string]any

// FlagToColumnName converts a CLI flag name to an ID column name.
// For example, "campaign-id" becomes "CAMPAIGN_ID".
func FlagToColumnName(flag string) string {
	return columnname.FromFlag(flag)
}

// CollectStdinFlags returns the flags whose current value is "-",
// preserving the input order (which should match the ID hierarchy).
func CollectStdinFlags(flags ...StdinFlag) []StdinFlag {
	var result []StdinFlag
	for _, f := range flags {
		if f.Ptr != nil && *f.Ptr == "-" {
			result = append(result, f)
		}
	}
	return result
}

// HasStdinFlags returns true if any flag value is "-".
func HasStdinFlags(flags ...StdinFlag) bool {
	for _, f := range flags {
		if f.Ptr != nil && *f.Ptr == "-" {
			return true
		}
	}
	return false
}

// readStdinLines reads stdin rows and calls fn for each.
// Supported inputs:
//   - tab-separated rows, optionally with a header row
//   - JSON objects/arrays containing ID fields
//
// Returns collected results, processed data row count, and failure count.
func readStdinLines(stdinFlags []StdinFlag, fn func() (any, error)) ([]any, int, int) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading stdin: %v\n", err)
		return nil, 0, 1
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, 0, 0
	}

	if trimmed[0] == '{' || trimmed[0] == '[' {
		return readJSONStdin(trimmed, stdinFlags, fn)
	}

	return readTSVStdin(bytes.NewReader(data), stdinFlags, fn)
}

func readTSVStdin(r io.Reader, stdinFlags []StdinFlag, fn func() (any, error)) ([]any, int, int) {
	scanner := bufio.NewScanner(r)
	var results []any
	lineNum := 0
	dataRows := 0
	failures := 0

	var colMap []int // colMap[flagIdx] = column index; nil for positional
	var headerCols []string
	firstLine := true

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		cols := strings.Split(line, "\t")

		// On the first non-empty line, check if it's a header row.
		if firstLine {
			firstLine = false
			if mapping, ok := parseStdinHeader(cols, stdinFlags); ok {
				colMap = mapping
				headerCols = append([]string(nil), cols...)
				continue // skip header, read data lines
			}
			// Not a header; fall through to process as data.
		}

		// Assign column values to flag pointers.
		dataRows++
		if colMap != nil {
			for i, sf := range stdinFlags {
				if colMap[i] < len(cols) {
					*sf.Ptr = strings.TrimSpace(cols[colMap[i]])
				} else {
					*sf.Ptr = ""
				}
			}
		} else {
			if len(cols) < len(stdinFlags) {
				fmt.Fprintf(os.Stderr, "line %d: expected %d columns, got %d: %q\n", lineNum, len(stdinFlags), len(cols), line)
				failures++
				continue
			}
			for i, sf := range stdinFlags {
				*sf.Ptr = strings.TrimSpace(cols[i])
			}
		}

		rowRecord := buildTSVRecord(cols, headerCols, colMap, stdinFlags)
		ctx := deriveStdinContextFromRecord(rowRecord, stdinFlags)
		setCurrentStdinContext(ctx)
		result, err := fn()
		clearCurrentStdinContext()
		if err != nil {
			fmt.Fprintf(os.Stderr, "line %d: %v\n", lineNum, err)
			failures++
			continue
		}
		if result != nil {
			results = append(results, augmentWithContext(result, ctx))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading stdin: %v\n", err)
		failures++
	}

	return results, dataRows, failures
}

func buildTSVRecord(cols []string, headerCols []string, colMap []int, stdinFlags []StdinFlag) map[string]any {
	record := make(map[string]any, len(cols))
	if colMap != nil {
		for i, header := range headerCols {
			if i < len(cols) {
				record[columnname.ToCamelCase(header)] = strings.TrimSpace(cols[i])
			}
		}
		for i, sf := range stdinFlags {
			if colMap[i] >= 0 && colMap[i] < len(cols) {
				record[columnname.ToCamelCase(sf.Name)] = strings.TrimSpace(cols[colMap[i]])
			}
		}
		return record
	}
	for i, sf := range stdinFlags {
		if i < len(cols) {
			record[columnname.ToCamelCase(sf.Name)] = strings.TrimSpace(cols[i])
		}
	}
	return record
}

func readJSONStdin(raw []byte, stdinFlags []StdinFlag, fn func() (any, error)) ([]any, int, int) {
	records, err := parseJSONStdinRecords(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading stdin JSON: %v\n", err)
		return nil, 1, 1
	}

	var results []any
	dataRows := 0
	failures := 0

	for i, record := range records {
		dataRows++
		if err := assignJSONRecord(stdinFlags, record); err != nil {
			fmt.Fprintf(os.Stderr, "record %d: %v\n", i+1, err)
			failures++
			continue
		}

		ctx := deriveStdinContextFromRecord(record, stdinFlags)
		setCurrentStdinContext(ctx)
		result, err := fn()
		clearCurrentStdinContext()
		if err != nil {
			fmt.Fprintf(os.Stderr, "record %d: %v\n", i+1, err)
			failures++
			continue
		}
		if result != nil {
			results = append(results, augmentWithContext(result, ctx))
		}
	}

	return results, dataRows, failures
}

func parseJSONStdinRecords(raw []byte) ([]map[string]any, error) {
	var decoded any
	if err := output.UnmarshalUseNumber(raw, &decoded); err != nil {
		return nil, err
	}

	switch v := decoded.(type) {
	case map[string]any:
		if data, ok := v["data"]; ok {
			switch data.(type) {
			case map[string]any, []any:
				return jsonRecordsFromValue(data)
			}
		}
		return []map[string]any{v}, nil
	case []any:
		return jsonRecordsFromValue(v)
	default:
		return nil, fmt.Errorf("expected JSON object or array")
	}
}

func jsonRecordsFromValue(v any) ([]map[string]any, error) {
	switch data := v.(type) {
	case map[string]any:
		return []map[string]any{data}, nil
	case []any:
		records := make([]map[string]any, 0, len(data))
		for i, item := range data {
			record, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("array item %d is not an object", i+1)
			}
			records = append(records, record)
		}
		return records, nil
	default:
		return nil, fmt.Errorf("expected JSON object or array")
	}
}

func assignJSONRecord(stdinFlags []StdinFlag, record map[string]any) error {
	values, err := mapJSONRecordValues(stdinFlags, record)
	if err != nil {
		return err
	}
	for i, sf := range stdinFlags {
		*sf.Ptr = values[i]
	}
	return nil
}

func setCurrentStdinContext(ctx map[string]any) {
	if len(ctx) == 0 {
		currentStdinContext = nil
		return
	}
	currentStdinContext = cloneAnyMap(ctx)
}

func clearCurrentStdinContext() {
	currentStdinContext = nil
}

func currentSyntheticContext() map[string]any {
	if len(currentStdinContext) == 0 {
		return nil
	}
	return cloneAnyMap(currentStdinContext)
}

func mapJSONRecordValues(stdinFlags []StdinFlag, record map[string]any) ([]string, error) {
	values := make([]string, len(stdinFlags))
	matched := make([]bool, len(stdinFlags))
	available := make(map[string]string, len(record)*2)
	presentCols := make([]string, 0, len(record))
	idValue := ""
	hasID := false

	for key, raw := range record {
		col := columnname.FromField(strings.TrimSpace(key))
		value := formatStdinJSONValue(raw)
		presentCols = append(presentCols, col)
		if col == "ID" {
			idValue = value
			hasID = value != ""
			continue
		}
		available[col] = value
		available[columnname.Compact(col)] = value
	}

	for i, sf := range stdinFlags {
		expected := FlagToColumnName(sf.Name)
		if value, ok := available[expected]; ok && value != "" {
			values[i] = value
			matched[i] = true
			continue
		}
		compact := columnname.Compact(expected)
		if value, ok := available[compact]; ok && value != "" {
			values[i] = value
			matched[i] = true
		}
	}

	if hasID {
		var unmatched []int
		for i := range stdinFlags {
			if !matched[i] {
				unmatched = append(unmatched, i)
			}
		}
		if len(unmatched) == 1 {
			values[unmatched[0]] = idValue
			matched[unmatched[0]] = true
		} else if len(unmatched) > 1 {
			resolved := resolveIDColumnName(presentCols)
			for i, sf := range stdinFlags {
				if matched[i] {
					continue
				}
				expected := FlagToColumnName(sf.Name)
				if expected == resolved || columnname.Compact(expected) == resolved {
					values[i] = idValue
					matched[i] = true
					break
				}
			}
		}
	}

	for i, sf := range stdinFlags {
		if !matched[i] || values[i] == "" {
			return nil, fmt.Errorf("missing %q field in JSON record", FlagToColumnName(sf.Name))
		}
	}

	return values, nil
}

func formatStdinJSONValue(v any) string {
	switch value := v.(type) {
	case json.Number:
		return value.String()
	case float64:
		return fmt.Sprintf("%d", int64(value))
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}

func deriveStdinContextFromRecord(record map[string]any, stdinFlags []StdinFlag) map[string]any {
	ctx := captureIDContext(stdinFlags)
	for _, key := range []string{"campaignId", "adGroupId"} {
		if raw, ok := lookupField(record, key); ok {
			ctx[key] = raw
		}
	}
	for _, key := range []string{"adamId", "appName", "creativeId", "campaignName", "budgetAmount", "dailyBudgetAmount", "productPageId", "adGroupName", "defaultBidAmount", "cpaGoal"} {
		if raw, ok := lookupField(record, key); ok {
			ctx[key] = raw
		}
	}
	if raw, ok := ctx["appName"]; ok {
		if short := deriveAppNameShort(raw); short != "" {
			ctx["appNameShort"] = short
		}
	}

	switch inferRecordEntityKind(record) {
	case fieldmeta.EntityCampaign:
		if raw, ok := lookupField(record, "name"); ok {
			ctx["campaignName"] = raw
		}
		if raw, ok := lookupField(record, "budgetAmount"); ok {
			ctx["budgetAmount"] = raw
		}
		if raw, ok := lookupField(record, "dailyBudgetAmount"); ok {
			ctx["dailyBudgetAmount"] = raw
		}
	case fieldmeta.EntityAdGroup:
		if raw, ok := lookupField(record, "name"); ok {
			ctx["adGroupName"] = raw
		}
		if raw, ok := lookupField(record, "defaultBidAmount"); ok {
			ctx["defaultBidAmount"] = raw
		}
		if raw, ok := lookupField(record, "cpaGoal"); ok {
			ctx["cpaGoal"] = raw
		}
	case fieldmeta.EntityCreative:
		if raw, ok := lookupField(record, "id"); ok {
			ctx["creativeId"] = raw
		}
	case fieldmeta.EntityProductPage:
		if raw, ok := lookupField(record, "id"); ok {
			ctx["productPageId"] = raw
		}
	}

	return ctx
}

func deriveAppNameShort(raw any) string {
	name := strings.TrimSpace(formatStdinJSONValue(raw))
	if name == "" {
		return ""
	}
	cut := len(name)
	for _, sep := range []string{"-", ":", ",", "|", "•", "–"} {
		for start := 0; start < len(name); {
			rel := strings.Index(name[start:], sep)
			if rel < 0 {
				break
			}
			idx := start + rel
			beforeSpace := idx > 0 && isAppNameShortSpace(name[idx-1])
			afterPos := idx + len(sep)
			afterSpace := afterPos < len(name) && isAppNameShortSpace(name[afterPos])
			if !beforeSpace && !afterSpace {
				start = idx + len(sep)
				continue
			}
			if idx < cut {
				cut = idx
			}
			break
		}
	}
	return strings.TrimSpace(name[:cut])
}

func isAppNameShortSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func lookupField(record map[string]any, name string) (any, bool) {
	want := columnname.Compact(columnname.FromField(name))
	for key, value := range record {
		if columnname.Compact(columnname.FromField(key)) == want {
			return value, true
		}
	}
	return nil, false
}

func inferRecordEntityKind(record map[string]any) fieldmeta.EntityKind {
	hasCampaign := false
	hasAdGroup := false
	hasID := false
	hasAdamID := false
	hasType := false
	hasName := false
	idValue := ""
	for key := range record {
		switch compact := columnname.Compact(columnname.FromField(key)); compact {
		case columnname.Compact(columnname.FromField("campaignId")):
			hasCampaign = true
		case columnname.Compact(columnname.FromField("adGroupId")):
			hasAdGroup = true
		case "ID":
			hasID = true
			if raw, ok := lookupField(record, "id"); ok {
				idValue = formatStdinJSONValue(raw)
			}
		case columnname.Compact(columnname.FromField("adamId")):
			hasAdamID = true
		case columnname.Compact(columnname.FromField("type")):
			hasType = true
		case columnname.Compact(columnname.FromField("name")):
			hasName = true
		}
	}
	if hasAdGroup {
		return fieldmeta.EntityKeyword
	}
	if hasID && hasAdamID && hasType {
		return fieldmeta.EntityCreative
	}
	if hasID && hasName && idValue != "" {
		if _, err := fmt.Sscan(idValue, new(int64)); err != nil {
			return fieldmeta.EntityProductPage
		}
	}
	if hasCampaign && hasID {
		return fieldmeta.EntityAdGroup
	}
	if hasCampaign {
		return fieldmeta.EntityAdGroup
	}
	if hasID {
		return fieldmeta.EntityCampaign
	}
	return fieldmeta.EntityUnknown
}

// parseStdinHeader checks if cols form a header row and returns a mapping
// from stdinFlag index to column index. Returns nil, false if not a header.
func parseStdinHeader(cols []string, stdinFlags []StdinFlag) ([]int, bool) {
	// Header columns are alphabetic names; data columns start with digits.
	for _, col := range cols {
		col = strings.TrimSpace(col)
		if col == "" || (col[0] >= '0' && col[0] <= '9') {
			return nil, false
		}
	}

	// Build map from expected column name to flag index.
	flagByCol := make(map[string]int, len(stdinFlags))
	for i, sf := range stdinFlags {
		name := FlagToColumnName(sf.Name)
		flagByCol[name] = i
		flagByCol[columnname.Compact(name)] = i
	}

	// Match specific columns.
	mapping := make([]int, len(stdinFlags))
	for i := range mapping {
		mapping[i] = -1
	}
	matched := make(map[int]bool)
	idColIdx := -1

	for ci, col := range cols {
		col = columnname.FromField(strings.TrimSpace(col))
		if col == "ID" {
			idColIdx = ci
			continue
		}
		if fi, ok := flagByCol[col]; ok && !matched[fi] {
			mapping[fi] = ci
			matched[fi] = true
			continue
		}
		if fi, ok := flagByCol[columnname.Compact(col)]; ok && !matched[fi] {
			mapping[fi] = ci
			matched[fi] = true
		}
	}

	// Resolve the "ID" column if present.
	if idColIdx >= 0 {
		var unmatched []int
		for fi := range stdinFlags {
			if !matched[fi] {
				unmatched = append(unmatched, fi)
			}
		}
		if len(unmatched) == 1 {
			// Single unmatched flag: ID maps to it.
			mapping[unmatched[0]] = idColIdx
			matched[unmatched[0]] = true
		} else if len(unmatched) > 1 {
			// Multiple unmatched: resolve by hierarchy.
			resolved := resolveIDColumnName(cols)
			if fi, ok := flagByCol[resolved]; ok && !matched[fi] {
				mapping[fi] = idColIdx
				matched[fi] = true
			}
		}
	}

	// All flags must have a matched column.
	for fi := range stdinFlags {
		if mapping[fi] < 0 {
			return nil, false
		}
	}

	return mapping, true
}

// resolveIDColumnName determines what a generic "ID" column represents
// based on what other ID columns are present in the header.
func resolveIDColumnName(cols []string) string {
	hasCampaign := false
	hasAdGroup := false
	for _, col := range cols {
		col = columnname.FromField(strings.TrimSpace(col))
		switch col {
		case "CAMPAIGN_ID":
			hasCampaign = true
		case "AD_GROUP_ID":
			hasAdGroup = true
		}
	}
	if hasAdGroup {
		return columnname.FromFlag("keyword-id")
	}
	if hasCampaign {
		return columnname.FromFlag("adgroup-id")
	}
	return columnname.FromFlag("campaign-id")
}

// RunWithStdin reads tab-separated lines from stdin and executes fn once per line.
// Results are collected and returned as merged output, printed once at the end.
// On error per line, prints to stderr and continues.
// Optional entityIDName sets the column name for the "id" field in ids output.
func RunWithStdin(stdinFlags []StdinFlag, fn func() (any, error), format string, fields string, pretty bool, entityIDName ...string) error {
	return RunWithStdinTransform(stdinFlags, fn, nil, format, fields, pretty, entityIDName...)
}

func RunWithStdinTransform(stdinFlags []StdinFlag, fn func() (any, error), transform func(any) (any, error), format string, fields string, pretty bool, entityIDName ...string) error {
	results, dataRows, failures := CollectResultsWithStdin(stdinFlags, fn)

	if len(results) > 0 {
		merged := MergeResults(results)
		if transform != nil {
			var err error
			merged, err = transform(merged)
			if err != nil {
				return err
			}
		}
		if err := PrintOutput(merged, format, fields, pretty, entityIDName...); err != nil {
			return err
		}
	}

	if failures > 0 {
		return ReportError(fmt.Errorf("%d of %d lines failed", failures, dataRows))
	}
	return nil
}

// CollectResultsWithStdin reads tab-separated stdin rows and executes fn once
// per row, returning the collected results plus row and failure counts.
func CollectResultsWithStdin(stdinFlags []StdinFlag, fn func() (any, error)) ([]any, int, int) {
	return readStdinLines(stdinFlags, fn)
}

// MergeResults merges multiple API responses into a single response.
// For json.RawMessage values with "data" arrays, concatenates all data arrays.
// For other types, collects into a slice.
func MergeResults(results []any) any {
	if len(results) == 0 {
		return nil
	}
	if len(results) == 1 {
		return results[0]
	}

	// Try to merge as JSON responses with "data" envelopes.
	var allRecords []json.RawMessage
	allEnveloped := true

	for _, r := range results {
		raw, ok := r.(json.RawMessage)
		if !ok {
			allEnveloped = false
			break
		}

		records, ok := extractDataRecords(raw)
		if !ok {
			allEnveloped = false
			break
		}
		allRecords = append(allRecords, records...)
	}

	if allEnveloped {
		dataArr, err := json.Marshal(allRecords)
		if err != nil {
			return results
		}
		merged, err := json.Marshal(map[string]json.RawMessage{"data": dataArr})
		if err != nil {
			return results
		}
		return json.RawMessage(merged)
	}

	return results
}

// captureIDContext snapshots the current stdin flag values as a map
// of camelCase JSON key → value, for injecting into result rows.
func captureIDContext(stdinFlags []StdinFlag) map[string]any {
	ctx := make(map[string]any, len(stdinFlags))
	for _, sf := range stdinFlags {
		if sf.Ptr != nil && *sf.Ptr != "" {
			key := columnname.ToCamelCase(sf.Name)
			ctx[key] = *sf.Ptr
		}
	}
	return ctx
}

// augmentWithContext injects parent ID context into each row of an API result.
// For {"data":[...]} envelopes, each record gets the context fields added.
// For single objects, context fields are added directly.
func augmentWithContext(result any, ctx map[string]any) any {
	if len(ctx) == 0 {
		return result
	}

	raw, ok := result.(json.RawMessage)
	if !ok {
		return result
	}

	// Try envelope: {"data": [...]}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Data) > 0 {
		augmented := injectContext(envelope.Data, ctx)
		if augmented != nil {
			out, err := json.Marshal(map[string]json.RawMessage{"data": augmented})
			if err == nil {
				return json.RawMessage(out)
			}
		}
		return result
	}

	// Try bare array or single object
	augmented := injectContext(raw, ctx)
	if augmented != nil {
		return json.RawMessage(augmented)
	}
	return result
}

// injectContext adds context fields to each object in a JSON array,
// or to a single JSON object. Returns nil if injection fails.
func injectContext(data json.RawMessage, ctx map[string]any) json.RawMessage {
	// Try as array of objects
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		augmented := make([]json.RawMessage, 0, len(arr))
		for _, item := range arr {
			augmented = append(augmented, injectIntoObject(item, ctx))
		}
		out, err := json.Marshal(augmented)
		if err != nil {
			return nil
		}
		return out
	}

	// Try as single object
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err == nil {
		return injectIntoObject(data, ctx)
	}

	return nil
}

// injectIntoObject adds context key-value pairs into a JSON object.
// Existing keys are not overwritten.
func injectIntoObject(data json.RawMessage, ctx map[string]any) json.RawMessage {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return data
	}
	for key, val := range ctx {
		if _, exists := obj[key]; !exists {
			quoted, err := json.Marshal(val)
			if err == nil {
				obj[key] = quoted
			}
		}
	}
	out, err := json.Marshal(obj)
	if err != nil {
		return data
	}
	return out
}

func cloneAnyMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// extractDataRecords extracts individual records from a JSON response.
// Returns the records as raw JSON and true if successful.
func extractDataRecords(raw json.RawMessage) ([]json.RawMessage, bool) {
	// Try envelope: {"data": [...]}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Data) > 0 {
		// Try as array
		var arr []json.RawMessage
		if err := json.Unmarshal(envelope.Data, &arr); err == nil {
			return arr, true
		}
		// Single object in data
		return []json.RawMessage{envelope.Data}, true
	}

	// Try bare array
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, true
	}

	// Single object
	return []json.RawMessage{raw}, true
}
