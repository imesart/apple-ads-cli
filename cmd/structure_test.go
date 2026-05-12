package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	apiPkg "github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

func TestStructureExport_Campaigns_NormalizedJSON(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"5","currency":"USD"},"status":"ENABLED"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[{"id":900401,"campaignId":900101,"text":"free workout","matchType":"EXACT","status":"ACTIVE"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","pricingModel":"CPC","defaultBidAmount":{"amount":"1.00","currency":"USD"},"status":"ENABLED"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords":
				return jsonResponse(`{"data":[{"id":900402,"campaignId":900101,"adGroupId":900201,"text":"protein powder","matchType":"BROAD","status":"ACTIVE"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[{"id":900301,"campaignId":900101,"adGroupId":900201,"text":"fitness coach","matchType":"EXACT","status":"ACTIVE","bidAmount":{"amount":"1.00","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		`"$schema":"https://raw.githubusercontent.com/imesart/apple-ads-cli/main/docs/schemas/aads-v1.schema.json"`,
		`"type":"structure"`,
		`"scope":"campaigns"`,
		`"campaigns":[{"campaign":{"adamId":900001`,
		`"name":"Core Search"`,
		`"defaultBidAmount":{"amount":"1.00","currency":"USD"}`,
		`"text":"free workout"`,
		`"matchType":"EXACT"`,
		`"text":"fitness coach"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %q", want, out)
		}
	}
	for _, unwanted := range []string{`"id":900101`, `"status":"ENABLED"`, `"bidAmount":{"amount":"1.00","currency":"USD"}`} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("output unexpectedly contains %q: %q", unwanted, out)
		}
	}
}

func TestStructureExport_Shareable_CampaignStatusNumericAliasFilterNormalizes(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/find":
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading campaigns body: %v", err)
				}
				if !bytes.Contains(body, []byte(`"field":"campaignStatus"`)) && !bytes.Contains(body, []byte(`"field":"status"`)) {
					t.Fatalf("campaign selector missing status field: %s", body)
				}
				if !bytes.Contains(body, []byte(`"ENABLED"`)) {
					t.Fatalf("campaign selector = %s, want status normalized to ENABLED", body)
				}
				if bytes.Contains(body, []byte(`"1"`)) {
					t.Fatalf("campaign selector should not contain raw numeric alias after normalization: %s", body)
				}
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"},"status":"ENABLED"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack US"}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--campaigns-filter", "status=1", "--shareable"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"scope":"campaigns"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestStructureExport_Help_DoesNotShowFormatOrGlobalFieldsFlags(t *testing.T) {
	out, code := captureRun(t, []string{"structure", "export", "--help"}, "")
	if code != ExitSuccess && code != ExitUsage {
		t.Fatalf("exit code = %d, want %d or %d; output=%q", code, ExitSuccess, ExitUsage, out)
	}
	for _, unwanted := range []string{"--format", "--fields"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("help should not include %q for structure export: %q", unwanted, out)
		}
	}
	if !strings.Contains(out, "-pretty") {
		t.Fatalf("help should still include -pretty: %q", out)
	}
}

func TestStructureExport_RejectsFormatAndGlobalFieldsFlags(t *testing.T) {
	for _, args := range [][]string{
		{"structure", "export", "--scope", "campaigns", "--format", "json"},
		{"structure", "export", "--scope", "campaigns", "-f", "json"},
		{"structure", "export", "--scope", "campaigns", "--fields", "name"},
	} {
		out, code := captureRun(t, args, "")
		if code != ExitUsage {
			t.Fatalf("args=%v exit code = %d, want %d; output=%q", args, code, ExitUsage, out)
		}
		if !strings.Contains(out, "flag provided but not defined") {
			t.Fatalf("args=%v expected unknown-flag error, got %q", args, out)
		}
	}
}

func TestStructureExport_NormalizedJSON_IncludesRequestedNonDefaultFields(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack Display","adChannelType":"DISPLAY","countriesOrRegions":["US"],"billingEvent":"IMPRESSIONS","dailyBudgetAmount":{"amount":"5","currency":"USD"},"targetCpa":{"amount":"7.50","currency":"USD"},"supplySources":["APPSTORE_SEARCH_TAB"],"biddingStrategy":"AUTO"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Display","pricingModel":"CPM","paymentModel":"LOC","defaultBidAmount":{"amount":"1.00","currency":"USD"},"automatedKeywordsOptIn":true,"cpaGoal":{"amount":"2.50","currency":"USD"},"biddingStrategy":"AUTO"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		`"adChannelType":"DISPLAY"`,
		`"billingEvent":"IMPRESSIONS"`,
		`"targetCpa":{"amount":"7.50","currency":"USD"}`,
		`"supplySources":["APPSTORE_SEARCH_TAB"]`,
		`"biddingStrategy":"AUTO"`,
		`"pricingModel":"CPM"`,
		`"paymentModel":"LOC"`,
		`"automatedKeywordsOptIn":true`,
		`"cpaGoal":{"amount":"2.50","currency":"USD"}`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %q", want, out)
		}
	}
}

func TestStructureExport_NoAdamID_OmitsCampaignAdamID(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--no-adam-id"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"adamId":900001`) {
		t.Fatalf("output should omit adamId when --no-adam-id is set: %q", out)
	}
	if !strings.Contains(out, `"name":"FitTrack US Search"`) || !strings.Contains(out, `"dailyBudgetAmount":{"amount":"5","currency":"USD"}`) {
		t.Fatalf("output should still include the rest of the campaign structure: %q", out)
	}
}

func TestStructureExport_NoTimes_StripsCampaignAndAdgroupTimes(t *testing.T) {
	past := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02T15:04:05.000")
	end := time.Now().UTC().Add(48 * time.Hour).Format("2006-01-02T15:04:05.000")

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"},"startTime":"` + past + `","endTime":"` + end + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"},"startTime":"` + past + `","endTime":"` + end + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--no-times"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"startTime":"`) || strings.Contains(out, `"endTime":"`) {
		t.Fatalf("output should omit campaign/adgroup times when --no-times is set: %q", out)
	}
}

