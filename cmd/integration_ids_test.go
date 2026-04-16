package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/imesart/apple-ads-cli/internal/testutil"
)

func TestLive_DiscoverLiveIDsFromListEndpoints(t *testing.T) {
	ids := testutil.RequireLiveIDs(t)

	if ids.OrgID == "" {
		t.Fatal("OrgID is empty")
	}
	if ids.AdamID == "" {
		t.Fatal("AdamID is empty")
	}
	if ids.CampaignID == "" {
		t.Fatal("CampaignID is empty")
	}
	if ids.AdGroupID == "" {
		t.Fatal("AdGroupID is empty")
	}
	if ids.KeywordID == "" {
		t.Fatal("KeywordID is empty")
	}
}

func TestLive_ReadCommandCoverage(t *testing.T) {
	ids := testutil.RequireLiveIDs(t)

	tests := []struct {
		name         string
		args         []string
		allowEmpty   bool
		skipOnOutput []string
		wantContains []string
	}{
		{name: "orgs list", args: []string{"orgs", "list", "-f", "json"}, wantContains: []string{`"orgId"`}},
		{name: "orgs user", args: []string{"orgs", "user", "-f", "json"}, wantContains: []string{`"userId"`}},
		{name: "campaigns list", args: []string{"campaigns", "list", "--limit", "1", "-f", "json"}, wantContains: []string{`"id"`}},
		{name: "campaigns get", args: []string{"campaigns", "get", "--campaign-id", ids.CampaignID, "-f", "json"}, wantContains: []string{ids.CampaignID}},
		{name: "adgroups list", args: []string{"adgroups", "list", "--campaign-id", ids.CampaignID, "--limit", "1", "-f", "json"}, wantContains: []string{`"id"`}},
		{name: "adgroups get", args: []string{"adgroups", "get", "--campaign-id", ids.CampaignID, "--adgroup-id", ids.AdGroupID, "-f", "json"}, wantContains: []string{ids.AdGroupID}},
		{name: "keywords list", args: []string{"keywords", "list", "--campaign-id", ids.CampaignID, "--adgroup-id", ids.AdGroupID, "--limit", "1", "-f", "json"}, wantContains: []string{`"id"`}},
		{name: "keywords get", args: []string{"keywords", "get", "--campaign-id", ids.CampaignID, "--adgroup-id", ids.AdGroupID, "--keyword-id", ids.KeywordID, "-f", "json"}, wantContains: []string{ids.KeywordID}},
		{name: "apps search", args: []string{"apps", "search", "--query", "fitness", "--limit", "1", "-f", "json"}, wantContains: []string{`"adamId"`}},
		{name: "apps details", args: []string{"apps", "details", "--adam-id", ids.AdamID, "-f", "json"}, wantContains: []string{ids.AdamID}},
		{name: "apps localized", args: []string{"apps", "localized", "--adam-id", ids.AdamID, "-f", "json"}, skipOnOutput: []string{`HTTP 404: Resource not found`}, wantContains: []string{`"language`}},
		{name: "geo search", args: []string{"geo", "search", "--query", "luxembourg", "--limit", "1", "-f", "json"}, wantContains: []string{`"displayName"`}},
		{name: "geo get", args: []string{"geo", "get", "--entity", "Country", "--geo-id", "US", "-f", "json"}, allowEmpty: true, wantContains: []string{`"countryOrRegion"`}},
		{name: "reports campaigns", args: []string{"reports", "campaigns", "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"}, wantContains: []string{`"campaignId"`}},
		{name: "reports adgroups", args: []string{"reports", "adgroups", "--campaign-id", ids.CampaignID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"}, wantContains: []string{`"adGroupId"`}},
		{name: "reports keywords campaign scoped", args: []string{"reports", "keywords", "--campaign-id", ids.CampaignID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"}, allowEmpty: true, wantContains: []string{`"keyword`}},
		{name: "reports keywords adgroup scoped", args: []string{"reports", "keywords", "--campaign-id", ids.CampaignID, "--adgroup-id", ids.AdGroupID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"}, allowEmpty: true, wantContains: []string{`"keyword`}},
		{name: "reports searchterms campaign scoped", args: []string{"reports", "searchterms", "--campaign-id", ids.CampaignID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"}, allowEmpty: true, wantContains: []string{`"searchTerm`}},
		{name: "reports searchterms adgroup scoped", args: []string{"reports", "searchterms", "--campaign-id", ids.CampaignID, "--adgroup-id", ids.AdGroupID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"}, allowEmpty: true, wantContains: []string{`"searchTerm`}},
		{name: "reports ads", args: []string{"reports", "ads", "--campaign-id", ids.CampaignID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"}, allowEmpty: true, wantContains: []string{`"adId"`}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, code := captureRun(t, tc.args, "")
			for _, marker := range tc.skipOnOutput {
				if strings.Contains(out, marker) {
					t.Skipf("skipping due to live API response: %s", marker)
				}
			}
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			if tc.allowEmpty && strings.TrimSpace(out) == "[]" {
				return
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(out, want) {
					t.Fatalf("output missing %q: %q", want, out)
				}
			}
		})
	}
}

func TestLive_MutationCheckCoverage(t *testing.T) {
	ids := testutil.RequireLiveIDs(t)

	tests := []struct {
		name         string
		args         []string
		wantContains []string
	}{
		{
			name:         "campaigns delete check",
			args:         []string{"campaigns", "delete", "--campaign-id", ids.CampaignID, "--check", "-f", "json"},
			wantContains: []string{`"action":"delete campaign"`, `"target":"campaign ` + ids.CampaignID + `"`},
		},
		{
			name:         "keywords delete keyword-id check",
			args:         []string{"keywords", "delete", "--campaign-id", ids.CampaignID, "--adgroup-id", ids.AdGroupID, "--keyword-id", ids.KeywordID, "--check", "-f", "json"},
			wantContains: []string{`"action":"delete keyword"`, `"target":"campaign ` + ids.CampaignID},
		},
		{
			name:         "impression share create check",
			args:         []string{"impression-share", "create", "--from-json", `{"name":"Weekly Share Report","dateRange":"LAST_WEEK","granularity":"DAILY"}`, "--check", "-f", "json"},
			wantContains: []string{`"action":"create impression share report"`, `"name: Weekly Share Report"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, code := captureRun(t, tc.args, "")
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(out, want) {
					t.Fatalf("output missing %q: %q", want, out)
				}
			}
		})
	}
}

func TestLive_StructureExportImportCheckCoverage(t *testing.T) {
	ids := testutil.RequireLiveIDs(t)

	structureJSON, code := captureRun(t, []string{
		"structure", "export",
		"--scope", "adgroups",
		"--campaign-id", ids.CampaignID,
		"--adgroups-filter", "id=" + ids.AdGroupID,
		"--adgroups-fields", "name,defaultBidAmount",
		"--keywords-filter", "id=" + ids.KeywordID,
		"--keywords-fields", "text,matchType",
		"--no-negatives",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("structure export exit code = %d, want %d; output=%q", code, ExitSuccess, structureJSON)
	}
	for _, want := range []string{
		`"type":"structure"`,
		`"scope":"adgroups"`,
		`"adgroups":[`,
		`"adgroup":`,
		`"keywords":[`,
	} {
		if !strings.Contains(structureJSON, want) {
			t.Fatalf("structure export output missing %q: %q", want, structureJSON)
		}
	}

	nameTemplate := fmt.Sprintf("aads-live-check-%d-%%(name)", time.Now().UnixNano())
	out, code := captureRun(t, []string{
		"structure", "import",
		"--from-structure", structureJSON,
		"--campaign-id", ids.CampaignID,
		"--adgroups-name", nameTemplate,
		"--check",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("structure import --check exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		`"type":"mapping"`,
		`"scope":"adgroups"`,
		`"adgroups":[`,
		`"created":{"id":1,"name":"aads-live-check-`,
		`"keywords":[`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("structure import --check output missing %q: %q", want, out)
		}
	}
}
