package cmd

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apiPkg "github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

const (
	testKeywordID        = "900301"
	testCampaignNegID    = "900401"
	testAdGroupNegID     = "900402"
	testAdID             = "900501"
	testCreativeID       = "900601"
	testBudgetOrderID    = "900701"
	testProductPageID    = "cpp-fitness-strength"
	testAdRejectionID    = "900801"
	testImpressionReport = "900901"
)

func newCoverageClient(t *testing.T, wantMethod, wantPath, response string) *apiPkg.Client {
	return newCoverageClientWithCheck(t, wantMethod, wantPath, response, nil)
}

func newCoverageClientWithCheck(t *testing.T, wantMethod, wantPath, response string, check func(*http.Request)) *apiPkg.Client {
	t.Helper()

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != wantMethod {
				t.Fatalf("method = %s, want %s", req.Method, wantMethod)
			}
			if req.URL.Path != wantPath {
				t.Fatalf("path = %s, want %s", req.URL.Path, wantPath)
			}
			if check != nil {
				check(req)
			}
			return jsonResponse(response), nil
		}),
	})
	return client
}

func reportResponseJSON() string {
	return `{
		"data": {
			"reportingDataResponse": {
				"row": [
					{
						"metadata": {
							"campaignId": 900101,
							"adGroupId": 900201,
							"keywordId": 900301,
							"adId": 900501,
							"campaignName": "FitTrack US Search",
							"adGroupName": "Core Search",
							"keyword": "fitness coach",
							"adName": "FitTrack Core Ad"
						},
						"total": {
							"impressions": 123,
							"taps": 12,
							"localSpend": {"amount":"4.56","currency":"USD"}
						}
					}
				]
			}
		}
	}`
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("writing config.yaml: %v", err)
	}
	return configDir
}

func writeTempECPrivateKey(t *testing.T) string {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating EC key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshalling EC key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})
	path := filepath.Join(t.TempDir(), "test-key.pem")
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		t.Fatalf("writing private key: %v", err)
	}
	return path
}

