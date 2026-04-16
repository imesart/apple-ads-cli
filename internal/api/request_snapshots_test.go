package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/imesart/apple-ads-cli/internal/api"
	aclsReq "github.com/imesart/apple-ads-cli/internal/api/requests/acls"
	adRejectionsReq "github.com/imesart/apple-ads-cli/internal/api/requests/ad_rejections"
	adgroupsReq "github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	adsReq "github.com/imesart/apple-ads-cli/internal/api/requests/ads"
	appsReq "github.com/imesart/apple-ads-cli/internal/api/requests/apps"
	budgetordersReq "github.com/imesart/apple-ads-cli/internal/api/requests/budgetorders"
	campaignsReq "github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	creativesReq "github.com/imesart/apple-ads-cli/internal/api/requests/creatives"
	geoReq "github.com/imesart/apple-ads-cli/internal/api/requests/geo"
	impressionShareReq "github.com/imesart/apple-ads-cli/internal/api/requests/impression_share"
	keywordsReq "github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	negAdgroupReq "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_adgroup"
	negCampaignReq "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_campaign"
	productPagesReq "github.com/imesart/apple-ads-cli/internal/api/requests/product_pages"
	reportsReq "github.com/imesart/apple-ads-cli/internal/api/requests/reports"
	"github.com/imesart/apple-ads-cli/internal/testutil"
)

type snapshotBuilder func(t *testing.T, snapshot *testutil.HTTPSnapshot) (string, error)

func TestRequestSnapshots(t *testing.T) {
	baseDir := filepath.Join("testdata", "request_snapshots")
	builders := snapshotBuilders()

	files, err := goldenFiles(baseDir)
	if err != nil {
		t.Fatalf("listing golden files: %v", err)
	}

	seen := make(map[string]bool, len(files))
	for _, rel := range files {
		builder, ok := builders[rel]
		if !ok {
			t.Fatalf("missing snapshot builder for %s", rel)
		}

		t.Run(rel, func(t *testing.T) {
			snapshot, err := testutil.ReadHTTPSnapshot(filepath.Join(baseDir, rel))
			if err != nil {
				t.Fatalf("reading snapshot: %v", err)
			}
			seen[rel] = true

			got, err := builder(t, snapshot)
			if err != nil {
				t.Fatalf("building snapshot: %v", err)
			}
			testutil.AssertGoldenSnapshot(t, got, filepath.Join(baseDir, rel))
		})
	}

	for rel := range builders {
		if !seen[rel] {
			t.Fatalf("snapshot builder without golden file: %s", rel)
		}
	}
}

