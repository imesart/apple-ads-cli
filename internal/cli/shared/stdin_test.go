package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// captureStdout replaces os.Stdout with a pipe and returns a cleanup
// function that restores os.Stdout and the read end of the pipe.
// The caller must call restore() before reading from the returned
// *os.File, then close it when done. If the captured output is not
// needed, simply call restore() — it drains and closes the pipe.
func captureStdout(t *testing.T) (restore func(), stdoutR *os.File) {
	t.Helper()
	oldStdout := os.Stdout
	r, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdout pipe: %v", err)
	}
	os.Stdout = stdoutW
	return func() {
		stdoutW.Close()
		os.Stdout = oldStdout
	}, r
}

func mustWriteString(t *testing.T, w *os.File, s string) {
	t.Helper()
	if _, err := w.WriteString(s); err != nil {
		t.Fatalf("writing pipe contents: %v", err)
	}
}

func TestCollectStdinFlags_None(t *testing.T) {
	a := "hello"
	b := "world"
	flags := CollectStdinFlags(
		StdinFlag{"a", &a},
		StdinFlag{"b", &b},
	)
	if len(flags) != 0 {
		t.Errorf("expected 0 stdin flags, got %d", len(flags))
	}
}

func TestCollectStdinFlags_Some(t *testing.T) {
	a := "-"
	b := "world"
	c := "-"
	flags := CollectStdinFlags(
		StdinFlag{"a", &a},
		StdinFlag{"b", &b},
		StdinFlag{"c", &c},
	)
	if len(flags) != 2 {
		t.Fatalf("expected 2 stdin flags, got %d", len(flags))
	}
	if flags[0].Name != "a" || flags[1].Name != "c" {
		t.Errorf("got flags %v %v, want a and c", flags[0].Name, flags[1].Name)
	}
}

func TestHasStdinFlags(t *testing.T) {
	a := "hello"
	b := "-"
	if HasStdinFlags(StdinFlag{"a", &a}) {
		t.Error("expected false")
	}
	if !HasStdinFlags(StdinFlag{"a", &a}, StdinFlag{"b", &b}) {
		t.Error("expected true")
	}
}

func TestFlagToColumnName(t *testing.T) {
	tests := []struct {
		flag string
		want string
	}{
		{"campaign-id", "CAMPAIGN_ID"},
		{"adgroup-id", "AD_GROUP_ID"},
		{"keyword-id", "KEYWORD_ID"},
		{"ad-id", "AD_ID"},
		{"budget-order-id", "BUDGET_ORDER_ID"},
		{"id", "ID"},
	}
	for _, tt := range tests {
		got := FlagToColumnName(tt.flag)
		if got != tt.want {
			t.Errorf("FlagToColumnName(%q) = %q, want %q", tt.flag, got, tt.want)
		}
	}
}

func TestRunWithStdin_Basic(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "111\t5001\n222\t5002\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Capture stdout (RunWithStdin prints merged output)
	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID, adgroupID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
	}

	var calls []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		calls = append(calls, campaignID+":"+adgroupID)
		return map[string]string{"cid": campaignID, "agid": adgroupID}, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0] != "111:5001" || calls[1] != "222:5002" {
		t.Errorf("got calls %v", calls)
	}
}

func TestRunWithStdin_SingleColumn(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "AAA\nBBB\nCCC\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var id string
	stdinFlags := []StdinFlag{{"id", &id}}

	var ids []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		ids = append(ids, id)
		return id, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 || ids[0] != "AAA" || ids[1] != "BBB" || ids[2] != "CCC" {
		t.Errorf("got ids %v", ids)
	}
}

func TestRunWithStdin_SkipsEmptyLines(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "111\n\n222\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var id string
	stdinFlags := []StdinFlag{{"id", &id}}

	count := 0
	err := RunWithStdin(stdinFlags, func() (any, error) {
		count++
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 calls, got %d", count)
	}
}

func TestRunWithStdin_TooFewColumns(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "111\n") // only 1 column but need 2
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	var a, b string
	stdinFlags := []StdinFlag{{"a", &a}, {"b", &b}}

	oldStderr := os.Stderr
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	err := RunWithStdin(stdinFlags, func() (any, error) {
		t.Fatal("should not be called")
		return nil, nil
	}, "json", "", false)

	stderrW.Close()
	os.Stderr = oldStderr
	buf := make([]byte, 4096)
	n, _ := stderrR.Read(buf)
	stderrR.Close()
	stderrOut := string(buf[:n])

	if err == nil {
		t.Fatal("expected error for too few columns")
	}
	if !strings.Contains(stderrOut, "expected 2 columns") {
		t.Errorf("stderr: %q", stderrOut)
	}
}

