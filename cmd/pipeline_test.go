package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	apiPkg "github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newTestClient creates an API client with the given round-trip function.
func newTestClient(fn roundTripFunc) *apiPkg.Client {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{Transport: fn})
	return client
}

// recordingTransport returns a roundTripFunc that validates method+path and
// passes the request body to fn for payload assertions.
func recordingTransport(t *testing.T, wantMethod, wantPath string, fn func(body []byte) *http.Response) roundTripFunc {
	t.Helper()
	return func(req *http.Request) (*http.Response, error) {
		if req.Method != wantMethod || req.URL.Path != wantPath {
			t.Fatalf("unexpected request %s %s, want %s %s", req.Method, req.URL.Path, wantMethod, wantPath)
		}
		var body []byte
		if req.Body != nil {
			body, _ = io.ReadAll(req.Body)
		}
		return fn(body), nil
	}
}

// mustUnmarshalMap unmarshals body as JSON object.
func mustUnmarshalMap(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("unmarshal body: %v\nbody: %s", err, body)
	}
	return m
}

// mustUnmarshalSlice unmarshals body as JSON array.
func mustUnmarshalSlice(t *testing.T, body []byte) []any {
	t.Helper()
	var s []any
	if err := json.Unmarshal(body, &s); err != nil {
		t.Fatalf("unmarshal body: %v\nbody: %s", err, body)
	}
	return s
}

// emptyOKResponse returns an HTTP 200 with no body (for DELETE endpoints).
func emptyOKResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}
}

// ---------------------------------------------------------------------------
// IDs format output
// ---------------------------------------------------------------------------

