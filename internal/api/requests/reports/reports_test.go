package reports

import (
	"encoding/json"
	"testing"
)

func TestCampaignsRequest(t *testing.T) {
	body := json.RawMessage(`{"startTime":"2024-01-01","endTime":"2024-01-31","granularity":"DAILY"}`)
	req := CampaignsRequest{RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/reports/campaigns" {
		t.Errorf("Path() = %q, want %q", got, "/reports/campaigns")
	}
	if req.Body() == nil {
		t.Error("Body() is nil, want non-nil")
	}
	if got, ok := req.Body().(json.RawMessage); !ok {
		t.Errorf("Body() type = %T, want json.RawMessage", req.Body())
	} else if string(got) != string(body) {
		t.Errorf("Body() = %s, want %s", got, body)
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestKeywordsRequest_WithAdGroupID(t *testing.T) {
	body := json.RawMessage(`{"startTime":"2024-01-01","endTime":"2024-01-31"}`)
	req := KeywordsRequest{CampaignID: "100", AdGroupID: "200", RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/reports/campaigns/100/adgroups/200/keywords" {
		t.Errorf("Path() = %q, want %q", got, "/reports/campaigns/100/adgroups/200/keywords")
	}
	if req.Body() == nil {
		t.Error("Body() is nil, want non-nil")
	}
	if got, ok := req.Body().(json.RawMessage); !ok {
		t.Errorf("Body() type = %T, want json.RawMessage", req.Body())
	} else if string(got) != string(body) {
		t.Errorf("Body() = %s, want %s", got, body)
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestKeywordsRequest_WithoutAdGroupID(t *testing.T) {
	body := json.RawMessage(`{"startTime":"2024-01-01","endTime":"2024-01-31"}`)
	req := KeywordsRequest{CampaignID: "100", RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/reports/campaigns/100/keywords" {
		t.Errorf("Path() = %q, want %q", got, "/reports/campaigns/100/keywords")
	}
	if req.Body() == nil {
		t.Error("Body() is nil, want non-nil")
	}
}

func TestKeywordsRequest_PathChangesWithAdGroupID(t *testing.T) {
	// Without AdGroupID - reports at campaign level
	reqCampaignLevel := KeywordsRequest{CampaignID: "50"}
	if got := reqCampaignLevel.Path(); got != "/reports/campaigns/50/keywords" {
		t.Errorf("Path() without AdGroupID = %q, want %q", got, "/reports/campaigns/50/keywords")
	}

	// With AdGroupID - reports at ad group level
	reqAdGroupLevel := KeywordsRequest{CampaignID: "50", AdGroupID: "60"}
	if got := reqAdGroupLevel.Path(); got != "/reports/campaigns/50/adgroups/60/keywords" {
		t.Errorf("Path() with AdGroupID = %q, want %q", got, "/reports/campaigns/50/adgroups/60/keywords")
	}
}

func TestKeywordsRequest_EmptyAdGroupID(t *testing.T) {
	// Empty string AdGroupID should use campaign-level path
	req := KeywordsRequest{CampaignID: "100", AdGroupID: ""}

	if got := req.Path(); got != "/reports/campaigns/100/keywords" {
		t.Errorf("Path() with empty AdGroupID = %q, want %q", got, "/reports/campaigns/100/keywords")
	}
}

func TestKeywordsRequest_DifferentIDs(t *testing.T) {
	tests := []struct {
		name       string
		campaignID string
		adGroupID  string
		wantPath   string
	}{
		{
			name:       "campaign level",
			campaignID: "111",
			wantPath:   "/reports/campaigns/111/keywords",
		},
		{
			name:       "ad group level",
			campaignID: "222",
			adGroupID:  "333",
			wantPath:   "/reports/campaigns/222/adgroups/333/keywords",
		},
		{
			name:       "different campaign",
			campaignID: "999",
			adGroupID:  "888",
			wantPath:   "/reports/campaigns/999/adgroups/888/keywords",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := KeywordsRequest{CampaignID: tt.campaignID, AdGroupID: tt.adGroupID}
			if got := req.Path(); got != tt.wantPath {
				t.Errorf("Path() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

// TestAllReportRequestsArePOST verifies both report endpoints use POST.
func TestAllReportRequestsArePOST(t *testing.T) {
	body := json.RawMessage(`{}`)

	campaignsReq := CampaignsRequest{RawBody: body}
	if got := campaignsReq.Method(); got != "POST" {
		t.Errorf("CampaignsRequest.Method() = %q, want %q", got, "POST")
	}

	keywordsReq := KeywordsRequest{CampaignID: "1", RawBody: body}
	if got := keywordsReq.Method(); got != "POST" {
		t.Errorf("KeywordsRequest.Method() = %q, want %q", got, "POST")
	}
}