func TestRunWithStdin_EmptyInput(t *testing.T) {
	r, w, _ := os.Pipe()
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	var id string
	stdinFlags := []StdinFlag{{"id", &id}}

	err := RunWithStdin(stdinFlags, func() (any, error) {
		t.Fatal("should not be called")
		return nil, nil
	}, "json", "", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunWithStdin_ErrorContinues(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "111\n222\n333\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	oldStderr := os.Stderr
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var id string
	stdinFlags := []StdinFlag{{"id", &id}}

	count := 0
	err := RunWithStdin(stdinFlags, func() (any, error) {
		count++
		if id == "222" {
			return nil, fmt.Errorf("simulated error for %s", id)
		}
		return id, nil
	}, "json", "", false)

	restoreStdout()

	stderrW.Close()
	os.Stderr = oldStderr
	buf := make([]byte, 4096)
	n, _ := stderrR.Read(buf)
	stderrR.Close()
	stderrOut := string(buf[:n])

	if count != 3 {
		t.Errorf("expected 3 calls, got %d", count)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "1 of 3 lines failed") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(stderrOut, "simulated error for 222") {
		t.Errorf("stderr missing error detail: %q", stderrOut)
	}
}

// --- Header detection tests ---

func TestRunWithStdin_WithHeader(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "CAMPAIGNID\tADGROUPID\n111\t5001\n222\t5002\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID, adgroupID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
	}

	var calls []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		calls = append(calls, campaignID+":"+adgroupID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0] != "111:5001" || calls[1] != "222:5002" {
		t.Errorf("got calls %v", calls)
	}
}

func TestRunWithStdin_WithHeader_ErrorCountsDataRowsOnly(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "CAMPAIGNID\tADGROUPID\n111\t5001\n222\t5002\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	oldStderr := os.Stderr
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	err := RunWithStdin([]StdinFlag{
		{"campaign-id", new(string)},
		{"adgroup-id", new(string)},
	}, func() (any, error) {
		return nil, fmt.Errorf("simulated failure")
	}, "json", "", false)

	stderrW.Close()
	os.Stderr = oldStderr
	_, _ = stderrR.Read(make([]byte, 4096))
	stderrR.Close()

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "2 of 2 lines failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunWithStdin_HeaderReorder(t *testing.T) {
	// Header columns in reverse order compared to flags
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "ADGROUPID\tCAMPAIGNID\n5001\t111\n5002\t222\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID, adgroupID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
	}

	var calls []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		calls = append(calls, campaignID+":"+adgroupID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	// campaign-id should get 111 (from col 1), adgroup-id should get 5001 (from col 0)
	if calls[0] != "111:5001" || calls[1] != "222:5002" {
		t.Errorf("got calls %v, want [111:5001 222:5002]", calls)
	}
}

func TestRunWithStdin_HeaderWithIDColumn(t *testing.T) {
	// "ID" column should resolve to campaign-id (only flag)
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "ID\n111\n222\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID string
	stdinFlags := []StdinFlag{{"campaign-id", &campaignID}}

	var ids []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		ids = append(ids, campaignID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != "111" || ids[1] != "222" {
		t.Errorf("got ids %v, want [111 222]", ids)
	}
}

func TestRunWithStdin_IDResolvesToAdGroup(t *testing.T) {
	// CAMPAIGNID + ID: ID should resolve to adgroup-id
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "CAMPAIGN_ID\tID\n111\t5001\n222\t5002\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID, adgroupID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
	}

	var calls []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		calls = append(calls, campaignID+":"+adgroupID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0] != "111:5001" || calls[1] != "222:5002" {
		t.Errorf("got calls %v, want [111:5001 222:5002]", calls)
	}
}

func TestRunWithStdin_LegacyCompactHeadersStillWork(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "CAMPAIGNID\tADGROUPID\n111\t5001\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID, adgroupID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
	}

	var calls []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		calls = append(calls, campaignID+":"+adgroupID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 1 || calls[0] != "111:5001" {
		t.Errorf("got calls %v, want [111:5001]", calls)
	}
}

func TestRunWithStdin_JSONSingleIDArray(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, `[{"id":111},{"id":222}]`)
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID string
	stdinFlags := []StdinFlag{{"campaign-id", &campaignID}}

	var ids []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		ids = append(ids, campaignID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != "111" || ids[1] != "222" {
		t.Fatalf("got ids %v, want [111 222]", ids)
	}
}

func TestRunWithStdin_JSONEnvelopeIDResolvesToAdGroup(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "{\n  \"data\": [\n    {\"campaignId\": 111, \"id\": 5001},\n    {\"campaignId\": 222, \"id\": 5002}\n  ]\n}\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID, adgroupID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
	}

	var calls []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		calls = append(calls, campaignID+":"+adgroupID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 2 || calls[0] != "111:5001" || calls[1] != "222:5002" {
		t.Fatalf("got calls %v, want [111:5001 222:5002]", calls)
	}
}

func TestRunWithStdin_JSONBareArrayIDResolvesToKeyword(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "[\n  {\"campaignId\": 111, \"adGroupId\": 5001, \"id\": 9001}\n]\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID, adgroupID, keywordID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
		{"keyword-id", &keywordID},
	}

	var calls []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		calls = append(calls, campaignID+":"+adgroupID+":"+keywordID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 1 || calls[0] != "111:5001:9001" {
		t.Fatalf("got calls %v, want [111:5001:9001]", calls)
	}
}

func TestRunWithStdin_IDResolvesToKeyword(t *testing.T) {
	// CAMPAIGNID + ADGROUPID + ID: ID should resolve to keyword-id
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "CAMPAIGN_ID\tADGROUP_ID\tID\n111\t5001\t9001\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID, adgroupID, keywordID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
		{"keyword-id", &keywordID},
	}

	var calls []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		calls = append(calls, campaignID+":"+adgroupID+":"+keywordID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 1 || calls[0] != "111:5001:9001" {
		t.Errorf("got calls %v, want [111:5001:9001]", calls)
	}
}

func TestRunWithStdin_HeaderWithExtraColumns(t *testing.T) {
	// Extra columns in header beyond what flags need — should be ignored
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "CAMPAIGN_ID\tADGROUP_ID\tKEYWORD_ID\n111\t5001\t9001\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	// Only need campaign-id
	var campaignID string
	stdinFlags := []StdinFlag{{"campaign-id", &campaignID}}

	var ids []string
	err := RunWithStdin(stdinFlags, func() (any, error) {
		ids = append(ids, campaignID)
		return nil, nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "111" {
		t.Errorf("got ids %v, want [111]", ids)
	}
}

func TestRunWithStdin_CommandStylePipelineForReports(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "CAMPAIGN_ID\n111\n222\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)
	defer stdoutR.Close()

	var campaignID string
	stdinFlags := []StdinFlag{{"campaign-id", &campaignID}}

	var gotCampaignIDs []string
	var gotBodies []map[string]any
	err := RunWithStdin(stdinFlags, func() (any, error) {
		gotCampaignIDs = append(gotCampaignIDs, campaignID)

		start, err := parseDateFlagAt("-7d", mustDateTime(t, 2026, 3, 25))
		if err != nil {
			return nil, err
		}
		end, err := parseDateFlagAt("now", mustDateTime(t, 2026, 3, 25))
		if err != nil {
			return nil, err
		}

		body := buildReportRequest(start, end, "", "", "UTC", true, true, false, "", "impressions:desc")
		gotBodies = append(gotBodies, body)
		return json.RawMessage(`{"data":[]}`), nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotCampaignIDs) != 2 || gotCampaignIDs[0] != "111" || gotCampaignIDs[1] != "222" {
		t.Fatalf("got campaign IDs %v, want [111 222]", gotCampaignIDs)
	}
	if len(gotBodies) != 2 {
		t.Fatalf("got %d request bodies, want 2", len(gotBodies))
	}
	selector := gotBodies[0]["selector"].(map[string]any)
	orderBy := selector["orderBy"].([]map[string]any)
	if orderBy[0]["field"] != "impressions" {
		t.Fatalf("field = %v, want impressions", orderBy[0]["field"])
	}
	if orderBy[0]["sortOrder"] != "DESCENDING" {
		t.Fatalf("sortOrder = %v, want DESCENDING", orderBy[0]["sortOrder"])
	}
	if gotBodies[0]["startTime"] != "2026-03-18" || gotBodies[0]["endTime"] != "2026-03-25" {
		t.Fatalf("got start/end = %v/%v, want 2026-03-18/2026-03-25", gotBodies[0]["startTime"], gotBodies[0]["endTime"])
	}
}

func TestParseStdinHeader_NotAHeader(t *testing.T) {
	// Numeric first column → not a header
	stdinFlags := []StdinFlag{{"campaign-id", new(string)}}
	_, ok := parseStdinHeader([]string{"12345"}, stdinFlags)
	if ok {
		t.Error("expected false for numeric data")
	}
}

func mustDateTime(t *testing.T, year int, month int, day int) time.Time {
	t.Helper()
	loc := time.FixedZone("UTC+2", 2*60*60)
	return time.Date(year, time.Month(month), day, 12, 0, 0, 0, loc)
}

func TestParseStdinHeader_UnmatchableHeader(t *testing.T) {
	// Header that doesn't match any flags → not a header
	stdinFlags := []StdinFlag{{"campaign-id", new(string)}}
	_, ok := parseStdinHeader([]string{"FOOBAR"}, stdinFlags)
	if ok {
		t.Error("expected false for unmatched header")
	}
}

// --- Merge tests ---

func TestMergeResults_SingleResult(t *testing.T) {
	data := json.RawMessage(`{"data":[{"id":1}]}`)
	result := MergeResults([]any{data})
	// Single result should be returned as-is
	raw, ok := result.(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage, got %T", result)
	}
	if string(raw) != string(data) {
		t.Errorf("got %s", string(raw))
	}
}

func TestMergeResults_MultipleEnvelopes(t *testing.T) {
	r1 := json.RawMessage(`{"data":[{"id":1,"name":"A"},{"id":2,"name":"B"}]}`)
	r2 := json.RawMessage(`{"data":[{"id":3,"name":"C"}]}`)

	result := MergeResults([]any{r1, r2})
	raw, ok := result.(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage, got %T", result)
	}

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(envelope.Data) != 3 {
		t.Fatalf("expected 3 records, got %d", len(envelope.Data))
	}
	if envelope.Data[0]["name"] != "A" || envelope.Data[2]["name"] != "C" {
		t.Errorf("unexpected data: %v", envelope.Data)
	}
}

func TestMergeResults_SingleObjectEnvelopes(t *testing.T) {
	r1 := json.RawMessage(`{"data":{"id":1}}`)
	r2 := json.RawMessage(`{"data":{"id":2}}`)

	result := MergeResults([]any{r1, r2})
	raw, ok := result.(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage, got %T", result)
	}

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(envelope.Data) != 2 {
		t.Fatalf("expected 2 records, got %d", len(envelope.Data))
	}
}

func TestMergeResults_Empty(t *testing.T) {
	result := MergeResults(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestRunWithStdin_CheckSummariesRenderStableHeaders(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "CAMPAIGN_ID\tADGROUP_ID\n1111111111\t5001000000\n2222222222\t5002000000\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)

	var campaignID, adgroupID string
	stdinFlags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adgroupID},
	}

	err := RunWithStdin(stdinFlags, func() (any, error) {
		return MutationCheckSummary{
			Result:          "Check passed.",
			Action:          "update adgroup",
			Target:          fmt.Sprintf("campaign %s, adgroup %s", campaignID, adgroupID),
			WouldAffect:     "1 object",
			ResolvedChanges: []string{"defaultBidAmount: 1.15 USD"},
			Safety:          []string{"bid and CPA goal limits ok"},
			ReadOnlyChecks:  []string{"fetched current adgroup"},
		}, nil
	}, "table", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := stdoutR.Read(buf)
	stdoutR.Close()
	lines := strings.Split(strings.TrimRight(string(buf[:n]), "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header and at least one row, got %q", string(buf[:n]))
	}

	want := "RESULT         ACTION          TARGET                                   WOULD_AFFECT  RESOLVED_CHANGES            SAFETY                      READ_ONLY_CHECKS         WARNINGS"
	if got := strings.TrimRight(lines[0], " "); got != want {
		t.Fatalf("header = %q, want %q", got, want)
	}
}

// --- Context augmentation tests ---

func TestAugmentWithContext_EnvelopeArray(t *testing.T) {
	result := json.RawMessage(`{"data":[{"id":1,"name":"A"},{"id":2,"name":"B"}]}`)
	ctx := map[string]any{"campaignId": "111"}

	augmented := augmentWithContext(result, ctx)
	raw, ok := augmented.(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage, got %T", augmented)
	}

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(envelope.Data) != 2 {
		t.Fatalf("expected 2 records, got %d", len(envelope.Data))
	}
	for i, rec := range envelope.Data {
		if rec["campaignId"] != "111" {
			t.Errorf("record %d: campaignId = %v, want 111", i, rec["campaignId"])
		}
	}
	// Original fields preserved
	if envelope.Data[0]["name"] != "A" {
		t.Errorf("record 0 name = %v, want A", envelope.Data[0]["name"])
	}
}

func TestAugmentWithContext_DoesNotOverwrite(t *testing.T) {
	// If the result already has campaignId, don't overwrite it
	result := json.RawMessage(`{"data":[{"id":1,"campaignId":"existing"}]}`)
	ctx := map[string]any{"campaignId": "999"}

	augmented := augmentWithContext(result, ctx)
	raw := augmented.(json.RawMessage)

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data[0]["campaignId"] != "existing" {
		t.Errorf("campaignId = %v, want existing (should not overwrite)", envelope.Data[0]["campaignId"])
	}
}

func TestAugmentWithContext_EmptyContext(t *testing.T) {
	result := json.RawMessage(`{"data":[{"id":1}]}`)
	augmented := augmentWithContext(result, nil)
	// Should return original unchanged
	raw := augmented.(json.RawMessage)
	if string(raw) != `{"data":[{"id":1}]}` {
		t.Errorf("got %s, want original unchanged", string(raw))
	}
}

func TestAugmentWithContext_MultipleIDs(t *testing.T) {
	result := json.RawMessage(`{"data":[{"id":9001}]}`)
	ctx := map[string]any{"campaignId": "111", "adGroupId": "222"}

	augmented := augmentWithContext(result, ctx)
	raw := augmented.(json.RawMessage)

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data[0]["campaignId"] != "111" {
		t.Errorf("campaignId = %v, want 111", envelope.Data[0]["campaignId"])
	}
	if envelope.Data[0]["adGroupId"] != "222" {
		t.Errorf("adGroupId = %v, want 222", envelope.Data[0]["adGroupId"])
	}
}

func TestCaptureIDContext(t *testing.T) {
	campaignID := "111"
	adGroupID := "222"
	flags := []StdinFlag{
		{"campaign-id", &campaignID},
		{"adgroup-id", &adGroupID},
	}
	ctx := captureIDContext(flags)
	if ctx["campaignId"] != "111" {
		t.Errorf("campaignId = %q, want 111", ctx["campaignId"])
	}
	if ctx["adGroupId"] != "222" {
		t.Errorf("adGroupId = %q, want 222", ctx["adGroupId"])
	}
}

func TestDeriveStdinContextFromRecord_CapturesExplicitAppFields(t *testing.T) {
	record := map[string]any{
		"adamId":  json.Number("900001"),
		"appName": "FitTrack SE: fitness calories",
	}

	ctx := deriveStdinContextFromRecord(record, nil)
	if got := ctx["adamId"]; got != json.Number("900001") {
		t.Fatalf("adamId = %#v, want 900001", got)
	}
	if got := ctx["appName"]; got != "FitTrack SE: fitness calories" {
		t.Fatalf("appName = %#v, want FitTrack SE: fitness calories", got)
	}
	if got := ctx["appNameShort"]; got != "FitTrack SE" {
		t.Fatalf("appNameShort = %#v, want FitTrack SE", got)
	}
}

func TestDeriveStdinContextFromRecord_CapturesExplicitCreativeAndProductPageFields(t *testing.T) {
	record := map[string]any{
		"creativeId":    json.Number("900601"),
		"productPageId": "cpp-fitness-strength",
	}

	ctx := deriveStdinContextFromRecord(record, nil)
	if got := ctx["creativeId"]; got != json.Number("900601") {
		t.Fatalf("creativeId = %#v, want 900601", got)
	}
	if got := ctx["productPageId"]; got != "cpp-fitness-strength" {
		t.Fatalf("productPageId = %#v, want cpp-fitness-strength", got)
	}
}

func TestDeriveStdinContextFromRecord_DoesNotMapIDToAdamID(t *testing.T) {
	record := map[string]any{
		"id": "app-store-id",
	}

	ctx := deriveStdinContextFromRecord(record, nil)
	if _, ok := ctx["adamId"]; ok {
		t.Fatalf("adamId unexpectedly present in context: %#v", ctx["adamId"])
	}
}

func TestDeriveStdinContextFromRecord_MapsCreativeAndProductPageIDFromRowShape(t *testing.T) {
	creativeCtx := deriveStdinContextFromRecord(map[string]any{
		"id":     json.Number("900601"),
		"adamId": json.Number("900001"),
		"type":   "CUSTOM_PRODUCT_PAGE",
	}, nil)
	if got := creativeCtx["creativeId"]; got != json.Number("900601") {
		t.Fatalf("creativeId = %#v, want 900601", got)
	}

	productPageCtx := deriveStdinContextFromRecord(map[string]any{
		"id":   "cpp-fitness-strength",
		"name": "FitTrack Strength Page",
	}, nil)
	if got := productPageCtx["productPageId"]; got != "cpp-fitness-strength" {
		t.Fatalf("productPageId = %#v, want cpp-fitness-strength", got)
	}
}

func TestDeriveStdinContextFromRecord_AppNameShortSeparators(t *testing.T) {
	tests := []struct {
		name    string
		appName string
		want    string
	}{
		{name: "colon", appName: "FitTrack SE: fitness calories", want: "FitTrack SE"},
		{name: "comma", appName: "FitTrack SE, fitness calories", want: "FitTrack SE"},
		{name: "hyphen with spaces", appName: "FitTracker - SE", want: "FitTracker"},
		{name: "hyphen without spaces stays intact", appName: "K-Mart", want: "K-Mart"},
		{name: "pipe", appName: "FitTrack | fitness calories", want: "FitTrack"},
		{name: "bullet", appName: "FitTrack • fitness calories", want: "FitTrack"},
		{name: "en dash", appName: "FitTrack – fitness calories", want: "FitTrack"},
		{name: "leading separator space", appName: "FitTrack: fitness calories", want: "FitTrack"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := deriveStdinContextFromRecord(map[string]any{"appName": tt.appName}, nil)
			if got := ctx["appNameShort"]; got != tt.want {
				t.Fatalf("appNameShort = %#v, want %q", got, tt.want)
			}
		})
	}
}

func TestRunWithStdin_AugmentsResultsWithContext(t *testing.T) {
	r, w, _ := os.Pipe()
	mustWriteString(t, w, "111\n222\n")
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	restoreStdout, stdoutR := captureStdout(t)

	var campaignID string
	stdinFlags := []StdinFlag{{"campaign-id", &campaignID}}

	err := RunWithStdin(stdinFlags, func() (any, error) {
		return json.RawMessage(fmt.Sprintf(`{"data":[{"id":1,"name":"adgroup for %s"}]}`, campaignID)), nil
	}, "json", "", false)

	restoreStdout()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := stdoutR.Read(buf)
	stdoutR.Close()

	var rows []map[string]any
	if err := json.Unmarshal(buf[:n], &rows); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 records, got %d", len(rows))
	}
	// First result should have campaignId=111, second campaignId=222
	if rows[0]["campaignId"] != "111" {
		t.Errorf("record 0 campaignId = %v, want 111", rows[0]["campaignId"])
	}
	if rows[1]["campaignId"] != "222" {
		t.Errorf("record 1 campaignId = %v, want 222", rows[1]["campaignId"])
	}
}
