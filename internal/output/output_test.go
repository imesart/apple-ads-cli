package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// captureStdout redirects os.Stdout to a pipe and returns the captured output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return buf.String()
}

func TestPrintJSON(t *testing.T) {
	data := map[string]string{"name": "test", "value": "hello"}

	output := captureStdout(t, func() {
		if err := PrintJSON(data, false); err != nil {
			t.Fatal(err)
		}
	})

	// Verify valid JSON
	var decoded map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %q, error: %v", output, err)
	}
	if decoded["name"] != "test" {
		t.Errorf("decoded[name] = %q, want %q", decoded["name"], "test")
	}
	if decoded["value"] != "hello" {
		t.Errorf("decoded[value] = %q, want %q", decoded["value"], "hello")
	}
}

func TestPrintJSON_Slice(t *testing.T) {
	data := []int{1, 2, 3}

	output := captureStdout(t, func() {
		if err := PrintJSON(data, false); err != nil {
			t.Fatal(err)
		}
	})

	var decoded []int
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %q, error: %v", output, err)
	}
	if len(decoded) != 3 {
		t.Errorf("decoded length = %d, want 3", len(decoded))
	}
}

func TestPrintJSON_Nil(t *testing.T) {
	output := captureStdout(t, func() {
		if err := PrintJSON(nil, false); err != nil {
			t.Fatal(err)
		}
	})

	trimmed := strings.TrimSpace(output)
	if trimmed != "null" {
		t.Errorf("PrintJSON(nil) = %q, want %q", trimmed, "null")
	}
}

func TestPrintJSON_PrettyOverride(t *testing.T) {
	data := map[string]any{
		"bidAmount": map[string]string{
			"amount":   "1.06",
			"currency": "EUR",
		},
	}

	output := captureStdout(t, func() {
		if err := PrintJSON(data, true); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "\n") {
		t.Fatalf("expected pretty JSON with newlines, got %q", output)
	}
	if !strings.Contains(output, `"bidAmount": {`) {
		t.Fatalf("expected pretty JSON object output, got %q", output)
	}
}

func TestPrintYAML(t *testing.T) {
	data := map[string]string{"name": "test", "value": "hello"}

	output := captureStdout(t, func() {
		if err := PrintYAML(data); err != nil {
			t.Fatal(err)
		}
	})

	// Verify valid YAML
	var decoded map[string]string
	if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid YAML: %q, error: %v", output, err)
	}
	if decoded["name"] != "test" {
		t.Errorf("decoded[name] = %q, want %q", decoded["name"], "test")
	}
	if decoded["value"] != "hello" {
		t.Errorf("decoded[value] = %q, want %q", decoded["value"], "hello")
	}
}

func TestPrintYAML_Slice(t *testing.T) {
	data := []string{"a", "b", "c"}

	output := captureStdout(t, func() {
		if err := PrintYAML(data); err != nil {
			t.Fatal(err)
		}
	})

	var decoded []string
	if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid YAML: %q, error: %v", output, err)
	}
	if len(decoded) != 3 {
		t.Errorf("decoded length = %d, want 3", len(decoded))
	}
}

func TestDefaultFormat(t *testing.T) {
	// When running in tests, stdout is a pipe, not a TTY
	f := DefaultFormat()
	if f != FormatJSON {
		t.Errorf("DefaultFormat() = %q, want %q (stdout is pipe in tests)", f, FormatJSON)
	}
}

func TestPrint_JSON(t *testing.T) {
	data := map[string]int{"count": 42}

	output := captureStdout(t, func() {
		if err := Print(FormatJSON, data); err != nil {
			t.Fatal(err)
		}
	})

	var decoded map[string]int
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &decoded); err != nil {
		t.Fatalf("Print(JSON) produced invalid JSON: %q", output)
	}
	if decoded["count"] != 42 {
		t.Errorf("decoded[count] = %d, want 42", decoded["count"])
	}
}