func TestCampaignsListIDsFormatOutput(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(`{"data":[{"id":101,"name":"ABC One","status":"ENABLED"},{"id":202,"name":"ABC Two","status":"PAUSED"}]}`), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{"campaigns", "list", "-f", "ids"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d: %q", len(lines), out)
	}
	if !strings.Contains(lines[0], "CAMPAIGN_ID") {
		t.Fatalf("expected CAMPAIGN_ID header, got %q", lines[0])
	}
	if !strings.Contains(out, "101") || !strings.Contains(out, "202") {
		t.Fatalf("expected campaign IDs 101 and 202, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Multi-level stdin piping
// ---------------------------------------------------------------------------

func TestAdGroupsListStdinPipeFromCampaigns(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns":
			return jsonResponse(`{"data":[{"id":101,"name":"C1","status":"ENABLED"},{"id":202,"name":"C2","status":"PAUSED"}]}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/101/adgroups":
			return jsonResponse(`{"data":[{"id":1001,"name":"Group A","status":"ENABLED"},{"id":1002,"name":"Group B","status":"PAUSED"}]}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/202/adgroups":
			return jsonResponse(`{"data":[{"id":2001,"name":"Group C","status":"ENABLED"}]}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	// Step 1: get campaign IDs
	idsOut, code := captureRun(t, []string{"campaigns", "list", "-f", "ids"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, idsOut)
	}

	// Step 2: pipe into adgroups list
	agOut, code := captureRun(t, []string{"adgroups", "list", "--campaign-id", "-", "-f", "ids"}, idsOut)
	if code != ExitSuccess {
		t.Fatalf("adgroups list exit code = %d; output=%q", code, agOut)
	}
	// Should have a single header row and all ad group IDs
	headerCount := strings.Count(agOut, "AD_GROUP_ID")
	if headerCount != 1 {
		t.Fatalf("expected a single header row, got %d occurrences of AD_GROUP_ID: %q", headerCount, agOut)
	}
	for _, want := range []string{"1001", "1002", "2001"} {
		if !strings.Contains(agOut, want) {
			t.Fatalf("adgroups ids output missing %q: %q", want, agOut)
		}
	}
}

func TestAdGroupsListStdinPipeFromCampaignsJSONDefaultOutput(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[{"id":101,"name":"C1","status":"ENABLED"}]}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/101/adgroups":
			return jsonResponse(`{"data":[{"id":1001,"name":"Group A","status":"ENABLED"}]}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	jsonOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status = ENABLED"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, jsonOut)
	}
	if !strings.Contains(jsonOut, `"id":101`) {
		t.Fatalf("expected json output with campaign id, got %q", jsonOut)
	}

	agOut, code := captureRun(t, []string{"adgroups", "list", "--campaign-id", "-", "-f", "ids"}, jsonOut)
	if code != ExitSuccess {
		t.Fatalf("adgroups list exit code = %d; output=%q", code, agOut)
	}
	if !strings.Contains(agOut, "CAMPAIGN_ID") || !strings.Contains(agOut, "AD_GROUP_ID") {
		t.Fatalf("expected ids header in output, got %q", agOut)
	}
	for _, want := range []string{"101", "1001"} {
		if !strings.Contains(agOut, want) {
			t.Fatalf("adgroups ids output missing %q: %q", want, agOut)
		}
	}
}

func TestAdGroupsList_FieldsIncludeSyntheticCampaignNameFromStdin(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[{"id":101,"name":"FitTrack","status":"ENABLED"}]}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/101/adgroups":
			return jsonResponse(`{"data":[{"id":1001,"name":"Brand Search","status":"ENABLED"}]}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	campaignOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED", "--fields", "id,name"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, campaignOut)
	}

	agOut, code := captureRun(t, []string{"adgroups", "list", "--campaign-id", "-", "--fields", "campaignName,name", "-f", "json"}, campaignOut)
	if code != ExitSuccess {
		t.Fatalf("adgroups list exit code = %d; output=%q", code, agOut)
	}
	if !strings.Contains(agOut, `"campaignName":"FitTrack"`) || !strings.Contains(agOut, `"name":"Brand Search"`) {
		t.Fatalf("unexpected output: %q", agOut)
	}
}

func TestAdGroupsList_SortUsesSyntheticCampaignNameFromStdinAfterMerge(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[{"id":101,"name":"Zeta","status":"ENABLED"},{"id":102,"name":"Alpha","status":"ENABLED"}]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/101/adgroups/find":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading adgroups find body: %v", err)
			}
			if bytes.Contains(body, []byte(`campaignName`)) {
				t.Fatalf("selector body should not send synthetic sort field: %s", body)
			}
			return jsonResponse(`{"data":[{"id":1001,"name":"Zeta Ad Group","status":"ENABLED"}]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/102/adgroups/find":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading adgroups find body: %v", err)
			}
			if bytes.Contains(body, []byte(`campaignName`)) {
				t.Fatalf("selector body should not send synthetic sort field: %s", body)
			}
			return jsonResponse(`{"data":[{"id":1002,"name":"Alpha Ad Group","status":"ENABLED"}]}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	campaignOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED", "--fields", "id,name"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, campaignOut)
	}

	agOut, code := captureRun(t, []string{
		"adgroups", "list",
		"--campaign-id", "-",
		"--sort", "campaignName:asc",
		"--fields", "campaignName,name",
		"-f", "json",
	}, campaignOut)
	if code != ExitSuccess {
		t.Fatalf("adgroups list exit code = %d; output=%q", code, agOut)
	}
	alpha := strings.Index(agOut, `"campaignName":"Alpha"`)
	zeta := strings.Index(agOut, `"campaignName":"Zeta"`)
	if alpha < 0 || zeta < 0 || alpha > zeta {
		t.Fatalf("expected merged rows sorted by campaignName asc, got %q", agOut)
	}
}

func TestAdGroupsList_SyntheticSortRejectsExplicitLimit(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API request for invalid local sort: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{"adgroups", "list", "--campaign-id", "101", "--sort", "campaignName:asc", "--limit", "1"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "requires fetching all rows") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestAdGroupsList_RemoteFilterUsesSyntheticCampaignNameFromStdin(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[{"id":101,"name":"Brand","status":"ENABLED"}]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/101/adgroups/find":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading adgroups find body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"name"`)) || !bytes.Contains(body, []byte(`"Brand"`)) {
				t.Fatalf("selector body missing resolved synthetic campaignName: %s", body)
			}
			return jsonResponse(`{"data":[{"id":1001,"campaignId":101,"name":"Brand Search","status":"ENABLED"}]}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	campaignOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED", "--fields", "id,name"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, campaignOut)
	}

	agOut, code := captureRun(t, []string{"adgroups", "list", "--campaign-id", "-", "--filter", "name CONTAINS campaignName", "-f", "json"}, campaignOut)
	if code != ExitSuccess {
		t.Fatalf("adgroups list exit code = %d; output=%q", code, agOut)
	}
	if !strings.Contains(agOut, `"name":"Brand Search"`) {
		t.Fatalf("unexpected output: %q", agOut)
	}
}

func TestReportsAdGroups_SortUsesSyntheticCampaignNameFromStdinAfterMerge(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[{"id":101,"name":"Zeta","status":"ENABLED"},{"id":102,"name":"Alpha","status":"ENABLED"}]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/reports/campaigns/101/adgroups":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading report body: %v", err)
			}
			if bytes.Contains(body, []byte(`campaignName`)) {
				t.Fatalf("report body should not send synthetic sort field: %s", body)
			}
			return jsonResponse(`{
				"data": {"reportingDataResponse": {"row": [{
					"metadata": {"campaignId": 101, "adGroupId": 5001, "adGroupName": "Zeta Group"},
					"total": {"impressions": 100, "localSpend": {"amount":"10.00","currency":"EUR"}}
				}]}}
			}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/reports/campaigns/102/adgroups":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading report body: %v", err)
			}
			if bytes.Contains(body, []byte(`campaignName`)) {
				t.Fatalf("report body should not send synthetic sort field: %s", body)
			}
			return jsonResponse(`{
				"data": {"reportingDataResponse": {"row": [{
					"metadata": {"campaignId": 102, "adGroupId": 5002, "adGroupName": "Alpha Group"},
					"total": {"impressions": 200, "localSpend": {"amount":"20.00","currency":"EUR"}}
				}]}}
			}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	campaignOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED", "--fields", "id,name"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, campaignOut)
	}

	reportOut, code := captureRun(t, []string{
		"reports", "adgroups",
		"--campaign-id", "-",
		"--start", "-7d",
		"--end", "now",
		"--sort", "campaignName:asc",
		"--fields", "CAMPAIGN_NAME,ADGROUP_NAME,IMPRESSIONS",
		"-f", "json",
	}, campaignOut)
	if code != ExitSuccess {
		t.Fatalf("reports adgroups exit code = %d; output=%q", code, reportOut)
	}
	alpha := strings.Index(reportOut, `"CAMPAIGN_NAME":"Alpha"`)
	zeta := strings.Index(reportOut, `"CAMPAIGN_NAME":"Zeta"`)
	if alpha < 0 || zeta < 0 || alpha > zeta {
		t.Fatalf("expected merged report rows sorted by campaignName asc, got %q", reportOut)
	}
}

func TestReportsAdGroups_LocalFilterUsesSyntheticDailyBudgetAmountFromStdin(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[{"id":101,"dailyBudgetAmount":{"amount":"50","currency":"EUR"},"status":"ENABLED"}]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/reports/campaigns/101/adgroups":
			return jsonResponse(`{
				"data": {
					"reportingDataResponse": {
						"row": [
							{
								"metadata": {"campaignId": 101, "adGroupId": 5001, "adGroupName": "Low Spend"},
								"total": {
									"impressions": 100,
									"localSpend": {"amount":"10.00","currency":"EUR"},
									"totalAvgCPI": {"amount":"0.50","currency":"EUR"}
								}
							},
							{
								"metadata": {"campaignId": 101, "adGroupId": 5002, "adGroupName": "High Spend"},
								"total": {
									"impressions": 200,
									"localSpend": {"amount":"60.00","currency":"EUR"},
									"totalAvgCPI": {"amount":"0.75","currency":"EUR"}
								}
							}
						]
					}
				}
			}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	campaignOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED", "--fields", "id,dailyBudgetAmount"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, campaignOut)
	}

	reportOut, code := captureRun(t, []string{
		"reports", "adgroups",
		"--campaign-id", "-",
		"--start", "-7d",
		"--end", "now",
		"--filter", "localSpend > dailyBudgetAmount",
		"--fields", "CAMPAIGN_ID,ADGROUP_ID,ADGROUP_NAME,IMPRESSIONS,LOCAL_SPEND,TOTAL_AVG_CPI",
		"-f", "json",
	}, campaignOut)
	if code != ExitSuccess {
		t.Fatalf("reports adgroups exit code = %d; output=%q", code, reportOut)
	}
	if strings.Contains(reportOut, "Low Spend") {
		t.Fatalf("low-spend row should have been filtered out: %q", reportOut)
	}
	if !strings.Contains(reportOut, "High Spend") || !strings.Contains(reportOut, `"LOCAL_SPEND":{"amount":"60.00","currency":"EUR"}`) {
		t.Fatalf("expected only high-spend row, got %q", reportOut)
	}
}

func TestReportsAdGroups_FieldsCampaignIDFromStdinWhenReportRowOmitsCampaignID(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[{"id":101,"status":"ENABLED"}]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/reports/campaigns/101/adgroups":
			return jsonResponse(`{
				"data": {
					"reportingDataResponse": {
						"row": [
							{
								"metadata": {"adGroupId": 5002, "adGroupName": "Scoped Row"},
								"total": {
									"impressions": 200,
									"localSpend": {"amount":"60.00","currency":"EUR"},
									"totalAvgCPI": {"amount":"0.75","currency":"EUR"}
								}
							}
						]
					}
				}
			}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	campaignOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED", "--fields", "id"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, campaignOut)
	}

	reportOut, code := captureRun(t, []string{
		"reports", "adgroups",
		"--campaign-id", "-",
		"--start", "-7d",
		"--end", "now",
		"--fields", "CAMPAIGN_ID,ADGROUP_ID,ADGROUP_NAME,IMPRESSIONS,LOCAL_SPEND,TOTAL_AVG_CPI",
		"-f", "json",
	}, campaignOut)
	if code != ExitSuccess {
		t.Fatalf("reports adgroups exit code = %d; output=%q", code, reportOut)
	}
	if !strings.Contains(reportOut, `"CAMPAIGN_ID":"101"`) || !strings.Contains(reportOut, `"ADGROUP_ID":5002`) {
		t.Fatalf("expected campaign id from stdin context, got %q", reportOut)
	}
}

func TestReportsAdGroups_PipeFormatCarriesCampaignNameThroughAdGroups(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[{"id":101,"name":"FitTrack","status":"ENABLED"}]}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/101/adgroups":
			return jsonResponse(`{"data":[{"id":1001,"name":"Brand Search","status":"ENABLED"}]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/reports/campaigns/101/adgroups":
			return jsonResponse(`{
				"data": {
					"reportingDataResponse": {
						"row": [
							{
								"metadata": {"campaignId": 101, "adGroupId": 1001, "adGroupName": "Brand Search"},
								"total": {
									"impressions": 100,
									"localSpend": {"amount":"10.00","currency":"EUR"},
									"totalAvgCPI": {"amount":"0.50","currency":"EUR"}
								}
							}
						]
					}
				}
			}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	campaignOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, campaignOut)
	}

	adgroupOut, code := captureRun(t, []string{"adgroups", "list", "--campaign-id", "-", "-f", "pipe"}, campaignOut)
	if code != ExitSuccess {
		t.Fatalf("adgroups list exit code = %d; output=%q", code, adgroupOut)
	}
	if !strings.Contains(adgroupOut, "CAMPAIGN_NAME") || !strings.Contains(adgroupOut, "FitTrack") {
		t.Fatalf("expected pipe output to include carried campaign name, got %q", adgroupOut)
	}

	reportOut, code := captureRun(t, []string{
		"reports", "adgroups",
		"--campaign-id", "-",
		"--adgroup-id", "-",
		"--start", "-7d",
		"--end", "now",
		"--fields", "campaignId,campaignName,adGroupId,adGroupName,impressions,localSpend,totalAvgCPI",
		"-f", "json",
	}, adgroupOut)
	if code != ExitSuccess {
		t.Fatalf("reports adgroups exit code = %d; output=%q", code, reportOut)
	}
	if !strings.Contains(reportOut, `"campaignName":"FitTrack"`) || !strings.Contains(reportOut, `"adGroupName":"Brand Search"`) {
		t.Fatalf("expected report output to retain campaign/ad group names from pipe context, got %q", reportOut)
	}
}

func TestReportsAdGroups_StdinPipelineGloballyResortsRemoteSortableFieldAfterMerge(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/find":
			return jsonResponse(`{"data":[
				{"id":101,"name":"Campaign A","status":"ENABLED"},
				{"id":102,"name":"Campaign B","status":"ENABLED"}
			]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/reports/campaigns/101/adgroups":
			return jsonResponse(`{
				"data": {
					"reportingDataResponse": {
						"row": [
							{
								"metadata": {"campaignId": 101, "adGroupId": 5001, "adGroupName": "A1"},
								"total": {
									"impressions": 100,
									"localSpend": {"amount":"10.00","currency":"EUR"},
									"totalAvgCPI": {"amount":"1.00","currency":"EUR"}
								}
							}
						]
					}
				}
			}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/reports/campaigns/102/adgroups":
			return jsonResponse(`{
				"data": {
					"reportingDataResponse": {
						"row": [
							{
								"metadata": {"campaignId": 102, "adGroupId": 5002, "adGroupName": "B1"},
								"total": {
									"impressions": 200,
									"localSpend": {"amount":"20.00","currency":"EUR"},
									"totalAvgCPI": {"amount":"2.00","currency":"EUR"}
								}
							}
						]
					}
				}
			}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	campaignOut, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED"}, "")
	if code != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d; output=%q", code, campaignOut)
	}

	reportOut, code := captureRun(t, []string{
		"reports", "adgroups",
		"--campaign-id", "-",
		"--start", "-7d",
		"--end", "now",
		"--sort", "localSpend:desc",
		"--fields", "campaignId,adGroupId,localSpend",
		"-f", "json",
	}, campaignOut)
	if code != ExitSuccess {
		t.Fatalf("reports adgroups exit code = %d; output=%q", code, reportOut)
	}

	high := strings.Index(reportOut, `"adGroupId":5002`)
	low := strings.Index(reportOut, `"adGroupId":5001`)
	if high < 0 || low < 0 || high > low {
		t.Fatalf("expected merged report rows sorted by localSpend desc, got %q", reportOut)
	}
}

// ---------------------------------------------------------------------------
// Campaigns create/update flag payloads
// ---------------------------------------------------------------------------

func TestCampaignsCreateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		if got["name"] != "My Campaign" {
			t.Fatalf("name = %v, want My Campaign", got["name"])
		}
		if got["adamId"].(float64) != 900001 {
			t.Fatalf("adamId = %v, want 900001", got["adamId"])
		}
		if got["adChannelType"] != "SEARCH" {
			t.Fatalf("adChannelType = %v, want SEARCH", got["adChannelType"])
		}
		if got["billingEvent"] != "TAPS" {
			t.Fatalf("billingEvent = %v, want TAPS", got["billingEvent"])
		}
		supplySources, ok := got["supplySources"].([]any)
		if !ok || len(supplySources) != 1 || supplySources[0] != "APPSTORE_SEARCH_RESULTS" {
			t.Fatalf("supplySources = %v, want [APPSTORE_SEARCH_RESULTS]", got["supplySources"])
		}
		daily := got["dailyBudgetAmount"].(map[string]any)
		if daily["amount"] != "5" || daily["currency"] != "USD" {
			t.Fatalf("dailyBudgetAmount = %v", daily)
		}
		regions, ok := got["countriesOrRegions"].([]any)
		if !ok || len(regions) == 0 || regions[0] != "US" {
			t.Fatalf("countriesOrRegions = %v", got["countriesOrRegions"])
		}
		return jsonResponse(`{"data":{"id":123}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "My Campaign",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "5",
		"--countries-or-regions", "US",
		"--ad-channel-type", "SEARCH",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestCampaignsCreateSupplySourcesFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		supplySources, ok := got["supplySources"].([]any)
		if !ok || len(supplySources) != 2 || supplySources[0] != "APPSTORE_SEARCH_RESULTS" || supplySources[1] != "APPSTORE_SEARCH_TAB" {
			t.Fatalf("supplySources = %v, want [APPSTORE_SEARCH_RESULTS APPSTORE_SEARCH_TAB]", got["supplySources"])
		}
		return jsonResponse(`{"data":{"id":123}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "My Campaign",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "5",
		"--countries-or-regions", "US",
		"--supply-sources", "APPSTORE_SEARCH_RESULTS,APPSTORE_SEARCH_TAB",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestCampaignsCreateLOCInvoiceDetailsFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		loc, ok := got["locInvoiceDetails"].(map[string]any)
		if !ok {
			t.Fatalf("locInvoiceDetails = %v", got["locInvoiceDetails"])
		}
		if loc["orderNumber"] != "PO-123" {
			t.Fatalf("orderNumber = %v, want PO-123", loc["orderNumber"])
		}
		if loc["buyerName"] != "Ada Buyer" {
			t.Fatalf("buyerName = %v, want Ada Buyer", loc["buyerName"])
		}
		return jsonResponse(`{"data":{"id":123}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "My Campaign",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "5",
		"--countries-or-regions", "US",
		"--ad-channel-type", "SEARCH",
		"--loc-invoice-details", `{"orderNumber":"PO-123","buyerName":"Ada Buyer"}`,
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestCampaignsCreateTargetCPAFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		cpa, ok := got["targetCpa"].(map[string]any)
		if !ok {
			t.Fatalf("targetCpa = %v", got["targetCpa"])
		}
		if cpa["amount"] != "10.50" || cpa["currency"] != "USD" {
			t.Fatalf("targetCpa = %v", cpa)
		}
		return jsonResponse(`{"data":{"id":123}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "My Campaign",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "5",
		"--target-cpa", "10.50",
		"--countries-or-regions", "US",
		"--ad-channel-type", "SEARCH",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestCampaignsCreateNameTemplate(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		if got["name"] != "FitTrack DE,FR SEARCH" {
			t.Fatalf("name = %v, want FitTrack DE,FR SEARCH", got["name"])
		}
		return jsonResponse(`{"data":{"id":123}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "FitTrack %(COUNTRIES_OR_REGIONS) %(adChannelType)",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "5",
		"--countries-or-regions", "de,fr",
		"--ad-channel-type", "SEARCH",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestCampaignsUpdateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPut, "/api/v5/campaigns/123", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		// campaigns update wraps body in {"campaign": ...}
		campaign, ok := got["campaign"].(map[string]any)
		if !ok {
			t.Fatalf("expected campaign wrapper, got %v", got)
		}
		if campaign["status"] != "PAUSED" {
			t.Fatalf("status = %v, want PAUSED", campaign["status"])
		}
		if campaign["name"] != "Renamed" {
			t.Fatalf("name = %v, want Renamed", campaign["name"])
		}
		budget := campaign["budgetAmount"].(map[string]any)
		if budget["amount"] != "100" || budget["currency"] != "USD" {
			t.Fatalf("budgetAmount = %v", budget)
		}
		return jsonResponse(`{"data":{"id":123}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "update",
		"--campaign-id", "123",
		"--status", "PAUSED",
		"--name", "Renamed",
		"--budget-amount", "100",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestCampaignsUpdateTargetCPAFlagPayload(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123":
			return jsonResponse(`{"data":{"id":123,"adChannelType":"SEARCH"}}`), nil
		case req.Method == http.MethodPut && req.URL.Path == "/api/v5/campaigns/123":
			body, _ := io.ReadAll(req.Body)
			got := mustUnmarshalMap(t, body)
			campaign, ok := got["campaign"].(map[string]any)
			if !ok {
				t.Fatalf("expected campaign wrapper, got %v", got)
			}
			cpa, ok := campaign["targetCpa"].(map[string]any)
			if !ok {
				t.Fatalf("targetCpa = %v", campaign["targetCpa"])
			}
			if cpa["amount"] != "12.25" || cpa["currency"] != "USD" {
				t.Fatalf("targetCpa = %v", cpa)
			}
			return jsonResponse(`{"data":{"id":123}}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "update",
		"--campaign-id", "123",
		"--target-cpa", "12.25",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestCampaignsCreateTargetCPARejectsDisplayCampaign(t *testing.T) {
	restore := shared.SetClientForTesting(newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	}), &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "Display Campaign",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "5",
		"--target-cpa", "10",
		"--countries-or-regions", "US",
		"--ad-channel-type", "DISPLAY",
	}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "targetCpa is supported only for SEARCH campaigns") {
		t.Fatalf("expected targetCpa validation error, got %q", out)
	}
}

func TestCampaignsUpdateTargetCPARejectsDisplayCampaign(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"DISPLAY"}}`), nil
		}
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "update",
		"--campaign-id", "123",
		"--target-cpa", "12.25",
	}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "targetCpa is supported only for SEARCH campaigns") {
		t.Fatalf("expected targetCpa validation error, got %q", out)
	}
}

func TestCampaignsUpdateLOCInvoiceDetailsFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPut, "/api/v5/campaigns/123", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		campaign, ok := got["campaign"].(map[string]any)
		if !ok {
			t.Fatalf("expected campaign wrapper, got %v", got)
		}
		loc, ok := campaign["locInvoiceDetails"].(map[string]any)
		if !ok {
			t.Fatalf("locInvoiceDetails = %v", campaign["locInvoiceDetails"])
		}
		if loc["orderNumber"] != "PO-123" {
			t.Fatalf("orderNumber = %v, want PO-123", loc["orderNumber"])
		}
		if loc["buyerName"] != "Ada Buyer" {
			t.Fatalf("buyerName = %v, want Ada Buyer", loc["buyerName"])
		}
		return jsonResponse(`{"data":{"id":123}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "update",
		"--campaign-id", "123",
		"--loc-invoice-details", `{"orderNumber":"PO-123","buyerName":"Ada Buyer"}`,
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

// ---------------------------------------------------------------------------
// Ad groups create/update flag payloads
// ---------------------------------------------------------------------------

func TestAdGroupsCreateFlagPayload(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123":
			return jsonResponse(`{"data":{"id":123,"adChannelType":"SEARCH"}}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns/123/adgroups":
			body, _ := io.ReadAll(req.Body)
			got := mustUnmarshalMap(t, body)
			if got["name"] != "My Ad Group" {
				t.Fatalf("name = %v, want My Ad Group", got["name"])
			}
			bid := got["defaultBidAmount"].(map[string]any)
			if bid["amount"] != "1.25" || bid["currency"] != "USD" {
				t.Fatalf("defaultBidAmount = %v", bid)
			}
			if cpa, ok := got["cpaGoal"].(map[string]any); ok {
				if cpa["amount"] != "2.50" {
					t.Fatalf("cpaGoal = %v", cpa)
				}
			} else {
				t.Fatal("expected cpaGoal")
			}
			if got["startTime"] != "2026-03-25T09:30:00.000" {
				t.Fatalf("startTime = %v, want 2026-03-25T09:30:00.000", got["startTime"])
			}
			return jsonResponse(`{"data":{"id":456}}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:            "123",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "UTC",
		DefaultTimeOfDay: "09:30:00",
	})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 15, 4, 5, 0, time.UTC)
	})
	defer restoreNow()

	out, code := captureRun(t, []string{
		"adgroups", "create",
		"--campaign-id", "123",
		"--name", "My Ad Group",
		"--default-bid", "1.25",
		"--cpa-goal", "2.50",
		"--country-code", "US",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestAdGroupsCreateNameTemplate(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/adgroups", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		if got["name"] != "Search Group PAUSED" {
			t.Fatalf("name = %v, want Search Group PAUSED", got["name"])
		}
		return jsonResponse(`{"data":{"id":789}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "create",
		"--campaign-id", "123",
		"--name", "Search Group %(STATUS)",
		"--default-bid", "1.25",
		"--status", "paused",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestAdGroupsCreateFromJSONPreservesLargeIntegers(t *testing.T) {
	const body = `{"name":"My Ad Group","defaultBidAmount":{"amount":"1.25","currency":"USD"},"targetingDimensions":{"age":{"included":[{"minAge":9007199254740993,"maxAge":9007199254740995}]}}}`
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/adgroups", func(reqBody []byte) *http.Response {
		got := string(reqBody)
		if !strings.Contains(got, `"minAge":9007199254740993`) {
			t.Fatalf("request body lost minAge precision: %s", got)
		}
		if !strings.Contains(got, `"maxAge":9007199254740995`) {
			t.Fatalf("request body lost maxAge precision: %s", got)
		}
		return jsonResponse(`{"data":{"id":456}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:            "123",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "UTC",
		DefaultTimeOfDay: "09:30:00",
	})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 15, 4, 5, 0, time.UTC)
	})
	defer restoreNow()

	out, code := captureRun(t, []string{
		"adgroups", "create",
		"--campaign-id", "123",
		"--from-json", body,
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestAdGroupsCreateFromJSONRejectsNonStringStartTime(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("request should not be sent for malformed startTime")
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	tests := []struct {
		name      string
		startTime string
	}{
		{name: "number", startTime: `123`},
		{name: "object", startTime: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"name":"My Ad Group","defaultBidAmount":{"amount":"1.25","currency":"USD"},"startTime":` + tt.startTime + `}`
			out, code := captureRun(t, []string{
				"adgroups", "create",
				"--campaign-id", "123",
				"--from-json", body,
			}, "")
			if code != ExitUsage {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
			}
			if !strings.Contains(out, "startTime must be a string") {
				t.Fatalf("expected startTime validation error, got %q", out)
			}
		})
	}
}

func TestAdGroupsUpdateFlagPayload(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123":
			return jsonResponse(`{"data":{"id":123,"adChannelType":"SEARCH"}}`), nil
		case req.Method == http.MethodPut && req.URL.Path == "/api/v5/campaigns/123/adgroups/456":
			body, _ := io.ReadAll(req.Body)
			got := mustUnmarshalMap(t, body)
			if got["status"] != "ENABLED" {
				t.Fatalf("status = %v, want ENABLED", got["status"])
			}
			if got["name"] != "AG Name" {
				t.Fatalf("name = %v, want AG Name", got["name"])
			}
			bid := got["defaultBidAmount"].(map[string]any)
			if bid["amount"] != "1.50" {
				t.Fatalf("defaultBidAmount = %v", bid)
			}
			cpa := got["cpaGoal"].(map[string]any)
			if cpa["amount"] != "2.25" {
				t.Fatalf("cpaGoal = %v", cpa)
			}
			return jsonResponse(`{"data":{"id":456}}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--status", "ENABLED",
		"--default-bid", "1.50",
		"--cpa-goal", "2.25",
		"--name", "AG Name",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestAdGroupsUpdateCPAGoalRejectsNonSearchCampaign(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"DISPLAY"}}`), nil
		}
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--cpa-goal", "2.25",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for CPA goal on non-SEARCH campaign; output=%q", out)
	}
	if !strings.Contains(out, "SEARCH") {
		t.Fatalf("expected error mentioning SEARCH, got %q", out)
	}
	if !strings.Contains(out, "--cpa-goal") {
		t.Fatalf("expected error to cite the --cpa-goal flag, got %q", out)
	}
}

// TestAdGroupsUpdateCPAGoalCheckAnnouncesFetch ensures --cpa-goal with --check
// both fetches the campaign to verify SEARCH and announces that fetch in
// readOnlyChecks.
func TestAdGroupsUpdateCPAGoalCheckAnnouncesFetch(t *testing.T) {
	requests := 0
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"SEARCH"}}`), nil
		}
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--cpa-goal", "2.25",
		"--check",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 1 {
		t.Fatalf("expected exactly one campaign GET, got %d", requests)
	}
	got := mustUnmarshalMap(t, []byte(out))
	checks, ok := got["readOnlyChecks"].([]any)
	if !ok {
		t.Fatalf("readOnlyChecks missing or wrong type: %v", got["readOnlyChecks"])
	}
	found := false
	for _, c := range checks {
		if s, _ := c.(string); s == "fetched campaign to verify SEARCH channel" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected readOnlyChecks to announce SEARCH-channel fetch; got %v", checks)
	}
}

// TestAdGroupsCreateCPAGoalFromJSONCheckAnnouncesFetch ensures cpaGoal supplied
// via --from-json (not the --cpa-goal flag) still triggers the campaign GET
// and is announced in --check readOnlyChecks.
func TestAdGroupsCreateCPAGoalFromJSONCheckAnnouncesFetch(t *testing.T) {
	requests := 0
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"SEARCH"}}`), nil
		}
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:            "123",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "UTC",
		DefaultTimeOfDay: "09:30:00",
	})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 15, 4, 5, 0, time.UTC)
	})
	defer restoreNow()

	body := `{"name":"x","defaultBidAmount":{"amount":"1.50","currency":"USD"},"cpaGoal":{"amount":"2.00","currency":"USD"}}`
	out, code := captureRun(t, []string{
		"adgroups", "create",
		"--campaign-id", "123",
		"--from-json", body,
		"--check",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 1 {
		t.Fatalf("expected exactly one campaign GET, got %d", requests)
	}
	got := mustUnmarshalMap(t, []byte(out))
	checks, ok := got["readOnlyChecks"].([]any)
	if !ok {
		t.Fatalf("readOnlyChecks missing or wrong type: %v", got["readOnlyChecks"])
	}
	found := false
	for _, c := range checks {
		if s, _ := c.(string); s == "fetched campaign to verify SEARCH channel" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected readOnlyChecks to announce SEARCH-channel fetch; got %v", checks)
	}
}

func TestAdGroupsCreateCPAGoalFromJSONUsesJSONLabelOnNonSearchCampaign(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"DISPLAY"}}`), nil
		}
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	body := `{"name":"x","defaultBidAmount":{"amount":"1.50","currency":"USD"},"cpaGoal":{"amount":"2.00","currency":"USD"}}`
	out, code := captureRun(t, []string{
		"adgroups", "create",
		"--campaign-id", "123",
		"--from-json", body,
		"--check",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for cpaGoal on non-SEARCH campaign; output=%q", out)
	}
	if !strings.Contains(out, "cpaGoal requires a SEARCH campaign") {
		t.Fatalf("expected JSON label in error, got %q", out)
	}
	if strings.Contains(out, "--cpa-goal") {
		t.Fatalf("did not expect --cpa-goal flag label, got %q", out)
	}
}

// TestAdGroupsUpdateCPAGoalFromJSONCheckAnnouncesFetch ensures cpaGoal supplied
// via --from-json (not the --cpa-goal flag) still triggers the campaign GET
// and is announced in --check readOnlyChecks.
func TestAdGroupsUpdateCPAGoalFromJSONCheckAnnouncesFetch(t *testing.T) {
	requests := 0
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"SEARCH"}}`), nil
		}
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	body := `{"cpaGoal":{"amount":"2.00","currency":"USD"}}`
	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--from-json", body,
		"--check",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 1 {
		t.Fatalf("expected exactly one campaign GET, got %d", requests)
	}
	got := mustUnmarshalMap(t, []byte(out))
	checks, ok := got["readOnlyChecks"].([]any)
	if !ok {
		t.Fatalf("readOnlyChecks missing or wrong type: %v", got["readOnlyChecks"])
	}
	found := false
	for _, c := range checks {
		if s, _ := c.(string); s == "fetched campaign to verify SEARCH channel" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected readOnlyChecks to announce SEARCH-channel fetch; got %v", checks)
	}
}