func snapshotBuilders() map[string]snapshotBuilder {
	return map[string]snapshotBuilder{
		"context/access_token.golden": authTokenBuilder(),
		"context/me_details.golden":   apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request { return aclsReq.MeRequest{} }),
		"context/user_acl.golden":     apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request { return aclsReq.ListRequest{} }),

		"campaigns/create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return campaignsReq.CreateRequest{RawBody: rawBody(s)}
		}),
		"campaigns/update.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return campaignsReq.UpdateRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"campaigns/delete.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return campaignsReq.DeleteRequest{CampaignID: segment(t, s, 1)}
		}),
		"campaigns/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return campaignsReq.GetRequest{CampaignID: segment(t, s, 1)}
		}),
		"campaigns/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return campaignsReq.ListRequest{Limit: queryInt(t, s, "limit"), Offset: queryInt(t, s, "offset")}
		}),
		"campaigns/find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return campaignsReq.FindRequest{RawBody: rawBody(s)}
		}),

		"ad_groups/create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adgroupsReq.CreateRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"ad_groups/delete.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adgroupsReq.DeleteRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3)}
		}),
		"ad_groups/find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adgroupsReq.FindRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"ad_groups/find_org_wide.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adgroupsReq.FindAllRequest{RawBody: rawBody(s)}
		}),
		"ad_groups/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adgroupsReq.GetRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3)}
		}),
		"ad_groups/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adgroupsReq.ListRequest{CampaignID: segment(t, s, 1), Limit: queryInt(t, s, "limit"), Offset: queryInt(t, s, "offset")}
		}),
		"ad_groups/update.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adgroupsReq.UpdateRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),

		"ads/create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adsReq.CreateRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),
		"ads/delete.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adsReq.DeleteRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), AdID: segment(t, s, 5)}
		}),
		"ads/find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adsReq.FindRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),
		"ads/find_org_wide.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adsReq.FindAllRequest{RawBody: rawBody(s)}
		}),
		"ads/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adsReq.GetRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), AdID: segment(t, s, 5)}
		}),
		"ads/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adsReq.ListRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), Limit: queryInt(t, s, "limit"), Offset: queryInt(t, s, "offset")}
		}),
		"ads/update.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adsReq.UpdateRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), AdID: segment(t, s, 5), RawBody: rawBody(s)}
		}),

		"targeting_keywords/create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return keywordsReq.CreateRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),
		"targeting_keywords/update.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return keywordsReq.UpdateRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),
		"targeting_keywords/delete.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return keywordsReq.DeleteOneRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), KeywordID: segment(t, s, 5)}
		}),
		"targeting_keywords/delete_bulk.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return keywordsReq.DeleteBulkRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),
		"targeting_keywords/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return keywordsReq.GetRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), KeywordID: segment(t, s, 5)}
		}),
		"targeting_keywords/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return keywordsReq.ListRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), Limit: queryInt(t, s, "limit"), Offset: queryInt(t, s, "offset")}
		}),
		"targeting_keywords/find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return keywordsReq.FindRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),

		"campaign_keywords/create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negCampaignReq.CreateRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"campaign_keywords/update.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negCampaignReq.UpdateRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"campaign_keywords/delete.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negCampaignReq.DeleteBulkRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"campaign_keywords/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negCampaignReq.GetRequest{CampaignID: segment(t, s, 1), KeywordID: segment(t, s, 3)}
		}),
		"campaign_keywords/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negCampaignReq.ListRequest{CampaignID: segment(t, s, 1), Limit: queryInt(t, s, "limit"), Offset: queryInt(t, s, "offset")}
		}),
		"campaign_keywords/find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negCampaignReq.FindRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),

		"ad_group_keywords/create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negAdgroupReq.CreateRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),
		"ad_group_keywords/update.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negAdgroupReq.UpdateRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),
		"ad_group_keywords/delete.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negAdgroupReq.DeleteBulkRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), RawBody: rawBody(s)}
		}),
		"ad_group_keywords/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negAdgroupReq.GetRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), KeywordID: segment(t, s, 5)}
		}),
		"ad_group_keywords/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negAdgroupReq.ListRequest{CampaignID: segment(t, s, 1), AdGroupID: segment(t, s, 3), Limit: queryInt(t, s, "limit"), Offset: queryInt(t, s, "offset")}
		}),
		"ad_group_keywords/find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return negAdgroupReq.FindRequest{CampaignID: segment(t, s, 1), RawBody: rawBody(s)}
		}),

		"creatives/create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return creativesReq.CreateRequest{RawBody: rawBody(s)}
		}),
		"creatives/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return creativesReq.GetRequest{
				CreativeID:                      segment(t, s, 1),
				IncludeDeletedCreativeSetAssets: s.URL.Query().Get("includeDeletedCreativeSetAssets") == "true",
			}
		}),
		"creatives/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return creativesReq.ListRequest{Limit: queryInt(t, s, "limit"), Offset: queryInt(t, s, "offset")}
		}),
		"creatives/find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return creativesReq.FindRequest{RawBody: rawBody(s)}
		}),

		"budget_orders/create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return budgetordersReq.CreateRequest{RawBody: rawBody(s)}
		}),
		"budget_orders/update.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return budgetordersReq.UpdateRequest{BudgetOrderID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"budget_orders/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return budgetordersReq.GetRequest{BudgetOrderID: segment(t, s, 1)}
		}),
		"budget_orders/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return budgetordersReq.ListRequest{Limit: queryInt(t, s, "limit"), Offset: queryInt(t, s, "offset")}
		}),

		"custom_product_pages/app_preview_device_sizes.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return productPagesReq.DevicesRequest{}
		}),
		"custom_product_pages/product_page_locales.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return productPagesReq.LocalesRequest{AdamID: segment(t, s, 1), ProductPageID: segment(t, s, 3)}
		}),
		"custom_product_pages/supported_countries_or_regions.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return productPagesReq.CountriesRequest{CountriesOrRegions: s.URL.Query().Get("countriesOrRegions")}
		}),
		"custom_product_pages/supported_countries_or_regions_no_query.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return productPagesReq.CountriesRequest{}
		}),
		"custom_product_pages/product_page.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return productPagesReq.GetRequest{AdamID: segment(t, s, 1), ProductPageID: segment(t, s, 3)}
		}),
		"custom_product_pages/product_pages.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return productPagesReq.ListRequest{
				AdamID: segment(t, s, 1),
				Name:   s.URL.Query().Get("name"),
				State:  s.URL.Query().Get("state"),
				Limit:  queryInt(t, s, "limit"),
				Offset: queryInt(t, s, "offset"),
			}
		}),

		"ad_rejections/app_assets_find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adRejectionsReq.FindAssetsRequest{AdamID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"ad_rejections/get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adRejectionsReq.GetRequest{ID: segment(t, s, 1)}
		}),
		"ad_rejections/find.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return adRejectionsReq.FindRequest{RawBody: rawBody(s)}
		}),

		"apps/eligibility.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return appsReq.EligibilityRequest{AdamID: segment(t, s, 1), RawBody: rawBody(s)}
		}),
		"apps/search.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return appsReq.SearchRequest{
				SearchQuery:     s.URL.Query().Get("query"),
				ReturnOwnedApps: s.URL.Query().Get("returnOwnedApps") == "true",
				Limit:           queryInt(t, s, "limit"),
				Offset:          queryInt(t, s, "offset"),
			}
		}),
		"apps/details.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return appsReq.DetailsRequest{AdamID: segment(t, s, 1)}
		}),
		"apps/locale_details.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return appsReq.LocalizedRequest{AdamID: segment(t, s, 1)}
		}),

		"geolocations/list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			body := decodeJSONArray(t, s.Body)
			if len(body) == 0 {
				t.Fatal("expected geo list body")
			}
			first := body[0]
			return geoReq.GetRequest{
				ID:     stringField(t, first, "id"),
				Entity: stringField(t, first, "entity"),
				Limit:  queryInt(t, s, "limit"),
				Offset: queryInt(t, s, "offset"),
			}
		}),
		"geolocations/search.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return geoReq.SearchRequest{
				SearchQuery: s.URL.Query().Get("query"),
				Entity:      s.URL.Query().Get("entity"),
				CountryCode: firstQueryValue(s.URL.Query(), "countryCode", "countrycode"),
				Limit:       queryInt(t, s, "limit"),
				Offset:      queryInt(t, s, "offset"),
			}
		}),

		"reports/ad_group_report.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return reportsReq.AdGroupsRequest{CampaignID: segment(t, s, 2), RawBody: rawBody(s)}
		}),
		"reports/ad_report.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return reportsReq.AdsRequest{CampaignID: segment(t, s, 2), RawBody: rawBody(s)}
		}),
		"reports/campaign_report.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return reportsReq.CampaignsRequest{RawBody: rawBody(s)}
		}),
		"reports/keyword_report.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return reportsReq.KeywordsRequest{CampaignID: segment(t, s, 2), AdGroupID: segment(t, s, 4), RawBody: rawBody(s)}
		}),
		"reports/keyword_campaign_wide_report.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return reportsReq.KeywordsRequest{CampaignID: segment(t, s, 2), RawBody: rawBody(s)}
		}),
		"reports/search_term_report.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return reportsReq.SearchTermsRequest{CampaignID: segment(t, s, 2), AdGroupID: segment(t, s, 4), RawBody: rawBody(s)}
		}),
		"reports/search_term_campaign_wide_report.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return reportsReq.SearchTermsRequest{CampaignID: segment(t, s, 2), RawBody: rawBody(s)}
		}),
		"reports/impression_share_report_create.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return impressionShareReq.CreateRequest{RawBody: rawBody(s)}
		}),
		"reports/impression_share_report_get.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return impressionShareReq.GetRequest{ReportID: segment(t, s, 1)}
		}),
		"reports/impression_share_report_list.golden": apiBuilder(func(t *testing.T, s *testutil.HTTPSnapshot) api.Request {
			return impressionShareReq.ListRequest{
				Field:     s.URL.Query().Get("field"),
				SortOrder: s.URL.Query().Get("sortOrder"),
				Limit:     queryInt(t, s, "limit"),
				Offset:    queryInt(t, s, "offset"),
			}
		}),
	}
}