func TestStructureExport_NoTimes_ExplicitFieldsKeepTimes(t *testing.T) {
	start := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02T15:04:05.000")
	end := time.Now().UTC().Add(48 * time.Hour).Format("2006-01-02T15:04:05.000")

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"},"startTime":"` + start + `","endTime":"` + end + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"},"startTime":"` + start + `","endTime":"` + end + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{
		"structure", "export", "--scope", "campaigns", "--no-times",
		"--campaigns-fields", "startTime,endTime",
		"--adgroups-fields", "",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"startTime":"`+start+`"`) || !strings.Contains(out, `"endTime":"`+end+`"`) {
		t.Fatalf("explicit field selection should keep times with --no-times: %q", out)
	}
}

func TestStructureExport_RedactNames_RedactsCampaignAppNameAndCountries(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack SE: fitness calories - DE,FR - Search","countriesOrRegions":["DE","FR"],"dailyBudgetAmount":{"amount":"5","currency":"EUR"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack SE: fitness calories"}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "EUR"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--redact-names"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"name":"%(appName) - %(countriesOrRegions) - Search"`) {
		t.Fatalf("expected redacted campaign name, got %q", out)
	}
}

func TestStructureExport_RedactNames_RedactsFuzzyAppNameShortAndAdgroupCampaignName(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack - DE - Discovery","countriesOrRegions":["DE"],"dailyBudgetAmount":{"amount":"5","currency":"EUR"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack SE: fitness calories"}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"FitTrack - DE - Discovery - Brand","defaultBidAmount":{"amount":"1.00","currency":"EUR"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "EUR"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--redact-names"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"name":"%(appNameShort) - %(countriesOrRegions) - Discovery"`) {
		t.Fatalf("expected fuzzy appNameShort redaction in campaign name, got %q", out)
	}
	if !strings.Contains(out, `"name":"%(campaignName) - Brand"`) {
		t.Fatalf("expected campaignName redaction in adgroup name, got %q", out)
	}
}

func TestStructureExport_RedactNames_UsesSharedAppNameShortSeparators(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack - DE - Discovery","countriesOrRegions":["DE"],"dailyBudgetAmount":{"amount":"5","currency":"EUR"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack / Search"}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "EUR"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--redact-names"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"name":"%(appNameShort) - %(countriesOrRegions) - Discovery"`) {
		t.Fatalf("expected slash-derived appNameShort redaction in campaign name, got %q", out)
	}
}

