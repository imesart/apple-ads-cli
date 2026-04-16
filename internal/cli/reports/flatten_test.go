package reports

import (
	"encoding/json"
	"testing"
)

func TestFlattenReportResponse_FlattensTotalsRows(t *testing.T) {
	raw := json.RawMessage(`{
		"reportingDataResponse": {
			"row": [
				{
					"metadata": {
						"campaignId": 111,
						"campaignName": "Alpha",
						"app": {
							"adamId": 123456,
							"appName": "My App"
						}
					},
					"total": {
						"impressions": 1000,
						"taps": 50,
						"localSpend": {"amount":"12.34","currency":"USD"}
					}
				}
			],
			"grandTotals": {
				"total": {
					"impressions": 1000,
					"taps": 50
				}
			}
		}
	}`)

	got, err := flattenReportResponse(raw)
	if err != nil {
		t.Fatalf("flattenReportResponse error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got.(json.RawMessage), &envelope); err != nil {
		t.Fatalf("unmarshal flattened envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
	row := data[0].(map[string]any)
	if row["campaignId"] != float64(111) {
		t.Fatalf("campaignId = %v, want 111", row["campaignId"])
	}
	if row["campaignName"] != "Alpha" {
		t.Fatalf("campaignName = %v, want Alpha", row["campaignName"])
	}
	if row["appAdamId"] != float64(123456) {
		t.Fatalf("appAdamId = %v, want 123456", row["appAdamId"])
	}
	if row["appAppName"] != "My App" {
		t.Fatalf("appAppName = %v, want My App", row["appAppName"])
	}
	if row["impressions"] != float64(1000) {
		t.Fatalf("impressions = %v, want 1000", row["impressions"])
	}
	localSpend := row["localSpend"].(map[string]any)
	if localSpend["amount"] != "12.34" || localSpend["currency"] != "USD" {
		t.Fatalf("localSpend = %v, want amount=12.34 currency=USD", localSpend)
	}

	grandTotals := envelope["grandTotals"].(map[string]any)
	if grandTotals["totalImpressions"] != float64(1000) {
		t.Fatalf("grandTotals.totalImpressions = %v, want 1000", grandTotals["totalImpressions"])
	}
}

func TestFlattenReportResponse_ExplodesGranularityRows(t *testing.T) {
	raw := json.RawMessage(`{
		"reportingDataResponse": {
			"row": [
				{
					"metadata": {
						"campaignId": 111,
						"adGroupId": 222,
						"adGroupName": "Group A"
					},
					"granularity": [
						{"date": "2026-03-20", "impressions": 10, "taps": 1},
						{"date": "2026-03-21", "impressions": 20, "taps": 2}
					]
				}
			]
		}
	}`)

	got, err := flattenReportResponse(raw)
	if err != nil {
		t.Fatalf("flattenReportResponse error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got.(json.RawMessage), &envelope); err != nil {
		t.Fatalf("unmarshal flattened envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 2 {
		t.Fatalf("data len = %d, want 2", len(data))
	}
	row0 := data[0].(map[string]any)
	row1 := data[1].(map[string]any)
	if row0["campaignId"] != float64(111) || row0["adGroupId"] != float64(222) {
		t.Fatalf("first row metadata = %v", row0)
	}
	if row0["date"] != "2026-03-20" || row0["impressions"] != float64(10) {
		t.Fatalf("first row = %v", row0)
	}
	if row1["date"] != "2026-03-21" || row1["taps"] != float64(2) {
		t.Fatalf("second row = %v", row1)
	}
}

func TestFlattenReportResponse_UnwrapsDataEnvelope(t *testing.T) {
	raw := json.RawMessage(`{
		"data": {
			"reportingDataResponse": {
				"row": [
					{
						"metadata": {
							"campaignId": 111,
							"adGroupId": 5001,
							"adGroupName": "Brand Exact"
						},
						"total": {
							"impressions": 1234,
							"taps": 56
						}
					}
				]
			}
		}
	}`)

	got, err := flattenReportResponse(raw)
	if err != nil {
		t.Fatalf("flattenReportResponse error: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(got.(json.RawMessage), &envelope); err != nil {
		t.Fatalf("unmarshal flattened envelope: %v", err)
	}
	data := envelope["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data len = %d, want 1", len(data))
	}
	row := data[0].(map[string]any)
	if row["campaignId"] != float64(111) || row["adGroupId"] != float64(5001) {
		t.Fatalf("row metadata = %v", row)
	}
	if row["adGroupName"] != "Brand Exact" || row["impressions"] != float64(1234) || row["taps"] != float64(56) {
		t.Fatalf("row = %v", row)
	}
}

func TestFlattenReportResponse_PassthroughNonReportEnvelope(t *testing.T) {
	raw := json.RawMessage(`{"data":[{"id":1}]}`)

	got, err := flattenReportResponse(raw)
	if err != nil {
		t.Fatalf("flattenReportResponse error: %v", err)
	}

	passthrough, ok := got.(json.RawMessage)
	if !ok {
		t.Fatalf("type = %T, want json.RawMessage", got)
	}
	if string(passthrough) != string(raw) {
		t.Fatalf("got %s, want %s", passthrough, raw)
	}
}

func TestFlattenReportResponse_NullDataBecomesEmptyArray(t *testing.T) {
	raw := json.RawMessage(`{"data":null}`)

	got, err := flattenReportResponse(raw)
	if err != nil {
		t.Fatalf("flattenReportResponse error: %v", err)
	}

	flattened, ok := got.(json.RawMessage)
	if !ok {
		t.Fatalf("type = %T, want json.RawMessage", got)
	}

	var envelope map[string]any
	if err := json.Unmarshal(flattened, &envelope); err != nil {
		t.Fatalf("unmarshal flattened envelope: %v", err)
	}
	data, ok := envelope["data"].([]any)
	if !ok {
		t.Fatalf("data type = %T, want []any", envelope["data"])
	}
	if len(data) != 0 {
		t.Fatalf("data len = %d, want 0", len(data))
	}
}