func apiBuilder(build func(t *testing.T, snapshot *testutil.HTTPSnapshot) api.Request) snapshotBuilder {
	return func(t *testing.T, snapshot *testutil.HTTPSnapshot) (string, error) {
		t.Helper()
		client := testutil.NewSnapshotAPIClient()
		return testutil.CaptureAPIRequestSnapshot(context.Background(), client, build(t, snapshot))
	}
}

func authTokenBuilder() snapshotBuilder {
	return func(t *testing.T, snapshot *testutil.HTTPSnapshot) (string, error) {
		t.Helper()
		form := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"client id"},
			"client_secret": {"client secret"},
			"scope":         {"searchadsorg"},
		}
		req, err := http.NewRequest(http.MethodPost, "https://appleid.apple.com/auth/oauth2/token", strings.NewReader(form.Encode()))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return testutil.FormatHTTPRequestSnapshot(req.Method, req.URL, req.Header, []byte(form.Encode()))
	}
}

func goldenFiles(baseDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".golden" {
			return nil
		}
		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func rawBody(snapshot *testutil.HTTPSnapshot) json.RawMessage {
	if len(snapshot.Body) == 0 {
		return nil
	}
	return json.RawMessage(snapshot.Body)
}

func segment(t *testing.T, snapshot *testutil.HTTPSnapshot, index int) string {
	t.Helper()
	segments := apiSegments(snapshot.URL)
	if index >= len(segments) {
		t.Fatalf("path %s missing segment %d", snapshot.URL.Path, index)
	}
	return segments[index]
}

func apiSegments(u *url.URL) []string {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "api" && strings.HasPrefix(parts[1], "v") {
		return parts[2:]
	}
	return parts
}

func queryInt(t *testing.T, snapshot *testutil.HTTPSnapshot, key string) int {
	t.Helper()
	value := snapshot.URL.Query().Get(key)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("invalid int query %s=%q: %v", key, value, err)
	}
	return parsed
}

func decodeJSONArray(t *testing.T, body []byte) []map[string]any {
	t.Helper()
	var decoded []map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decoding JSON array: %v", err)
	}
	return decoded
}

func stringField(t *testing.T, value map[string]any, key string) string {
	t.Helper()
	got, ok := value[key].(string)
	if !ok {
		t.Fatalf("field %s missing or not string in %#v", key, value)
	}
	return got
}

func firstQueryValue(values url.Values, keys ...string) string {
	for _, key := range keys {
		if got := values.Get(key); got != "" {
			return got
		}
	}
	return ""
}