func TestStructureExport_RedactNames_FailsWhenAppLookupFails(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack - DE - Discovery","countriesOrRegions":["DE"],"dailyBudgetAmount":{"amount":"5","currency":"EUR"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/apps/900001":
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"error":{"errors":[{"message":"The referenced resource was not found."}]}}`)),
				}, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "EUR"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--redact-names"}, "")
	if code != ExitAPIError {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitAPIError, out)
	}
	if !strings.Contains(out, "referenced resource was not found") && !strings.Contains(out, "resource was not found") {
		t.Fatalf("expected app lookup failure in output, got %q", out)
	}
}

func TestStructureExport_NoBudgets_OmitsBudgetBidAndInvoiceFields(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"},"budgetAmount":{"amount":"100","currency":"USD"},"targetCpa":{"amount":"7.50","currency":"USD"},"locInvoiceDetails":{"orderNumber":"PO-123"},"budgetOrders":[11,12]}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"},"cpaGoal":{"amount":"2.50","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[{"id":900301,"campaignId":900101,"adGroupId":900201,"text":"fitness coach","matchType":"EXACT","bidAmount":{"amount":"2.00","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--no-budgets"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, unwanted := range []string{`"dailyBudgetAmount":`, `"budgetAmount":`, `"targetCpa":`, `"locInvoiceDetails":`, `"budgetOrders":`, `"defaultBidAmount":`, `"cpaGoal":`, `"bidAmount":`} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("output should omit %q when --no-budgets is set: %q", unwanted, out)
		}
	}
}

func TestStructureExport_Shareable_AppliesSharePreset(t *testing.T) {
	start := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02T15:04:05.000")
	end := time.Now().UTC().Add(48 * time.Hour).Format("2006-01-02T15:04:05.000")

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack SE: fitness calories - DE - Discovery","countriesOrRegions":["DE"],"dailyBudgetAmount":{"amount":"5","currency":"EUR"},"budgetAmount":{"amount":"100","currency":"EUR"},"targetCpa":{"amount":"7.50","currency":"EUR"},"locInvoiceDetails":{"orderNumber":"PO-123"},"budgetOrders":[11,12],"startTime":"` + start + `","endTime":"` + end + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack SE: fitness calories"}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"FitTrack SE: fitness calories - DE - Discovery - Brand","defaultBidAmount":{"amount":"1.00","currency":"EUR"},"cpaGoal":{"amount":"2.50","currency":"EUR"},"startTime":"` + start + `","endTime":"` + end + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				t.Fatalf("shareable export should skip negatives and keywords unless explicitly requested: %s", req.URL.Path)
				return nil, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "EUR"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--shareable"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, unwanted := range []string{`"adamId":`, `"dailyBudgetAmount":`, `"budgetAmount":`, `"targetCpa":`, `"locInvoiceDetails":`, `"budgetOrders":`, `"defaultBidAmount":`, `"cpaGoal":`, `"startTime":`, `"endTime":`, `"campaignNegativeKeywords"`, `"adgroupNegativeKeywords"`, `"keywords":[`} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("shareable output unexpectedly contains %q: %q", unwanted, out)
		}
	}
	for _, want := range []string{`"%(appName) - %(countriesOrRegions) - Discovery"`, `%(campaignName) - Brand`} {
		if !strings.Contains(out, want) {
			t.Fatalf("shareable output missing %q: %q", want, out)
		}
	}
}

func TestStructureExport_Shareable_ExplicitFieldsReincludeOmittedContent(t *testing.T) {
	start := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02T15:04:05.000")
	end := time.Now().UTC().Add(48 * time.Hour).Format("2006-01-02T15:04:05.000")

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack SE: fitness calories - DE - Discovery","countriesOrRegions":["DE"],"dailyBudgetAmount":{"amount":"5","currency":"EUR"},"budgetAmount":{"amount":"100","currency":"EUR"},"targetCpa":{"amount":"7.50","currency":"EUR"},"locInvoiceDetails":{"orderNumber":"PO-123"},"budgetOrders":[11,12],"startTime":"` + start + `","endTime":"` + end + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack SE: fitness calories"}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[{"id":900401,"campaignId":900101,"text":"free workout","matchType":"EXACT"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"FitTrack SE: fitness calories - DE - Discovery - Brand","defaultBidAmount":{"amount":"1.00","currency":"EUR"},"cpaGoal":{"amount":"2.50","currency":"EUR"},"startTime":"` + start + `","endTime":"` + end + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords":
				return jsonResponse(`{"data":[{"id":900402,"campaignId":900101,"adGroupId":900201,"text":"protein powder","matchType":"BROAD"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[{"id":900301,"campaignId":900101,"adGroupId":900201,"text":"fitness coach","matchType":"EXACT","bidAmount":{"amount":"2.00","currency":"EUR"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "EUR"})
	defer restore()

	out, code := captureRun(t, []string{
		"structure", "export", "--scope", "campaigns", "--shareable",
		"--campaigns-fields", "adamId,dailyBudgetAmount,budgetAmount,targetCpa,locInvoiceDetails,budgetOrders,startTime,endTime",
		"--adgroups-fields", "defaultBidAmount,cpaGoal,startTime,endTime",
		"--keywords-fields", "text,matchType,bidAmount",
		"--campaigns-negatives-fields", "text,matchType",
		"--adgroups-negatives-fields", "text,matchType",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		`"adamId":900001`,
		`"dailyBudgetAmount":{"amount":"5","currency":"EUR"}`,
		`"budgetAmount":{"amount":"100","currency":"EUR"}`,
		`"targetCpa":{"amount":"7.50","currency":"EUR"}`,
		`"locInvoiceDetails":{"orderNumber":"PO-123"}`,
		`"budgetOrders":[11,12]`,
		`"defaultBidAmount":{"amount":"1.00","currency":"EUR"}`,
		`"cpaGoal":{"amount":"2.50","currency":"EUR"}`,
		`"startTime":"` + start + `"`,
		`"endTime":"` + end + `"`,
		`"bidAmount":{"amount":"2.00","currency":"EUR"}`,
		`"campaignNegativeKeywords":[{"matchType":"EXACT","text":"free workout"}]`,
		`"adgroupNegativeKeywords":[{"matchType":"BROAD","text":"protein powder"}]`,
		`"keywords":[{"bidAmount":{"amount":"2.00","currency":"EUR"},"matchType":"EXACT","text":"fitness coach"}]`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("shareable output missing explicit override %q: %q", want, out)
		}
	}
}

func TestStructureExport_Pretty_PrintsIndentedJSONWhenStdoutIsNotTTY(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--pretty"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "\n  \"schemaVersion\": 1,\n") {
		t.Fatalf("expected pretty JSON output, got %q", out)
	}
}

func TestStructureExport_KeywordsFilter_UsesKeywordFindWithinSelectedAdgroups(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/targetingkeywords/find":
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading request body: %v", err)
				}
				if !bytes.Contains(body, []byte(`"field":"text"`)) || !bytes.Contains(body, []byte(`"fitness"`)) {
					t.Fatalf("selector body missing expected keyword filter: %s", body)
				}
				if !bytes.Contains(body, []byte(`"field":"adGroupId"`)) || !bytes.Contains(body, []byte(`"900201"`)) {
					t.Fatalf("selector body missing adGroupId condition: %s", body)
				}
				return jsonResponse(`{"data":[{"id":900301,"campaignId":900101,"adGroupId":900201,"text":"fitness coach","matchType":"EXACT","status":"ACTIVE"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--keywords-filter", "text CONTAINS fitness"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"text":"fitness coach"`) {
		t.Fatalf("expected filtered keyword in output, got %q", out)
	}
}

func TestStructureExport_NoNegatives_SkipsNegativeRequests(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[{"id":900301,"campaignId":900101,"adGroupId":900201,"text":"fitness coach","matchType":"EXACT"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/negativekeywords":
				t.Fatalf("negative endpoint should not be requested when --no-negatives is set: %s", req.URL.Path)
				return nil, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--no-negatives"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"campaignNegativeKeywords"`) || strings.Contains(out, `"adgroupNegativeKeywords"`) {
		t.Fatalf("output should not include negatives when --no-negatives is set: %q", out)
	}
	if !strings.Contains(out, `"text":"fitness coach"`) {
		t.Fatalf("output should still include keywords when --no-negatives is set: %q", out)
	}
}

func TestStructureExport_NoKeywords_SkipsKeywordRequests(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[{"id":900401,"campaignId":900101,"text":"free workout","matchType":"EXACT"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords":
				return jsonResponse(`{"data":[{"id":900402,"campaignId":900101,"adGroupId":900201,"text":"protein powder","matchType":"BROAD"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				t.Fatalf("keyword endpoint should not be requested when --no-keywords is set")
				return nil, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--no-keywords"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"keywords":[`) || strings.Contains(out, `"text":"fitness coach"`) {
		t.Fatalf("output should not include keywords when --no-keywords is set: %q", out)
	}
	if !strings.Contains(out, `"campaignNegativeKeywords"`) || !strings.Contains(out, `"adgroupNegativeKeywords"`) {
		t.Fatalf("output should still include negatives when --no-keywords is set: %q", out)
	}
}

func TestStructureExport_AdgroupTargetingDimensions_OmitsDefaultValues(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"},"targetingDimensions":{"adminArea":null,"age":null,"appCategories":null,"appDownloaders":null,"country":null,"daypart":null,"deviceClass":{"included":["IPHONE","IPAD"]},"gender":null,"locality":null}},{"id":900202,"campaignId":900101,"name":"FR Search","defaultBidAmount":{"amount":"1.00","currency":"USD"},"targetingDimensions":{"adminArea":null,"age":null,"appCategories":null,"appDownloaders":null,"country":{"included":["FR"]},"daypart":null,"deviceClass":{"included":["IPHONE","IPAD"]},"gender":null,"locality":null}}],"pagination":{"totalResults":2,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900101/adgroups/900202/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords", "/api/v5/campaigns/900101/adgroups/900202/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"deviceClass":{"included":["IPHONE","IPAD"]}`) {
		t.Fatalf("default deviceClass should be omitted from targetingDimensions: %q", out)
	}
	if strings.Contains(out, `"adminArea":null`) || strings.Contains(out, `"gender":null`) {
		t.Fatalf("null targetingDimensions fields should be omitted: %q", out)
	}
	if strings.Contains(out, `"name":"Core Search","targetingDimensions"`) {
		t.Fatalf("fully default targetingDimensions should be omitted: %q", out)
	}
	if !strings.Contains(out, `"name":"FR Search"`) || !strings.Contains(out, `"targetingDimensions":{"country":{"included":["FR"]}}`) {
		t.Fatalf("non-default targetingDimensions should retain only meaningful fields: %q", out)
	}
}

func TestStructureExport_AdgroupTargetingDimensions_PreservedWhenFieldsRequestFullExport(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack US Search","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"},"targetingDimensions":{"adminArea":null,"age":null,"appCategories":null,"appDownloaders":null,"country":null,"daypart":null,"deviceClass":{"included":["IPHONE","IPAD"]},"gender":null,"locality":null}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--adgroups-fields", ""}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"targetingDimensions":{"adminArea":null,"age":null,"appCategories":null,"appDownloaders":null,"country":null,"daypart":null,"deviceClass":{"included":["IPHONE","IPAD"]},"gender":null,"locality":null}`) {
		t.Fatalf("full adgroup field export should preserve raw targetingDimensions: %q", out)
	}
}

func TestStructureExport_OmitsPastStartTimeByDefault(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02T15:04:05.000")
	future := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02T15:04:05.000")

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"Past Campaign","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"},"startTime":"` + past + `"},{"id":900102,"adamId":900002,"name":"Future Campaign","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"6","currency":"USD"},"startTime":"` + future + `"}],"pagination":{"totalResults":2,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords", "/api/v5/campaigns/900102/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Past Group","defaultBidAmount":{"amount":"1.00","currency":"USD"},"startTime":"` + past + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900102/adgroups":
				return jsonResponse(`{"data":[{"id":900202,"campaignId":900102,"name":"Future Group","defaultBidAmount":{"amount":"1.00","currency":"USD"},"startTime":"` + future + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900102/adgroups/900202/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords", "/api/v5/campaigns/900102/adgroups/900202/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"name":"Past Campaign","startTime":"`+past+`"`) {
		t.Fatalf("past campaign startTime should be omitted by default: %q", out)
	}
	if strings.Contains(out, `"name":"Past Group","startTime":"`+past+`"`) {
		t.Fatalf("past adgroup startTime should be omitted by default: %q", out)
	}
	if !strings.Contains(out, `"name":"Future Campaign","startTime":"`+future+`"`) {
		t.Fatalf("future campaign startTime should be retained: %q", out)
	}
	if !strings.Contains(out, `"name":"Future Group","startTime":"`+future+`"`) {
		t.Fatalf("future adgroup startTime should be retained: %q", out)
	}
}

func TestStructureExport_PastStartTime_PreservedWhenFieldsRequested(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02T15:04:05.000")

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"Past Campaign","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"},"startTime":"` + past + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"campaignId":900101,"name":"Past Group","defaultBidAmount":{"amount":"1.00","currency":"USD"},"startTime":"` + past + `"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups/900201/negativekeywords", "/api/v5/campaigns/900101/adgroups/900201/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--campaigns-fields", "startTime", "--adgroups-fields", "startTime"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"name":"Past Campaign","startTime":"`+past+`"`) {
		t.Fatalf("requested campaign startTime should be preserved: %q", out)
	}
	if !strings.Contains(out, `"name":"Past Group","startTime":"`+past+`"`) {
		t.Fatalf("requested adgroup startTime should be preserved: %q", out)
	}
}

func TestStructureExport_CampaignIDFromStdin_AggregatesCampaigns(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/900101":
				return jsonResponse(`{"data":{"id":900101,"adamId":900001,"name":"Campaign One","countriesOrRegions":["US"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}}`), nil
			case "/api/v5/campaigns/900102":
				return jsonResponse(`{"data":{"id":900102,"adamId":900002,"name":"Campaign Two","countriesOrRegions":["FR"],"dailyBudgetAmount":{"amount":"6","currency":"USD"}}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords", "/api/v5/campaigns/900102/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups", "/api/v5/campaigns/900102/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	stdin := "CAMPAIGN_ID\n900101\n900102\n"
	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--campaign-id", "-"}, stdin)
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		`"type":"structure"`,
		`"name":"Campaign One"`,
		`"name":"Campaign Two"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %q", want, out)
		}
	}
}

func TestStructureExport_CampaignsFilter_RequestsRequiredFieldsForNestedExport(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/find":
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading request body: %v", err)
				}
				if !bytes.Contains(body, []byte(`"field":"name"`)) || !bytes.Contains(body, []byte(`"STARTSWITH"`)) {
					t.Fatalf("selector body missing expected filter: %s", body)
				}
				if !bytes.Contains(body, []byte(`"fields"`)) || !bytes.Contains(body, []byte(`"id"`)) || !bytes.Contains(body, []byte(`"adamId"`)) || !bytes.Contains(body, []byte(`"dailyBudgetAmount"`)) || !bytes.Contains(body, []byte(`"countriesOrRegions"`)) {
					t.Fatalf("selector body missing required export fields: %s", body)
				}
				return jsonResponse(`{"data":[{"id":900101,"adamId":900001,"name":"FitTrack SE - LU","countriesOrRegions":["LU"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/900101/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/<nil>/negativekeywords", "/api/v5/campaigns/<nil>/adgroups":
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"error":{"errors":[{"message":"The referenced resource was not found."}]}}`)),
				}, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--campaigns-filter", "name STARTSWITH FitTrack SE - LU"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"name":"FitTrack SE - LU"`) {
		t.Fatalf("expected filtered campaign in output, got %q", out)
	}
}

