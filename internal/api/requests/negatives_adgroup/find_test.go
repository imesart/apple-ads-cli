package negatives_adgroup

import (
	"encoding/json"
	"testing"
)

func TestFindRequest_IsCampaignScoped(t *testing.T) {
	body := json.RawMessage(`{"conditions":[{"field":"status","operator":"EQUALS","values":["ACTIVE"]}]}`)
	req := FindRequest{CampaignID: "100", RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Fatalf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/negativekeywords/find" {
		t.Fatalf("Path() = %q, want %q", got, "/campaigns/100/adgroups/negativekeywords/find")
	}
	if req.Body() == nil {
		t.Fatal("Body() is nil, want non-nil")
	}
	if got, ok := req.Body().(json.RawMessage); !ok {
		t.Fatalf("Body() type = %T, want json.RawMessage", req.Body())
	} else if string(got) != string(body) {
		t.Fatalf("Body() = %s, want %s", got, body)
	}
	if req.Query() != nil {
		t.Fatalf("Query() = %v, want nil", req.Query())
	}
}