func TestPrint_JSON_UnwrapsDataEnvelope(t *testing.T) {
	data := json.RawMessage(`{"data":[{"id":1},{"id":2}],"pagination":{"totalResults":2}}`)

	output := captureStdout(t, func() {
		if err := Print(FormatJSON, data); err != nil {
			t.Fatal(err)
		}
	})

	var decoded []map[string]int
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &decoded); err != nil {
		t.Fatalf("Print(JSON) produced invalid JSON: %q", output)
	}
	if len(decoded) != 2 || decoded[0]["id"] != 1 || decoded[1]["id"] != 2 {
		t.Fatalf("decoded = %#v, want ids [1 2]", decoded)
	}
}

func TestPrint_JSON_NullDataEnvelopeBecomesEmptyArray(t *testing.T) {
	data := json.RawMessage(`{"data":null}`)

	output := captureStdout(t, func() {
		if err := Print(FormatJSON, data); err != nil {
			t.Fatal(err)
		}
	})

	if strings.TrimSpace(output) != "[]" {
		t.Fatalf("output = %q, want []", strings.TrimSpace(output))
	}
}

func TestPrint_YAML(t *testing.T) {
	data := map[string]int{"count": 42}

	output := captureStdout(t, func() {
		if err := Print(FormatYAML, data); err != nil {
			t.Fatal(err)
		}
	})

	var decoded map[string]int
	if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("Print(YAML) produced invalid YAML: %q", output)
	}
	if decoded["count"] != 42 {
		t.Errorf("decoded[count] = %d, want 42", decoded["count"])
	}
}

func TestPrint_YAML_UnwrapsDataEnvelope(t *testing.T) {
	data := json.RawMessage(`{"data":{"id":42,"name":"Alpha"},"pagination":{"totalResults":1}}`)

	output := captureStdout(t, func() {
		if err := Print(FormatYAML, data); err != nil {
			t.Fatal(err)
		}
	})

	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("Print(YAML) produced invalid YAML: %q", output)
	}
	if decoded["id"] != "42" || decoded["name"] != "Alpha" {
		t.Fatalf("decoded = %#v, want map with id=42 name=Alpha", decoded)
	}
}

func TestPrint_Markdown(t *testing.T) {
	data := map[string]int{"count": 42}

	output := captureStdout(t, func() {
		if err := Print(FormatMarkdown, data); err != nil {
			t.Fatal(err)
		}
	})

	if strings.Contains(output, "| KEY | VALUE |") {
		t.Fatalf("markdown output should not use key/value layout: %q", output)
	}
	for _, want := range []string{"| COUNT |", "| 42 |"} {
		if !strings.Contains(output, want) {
			t.Fatalf("markdown output missing %q: %q", want, output)
		}
	}
}

func TestPrint_TableSingleObjectUsesColumnHeaders(t *testing.T) {
	data := json.RawMessage(`{"data":{"id":63835007,"name":"Weekly Share Report","granularity":"DAILY","startTime":"2026-03-20","endTime":"2026-03-27"}}`)

	output := captureStdout(t, func() {
		if err := Print(FormatTable, data); err != nil {
			t.Fatal(err)
		}
	})

	if strings.Contains(output, "KEY") || strings.Contains(output, "VALUE") {
		t.Fatalf("table output should not use key/value layout: %q", output)
	}
	for _, want := range []string{"ID", "NAME", "GRANULARITY", "START_TIME", "END_TIME", "63835007", "Weekly Share Report"} {
		if !strings.Contains(output, want) {
			t.Fatalf("table output missing %q: %q", want, output)
		}
	}
}