func TestStructureExport_PreservesLargeNestedIDs(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/find":
				return jsonResponse(`{"data":[{"id":1234567890,"adamId":900001,"name":"FitTrack SE - LU - Discovery","countriesOrRegions":["LU"],"dailyBudgetAmount":{"amount":"5","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/1234567890/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/1234567890/adgroups":
				return jsonResponse(`{"data":[{"id":9876543210,"campaignId":1234567890,"name":"Core Search","defaultBidAmount":{"amount":"1.00","currency":"USD"}}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/1234567890/adgroups/9876543210/negativekeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			case "/api/v5/campaigns/1234567890/adgroups/9876543210/targetingkeywords":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	out, code := captureRun(t, []string{"structure", "export", "--scope", "campaigns", "--campaigns-filter", "name STARTSWITH FitTrack SE - LU"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"name":"Core Search"`) {
		t.Fatalf("expected nested adgroup export, got %q", out)
	}
}

func TestStructureImport_Check_Adgroups_EmitsMappingJSON(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/500":
				return jsonResponse(`{"data":{"id":500,"name":"Destination Campaign","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			case "/api/v5/campaigns/500/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"Source Group","defaultBidAmount":{"amount":"1.20","currency":"USD"}},"adgroupNegativeKeywords":[{"text":"free","matchType":"EXACT"}],"keywords":[{"text":"brand","matchType":"EXACT"}]}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--check"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		`"type":"mapping"`,
		`"scope":"adgroups"`,
		`"created":{"id":1,"name":"Source Group"}`,
		`"created":{"id":2,"text":"free"}`,
		`"created":{"id":3,"text":"brand"}`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %q", want, out)
		}
	}
	if strings.Contains(out, "/api/v5/campaigns/500/adgroups/1") {
		t.Fatalf("check mode should not send mutating requests: %q", out)
	}
}

func TestStructureImport_AdgroupCPAGoalRejectedOnNonSearchCampaign(t *testing.T) {
	requests := 0
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.URL.Path == "/api/v5/campaigns/500" {
				return jsonResponse(`{"data":{"id":500,"name":"Destination","adChannelType":"DISPLAY","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			}
			t.Fatalf("unexpected request path: %s", req.URL.Path)
			return nil, nil
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"Source Group","defaultBidAmount":{"amount":"1.20","currency":"USD"},"cpaGoal":{"amount":"2.00","currency":"USD"}}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--check"}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for cpaGoal on non-SEARCH campaign; output=%q", out)
	}
	if !strings.Contains(out, "cpaGoal requires a SEARCH campaign") {
		t.Fatalf("expected error to use JSON label 'cpaGoal requires a SEARCH campaign', got %q", out)
	}
	if strings.Contains(out, "--cpa-goal") {
		t.Fatalf("structure import should not reference --cpa-goal flag label, got %q", out)
	}
	if !strings.Contains(out, "DISPLAY") {
		t.Fatalf("expected error to mention DISPLAY adChannelType, got %q", out)
	}
	if requests == 0 {
		t.Fatalf("expected campaign fetch to have happened")
	}
}

func TestStructureImport_AdgroupStartTimeFlagUsesImportFlagName(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/api/v5/campaigns/500" {
				return jsonResponse(`{"data":{"id":500,"name":"Destination","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			}
			t.Fatalf("unexpected request path: %s", req.URL.Path)
			return nil, nil
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"Source Group","defaultBidAmount":{"amount":"1.20","currency":"USD"}}}]}`
	out, code := captureRun(t, []string{
		"structure", "import",
		"--from-structure", structureJSON,
		"--campaign-id", "500",
		"--adgroups-start-time", "not-a-time",
		"--check",
	}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure for invalid adgroups start time; output=%q", out)
	}
	if !strings.Contains(out, "--adgroups-start-time") {
		t.Fatalf("expected import flag label in error, got %q", out)
	}
	if strings.Contains(out, "--start-time") {
		t.Fatalf("did not expect generic --start-time label, got %q", out)
	}
}

func TestStructureImport_Check_ReportsAllCollisions(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/500":
				return jsonResponse(`{"data":{"id":500,"name":"Destination Campaign","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			case "/api/v5/campaigns/500/adgroups":
				return jsonResponse(`{"data":[{"id":900201,"name":"Existing Name"}],"pagination":{"totalResults":1,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"Existing Name","defaultBidAmount":{"amount":"1.20","currency":"USD"}}},{"adgroup":{"name":"existing name","defaultBidAmount":{"amount":"1.20","currency":"USD"}}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--check"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	for _, want := range []string{
		`duplicate adgroup name in import batch for campaign Destination Campaign: Existing Name`,
		`adgroup name already exists in campaign Destination Campaign: Existing Name`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %q", want, out)
		}
	}
}

func TestStructureImport_OutputMapping_WritesFileAndSuppressesStdoutMapping(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/500":
				return jsonResponse(`{"data":{"id":500,"name":"Destination Campaign","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			case "/api/v5/campaigns/500/adgroups":
				if req.Method == http.MethodGet {
					return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
				}
				return jsonResponse(`{"data":{"id":700,"name":"Imported Group"}}`), nil
			case "/api/v5/campaigns/500/adgroups/700/negativekeywords/bulk":
				return jsonResponse(`{"data":[{"id":701,"text":"free"}]}`), nil
			case "/api/v5/campaigns/500/adgroups/700/targetingkeywords/bulk":
				return jsonResponse(`{"data":[{"id":702,"text":"brand"}]}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	dir := t.TempDir()
	mappingPath := filepath.Join(dir, "mapping.json")
	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"Imported Group","defaultBidAmount":{"amount":"1.20","currency":"USD"}},"adgroupNegativeKeywords":[{"text":"free","matchType":"EXACT"}],"keywords":[{"text":"brand","matchType":"EXACT"}]}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--output-mapping", mappingPath}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"type":"mapping"`) {
		t.Fatalf("stdout should not contain mapping JSON when --output-mapping is used: %q", out)
	}
	data, err := os.ReadFile(mappingPath)
	if err != nil {
		t.Fatalf("reading mapping file: %v", err)
	}
	for _, want := range []string{`"type":"mapping"`, `"created":{"id":700,"name":"Imported Group"}`, `"created":{"id":701,"text":"free"}`, `"created":{"id":702,"text":"brand"}`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("mapping file missing %q: %s", want, string(data))
		}
	}
}

func TestStructureImport_OutputMapping_PartialFailureTracksCreatedFailedAndNotAttempted(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				if req.Method == http.MethodGet {
					return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
				}
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading campaign create body: %v", err)
				}
				if bytes.Contains(body, []byte(`"name":"Campaign One"`)) {
					return jsonResponse(`{"data":{"id":700,"name":"Campaign One"}}`), nil
				}
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"error":{"errors":[{"message":"Second campaign failed."}]}}`)),
				}, nil
			case "/api/v5/campaigns/700/adgroups":
				return jsonResponse(`{"data":{"id":701,"name":"Group One"}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	dir := t.TempDir()
	mappingPath := filepath.Join(dir, "mapping.json")
	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"Campaign One","dailyBudgetAmount":{"amount":"10","currency":"USD"},"countriesOrRegions":["US"]},"adgroups":[{"adgroup":{"name":"Group One","defaultBidAmount":{"amount":"1.20","currency":"USD"}}}]},{"campaign":{"adamId":900002,"name":"Campaign Two","dailyBudgetAmount":{"amount":"12","currency":"USD"},"countriesOrRegions":["US"]},"adgroups":[{"adgroup":{"name":"Group Two","defaultBidAmount":{"amount":"1.50","currency":"USD"}}}]}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--output-mapping", mappingPath}, "")
	if code != ExitAPIError {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitAPIError, out)
	}
	data, err := os.ReadFile(mappingPath)
	if err != nil {
		t.Fatalf("reading mapping file: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		`"created":{"id":700,"name":"Campaign One"},"status":"created"`,
		`"created":{"id":701,"name":"Group One"},"status":"created"`,
		`"source":{"name":"Campaign Two"},"status":"failed","error":"api error: HTTP 400: Second campaign failed."`,
		`"source":{"name":"Group Two"},"status":"not_attempted"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("mapping file missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, `"source":{"name":"Group Two"},"created":`) {
		t.Fatalf("not_attempted items should not include created data: %s", got)
	}
}

func TestStructureImport_OutputMapping_BulkFailureMarksBatchFailedAndFollowingItemsNotAttempted(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/500":
				return jsonResponse(`{"data":{"id":500,"name":"Destination Campaign","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			case "/api/v5/campaigns/500/adgroups":
				if req.Method == http.MethodGet {
					return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
				}
				return jsonResponse(`{"data":{"id":700,"name":"Imported Group"}}`), nil
			case "/api/v5/campaigns/500/adgroups/700/negativekeywords/bulk":
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"error":{"errors":[{"message":"Negative keyword bulk failed."}]}}`)),
				}, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	dir := t.TempDir()
	mappingPath := filepath.Join(dir, "mapping.json")
	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"Imported Group","defaultBidAmount":{"amount":"1.20","currency":"USD"}},"adgroupNegativeKeywords":[{"text":"free","matchType":"EXACT"}],"keywords":[{"text":"brand","matchType":"EXACT"}]}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--output-mapping", mappingPath}, "")
	if code != ExitAPIError {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitAPIError, out)
	}
	data, err := os.ReadFile(mappingPath)
	if err != nil {
		t.Fatalf("reading mapping file: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		`"created":{"id":700,"name":"Imported Group"},"status":"created"`,
		`"source":{"text":"free"},"status":"failed","error":"api error: HTTP 400: Negative keyword bulk failed."`,
		`"source":{"text":"brand"},"status":"not_attempted"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("mapping file missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, `"source":{"text":"brand"},"created":`) {
		t.Fatalf("not_attempted keyword should not include created data: %s", got)
	}
}

func TestStructureImport_Pretty_PrintsIndentedJSONWhenStdoutIsNotTTY(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/500":
				return jsonResponse(`{"data":{"id":500,"name":"Destination Campaign","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			case "/api/v5/campaigns/500/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"Source Group","defaultBidAmount":{"amount":"1.20","currency":"USD"}}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--check", "--pretty"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "\n  \"schemaVersion\": 1,\n") {
		t.Fatalf("expected pretty JSON output, got %q", out)
	}
}

func TestStructureImport_Check_NoAdgroups_CreatesCampaignsOnly(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				if req.Method == http.MethodGet {
					return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
				}
				t.Fatalf("check mode should not send campaign create requests")
				return nil, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"Source Campaign","dailyBudgetAmount":{"amount":"10","currency":"USD"},"countriesOrRegions":["US"]},"campaignNegativeKeywords":[{"text":"free","matchType":"EXACT"}],"adgroups":[{"adgroup":{"name":"Group A","defaultBidAmount":{"amount":"1.20","currency":"USD"}},"adgroupNegativeKeywords":[{"text":"ignore me","matchType":"EXACT"}],"keywords":[{"text":"brand","matchType":"EXACT"}]}]}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--no-adgroups", "--check"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		`"type":"mapping"`,
		`"created":{"id":1,"name":"Source Campaign"}`,
		`"created":{"id":2,"text":"free"}`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %q", want, out)
		}
	}
	for _, unwanted := range []string{`"Group A"`, `"brand"`, `"ignore me"`} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("output unexpectedly contains %q: %q", unwanted, out)
		}
	}
}

func TestStructureImport_UsesCreateDefaultsWhenExportOmittedThem(t *testing.T) {
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 15, 4, 5, 0, time.UTC)
	})
	defer restoreNow()

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				if req.Method == http.MethodGet {
					return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
				}
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading campaign create body: %v", err)
				}
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("parsing campaign create body: %v", err)
				}
				if got["adChannelType"] != "SEARCH" {
					t.Fatalf("campaign adChannelType = %v, want SEARCH", got["adChannelType"])
				}
				if got["billingEvent"] != "TAPS" {
					t.Fatalf("campaign billingEvent = %v, want TAPS", got["billingEvent"])
				}
				supplySources, ok := got["supplySources"].([]any)
				if !ok || len(supplySources) != 1 || supplySources[0] != "APPSTORE_SEARCH_RESULTS" {
					t.Fatalf("campaign supplySources = %v, want [APPSTORE_SEARCH_RESULTS]", got["supplySources"])
				}
				return jsonResponse(`{"data":{"id":700,"name":"Imported Campaign"}}`), nil
			case "/api/v5/campaigns/700/adgroups":
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading adgroup create body: %v", err)
				}
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("parsing adgroup create body: %v", err)
				}
				if got["pricingModel"] != "CPC" {
					t.Fatalf("adgroup pricingModel = %v, want CPC", got["pricingModel"])
				}
				if got["automatedKeywordsOptIn"] != false {
					t.Fatalf("adgroup automatedKeywordsOptIn = %v, want false", got["automatedKeywordsOptIn"])
				}
				if got["startTime"] != "2026-03-25T09:30:00.000" {
					t.Fatalf("adgroup startTime = %v, want 2026-03-25T09:30:00.000", got["startTime"])
				}
				return jsonResponse(`{"data":{"id":701,"name":"Imported Group"}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:            "123",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "UTC",
		DefaultTimeOfDay: "09:30:00",
	})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"Imported Campaign","dailyBudgetAmount":{"amount":"10","currency":"USD"},"countriesOrRegions":["US"]},"adgroups":[{"adgroup":{"name":"Imported Group","defaultBidAmount":{"amount":"1.20","currency":"USD"}}}]}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestStructureImport_MissingCampaignRequiredFieldsSuggestsFlags(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--check"}, "")
	if code == ExitSuccess {
		t.Fatalf("expected missing required fields error; output=%q", out)
	}
	for _, want := range []string{
		"adamId (include it in the structure JSON or pass --adam-id)",
		"countriesOrRegions (include it in the structure JSON or pass --countries-or-regions)",
		"dailyBudgetAmount (include it in the structure JSON or pass --daily-budget-amount)",
		"name (include it in the structure JSON or pass --campaigns-name)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got %q", want, out)
		}
	}
}

func TestStructureImport_MissingAdgroupRequiredFieldsSuggestsFlags(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/500":
				return jsonResponse(`{"data":{"id":500,"name":"Destination Campaign","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--check"}, "")
	if code == ExitSuccess {
		t.Fatalf("expected missing required fields error; output=%q", out)
	}
	for _, want := range []string{
		"defaultBidAmount (include it in the structure JSON or pass --default-bid)",
		"name (include it in the structure JSON or pass --adgroups-name)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got %q", want, out)
		}
	}
}

func TestStructureImport_Check_CampaignNameTemplateSupportsSyntheticAppFields(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack Pro"}}`), nil
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"Brand Search","dailyBudgetAmount":{"amount":"10","currency":"USD"},"countriesOrRegions":["US"]}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaigns-name", "%(appName) - %(name)", "--check"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"created":{"id":1,"name":"FitTrack Pro - Brand Search"}`) {
		t.Fatalf("expected rendered campaign name with synthetic appName, got %q", out)
	}
}

func TestStructureImport_Check_CampaignSourceNameTemplateIsRendered(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack SE: fitness calories"}}`), nil
			case "/api/v5/campaigns":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "EUR"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"%(appNameShort) - %(countriesOrRegions)","dailyBudgetAmount":{"amount":"10","currency":"EUR"},"countriesOrRegions":["DE"]}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--countries-or-regions", "DE", "--check"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"created":{"id":1,"name":"FitTrack SE - DE"}`) {
		t.Fatalf("expected rendered campaign source name template with appNameShort, got %q", out)
	}
}