func TestIntegration_MockedAPICommandCoverage(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantMethod   string
		wantPath     string
		response     string
		wantContains []string
	}{
		{
			name:         "campaigns get",
			args:         []string{"campaigns", "get", "--campaign-id", testCampaignID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/campaigns/" + testCampaignID,
			response:     `{"data":{"id":900101,"adamId":900001,"name":"FitTrack US Search","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"5","currency":"USD"}}}`,
			wantContains: []string{`"FitTrack US Search"`},
		},
		{
			name:         "adgroups get",
			args:         []string{"adgroups", "get", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/campaigns/" + testCampaignID + "/adgroups/" + testAdGroupID,
			response:     `{"data":{"id":900201,"campaignId":900101,"name":"Core Search","pricingModel":"CPC","defaultBidAmount":{"amount":"1.00","currency":"USD"},"startTime":"2026-01-01T00:00:00.000"}}`,
			wantContains: []string{`"Core Search"`},
		},
		{
			name:         "keywords get",
			args:         []string{"keywords", "get", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID, "--keyword-id", testKeywordID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/campaigns/" + testCampaignID + "/adgroups/" + testAdGroupID + "/targetingkeywords/" + testKeywordID,
			response:     `{"data":{"id":900301,"campaignId":900101,"adGroupId":900201,"text":"fitness coach","matchType":"EXACT","status":"ACTIVE"}}`,
			wantContains: []string{`"fitness coach"`},
		},
		{
			name:         "negatives get campaign level",
			args:         []string{"negatives", "get", "--campaign-id", testCampaignID, "--keyword-id", testCampaignNegID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/campaigns/" + testCampaignID + "/negativekeywords/" + testCampaignNegID,
			response:     `{"data":{"id":900401,"campaignId":900101,"text":"free workout","matchType":"EXACT","status":"ACTIVE"}}`,
			wantContains: []string{`"free workout"`},
		},
		{
			name:         "negatives get adgroup level",
			args:         []string{"negatives", "get", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID, "--keyword-id", testAdGroupNegID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/campaigns/" + testCampaignID + "/adgroups/" + testAdGroupID + "/negativekeywords/" + testAdGroupNegID,
			response:     `{"data":{"id":900402,"campaignId":900101,"adGroupId":900201,"text":"protein powder","matchType":"BROAD","status":"ACTIVE"}}`,
			wantContains: []string{`"protein powder"`},
		},
		{
			name:         "ads list scoped",
			args:         []string{"ads", "list", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/campaigns/" + testCampaignID + "/adgroups/" + testAdGroupID + "/ads",
			response:     `{"data":[{"id":900501,"name":"FitTrack Core Ad","creativeId":900601,"status":"ENABLED"}]}`,
			wantContains: []string{`"FitTrack Core Ad"`},
		},
		{
			name:         "ads list search all campaigns",
			args:         []string{"ads", "list", "--filter", "name CONTAINS FitTrack", "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/ads/find",
			response:     `{"data":[{"id":900501,"campaignId":900101,"adGroupId":900201,"name":"FitTrack Core Ad","status":"ENABLED"}]}`,
			wantContains: []string{`"FitTrack Core Ad"`},
		},
		{
			name:         "ads get",
			args:         []string{"ads", "get", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID, "--ad-id", testAdID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/campaigns/" + testCampaignID + "/adgroups/" + testAdGroupID + "/ads/" + testAdID,
			response:     `{"data":{"id":900501,"campaignId":900101,"adGroupId":900201,"name":"FitTrack Core Ad","creativeId":900601,"status":"ENABLED"}}`,
			wantContains: []string{`"FitTrack Core Ad"`},
		},
		{
			name:         "creatives list",
			args:         []string{"creatives", "list", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/creatives",
			response:     `{"data":[{"id":900601,"adamId":900001,"name":"FitTrack Strength Page","type":"CUSTOM_PRODUCT_PAGE"}]}`,
			wantContains: []string{`"FitTrack Strength Page"`},
		},
		{
			name:         "creatives get",
			args:         []string{"creatives", "get", "--creative-id", testCreativeID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/creatives/" + testCreativeID,
			response:     `{"data":{"id":900601,"adamId":900001,"name":"FitTrack Strength Page","type":"CUSTOM_PRODUCT_PAGE"}}`,
			wantContains: []string{`"FitTrack Strength Page"`},
		},
		{
			name:         "budgetorders list",
			args:         []string{"budgetorders", "list", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/budgetorders",
			response:     `{"data":[{"id":900701,"name":"FitTrack Quarterly Budget","status":"ACTIVE"}]}`,
			wantContains: []string{`"FitTrack Quarterly Budget"`},
		},
		{
			name:         "budgetorders get",
			args:         []string{"budgetorders", "get", "--budget-order-id", testBudgetOrderID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/budgetorders/" + testBudgetOrderID,
			response:     `{"data":{"id":900701,"name":"FitTrack Quarterly Budget","status":"ACTIVE"}}`,
			wantContains: []string{`"FitTrack Quarterly Budget"`},
		},
		{
			name:         "product pages list",
			args:         []string{"product-pages", "list", "--adam-id", testAdamID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/apps/" + testAdamID + "/product-pages",
			response:     `{"data":[{"id":"cpp-fitness-strength","name":"FitTrack Strength Page"}]}`,
			wantContains: []string{`"FitTrack Strength Page"`},
		},
		{
			name:         "product pages get",
			args:         []string{"product-pages", "get", "--adam-id", testAdamID, "--product-page-id", testProductPageID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/apps/" + testAdamID + "/product-pages/" + testProductPageID,
			response:     `{"data":{"id":"cpp-fitness-strength","name":"FitTrack Strength Page"}}`,
			wantContains: []string{`"FitTrack Strength Page"`},
		},
		{
			name:         "product pages locales",
			args:         []string{"product-pages", "locales", "--adam-id", testAdamID, "--product-page-id", testProductPageID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/apps/" + testAdamID + "/product-pages/" + testProductPageID + "/locale-details",
			response:     `{"data":[{"language":"en-US","name":"FitTrack Strength Page"}]}`,
			wantContains: []string{`"en-US"`},
		},
		{
			name:         "product pages countries",
			args:         []string{"product-pages", "countries", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/countries-or-regions",
			response:     `{"data":[{"countryOrRegion":"US","name":"United States"}]}`,
			wantContains: []string{`"United States"`},
		},
		{
			name:         "product pages devices",
			args:         []string{"product-pages", "devices", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/creativeappmappings/devices",
			response:     `{"data":[{"deviceClass":"IPHONE","size":"6.7"}]}`,
			wantContains: []string{`"IPHONE"`},
		},
		{
			name:         "ad rejections list",
			args:         []string{"ad-rejections", "list", "--filter", "adamId=900001", "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/product-page-reasons/find",
			response:     `{"data":[{"id":900801,"adamId":900001,"reasonCode":"ASSET_TEXT","countryOrRegion":"US"}]}`,
			wantContains: []string{`"ASSET_TEXT"`},
		},
		{
			name:         "ad rejections get",
			args:         []string{"ad-rejections", "get", "--reason-id", testAdRejectionID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/product-page-reasons/" + testAdRejectionID,
			response:     `{"data":{"id":900801,"reasonCode":"ASSET_TEXT","countryOrRegion":"US"}}`,
			wantContains: []string{`"ASSET_TEXT"`},
		},
		{
			name:         "ad rejections assets",
			args:         []string{"ad-rejections", "assets", "--adam-id", testAdamID, "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/apps/" + testAdamID + "/assets/find",
			response:     `{"data":[{"assetGenId":"asset-1","languageCode":"en-US"}]}`,
			wantContains: []string{`"asset-1"`},
		},
		{
			name:         "reports keywords campaign scoped",
			args:         []string{"reports", "keywords", "--campaign-id", testCampaignID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/reports/campaigns/" + testCampaignID + "/keywords",
			response:     reportResponseJSON(),
			wantContains: []string{`"fitness coach"`},
		},
		{
			name:         "reports keywords adgroup scoped",
			args:         []string{"reports", "keywords", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/reports/campaigns/" + testCampaignID + "/adgroups/" + testAdGroupID + "/keywords",
			response:     reportResponseJSON(),
			wantContains: []string{`"fitness coach"`},
		},
		{
			name:         "reports searchterms campaign scoped",
			args:         []string{"reports", "searchterms", "--campaign-id", testCampaignID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/reports/campaigns/" + testCampaignID + "/searchterms",
			response:     reportResponseJSON(),
			wantContains: []string{`"FitTrack US Search"`},
		},
		{
			name:         "reports searchterms adgroup scoped",
			args:         []string{"reports", "searchterms", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/reports/campaigns/" + testCampaignID + "/adgroups/" + testAdGroupID + "/searchterms",
			response:     reportResponseJSON(),
			wantContains: []string{`"Core Search"`},
		},
		{
			name:         "reports ads",
			args:         []string{"reports", "ads", "--campaign-id", testCampaignID, "--start", "2026-03-18", "--end", "2026-03-25", "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/reports/campaigns/" + testCampaignID + "/ads",
			response:     reportResponseJSON(),
			wantContains: []string{`"FitTrack Core Ad"`},
		},
		{
			name:         "apps search",
			args:         []string{"apps", "search", "--query", "fittrack", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/search/apps",
			response:     `{"data":[{"adamId":900001,"name":"FitTrack"}]}`,
			wantContains: []string{`"FitTrack"`},
		},
		{
			name:         "apps details",
			args:         []string{"apps", "details", "--adam-id", testAdamID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/apps/" + testAdamID,
			response:     `{"data":{"adamId":900001,"name":"FitTrack","bundleId":"com.example.fittrack"}}`,
			wantContains: []string{`"com.example.fittrack"`},
		},
		{
			name:         "apps localized",
			args:         []string{"apps", "localized", "--adam-id", testAdamID, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/apps/" + testAdamID + "/locale-details",
			response:     `{"data":[{"language":"en-US","name":"FitTrack"}]}`,
			wantContains: []string{`"en-US"`},
		},
		{
			name:         "geo search",
			args:         []string{"geo", "search", "--query", "luxembourg", "--entity", "Locality", "--country-code", "LU", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/search/geo",
			response:     `{"data":[{"id":"geo-1","entity":"Locality","name":"Luxembourg"}]}`,
			wantContains: []string{`"Luxembourg"`},
		},
		{
			name:         "geo get",
			args:         []string{"geo", "get", "--entity", "Country", "--geo-id", "US", "-f", "json"},
			wantMethod:   http.MethodPost,
			wantPath:     "/api/v5/search/geo",
			response:     `{"data":[{"id":"US","entity":"Country","displayName":"United States","countryOrRegion":"US"}]}`,
			wantContains: []string{`"United States"`},
		},
		{
			name:         "orgs list",
			args:         []string{"orgs", "list", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/acls",
			response:     `{"data":[{"orgId":123,"orgName":"FitTrack Org","currency":"USD","timeZone":"UTC"}]}`,
			wantContains: []string{`"FitTrack Org"`},
		},
		{
			name:         "orgs user",
			args:         []string{"orgs", "user", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/me",
			response:     `{"data":{"emailAddress":"fittrack@example.com","firstName":"Fit","lastName":"Track"}}`,
			wantContains: []string{`"fittrack@example.com"`},
		},
		{
			name:         "impression share list",
			args:         []string{"impression-share", "list", "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/custom-reports",
			response:     `{"data":[{"id":900901,"name":"Weekly Share Report"}]}`,
			wantContains: []string{`"Weekly Share Report"`},
		},
		{
			name:         "impression share get",
			args:         []string{"impression-share", "get", "--report-id", testImpressionReport, "-f", "json"},
			wantMethod:   http.MethodGet,
			wantPath:     "/api/v5/custom-reports/" + testImpressionReport,
			response:     `{"data":{"id":900901,"name":"Weekly Share Report"}}`,
			wantContains: []string{`"Weekly Share Report"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newCoverageClient(t, tc.wantMethod, tc.wantPath, tc.response)
			restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
			defer restore()

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

func TestIntegration_MockedMutationCheckCoverage(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantContains []string
	}{
		{
			name:         "campaigns delete check",
			args:         []string{"campaigns", "delete", "--campaign-id", testCampaignID, "--check", "-f", "json"},
			wantContains: []string{`"action":"delete campaign"`, `"target":"campaign 900101"`},
		},
		{
			name:         "keywords delete keyword-id check",
			args:         []string{"keywords", "delete", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID, "--keyword-id", testKeywordID, "--check", "-f", "json"},
			wantContains: []string{`"action":"delete keyword"`, `"target":"campaign 900101, adgroup 900201, keyword 900301"`},
		},
		{
			name:         "impression share create check",
			args:         []string{"impression-share", "create", "--from-json", `{"name":"Weekly Share Report","dateRange":"LAST_WEEK","granularity":"DAILY"}`, "--check", "-f", "json"},
			wantContains: []string{`"action":"create impression share report"`, `"name: Weekly Share Report"`},
		},
		{
			name:         "profiles create check",
			args:         []string{"profiles", "create", "--name", "new-work", "--client-id", "SEARCHADS.mock", "--org-id", "123", "--check", "-f", "json"},
			wantContains: []string{`"action":"create profile"`, `"target":"name new-work"`},
		},
		{
			name:         "profiles update check",
			args:         []string{"profiles", "update", "--name", "work", "--org-id", "456", "--check", "-f", "json"},
			wantContains: []string{`"action":"update profile"`, `"orgId: 456"`},
		},
		{
			name:         "profiles delete check",
			args:         []string{"profiles", "delete", "--name", "work", "--check", "-f", "json"},
			wantContains: []string{`"action":"delete profile"`, `"target":"name work"`},
		},
		{
			name:         "profiles delete check with private key",
			args:         []string{"profiles", "delete", "--name", "work", "--delete-private-key", "--check", "-f", "json"},
			wantContains: []string{`"action":"delete profile"`, `"target":"name work"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configDir := writeTempConfig(t, `
default_profile: work
profiles:
  work:
    client_id: work-client
    team_id: work-team
    key_id: work-key
    org_id: "123"
    private_key_path: /tmp/mock-key.pem
`)
			t.Setenv("AADS_CONFIG_DIR", configDir)

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

func TestIntegration_AppsSearch_OnlyOwnedAppsQuery(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		wantQuery           string
		wantReturnOwnedApps string
		wantLimit           string
		wantOffset          string
		response            string
		wantContains        string
	}{
		{
			name:                "paged request",
			args:                []string{"apps", "search", "--query", "fittrack", "--only-owned-apps", "--limit", "2", "--offset", "4", "-f", "json"},
			wantQuery:           "fittrack",
			wantReturnOwnedApps: "true",
			wantLimit:           "2",
			wantOffset:          "4",
			response:            `{"data":[{"adamId":900001,"name":"FitTrack"}]}`,
			wantContains:        `"FitTrack"`,
		},
		{
			name:                "fetch all pages",
			args:                []string{"apps", "search", "--query", "fittrack", "--only-owned-apps", "-f", "json"},
			wantQuery:           "fittrack",
			wantReturnOwnedApps: "true",
			wantLimit:           "1000",
			wantOffset:          "0",
			response:            `{"data":[{"adamId":900001,"name":"FitTrack"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1}}`,
			wantContains:        `"FitTrack"`,
		},
		{
			name:                "owned apps without query",
			args:                []string{"apps", "search", "--only-owned-apps", "--limit", "2", "-f", "json"},
			wantQuery:           "",
			wantReturnOwnedApps: "true",
			wantLimit:           "2",
			wantOffset:          "",
			response:            `{"data":[{"adamId":900001,"name":"FitTrack"}]}`,
			wantContains:        `"FitTrack"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newCoverageClientWithCheck(
				t,
				http.MethodGet,
				"/api/v5/search/apps",
				tc.response,
				func(req *http.Request) {
					query := req.URL.Query()
					if got := query.Get("query"); got != tc.wantQuery {
						t.Fatalf("query = %q, want %q", got, tc.wantQuery)
					}
					if got := query.Get("returnOwnedApps"); got != tc.wantReturnOwnedApps {
						t.Fatalf("returnOwnedApps = %q, want %q", got, tc.wantReturnOwnedApps)
					}
					if got := query.Get("limit"); got != tc.wantLimit {
						t.Fatalf("limit = %q, want %q", got, tc.wantLimit)
					}
					if got := query.Get("offset"); got != tc.wantOffset {
						t.Fatalf("offset = %q, want %q", got, tc.wantOffset)
					}
				},
			)
			restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
			defer restore()

			out, code := captureRun(t, tc.args, "")
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			if !strings.Contains(out, tc.wantContains) {
				t.Fatalf("output missing %q: %q", tc.wantContains, out)
			}
		})
	}
}

func TestIntegration_AppsSearch_LocalSort(t *testing.T) {
	client := newCoverageClientWithCheck(
		t,
		http.MethodGet,
		"/api/v5/search/apps",
		`{"data":[{"adamId":2,"name":"Zulu"},{"adamId":1,"name":"Alpha"}]}`,
		func(req *http.Request) {
			query := req.URL.Query()
			if got := query.Get("query"); got != "fittrack" {
				t.Fatalf("query = %q, want %q", got, "fittrack")
			}
			if got := query.Get("limit"); got != "2" {
				t.Fatalf("limit = %q, want %q", got, "2")
			}
			if got := query.Get("sort"); got != "" {
				t.Fatalf("unexpected remote sort query %q", got)
			}
			if got := query.Get("sortOrder"); got != "" {
				t.Fatalf("unexpected remote sortOrder query %q", got)
			}
		},
	)
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"apps", "search", "--query", "fittrack", "--limit", "2", "--sort", "name:asc", "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	alpha := strings.Index(out, `"name":"Alpha"`)
	zulu := strings.Index(out, `"name":"Zulu"`)
	if alpha < 0 || zulu < 0 || alpha > zulu {
		t.Fatalf("expected local sort by name asc, got %q", out)
	}
}

func TestIntegration_AppsSearch_LocalSortDescending(t *testing.T) {
	client := newCoverageClientWithCheck(
		t,
		http.MethodGet,
		"/api/v5/search/apps",
		`{"data":[{"adamId":1,"name":"Alpha"},{"adamId":2,"name":"Zulu"}]}`,
		func(req *http.Request) {
			query := req.URL.Query()
			if got := query.Get("sort"); got != "" {
				t.Fatalf("unexpected remote sort query %q", got)
			}
			if got := query.Get("sortOrder"); got != "" {
				t.Fatalf("unexpected remote sortOrder query %q", got)
			}
		},
	)
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"apps", "search", "--query", "fittrack", "--limit", "2", "--sort", "name:desc", "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	alpha := strings.Index(out, `"name":"Alpha"`)
	zulu := strings.Index(out, `"name":"Zulu"`)
	if alpha < 0 || zulu < 0 || zulu > alpha {
		t.Fatalf("expected local sort by name desc, got %q", out)
	}
}

func TestIntegration_AppsSearch_LocalSortMultiple(t *testing.T) {
	client := newCoverageClientWithCheck(
		t,
		http.MethodGet,
		"/api/v5/search/apps",
		`{"data":[
			{"adamId":1,"name":"Alpha","rank":2},
			{"adamId":2,"name":"Alpha","rank":1},
			{"adamId":3,"name":"Beta","rank":5}
		]}`,
		func(req *http.Request) {
			query := req.URL.Query()
			if got := query.Get("sort"); got != "" {
				t.Fatalf("unexpected remote sort query %q", got)
			}
		},
	)
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"apps", "search", "--query", "fittrack", "--limit", "3", "--sort", "name:asc", "--sort", "rank:desc", "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	adam1 := strings.Index(out, `"adamId":1`) // Alpha, rank=2 — first inside Alpha group
	adam2 := strings.Index(out, `"adamId":2`) // Alpha, rank=1
	adam3 := strings.Index(out, `"adamId":3`) // Beta
	if adam1 < 0 || adam2 < 0 || adam3 < 0 || adam1 > adam2 || adam2 > adam3 {
		t.Fatalf("expected name asc then rank desc (Alpha rank=2, Alpha rank=1, Beta), got %q", out)
	}
}

func TestIntegration_CustomCollectionCommands_LocalSort(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantPath   string
		wantQuery  map[string]string
		response   string
		wantFirst  string
		wantSecond string
	}{
		{
			name:       "geo search",
			args:       []string{"geo", "search", "--query", "luxembourg", "--limit", "2", "--sort", "name:asc", "-f", "json"},
			wantPath:   "/api/v5/search/geo",
			wantQuery:  map[string]string{"query": "luxembourg", "limit": "2"},
			response:   `{"data":[{"id":"geo-2","name":"Zulu"},{"id":"geo-1","name":"Alpha"}]}`,
			wantFirst:  `"name":"Alpha"`,
			wantSecond: `"name":"Zulu"`,
		},
		{
			name:       "orgs list",
			args:       []string{"orgs", "list", "--sort", "orgName:asc", "-f", "json"},
			wantPath:   "/api/v5/acls",
			wantQuery:  map[string]string{},
			response:   `{"data":[{"orgId":2,"orgName":"Zulu Org"},{"orgId":1,"orgName":"Alpha Org"}]}`,
			wantFirst:  `"orgName":"Alpha Org"`,
			wantSecond: `"orgName":"Zulu Org"`,
		},
		{
			name:       "product pages list",
			args:       []string{"product-pages", "list", "--adam-id", testAdamID, "--limit", "2", "--sort", "name:asc", "-f", "json"},
			wantPath:   "/api/v5/apps/" + testAdamID + "/product-pages",
			wantQuery:  map[string]string{"limit": "2"},
			response:   `{"data":[{"id":"cpp-2","name":"Zulu Page"},{"id":"cpp-1","name":"Alpha Page"}]}`,
			wantFirst:  `"name":"Alpha Page"`,
			wantSecond: `"name":"Zulu Page"`,
		},
		{
			name:       "product pages locales",
			args:       []string{"product-pages", "locales", "--adam-id", testAdamID, "--product-page-id", testProductPageID, "--sort", "language:asc", "-f", "json"},
			wantPath:   "/api/v5/apps/" + testAdamID + "/product-pages/" + testProductPageID + "/locale-details",
			wantQuery:  map[string]string{},
			response:   `{"data":[{"language":"fr-FR","name":"Zulu"},{"language":"en-US","name":"Alpha"}]}`,
			wantFirst:  `"language":"en-US"`,
			wantSecond: `"language":"fr-FR"`,
		},
		{
			name:       "product pages countries",
			args:       []string{"product-pages", "countries", "--sort", "countryOrRegion:asc", "-f", "json"},
			wantPath:   "/api/v5/countries-or-regions",
			wantQuery:  map[string]string{},
			response:   `{"data":[{"countryOrRegion":"US","name":"United States"},{"countryOrRegion":"GB","name":"United Kingdom"}]}`,
			wantFirst:  `"countryOrRegion":"GB"`,
			wantSecond: `"countryOrRegion":"US"`,
		},
		{
			name:       "product pages devices",
			args:       []string{"product-pages", "devices", "--sort", "deviceClass:asc", "-f", "json"},
			wantPath:   "/api/v5/creativeappmappings/devices",
			wantQuery:  map[string]string{},
			response:   `{"data":[{"deviceClass":"IPHONE","size":"6.7"},{"deviceClass":"IPAD","size":"13"}]}`,
			wantFirst:  `"deviceClass":"IPAD"`,
			wantSecond: `"deviceClass":"IPHONE"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newCoverageClientWithCheck(
				t,
				http.MethodGet,
				tc.wantPath,
				tc.response,
				func(req *http.Request) {
					query := req.URL.Query()
					for key, want := range tc.wantQuery {
						if got := query.Get(key); got != want {
							t.Fatalf("%s = %q, want %q", key, got, want)
						}
					}
					if got := query.Get("sort"); got != "" {
						t.Fatalf("unexpected remote sort query %q", got)
					}
					if got := query.Get("sortOrder"); got != "" {
						t.Fatalf("unexpected remote sortOrder query %q", got)
					}
				},
			)
			restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
			defer restore()

			out, code := captureRun(t, tc.args, "")
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			first := strings.Index(out, tc.wantFirst)
			second := strings.Index(out, tc.wantSecond)
			if first < 0 || second < 0 || first > second {
				t.Fatalf("expected local sort order %q before %q, got %q", tc.wantFirst, tc.wantSecond, out)
			}
		})
	}
}

func TestIntegration_BudgetOrdersList_LocalSort(t *testing.T) {
	client := newCoverageClientWithCheck(
		t,
		http.MethodGet,
		"/api/v5/budgetorders",
		`{"data":[{"id":2,"name":"Zulu Budget","status":"ACTIVE"},{"id":1,"name":"Alpha Budget","status":"ACTIVE"}]}`,
		func(req *http.Request) {
			query := req.URL.Query()
			if got := query.Get("limit"); got != "2" {
				t.Fatalf("limit = %q, want %q", got, "2")
			}
			if got := query.Get("sort"); got != "" {
				t.Fatalf("unexpected remote sort query %q", got)
			}
			if got := query.Get("sortOrder"); got != "" {
				t.Fatalf("unexpected remote sortOrder query %q", got)
			}
		},
	)
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"budgetorders", "list", "--limit", "2", "--sort", "name:asc", "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	alpha := strings.Index(out, `"name":"Alpha Budget"`)
	zulu := strings.Index(out, `"name":"Zulu Budget"`)
	if alpha < 0 || zulu < 0 || alpha > zulu {
		t.Fatalf("expected local sort by name asc, got %q", out)
	}
}

func TestIntegration_ProductPagesLocales_AdamIDFromStdin(t *testing.T) {
	const adamID1 = "900001"
	const adamID2 = "900002"
	seen := map[string]int{}
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", req.Method)
		}
		switch req.URL.Path {
		case "/api/v5/apps/" + adamID1 + "/product-pages/" + testProductPageID + "/locale-details":
			seen[adamID1]++
			return jsonResponse(`{"data":[{"language":"en-US","name":"Alpha"}]}`), nil
		case "/api/v5/apps/" + adamID2 + "/product-pages/" + testProductPageID + "/locale-details":
			seen[adamID2]++
			return jsonResponse(`{"data":[{"language":"fr-FR","name":"Beta"}]}`), nil
		default:
			t.Fatalf("unexpected path %s", req.URL.Path)
			return nil, nil
		}
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	stdin := adamID1 + "\n" + adamID2 + "\n"
	out, code := captureRun(t, []string{
		"product-pages", "locales",
		"--adam-id", "-",
		"--product-page-id", testProductPageID,
		"-f", "json",
	}, stdin)
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if seen[adamID1] != 1 || seen[adamID2] != 1 {
		t.Fatalf("expected one request per adam-id, got %v", seen)
	}
	if !strings.Contains(out, `"en-US"`) || !strings.Contains(out, `"fr-FR"`) {
		t.Fatalf("expected merged locales from both adam-ids, got %q", out)
	}
}

func TestIntegration_ProfilesCreate_RequiresResolvedOrgID(t *testing.T) {
	configDir := writeTempConfig(t, `
default_profile: work
profiles:
  work:
    client_id: work-client
    team_id: work-team
    key_id: work-key
    org_id: "123"
    private_key_path: /tmp/mock-key.pem
`)
	t.Setenv("AADS_CONFIG_DIR", configDir)

	out, code := captureRun(t, []string{"profiles", "create", "--name", "new-work", "--client-id", "SEARCHADS.mock"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "--org-id is required unless it can be inferred") {
		t.Fatalf("output missing inferred-org usage error: %q", out)
	}
}

func TestIntegration_ProfilesCreate_InferOrgAndDefaultsFromACLs(t *testing.T) {
	t.Setenv("AADS_CONFIG_DIR", t.TempDir())

	keyPath := writeTempECPrivateKey(t)
	previousTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Host == "appleid.apple.com" && req.URL.Path == "/auth/oauth2/token":
			return jsonResponse(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`), nil
		case req.URL.Host == "api.searchads.apple.com" && req.URL.Path == "/api/v5/me":
			if got := req.Header.Get("X-AP-Context"); got != "" {
				t.Fatalf("/me should not send X-AP-Context during profile discovery, got %q", got)
			}
			return jsonResponse(`{"data":{"userId":111,"parentOrgId":456}}`), nil
		case req.URL.Host == "api.searchads.apple.com" && req.URL.Path == "/api/v5/acls":
			return jsonResponse(`{"data":[{"orgId":123,"orgName":"Other Org","currency":"USD","paymentModel":"PAYG","roleNames":["Admin"],"timeZone":"America/New_York"},{"orgId":456,"orgName":"Chosen Org","currency":"EUR","paymentModel":"PAYG","roleNames":["Admin"],"timeZone":"Europe/Paris"}]}`), nil
		default:
			t.Fatalf("unexpected HTTP request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})
	defer func() { http.DefaultTransport = previousTransport }()

	out, code := captureRun(t, []string{
		"profiles", "create",
		"--name", "new-work",
		"--client-id", "SEARCHADS.mock",
		"--team-id", "TEAM.mock",
		"--key-id", "KEY.mock",
		"--private-key-path", keyPath,
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}

	configPath := filepath.Join(os.Getenv("AADS_CONFIG_DIR"), "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", configPath, err)
	}
	content := string(data)
	if !strings.Contains(content, `org_id: "456"`) {
		t.Fatalf("config file missing inferred org_id: %s", content)
	}
	if !strings.Contains(content, `default_currency: EUR`) {
		t.Fatalf("config file missing inferred default_currency: %s", content)
	}
	if !strings.Contains(content, `default_timezone: Europe/Paris`) {
		t.Fatalf("config file missing inferred default_timezone: %s", content)
	}
}

func TestIntegration_ProfilesCreate_CLIFlagsOverrideACLDefaults(t *testing.T) {
	t.Setenv("AADS_CONFIG_DIR", t.TempDir())

	keyPath := writeTempECPrivateKey(t)
	previousTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Host == "appleid.apple.com" && req.URL.Path == "/auth/oauth2/token":
			return jsonResponse(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`), nil
		case req.URL.Host == "api.searchads.apple.com" && req.URL.Path == "/api/v5/acls":
			return jsonResponse(`{"data":[{"orgId":456,"orgName":"Chosen Org","currency":"EUR","paymentModel":"PAYG","roleNames":["Admin"],"timeZone":"Europe/Paris"}]}`), nil
		default:
			t.Fatalf("unexpected HTTP request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})
	defer func() { http.DefaultTransport = previousTransport }()

	out, code := captureRun(t, []string{
		"profiles", "create",
		"--name", "new-work",
		"--client-id", "SEARCHADS.mock",
		"--team-id", "TEAM.mock",
		"--key-id", "KEY.mock",
		"--private-key-path", keyPath,
		"--org-id", "456",
		"--default-currency", "USD",
		"--default-timezone", "Europe/Luxembourg",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}

	configPath := filepath.Join(os.Getenv("AADS_CONFIG_DIR"), "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", configPath, err)
	}
	content := string(data)
	if !strings.Contains(content, `org_id: "456"`) {
		t.Fatalf("config file missing org_id: %s", content)
	}
	if !strings.Contains(content, `default_currency: USD`) {
		t.Fatalf("config file should preserve CLI default_currency: %s", content)
	}
	if !strings.Contains(content, `default_timezone: Europe/Luxembourg`) {
		t.Fatalf("config file should preserve CLI default_timezone: %s", content)
	}
}

func TestIntegration_ProfilesCreate_WarnsWhenACLRowDoesNotMatchResolvedOrg(t *testing.T) {
	t.Setenv("AADS_CONFIG_DIR", t.TempDir())

	keyPath := writeTempECPrivateKey(t)
	previousTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Host == "appleid.apple.com" && req.URL.Path == "/auth/oauth2/token":
			return jsonResponse(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`), nil
		case req.URL.Host == "api.searchads.apple.com" && req.URL.Path == "/api/v5/me":
			return jsonResponse(`{"data":{"userId":111,"parentOrgId":456}}`), nil
		case req.URL.Host == "api.searchads.apple.com" && req.URL.Path == "/api/v5/acls":
			return jsonResponse(`{"data":[{"orgId":999,"orgName":"Other Org","currency":"USD","paymentModel":"PAYG","roleNames":["Admin"],"timeZone":"America/New_York"}]}`), nil
		default:
			t.Fatalf("unexpected HTTP request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})
	defer func() { http.DefaultTransport = previousTransport }()

	out, code := captureRun(t, []string{
		"profiles", "create",
		"--name", "new-work",
		"--client-id", "SEARCHADS.mock",
		"--team-id", "TEAM.mock",
		"--key-id", "KEY.mock",
		"--private-key-path", keyPath,
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "Warning: could not find org 456 in orgs list (Apple ACLs)") {
		t.Fatalf("expected ACL warning, got %q", out)
	}

	configPath := filepath.Join(os.Getenv("AADS_CONFIG_DIR"), "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", configPath, err)
	}
	content := string(data)
	if !strings.Contains(content, `org_id: "456"`) {
		t.Fatalf("config file missing inferred org_id: %s", content)
	}
	if !strings.Contains(content, `default_currency: ""`) {
		t.Fatalf("config file should keep default_currency empty without matching ACL row: %s", content)
	}
	if !strings.Contains(content, `default_timezone: ""`) {
		t.Fatalf("config file should keep default_timezone empty without matching ACL row: %s", content)
	}
}

func TestIntegration_LocalCommandCoverage(t *testing.T) {
	t.Run("profiles list get and set-default use temp config dir", func(t *testing.T) {
		configDir := writeTempConfig(t, `
default_profile: work
profiles:
  work:
    client_id: work-client
    team_id: work-team
    key_id: work-key
    org_id: "123"
    private_key_path: /tmp/work.pem
  backup:
    client_id: backup-client
    team_id: backup-team
    key_id: backup-key
    org_id: "456"
    private_key_path: /tmp/backup.pem
`)
		t.Setenv("AADS_CONFIG_DIR", configDir)

		out, code := captureRun(t, []string{"profiles", "list"}, "")
		if code != ExitSuccess || !strings.Contains(out, "work") || !strings.Contains(out, "backup") {
			t.Fatalf("profiles list failed: code=%d output=%q", code, out)
		}

		out, code = captureRun(t, []string{"profiles", "get", "--name", "backup"}, "")
		if code != ExitSuccess || !strings.Contains(out, "backup") {
			t.Fatalf("profiles get failed: code=%d output=%q", code, out)
		}

		out, code = captureRun(t, []string{"profiles", "set-default", "backup"}, "")
		if code != ExitSuccess || !strings.Contains(out, `Default profile set to "backup"`) {
			t.Fatalf("profiles set-default failed: code=%d output=%q", code, out)
		}

		data, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
		if err != nil {
			t.Fatalf("reading config after set-default: %v", err)
		}
		if !strings.Contains(string(data), "default_profile: backup") {
			t.Fatalf("config file missing updated default profile: %s", data)
		}
	})

	t.Run("completion bash", func(t *testing.T) {
		out, code := captureRun(t, []string{"completion", "bash"}, "")
		if code != ExitSuccess {
			t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
		}
		if !strings.Contains(out, "_aads") {
			t.Fatalf("completion output missing _aads: %q", out)
		}
	})

	t.Run("schema keyword query", func(t *testing.T) {
		out, code := captureRun(t, []string{"schema", "keyword"}, "")
		if code != ExitSuccess {
			t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
		}
		if !strings.Contains(strings.ToLower(out), "keyword") {
			t.Fatalf("schema output missing keyword query match: %q", out)
		}
	})
}