func TestPrint_MarkdownSingleObjectUsesColumnHeaders(t *testing.T) {
	data := json.RawMessage(`{"data":{"id":63835007,"name":"Weekly Share Report","granularity":"DAILY","startTime":"2026-03-20","endTime":"2026-03-27"}}`)

	output := captureStdout(t, func() {
		if err := Print(FormatMarkdown, data); err != nil {
			t.Fatal(err)
		}
	})

	if strings.Contains(output, "| KEY | VALUE |") {
		t.Fatalf("markdown output should not use key/value layout: %q", output)
	}
	for _, want := range []string{"| ID | NAME | END_TIME | GRANULARITY | START_TIME |", "63835007", "Weekly Share Report"} {
		if !strings.Contains(output, want) {
			t.Fatalf("markdown output missing %q: %q", want, output)
		}
	}
}

func TestPrint_Table(t *testing.T) {
	data := map[string]string{"key1": "val1"}

	output := captureStdout(t, func() {
		if err := Print(FormatTable, data); err != nil {
			t.Fatal(err)
		}
	})

	if output == "" {
		t.Error("Print(Table) produced empty output")
	}
}

func TestPrint_UnsupportedFormat(t *testing.T) {
	err := Print(Format("xml"), nil)
	if err == nil {
		t.Error("Print(xml) should return error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error message = %q, should contain 'unsupported'", err.Error())
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
	if FormatTable != "table" {
		t.Errorf("FormatTable = %q, want %q", FormatTable, "table")
	}
	if FormatYAML != "yaml" {
		t.Errorf("FormatYAML = %q, want %q", FormatYAML, "yaml")
	}
	if FormatMarkdown != "markdown" {
		t.Errorf("FormatMarkdown = %q, want %q", FormatMarkdown, "markdown")
	}
	if FormatPipe != "pipe" {
		t.Errorf("FormatPipe = %q, want %q", FormatPipe, "pipe")
	}
}

func TestPrint_Pipe(t *testing.T) {
	data := json.RawMessage(`{"data":[{"campaignId":101,"campaignName":"FitTrack","id":1001,"name":"Brand Search","status":"ENABLED"}]}`)

	output := captureStdout(t, func() {
		if err := Print(FormatPipe, data); err != nil {
			t.Fatal(err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header plus row, got %d lines: %q", len(lines), output)
	}
	for _, want := range []string{"CAMPAIGN_ID", "CAMPAIGN_NAME", "ID", "NAME", "STATUS"} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("pipe header missing %q: %q", want, lines[0])
		}
	}
	for _, want := range []string{"101", "FitTrack", "1001", "Brand Search", "ENABLED"} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("pipe row missing %q: %q", want, lines[1])
		}
	}
	if !strings.Contains(lines[0], "\t") || !strings.Contains(lines[1], "\t") {
		t.Fatalf("pipe output should be tab-separated: %q", output)
	}
}

func TestSelectFields_EnvelopeArray(t *testing.T) {
	raw := json.RawMessage(`{"data":[{"campaignId":1,"name":"Alpha","bidAmount":{"amount":"1.20","currency":"USD"}}]}`)

	selected, err := SelectFields(raw, "NAME,campaign_id,bid_amount.amount", "")
	if err != nil {
		t.Fatalf("SelectFields error: %v", err)
	}

	out, err := json.Marshal(selected)
	if err != nil {
		t.Fatalf("Marshal selected fields: %v", err)
	}

	want := `[{"NAME":"Alpha","campaign_id":1,"bid_amount.amount":"1.20"}]`
	if string(out) != want {
		t.Fatalf("filtered JSON = %s, want %s", out, want)
	}
}

func TestSelectFields_Unknown(t *testing.T) {
	raw := json.RawMessage(`{"data":{"name":"Alpha"}}`)

	_, err := SelectFields(raw, "missing", "")
	if err == nil {
		t.Fatal("SelectFields should fail for unknown field")
	}
}

func TestPrintTable_Nil(t *testing.T) {
	err := PrintTable(nil)
	if err != nil {
		t.Errorf("PrintTable(nil) = %v, want nil", err)
	}
}