func TestStructureImport_Check_AdgroupNameTemplateSupportsSyntheticAppAndCampaignFields(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/500":
				return jsonResponse(`{"data":{"id":500,"name":"Destination Campaign","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack Pro"}}`), nil
			case "/api/v5/campaigns/500/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"Source Group","defaultBidAmount":{"amount":"1.20","currency":"USD"}}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--adgroups-name", "%(appName) - %(campaignName) - %(name)", "--check"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"created":{"id":1,"name":"FitTrack Pro - Destination Campaign - Source Group"}`) {
		t.Fatalf("expected rendered adgroup name with synthetic appName/campaignName, got %q", out)
	}
}

func TestStructureImport_Check_AdgroupSourceNameTemplateIsRendered(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns/500":
				return jsonResponse(`{"data":{"id":500,"name":"Destination Campaign","adChannelType":"SEARCH","countriesOrRegions":["US"],"billingEvent":"TAPS","dailyBudgetAmount":{"amount":"10","currency":"USD"},"adamId":900001}}`), nil
			case "/api/v5/apps/900001":
				return jsonResponse(`{"data":{"adamId":900001,"appName":"FitTrack Pro"}}`), nil
			case "/api/v5/campaigns/500/adgroups":
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{OrgID: "123", DefaultCurrency: "USD"})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"adgroups","creationTime":"2026-03-31T00:00:00Z","adgroups":[{"adgroup":{"name":"%(appNameShort) - %(campaignName) - Base","defaultBidAmount":{"amount":"1.20","currency":"USD"}}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--campaign-id", "500", "--check"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"created":{"id":1,"name":"FitTrack Pro - Destination Campaign - Base"}`) {
		t.Fatalf("expected rendered adgroup source name template, got %q", out)
	}
}