func TestAdGroupsUpdateCPAGoalFromJSONUsesJSONLabelOnNonSearchCampaign(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"DISPLAY"}}`), nil
		}
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	body := `{"cpaGoal":{"amount":"2.00","currency":"USD"}}`
	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--from-json", body,
		"--check",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for cpaGoal on non-SEARCH campaign; output=%q", out)
	}
	if !strings.Contains(out, "cpaGoal requires a SEARCH campaign") {
		t.Fatalf("expected JSON label in error, got %q", out)
	}
	if strings.Contains(out, "--cpa-goal") {
		t.Fatalf("did not expect --cpa-goal flag label, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Keywords create/update flag payloads
// ---------------------------------------------------------------------------

func TestKeywordsCreateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/adgroups/456/targetingkeywords/bulk", func(body []byte) *http.Response {
		items := mustUnmarshalSlice(t, body)
		if len(items) != 2 {
			t.Fatalf("expected 2 keywords, got %d: %s", len(items), body)
		}
		kw0 := items[0].(map[string]any)
		if kw0["text"] != "hello" || kw0["matchType"] != "EXACT" {
			t.Fatalf("keyword[0] = %v", kw0)
		}
		if bid, ok := kw0["bidAmount"].(map[string]any); ok {
			if bid["amount"] != "0.75" {
				t.Fatalf("bidAmount = %v", bid)
			}
		} else {
			t.Fatal("expected bidAmount")
		}
		kw1 := items[1].(map[string]any)
		if kw1["text"] != "world" {
			t.Fatalf("keyword[1] text = %v", kw1["text"])
		}
		return jsonResponse(`{"data":[{"id":789},{"id":790}]}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "create",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--text", "hello,world",
		"--bid", "0.75",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestKeywordsUpdateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPut, "/api/v5/campaigns/123/adgroups/456/targetingkeywords/bulk", func(body []byte) *http.Response {
		items := mustUnmarshalSlice(t, body)
		if len(items) != 1 {
			t.Fatalf("expected 1 keyword, got %d: %s", len(items), body)
		}
		kw := items[0].(map[string]any)
		if kw["status"] != "PAUSED" {
			t.Fatalf("status = %v, want PAUSED", kw["status"])
		}
		if bid, ok := kw["bidAmount"].(map[string]any); ok {
			if bid["amount"] != "0.75" {
				t.Fatalf("bidAmount = %v", bid)
			}
		} else {
			t.Fatal("expected bidAmount")
		}
		return jsonResponse(`{"data":[{"id":789}]}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--keyword-id", "789",
		"--status", "PAUSED",
		"--bid", "0.75",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestKeywordsCreateRejectsInvalidMatchTypeEvenWhenTextParsesEmpty(t *testing.T) {
	restore := shared.SetClientForTesting(newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	}), &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "create",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--text", ",,,",
		"--match-type", "nope",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for invalid match type; output=%q", out)
	}
	if !strings.Contains(out, "invalid match type") {
		t.Fatalf("expected invalid-match-type error, got %q", out)
	}
}

func TestKeywordsCreateRejectsExplicitEmptyMatchType(t *testing.T) {
	restore := shared.SetClientForTesting(newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	}), &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "create",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--text", "fitness coach",
		"--match-type", "",
		"--check",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for empty match type; output=%q", out)
	}
	if !strings.Contains(out, "invalid match type") {
		t.Fatalf("expected invalid-match-type error, got %q", out)
	}
}

func TestKeywordsCreateRejectsEmptyTextListWithValidMatchType(t *testing.T) {
	restore := shared.SetClientForTesting(newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	}), &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "create",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--text", ",,,",
		"--match-type", "EXACT",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for empty keyword list; output=%q", out)
	}
	if !strings.Contains(out, "--text") {
		t.Fatalf("expected error mentioning --text, got %q", out)
	}
}

// TestFromJSONRejectsShortcutFlags ensures shortcut flags can't silently
// coexist with --from-json across the four mutation commands. The JSON path
// ignores shortcut flags, so accepting the combination would mask user error.
func TestFromJSONRejectsShortcutFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		conflict string
	}{
		{
			name: "adgroups create + --status",
			args: []string{
				"adgroups", "create",
				"--campaign-id", "123",
				"--from-json", `{"name":"x","defaultBidAmount":{"amount":"1.50","currency":"USD"}}`,
				"--status", "PAUSED",
			},
			conflict: "--status",
		},
		{
			name: "adgroups create + --cpa-goal",
			args: []string{
				"adgroups", "create",
				"--campaign-id", "123",
				"--from-json", `{"name":"x","defaultBidAmount":{"amount":"1.50","currency":"USD"}}`,
				"--cpa-goal", "2.00",
			},
			conflict: "--cpa-goal",
		},
		{
			name: "adgroups update + --default-bid (no --merge)",
			args: []string{
				"adgroups", "update",
				"--campaign-id", "123",
				"--adgroup-id", "456",
				"--from-json", `{"status":"PAUSED"}`,
				"--default-bid", "1.50",
			},
			conflict: "--default-bid",
		},
		{
			name: "keywords create + --match-type",
			args: []string{
				"keywords", "create",
				"--campaign-id", "123",
				"--adgroup-id", "456",
				"--from-json", `[{"text":"x","matchType":"EXACT"}]`,
				"--match-type", "BROAD",
			},
			conflict: "--match-type",
		},
		{
			name: "keywords update + --bid",
			args: []string{
				"keywords", "update",
				"--campaign-id", "123",
				"--adgroup-id", "456",
				"--keyword-id", "789",
				"--from-json", `[{"id":789,"status":"PAUSED"}]`,
				"--bid", "1.50",
			},
			conflict: "--bid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := shared.SetClientForTesting(newTestClient(func(req *http.Request) (*http.Response, error) {
				t.Fatalf("no request should be sent on flag-conflict rejection: %s %s", req.Method, req.URL.Path)
				return nil, nil
			}), &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
			defer restore()

			out, code := captureRun(t, tt.args, "")
			if code != ExitUsage {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
			}
			if !strings.Contains(out, tt.conflict) {
				t.Fatalf("expected error to cite %s, got %q", tt.conflict, out)
			}
			if !strings.Contains(out, "--from-json") {
				t.Fatalf("expected error to mention --from-json, got %q", out)
			}
		})
	}
}

// TestAdGroupsUpdateMergeAllowsFromJSONWithShortcuts confirms --merge keeps
// the documented overlay behavior: shortcut flags layer on top of --from-json.
func TestAdGroupsUpdateMergeAllowsFromJSONWithShortcuts(t *testing.T) {
	requests := 0
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		requests++
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123/adgroups/456":
			return jsonResponse(`{"data":{"id":456,"name":"original","defaultBidAmount":{"amount":"1.00","currency":"USD"}}}`), nil
		case req.Method == http.MethodPut && req.URL.Path == "/api/v5/campaigns/123/adgroups/456":
			return jsonResponse(`{"data":{"id":456}}`), nil
		}
		t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--merge",
		"--from-json", `{"name":"from-json"}`,
		"--status", "PAUSED",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestAdGroupsUpdateInheritedBids(t *testing.T) {
	requests := 0
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		requests++
		switch requests {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/api/v5/campaigns/123/adgroups/456" {
				t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			}
			return jsonResponse(`{"data":{"id":456,"defaultBidAmount":{"amount":"1.00","currency":"USD"}}}`), nil
		case 2:
			if req.Method != http.MethodGet || req.URL.Path != "/api/v5/campaigns/123/adgroups/456/targetingkeywords" {
				t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			}
			if got := req.URL.Query().Get("limit"); got != "1000" {
				t.Fatalf("limit = %q, want 1000", got)
			}
			return jsonResponse(`{"data":[
				{"id":101,"bidAmount":{"amount":"1.00","currency":"USD"}},
				{"id":102,"bidAmount":{"amount":"1.25","currency":"USD"}},
				{"id":103}
			],"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":1000}}`), nil
		case 3:
			if req.Method != http.MethodPut || req.URL.Path != "/api/v5/campaigns/123/adgroups/456" {
				t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			}
			body, _ := io.ReadAll(req.Body)
			got := mustUnmarshalMap(t, body)
			bid := got["defaultBidAmount"].(map[string]any)
			if bid["amount"] != "1.10" || bid["currency"] != "USD" {
				t.Fatalf("defaultBidAmount = %v", bid)
			}
			return jsonResponse(`{"data":{"id":456}}`), nil
		case 4:
			if req.Method != http.MethodPut || req.URL.Path != "/api/v5/campaigns/123/adgroups/456/targetingkeywords/bulk" {
				t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			}
			body, _ := io.ReadAll(req.Body)
			items := mustUnmarshalSlice(t, body)
			if len(items) != 1 {
				t.Fatalf("expected 1 keyword update, got %d: %s", len(items), body)
			}
			item := items[0].(map[string]any)
			if item["id"].(float64) != 101 {
				t.Fatalf("id = %v, want 101", item["id"])
			}
			bid := item["bidAmount"].(map[string]any)
			if bid["amount"] != "1.10" || bid["currency"] != "USD" {
				t.Fatalf("bidAmount = %v", bid)
			}
			return jsonResponse(`{"data":[{"id":101}]}`), nil
		default:
			t.Fatalf("unexpected extra request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--default-bid", "+10%",
		"--update-inherited-bids",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 4 {
		t.Fatalf("request count = %d, want 4", requests)
	}
}

func TestAdGroupsUpdateInheritedBidsCheck(t *testing.T) {
	requests := 0
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		requests++
		switch requests {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/api/v5/campaigns/123/adgroups/456" {
				t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			}
			return jsonResponse(`{"data":{"id":456,"defaultBidAmount":{"amount":"1.00","currency":"USD"}}}`), nil
		case 2:
			if req.Method != http.MethodGet || req.URL.Path != "/api/v5/campaigns/123/adgroups/456/targetingkeywords" {
				t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			}
			return jsonResponse(`{"data":[
				{"id":101,"bidAmount":{"amount":"1.00","currency":"USD"}},
				{"id":102,"bidAmount":{"amount":"1.00","currency":"USD"}}
			],"pagination":{"totalResults":2,"startIndex":0,"itemsPerPage":1000}}`), nil
		default:
			t.Fatalf("unexpected extra request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--default-bid", "+10%",
		"--update-inherited-bids",
		"--check",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want 2", requests)
	}

	got := mustUnmarshalMap(t, []byte(out))
	if got["wouldAffect"] != "3 objects" {
		t.Fatalf("wouldAffect = %v, want 3 objects", got["wouldAffect"])
	}
	checks, ok := got["readOnlyChecks"].([]any)
	if !ok || len(checks) != 2 {
		t.Fatalf("readOnlyChecks = %v", got["readOnlyChecks"])
	}
}

// ---------------------------------------------------------------------------
// Ads create/update flag payloads
// ---------------------------------------------------------------------------

func TestAdsCreateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/adgroups/456/ads", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		if got["name"] != "My Ad" {
			t.Fatalf("name = %v, want My Ad", got["name"])
		}
		if got["creativeId"].(float64) != 789 {
			t.Fatalf("creativeId = %v, want 789", got["creativeId"])
		}
		return jsonResponse(`{"data":{"id":789}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"ads", "create",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--creative-id", "789",
		"--name", "My Ad",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestAdsUpdateStatusFlag(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPut, "/api/v5/campaigns/123/adgroups/456/ads/789", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		if got["status"] != "PAUSED" {
			t.Fatalf("status = %v, want PAUSED", got["status"])
		}
		return jsonResponse(`{"data":{"id":789}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"ads", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--ad-id", "789",
		"--status", "PAUSED",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestAdsUpdateNumericStatusAlias(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPut, "/api/v5/campaigns/123/adgroups/456/ads/789", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		if got["status"] != "ENABLED" {
			t.Fatalf("status = %v, want ENABLED (from alias 1)", got["status"])
		}
		return jsonResponse(`{"data":{"id":789}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"ads", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--ad-id", "789",
		"--status", "1",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

// ---------------------------------------------------------------------------
// Negatives create/update flag payloads
// ---------------------------------------------------------------------------

func TestNegativesCreateCampaignLevelFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/negativekeywords/bulk", func(body []byte) *http.Response {
		items := mustUnmarshalSlice(t, body)
		if len(items) != 2 {
			t.Fatalf("expected 2 negatives, got %d: %s", len(items), body)
		}
		kw0 := items[0].(map[string]any)
		if kw0["text"] != "brand one" || kw0["matchType"] != "EXACT" {
			t.Fatalf("negative[0] = %v", kw0)
		}
		kw1 := items[1].(map[string]any)
		if kw1["text"] != "brand two" {
			t.Fatalf("negative[1] = %v", kw1)
		}
		return jsonResponse(`{"data":[{"id":99}]}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "create",
		"--campaign-id", "123",
		"--text", `"brand one","brand two"`,
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestNegativesCreateAdGroupLevelFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/adgroups/456/negativekeywords/bulk", func(body []byte) *http.Response {
		items := mustUnmarshalSlice(t, body)
		if len(items) != 1 {
			t.Fatalf("expected 1 negative, got %d: %s", len(items), body)
		}
		kw := items[0].(map[string]any)
		if kw["text"] != "blocked" {
			t.Fatalf("text = %v, want blocked", kw["text"])
		}
		return jsonResponse(`{"data":[{"id":77}]}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "create",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--text", "blocked",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestNegativesUpdateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPut, "/api/v5/campaigns/123/negativekeywords/bulk", func(body []byte) *http.Response {
		items := mustUnmarshalSlice(t, body)
		if len(items) != 1 {
			t.Fatalf("expected 1 negative, got %d: %s", len(items), body)
		}
		kw := items[0].(map[string]any)
		if kw["status"] != "PAUSED" {
			t.Fatalf("status = %v, want PAUSED", kw["status"])
		}
		return jsonResponse(`{"data":[{"id":99}]}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "update",
		"--campaign-id", "123",
		"--keyword-id", "99",
		"--status", "PAUSED",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestNegativesUpdateZeroStatusAlias(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPut, "/api/v5/campaigns/123/negativekeywords/bulk", func(body []byte) *http.Response {
		items := mustUnmarshalSlice(t, body)
		if len(items) != 1 {
			t.Fatalf("expected 1 negative: %s", body)
		}
		kw := items[0].(map[string]any)
		if kw["status"] != "PAUSED" {
			t.Fatalf("status = %v, want PAUSED (from alias 0)", kw["status"])
		}
		return jsonResponse(`{"data":[{"id":99}]}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "update",
		"--campaign-id", "123",
		"--keyword-id", "99",
		"--status", "0",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

// ---------------------------------------------------------------------------
// Budget orders create/update flag payloads
// ---------------------------------------------------------------------------

func TestBudgetOrdersCreateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/budgetorders", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		orgIds := got["orgIds"].([]any)
		if len(orgIds) != 1 || orgIds[0].(float64) != 123 {
			t.Fatalf("orgIds = %v, want [123]", orgIds)
		}
		bo := got["bo"].(map[string]any)
		if bo["name"] != "BO Name" {
			t.Fatalf("name = %v, want BO Name", bo["name"])
		}
		if bo["orderNumber"] != "PO-1" {
			t.Fatalf("orderNumber = %v, want PO-1", bo["orderNumber"])
		}
		if bo["primaryBuyerEmail"] != "buyer@example.com" {
			t.Fatalf("primaryBuyerEmail = %v", bo["primaryBuyerEmail"])
		}
		return jsonResponse(`{"data":{"id":321}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"budgetorders", "create",
		"--name", "BO Name",
		"--budget-amount", "400",
		"--order-number", "PO-1",
		"--primary-buyer-email", "buyer@example.com",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestBudgetOrdersUpdateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPut, "/api/v5/budgetorders/321", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		bo := got["bo"].(map[string]any)
		if bo["name"] != "BO Updated" {
			t.Fatalf("name = %v, want BO Updated", bo["name"])
		}
		budget := bo["budget"].(map[string]any)
		if budget["amount"] != "500" {
			t.Fatalf("budget amount = %v, want 500", budget["amount"])
		}
		return jsonResponse(`{"data":{"id":321}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"budgetorders", "update",
		"--budget-order-id", "321",
		"--name", "BO Updated",
		"--budget-amount", "500",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

// ---------------------------------------------------------------------------
// Creatives create flag payload
// ---------------------------------------------------------------------------

func TestCreativesCreateFlagPayload(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/creatives", func(body []byte) *http.Response {
		got := mustUnmarshalMap(t, body)
		if got["adamId"].(float64) != 456 {
			t.Fatalf("adamId = %v, want 456", got["adamId"])
		}
		if got["name"] != "Creative 1" {
			t.Fatalf("name = %v, want Creative 1", got["name"])
		}
		if got["type"] != "CUSTOM_PRODUCT_PAGE" {
			t.Fatalf("type = %v, want CUSTOM_PRODUCT_PAGE", got["type"])
		}
		return jsonResponse(`{"data":{"id":55}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"creatives", "create",
		"--adam-id", "456",
		"--name", "Creative 1",
		"--type", "custom_product_page",
		"--product-page-id", "pp1",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

// ---------------------------------------------------------------------------
// Safety limit tests
// ---------------------------------------------------------------------------

func TestCampaignsCreateDailyBudgetSafetyLimit(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call despite safety limit: %s", req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "USD",
		MaxDailyBudget:  config.DecimalText("10"),
	})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "Over Limit",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "15",
		"--countries-or-regions", "US",
		"--ad-channel-type", "SEARCH",
	}, "")
	if code != ExitSafetyLimit {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSafetyLimit, out)
	}
	if !strings.Contains(out, "exceeds limit") {
		t.Fatalf("expected safety limit error, got %q", out)
	}
}

func TestCampaignsCreateTargetCPASafetyLimit(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("request should not be sent when target CPA exceeds safety limit")
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "USD",
		MaxCPAGoal:      config.DecimalText("10"),
	})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "Over CPA Limit",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "15",
		"--target-cpa", "15",
		"--countries-or-regions", "US",
		"--ad-channel-type", "SEARCH",
	}, "")
	if code != ExitSafetyLimit {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSafetyLimit, out)
	}
	if !strings.Contains(out, "exceeds limit") {
		t.Fatalf("expected safety limit error, got %q", out)
	}
}

func TestCampaignsCreateDailyBudgetSafetyLimitForceBypass(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		called = true
		return jsonResponse(`{"data":{"id":123}}`), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "USD",
		MaxDailyBudget:  config.DecimalText("10"),
	})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "Force Bypass",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "15",
		"--countries-or-regions", "US",
		"--ad-channel-type", "SEARCH",
		"--force",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("--force should bypass safety limit; API not called")
	}
}

func TestCampaignsCreateStartTimeUsesProfileDefaults(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns", func(body []byte) *http.Response {
		payload := mustUnmarshalMap(t, body)
		if got := payload["startTime"]; got != "2026-03-30T09:30:00.000" {
			t.Fatalf("startTime = %v, want %q", got, "2026-03-30T09:30:00.000")
		}
		return jsonResponse(`{"data":{"id":123}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:            "123",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "America/New_York",
		DefaultTimeOfDay: "09:30",
	})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 15, 4, 5, 0, time.UTC)
	})
	defer restoreNow()

	out, code := captureRun(t, []string{
		"campaigns", "create",
		"--name", "Timed Campaign",
		"--adam-id", testAdamID,
		"--daily-budget-amount", "15 USD",
		"--countries-or-regions", "US",
		"--ad-channel-type", "SEARCH",
		"--start-time", "+5d",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestReportsCampaignsUTCUsesUTCDayBoundary(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/reports/campaigns", func(body []byte) *http.Response {
		payload := mustUnmarshalMap(t, body)
		if got := payload["startTime"]; got != "2026-03-26" {
			t.Fatalf("startTime = %v, want %q", got, "2026-03-26")
		}
		if got := payload["endTime"]; got != "2026-03-26" {
			t.Fatalf("endTime = %v, want %q", got, "2026-03-26")
		}
		return jsonResponse(`{"data":{"reportingDataResponse":{"row":[]}}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultTimezone: "America/Los_Angeles",
	})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		loc := time.FixedZone("UTC-7", -7*60*60)
		return time.Date(2026, time.March, 25, 23, 30, 0, 0, loc)
	})
	defer restoreNow()

	out, code := captureRun(t, []string{
		"reports", "campaigns",
		"--start", "now",
		"--end", "now",
		"--timezone", "UTC",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestImpressionShareCreateShortcutFlags(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/custom-reports", func(body []byte) *http.Response {
		payload := mustUnmarshalMap(t, body)
		if got := payload["name"]; got != "Weekly Share Report" {
			t.Fatalf("name = %v, want %q", got, "Weekly Share Report")
		}
		if _, ok := payload["dateRange"]; ok {
			t.Fatalf("dateRange unexpectedly set: %v", payload["dateRange"])
		}
		if got := payload["granularity"]; got != "DAILY" {
			t.Fatalf("granularity = %v, want %q", got, "DAILY")
		}
		if got := payload["startTime"]; got != "2026-03-20" {
			t.Fatalf("startTime = %v, want %q", got, "2026-03-20")
		}
		if got := payload["endTime"]; got != "2026-03-27" {
			t.Fatalf("endTime = %v, want %q", got, "2026-03-27")
		}
		return jsonResponse(`{"data":{"id":900901}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 27, 10, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	out, code := captureRun(t, []string{
		"impression-share", "create",
		"--name", "Weekly Share Report",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestImpressionShareCreateCustomDateRangeShortcutFlags(t *testing.T) {
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/custom-reports", func(body []byte) *http.Response {
		payload := mustUnmarshalMap(t, body)
		if got := payload["name"]; got != "Custom Share Report" {
			t.Fatalf("name = %v, want %q", got, "Custom Share Report")
		}
		if got := payload["dateRange"]; got != "CUSTOM" {
			t.Fatalf("dateRange = %v, want %q", got, "CUSTOM")
		}
		if got := payload["startTime"]; got != "2026-03-20" {
			t.Fatalf("startTime = %v, want %q", got, "2026-03-20")
		}
		if got := payload["endTime"]; got != "2026-03-27" {
			t.Fatalf("endTime = %v, want %q", got, "2026-03-27")
		}
		if got := payload["granularity"]; got != "WEEKLY" {
			t.Fatalf("granularity = %v, want %q", got, "WEEKLY")
		}
		return jsonResponse(`{"data":{"id":900902}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 27, 10, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	out, code := captureRun(t, []string{
		"impression-share", "create",
		"--name", "Custom Share Report",
		"--dateRange", "custom",
		"--startTime", "-1w",
		"--endTime", "now",
		"--granularity", "weekly",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestAdGroupsUpdateCPAGoalSafetyLimit(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"SEARCH"}}`), nil
		}
		t.Fatalf("unexpected API call: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "USD",
		MaxCPAGoal:      config.DecimalText("2"),
	})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--cpa-goal", "2.25",
	}, "")
	if code != ExitSafetyLimit {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSafetyLimit, out)
	}
	if !strings.Contains(out, "exceeds limit") {
		t.Fatalf("expected CPA goal safety limit error, got %q", out)
	}
}

func TestAdGroupsCreateBidSafetyLimit(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/api/v5/campaigns/123" {
			return jsonResponse(`{"data":{"id":123,"adChannelType":"SEARCH"}}`), nil
		}
		t.Fatalf("unexpected API call despite safety limit: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "USD",
		MaxBid:          config.DecimalText("1"),
	})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "create",
		"--campaign-id", "123",
		"--name", "Over Limit AG",
		"--default-bid", "1.25",
	}, "")
	if code != ExitSafetyLimit {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSafetyLimit, out)
	}
	if !strings.Contains(out, "exceeds limit") {
		t.Fatalf("expected bid safety limit error, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Currency requirement
// ---------------------------------------------------------------------------

func TestCampaignsUpdateRequiresCurrencyWhenNoDefault(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "", // no default
	})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "update",
		"--campaign-id", "123",
		"--budget-amount", "100",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for missing currency; output=%q", out)
	}
	if !strings.Contains(out, "currency") {
		t.Fatalf("expected error mentioning currency, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Delete confirmation
// ---------------------------------------------------------------------------

func TestCampaignsDeleteRequiresConfirm(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call without --confirm: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "delete",
		"--campaign-id", "123",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure without --confirm; output=%q", out)
	}
	if !strings.Contains(out, "confirm") {
		t.Fatalf("expected error mentioning --confirm, got %q", out)
	}
}

func TestCampaignsDeleteWithConfirm(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete || req.URL.Path != "/api/v5/campaigns/123" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		called = true
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(nil)),
		}, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"campaigns", "delete",
		"--campaign-id", "123",
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestAdGroupsDeleteRequiresConfirm(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call without --confirm: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure without --confirm; output=%q", out)
	}
	if !strings.Contains(out, "confirm") {
		t.Fatalf("expected error mentioning --confirm, got %q", out)
	}
}

func TestAdGroupsDeleteWithConfirm(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete || req.URL.Path != "/api/v5/campaigns/123/adgroups/456" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		called = true
		return emptyOKResponse(), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"adgroups", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestAdsDeleteRequiresConfirm(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call without --confirm: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"ads", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--ad-id", "789",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure without --confirm; output=%q", out)
	}
	if !strings.Contains(out, "confirm") {
		t.Fatalf("expected error mentioning --confirm, got %q", out)
	}
}

func TestAdsDeleteWithConfirm(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete || req.URL.Path != "/api/v5/campaigns/123/adgroups/456/ads/789" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		called = true
		return emptyOKResponse(), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"ads", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--ad-id", "789",
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestKeywordsDeleteRequiresConfirm(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call without --confirm: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--from-json", `[789, 790]`,
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure without --confirm; output=%q", out)
	}
	if !strings.Contains(out, "confirm") {
		t.Fatalf("expected error mentioning --confirm, got %q", out)
	}
}

func TestKeywordsDeleteWithConfirm(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/api/v5/campaigns/123/adgroups/456/targetingkeywords/delete/bulk" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		called = true
		return jsonResponse(`{"data":{"results":[]}}`), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--from-json", `[789, 790]`,
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestKeywordsDeleteKeywordIDUsesSingleDeleteEndpoint(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete || req.URL.Path != "/api/v5/campaigns/123/adgroups/456/targetingkeywords/789" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		called = true
		return jsonResponse(`{"data":{"id":789}}`), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--keyword-id", "789",
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestKeywordsDeleteKeywordIDListUsesBulkEndpoint(t *testing.T) {
	called := false
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/adgroups/456/targetingkeywords/delete/bulk", func(body []byte) *http.Response {
		ids := mustUnmarshalSlice(t, body)
		if len(ids) != 2 || ids[0] != "789" || ids[1] != "790" {
			t.Fatalf("body = %s, want keyword ID list", body)
		}
		called = true
		return jsonResponse(`{"data":{"results":[]}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--keyword-id", "789, 790",
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestKeywordsDeleteRejectsKeywordIDAndFromJSON(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call for invalid arguments: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--keyword-id", "789",
		"--from-json", `[790]`,
		"--confirm",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected usage failure; output=%q", out)
	}
	if !strings.Contains(out, "only one of --keyword-id or --from-json") {
		t.Fatalf("expected mutually exclusive flag error, got %q", out)
	}
}

func TestKeywordsDeleteRejectsKeywordIDStdinAndFromJSONStdin(t *testing.T) {
	out, code := captureRun(t, []string{
		"keywords", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--keyword-id", "-",
		"--from-json", "@-",
		"--confirm",
	}, "789\n")
	if code == ExitSuccess {
		t.Fatalf("expected usage failure; output=%q", out)
	}
	if !strings.Contains(out, "cannot use --from-json @- with stdin-piped ID flags") {
		t.Fatalf("expected stdin conflict error, got %q", out)
	}
}

func TestKeywordsDeleteKeywordIDFromStdin(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete || req.URL.Path != "/api/v5/campaigns/123/adgroups/456/targetingkeywords/789" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		called = true
		return jsonResponse(`{"data":{"id":789}}`), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"keywords", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--keyword-id", "-",
		"--confirm",
	}, "789\n")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestKeywordsDeleteRejectsDeleteOneCommand(t *testing.T) {
	out, code := captureRun(t, []string{
		"keywords", "delete-one",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--keyword-id", "789",
		"--check",
	}, "")
	if strings.Contains(out, "delete-one") {
		t.Fatalf("delete-one should be absent from keywords help; code=%d output=%q", code, out)
	}
}

func TestNegativesDeleteCampaignLevelRequiresConfirm(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call without --confirm: %s %s", req.Method, req.URL.Path)
		return nil, nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "delete",
		"--campaign-id", "123",
		"--from-json", `[99, 100]`,
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure without --confirm; output=%q", out)
	}
	if !strings.Contains(out, "confirm") {
		t.Fatalf("expected error mentioning --confirm, got %q", out)
	}
}

func TestNegativesDeleteCampaignLevelWithConfirm(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/api/v5/campaigns/123/negativekeywords/delete/bulk" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		called = true
		return jsonResponse(`{"data":{"results":[]}}`), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "delete",
		"--campaign-id", "123",
		"--from-json", `[99, 100]`,
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestNegativesDeleteKeywordIDCampaignLevelWithConfirm(t *testing.T) {
	called := false
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/negativekeywords/delete/bulk", func(body []byte) *http.Response {
		ids := mustUnmarshalSlice(t, body)
		if len(ids) != 1 || ids[0] != "99" {
			t.Fatalf("body = %s, want one negative keyword ID", body)
		}
		called = true
		return jsonResponse(`{"data":{"results":[]}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "delete",
		"--campaign-id", "123",
		"--keyword-id", "99",
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestNegativesDeleteAdGroupLevelWithConfirm(t *testing.T) {
	called := false
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/api/v5/campaigns/123/adgroups/456/negativekeywords/delete/bulk" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		called = true
		return jsonResponse(`{"data":{"results":[]}}`), nil
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--from-json", `[77]`,
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestNegativesDeleteKeywordIDListAdGroupLevelWithConfirm(t *testing.T) {
	called := false
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/adgroups/456/negativekeywords/delete/bulk", func(body []byte) *http.Response {
		ids := mustUnmarshalSlice(t, body)
		if len(ids) != 2 || ids[0] != "77" || ids[1] != "78" {
			t.Fatalf("body = %s, want negative keyword ID list", body)
		}
		called = true
		return jsonResponse(`{"data":{"results":[]}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "delete",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--keyword-id", "77, 78",
		"--confirm",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}

func TestNegativesDeleteKeywordIDFromStdin(t *testing.T) {
	called := false
	client := newTestClient(recordingTransport(t, http.MethodPost, "/api/v5/campaigns/123/negativekeywords/delete/bulk", func(body []byte) *http.Response {
		ids := mustUnmarshalSlice(t, body)
		if len(ids) != 2 || ids[0] != "99" || ids[1] != "100" {
			t.Fatalf("body = %s, want negative keyword ID list from stdin", body)
		}
		called = true
		return jsonResponse(`{"data":{"results":[]}}`)
	}))
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restore()

	out, code := captureRun(t, []string{
		"negatives", "delete",
		"--campaign-id", "123",
		"--keyword-id", "-",
		"--confirm",
	}, "99,100\n")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !called {
		t.Fatalf("delete with --confirm should call API")
	}
}
