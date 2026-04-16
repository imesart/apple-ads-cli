package output

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func captureIDs(t *testing.T, data any, entityIDName string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := PrintIDs(data, entityIDName)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("PrintIDs error: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}

func TestPrintIDs_CampaignsEnvelope(t *testing.T) {
	data := json.RawMessage(`{"data":[{"id":100,"name":"Camp A"},{"id":200,"name":"Camp B"}]}`)
	got := captureIDs(t, data, "CAMPAIGNID")
	want := "CAMPAIGN_ID\n100\n200\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_CampaignsInferred(t *testing.T) {
	data := json.RawMessage(`{"data":[{"id":100},{"id":200}]}`)
	got := captureIDs(t, data, "")
	want := "CAMPAIGN_ID\n100\n200\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_AdGroupsWithParent(t *testing.T) {
	data := json.RawMessage(`{"data":[
		{"id":5001,"campaignId":111,"name":"AG1"},
		{"id":5002,"campaignId":111,"name":"AG2"}
	]}`)
	got := captureIDs(t, data, "ADGROUPID")
	want := "CAMPAIGN_ID\tAD_GROUP_ID\n111\t5001\n111\t5002\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_AdGroupsInferred(t *testing.T) {
	data := json.RawMessage(`{"data":[
		{"id":5001,"campaignId":111}
	]}`)
	got := captureIDs(t, data, "")
	want := "CAMPAIGN_ID\tAD_GROUP_ID\n111\t5001\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_ThreeLevelHierarchy(t *testing.T) {
	data := json.RawMessage(`{"data":[
		{"id":1,"campaignId":10,"adGroupId":100}
	]}`)
	got := captureIDs(t, data, "KEYWORDID")
	want := "CAMPAIGN_ID\tAD_GROUP_ID\tKEYWORD_ID\n10\t100\t1\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_ThreeLevelInferred(t *testing.T) {
	data := json.RawMessage(`{"data":[
		{"id":1,"campaignId":10,"adGroupId":100}
	]}`)
	got := captureIDs(t, data, "")
	want := "CAMPAIGN_ID\tAD_GROUP_ID\tKEYWORD_ID\n10\t100\t1\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_AdsExplicitName(t *testing.T) {
	data := json.RawMessage(`{"data":[
		{"id":1,"campaignId":10,"adGroupId":100}
	]}`)
	got := captureIDs(t, data, "ADID")
	want := "CAMPAIGN_ID\tAD_GROUP_ID\tAD_ID\n10\t100\t1\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_SingleObject(t *testing.T) {
	data := json.RawMessage(`{"data":{"id":42,"campaignId":99}}`)
	got := captureIDs(t, data, "ADGROUPID")
	want := "CAMPAIGN_ID\tAD_GROUP_ID\n99\t42\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_EmptyData(t *testing.T) {
	data := json.RawMessage(`{"data":[]}`)
	got := captureIDs(t, data, "")
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestPrintIDs_NullData(t *testing.T) {
	data := json.RawMessage(`{"data":null}`)
	got := captureIDs(t, data, "")
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestPrintIDs_BareArray(t *testing.T) {
	data := json.RawMessage(`[{"id":1},{"id":2}]`)
	got := captureIDs(t, data, "CAMPAIGNID")
	want := "CAMPAIGN_ID\n1\n2\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_BareObject(t *testing.T) {
	data := json.RawMessage(`{"id":7,"campaignId":3}`)
	got := captureIDs(t, data, "ADGROUPID")
	want := "CAMPAIGN_ID\tAD_GROUP_ID\n3\t7\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintIDs_NoIDField(t *testing.T) {
	data := json.RawMessage(`{"data":[{"name":"no id here"}]}`)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := PrintIDs(data, "")
	w.Close()
	os.Stdout = old
	r.Close()

	if err == nil {
		t.Fatal("expected error for records with no id fields")
	}
	if !strings.Contains(err.Error(), "no id fields") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPrintIDs_IntegrationWithPrint(t *testing.T) {
	data := json.RawMessage(`{"data":[{"id":42}]}`)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Print(FormatIDs, data, "CAMPAIGNID")
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("Print(FormatIDs) error: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	r.Close()
	got := string(buf[:n])
	if got != "CAMPAIGN_ID\n42\n" {
		t.Errorf("got %q, want %q", got, "CAMPAIGN_ID\n42\n")
	}
}