func TestStructureImport_ResolvesRelativeTimesFromStructureJSON(t *testing.T) {
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 15, 4, 5, 0, time.UTC)
	})
	defer restoreNow()

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				if req.Method == http.MethodGet {
					return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
				}
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading campaign create body: %v", err)
				}
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("parsing campaign create body: %v", err)
				}
				if got["startTime"] != "2026-06-25T09:30:00.000" {
					t.Fatalf("campaign startTime = %v, want 2026-06-25T09:30:00.000", got["startTime"])
				}
				if got["endTime"] != "2026-07-25T09:30:00.000" {
					t.Fatalf("campaign endTime = %v, want 2026-07-25T09:30:00.000", got["endTime"])
				}
				return jsonResponse(`{"data":{"id":700,"name":"Imported Campaign"}}`), nil
			case "/api/v5/campaigns/700/adgroups":
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading adgroup create body: %v", err)
				}
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("parsing adgroup create body: %v", err)
				}
				if got["startTime"] != "2026-03-25T09:30:00.000" {
					t.Fatalf("adgroup startTime = %v, want 2026-03-25T09:30:00.000", got["startTime"])
				}
				if got["endTime"] != "2026-04-25T09:30:00.000" {
					t.Fatalf("adgroup endTime = %v, want 2026-04-25T09:30:00.000", got["endTime"])
				}
				return jsonResponse(`{"data":{"id":701,"name":"Imported Group"}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:            "123",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "UTC",
		DefaultTimeOfDay: "09:30:00",
	})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"Imported Campaign","dailyBudgetAmount":{"amount":"10","currency":"USD"},"countriesOrRegions":["US"],"startTime":"+3mo","endTime":"+4mo"},"adgroups":[{"adgroup":{"name":"Imported Group","defaultBidAmount":{"amount":"1.20","currency":"USD"},"startTime":"now","endTime":"+1mo"}}]}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestStructureImport_FailsWhenResolvedEndTimePrecedesStartTime(t *testing.T) {
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 15, 4, 5, 0, time.UTC)
	})
	defer restoreNow()

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/campaigns":
				if req.Method == http.MethodGet {
					return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
				}
				t.Fatalf("unexpected mutating request path: %s", req.URL.Path)
				return nil, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:            "123",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "UTC",
		DefaultTimeOfDay: "09:30:00",
	})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"Imported Campaign","dailyBudgetAmount":{"amount":"10","currency":"USD"},"countriesOrRegions":["US"],"startTime":"+3mo","endTime":"now"}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--check"}, "")
	if code == ExitSuccess {
		t.Fatalf("expected failure when endTime resolves before startTime; output=%q", out)
	}
	if !strings.Contains(out, "campaign endTime must not be earlier than startTime") {
		t.Fatalf("expected schedule ordering error, got %q", out)
	}
}

func TestStructureImport_TargetCPARejectsDisplayCampaign(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/api/v5/campaigns" && req.Method == http.MethodGet {
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			}
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "USD",
	})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"Imported Campaign","adChannelType":"DISPLAY","dailyBudgetAmount":{"amount":"10","currency":"USD"},"targetCpa":{"amount":"5","currency":"USD"},"countriesOrRegions":["US"]}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--check"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "targetCpa is supported only for SEARCH campaigns") {
		t.Fatalf("expected targetCpa validation error, got %q", out)
	}
}

func TestStructureImport_TargetCPASafetyLimit(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/api/v5/campaigns" && req.Method == http.MethodGet {
				return jsonResponse(`{"data":[],"pagination":{"totalResults":0,"startIndex":0,"itemsPerPage":1000}}`), nil
			}
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
			return nil, nil
		}),
	})
	restore := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "USD",
		MaxCPAGoal:      config.DecimalText("4"),
	})
	defer restore()

	structureJSON := `{"schemaVersion":1,"type":"structure","scope":"campaigns","creationTime":"2026-03-31T00:00:00Z","campaigns":[{"campaign":{"adamId":900001,"name":"Imported Campaign","dailyBudgetAmount":{"amount":"10","currency":"USD"},"targetCpa":{"amount":"5","currency":"USD"},"countriesOrRegions":["US"]}}]}`
	out, code := captureRun(t, []string{"structure", "import", "--from-structure", structureJSON, "--check"}, "")
	if code != ExitSafetyLimit {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSafetyLimit, out)
	}
	if !strings.Contains(out, "exceeds limit") {
		t.Fatalf("expected safety limit error, got %q", out)
	}
}