func TestPrintTable_Map(t *testing.T) {
	data := map[string]string{"key1": "val1", "key2": "val2"}

	output := captureStdout(t, func() {
		if err := PrintTable(data); err != nil {
			t.Fatal(err)
		}
	})

	if strings.Contains(output, "KEY  VALUE") || strings.Contains(output, "KEY   VALUE") {
		t.Errorf("table output should use column headers for string-keyed maps: %q", output)
	}
	for _, want := range []string{"KEY_1", "KEY_2", "val1", "val2"} {
		if !strings.Contains(output, want) {
			t.Errorf("table output missing %q: %q", want, output)
		}
	}
}

func TestPrintTable_Struct(t *testing.T) {
	type Item struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	data := Item{Name: "test", Count: 5}

	output := captureStdout(t, func() {
		if err := PrintTable(data); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "NAME") || !strings.Contains(output, "COUNT") {
		t.Errorf("table output missing NAME/COUNT headers: %q", output)
	}
	if !strings.Contains(output, "test") || !strings.Contains(output, "5") {
		t.Errorf("table output missing data: %q", output)
	}
}

func TestPrintTable_SliceOfStructs(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	data := []Item{
		{ID: 1, Name: "first"},
		{ID: 2, Name: "second"},
	}

	output := captureStdout(t, func() {
		if err := PrintTable(data); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "ID") || !strings.Contains(output, "NAME") {
		t.Errorf("table output missing headers: %q", output)
	}
	if !strings.Contains(output, "first") || !strings.Contains(output, "second") {
		t.Errorf("table output missing data: %q", output)
	}
}

func TestPrintTable_EmptySlice(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}
	data := []Item{}

	output := captureStdout(t, func() {
		if err := PrintTable(data); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "No results") {
		t.Errorf("expected 'No results.' for empty slice, got: %q", output)
	}
}

func TestPrintTable_EmptyMap(t *testing.T) {
	data := map[string]string{}

	output := captureStdout(t, func() {
		if err := PrintTable(data); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "No results") {
		t.Errorf("expected 'No results.' for empty map, got: %q", output)
	}
}

func TestPrintTable_RawMessage_APIEnvelope(t *testing.T) {
	// Simulates a typical Apple Ads API response with data envelope
	raw := json.RawMessage(`{
		"data": [
			{"id": 123, "name": "Campaign A", "status": "ENABLED"},
			{"id": 456, "name": "Campaign B", "status": "PAUSED"}
		],
		"pagination": {"totalResults": 2}
	}`)

	output := captureStdout(t, func() {
		if err := PrintTable(raw); err != nil {
			t.Fatal(err)
		}
	})

	// Should show column headers
	if !strings.Contains(output, "ID") || !strings.Contains(output, "NAME") || !strings.Contains(output, "STATUS") {
		t.Errorf("table output missing headers: %q", output)
	}
	// Should show actual data, not byte values
	if !strings.Contains(output, "Campaign A") || !strings.Contains(output, "Campaign B") {
		t.Errorf("table output missing data: %q", output)
	}
	if !strings.Contains(output, "123") || !strings.Contains(output, "456") {
		t.Errorf("table output missing IDs: %q", output)
	}
	if !strings.Contains(output, "ENABLED") || !strings.Contains(output, "PAUSED") {
		t.Errorf("table output missing status values: %q", output)
	}
}

func TestPrintTable_RawMessage_BareArray(t *testing.T) {
	raw := json.RawMessage(`[
		{"id": 1, "text": "keyword1", "matchType": "BROAD"},
		{"id": 2, "text": "keyword2", "matchType": "EXACT"}
	]`)

	output := captureStdout(t, func() {
		if err := PrintTable(raw); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "ID") || !strings.Contains(output, "TEXT") || !strings.Contains(output, "MATCH_TYPE") {
		t.Errorf("table output missing headers: %q", output)
	}
	if !strings.Contains(output, "keyword1") || !strings.Contains(output, "keyword2") {
		t.Errorf("table output missing data: %q", output)
	}
}

func TestPrintTable_RawMessage_SingleObject(t *testing.T) {
	raw := json.RawMessage(`{
		"data": {"id": 789, "name": "My Campaign", "status": "ENABLED"}
	}`)

	output := captureStdout(t, func() {
		if err := PrintTable(raw); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "ID") || !strings.Contains(output, "NAME") {
		t.Errorf("table output missing headers: %q", output)
	}
	if !strings.Contains(output, "789") || !strings.Contains(output, "My Campaign") {
		t.Errorf("table output missing data: %q", output)
	}
}

func TestPrintTable_RawMessage_MoneyFields(t *testing.T) {
	raw := json.RawMessage(`{
		"data": [
			{"id": 1, "name": "Test", "budget": {"amount": "100.50", "currency": "USD"}}
		]
	}`)

	output := captureStdout(t, func() {
		if err := PrintTable(raw); err != nil {
			t.Fatal(err)
		}
	})

	// Money should be flattened to "amount currency"
	if !strings.Contains(output, "100.50 USD") {
		t.Errorf("table output should flatten Money to 'amount currency', got: %q", output)
	}
}

func TestPrintTable_RawMessage_EmptyData(t *testing.T) {
	raw := json.RawMessage(`{"data": []}`)

	output := captureStdout(t, func() {
		if err := PrintTable(raw); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "No results") {
		t.Errorf("expected 'No results.' for empty data array, got: %q", output)
	}
}

func TestPrintTable_RawMessage_NoByteValues(t *testing.T) {
	// This is the bug regression test: json.RawMessage should NOT be printed as byte values.
	// The JSON `{"data":[{"id":1}]}` starts with byte 123 ('{').
	// Before the fix, PrintTable would show "123" (the byte value), not the actual data.
	raw := json.RawMessage(`{"data":[{"id":42,"name":"test"}]}`)

	output := captureStdout(t, func() {
		if err := PrintTable(raw); err != nil {
			t.Fatal(err)
		}
	})

	// Must contain actual field data
	if !strings.Contains(output, "42") || !strings.Contains(output, "test") {
		t.Errorf("expected actual data, got: %q", output)
	}
	// Must NOT contain "VALUE" header (which printScalarSlice would produce)
	if strings.Contains(output, "VALUE") {
		t.Errorf("output should NOT have scalar VALUE column (byte-by-byte rendering bug): %q", output)
	}
}

func TestPrintTable_RawMessage_ColumnOrdering(t *testing.T) {
	// id, name, status should appear before other fields
	raw := json.RawMessage(`[{"zebra": "z", "id": 1, "name": "test", "alpha": "a", "status": "ENABLED"}]`)

	output := captureStdout(t, func() {
		if err := PrintTable(raw); err != nil {
			t.Fatal(err)
		}
	})

	idIdx := strings.Index(output, "ID")
	nameIdx := strings.Index(output, "NAME")
	statusIdx := strings.Index(output, "STATUS")
	alphaIdx := strings.Index(output, "ALPHA")
	zebraIdx := strings.Index(output, "ZEBRA")

	if idIdx == -1 || nameIdx == -1 || statusIdx == -1 {
		t.Fatalf("missing priority headers in output: %q", output)
	}
	// Priority fields should come before alphabetical fields
	if idIdx > alphaIdx || nameIdx > alphaIdx || statusIdx > alphaIdx {
		t.Errorf("priority fields (id,name,status) should precede alpha: %q", output)
	}
	if alphaIdx > zebraIdx {
		t.Errorf("remaining fields should be alphabetical (alpha before zebra): %q", output)
	}
}

func TestFormatAny(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"integer float64", float64(42), "42"},
		{"decimal float64", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"money map", map[string]any{"amount": "10.00", "currency": "USD"}, "10.00 USD"},
		{"nested array", []any{"a", "b", "c"}, "a, b, c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAny(tt.val)
			if got != tt.want {
				t.Errorf("formatAny(%v) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}
