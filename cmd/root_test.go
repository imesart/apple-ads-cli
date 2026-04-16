package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/imesart/apple-ads-cli/internal/api"
	apiPkg "github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/config"
)

const (
	testAdamID     = "900001"
	testCampaignID = "900101"
	testAdGroupID  = "900201"
)

func TestRun_Version(t *testing.T) {
	out, code := captureRun(t, []string{"--version"}, "")
	if code != ExitSuccess {
		t.Errorf("Run(--version) = %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(out, "1.0.0-test") {
		t.Fatalf("version output missing CLI version: %q", out)
	}
	if !strings.Contains(out, "Target API: Apple Ads Campaign Management API v5.5") {
		t.Fatalf("version output missing target API: %q", out)
	}
}

func TestRun_VersionCommand_IncludesTargetAPI(t *testing.T) {
	out, code := captureRun(t, []string{"version"}, "")
	if code != ExitSuccess {
		t.Fatalf("Run(version) = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "1.0.0-test") {
		t.Fatalf("version command output missing CLI version: %q", out)
	}
	if !strings.Contains(out, "Target API: Apple Ads Campaign Management API v5.5") {
		t.Fatalf("version command output missing target API: %q", out)
	}
}

func TestRun_Help(t *testing.T) {
	out, code := captureRun(t, []string{"--help"}, "")
	if code != ExitSuccess {
		t.Errorf("Run(--help) = %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(out, "aads <subcommand>") {
		t.Errorf("Run(--help) output missing root usage: %q", out)
	}
}

func TestRun_Help_LeafCommand(t *testing.T) {
	out, code := captureRun(t, []string{"campaigns", "list", "--help"}, "")
	if code != ExitSuccess {
		t.Errorf("Run(campaigns list --help) = %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(out, "aads campaigns list") {
		t.Errorf("Run(campaigns list --help) output missing leaf usage: %q", out)
	}
}

func TestRun_Help_ShortFlag(t *testing.T) {
	out, code := captureRun(t, []string{"-h"}, "")
	if code != ExitSuccess {
		t.Errorf("Run(-h) = %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(out, "aads <subcommand>") {
		t.Errorf("Run(-h) output missing root usage: %q", out)
	}
}

func TestRun_Help_CampaignsCreate_NameFlagMentionsTemplates(t *testing.T) {
	out, code := captureRun(t, []string{"help", "campaigns", "create"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "Campaign name (accepts template variables") {
		t.Fatalf("campaigns create help missing template-variable guidance on --name: %q", out)
	}
}

func TestRun_Help_BudgetAmountFlagsAreDeprecated(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "campaigns create",
			args: []string{"help", "campaigns", "create"},
			want: "DEPRECATED: Total budget",
		},
		{
			name: "campaigns update",
			args: []string{"help", "campaigns", "update"},
			want: "DEPRECATED: Total budget",
		},
		{
			name: "structure import",
			args: []string{"help", "structure", "import"},
			want: "DEPRECATED: Override budgetAmount for created campaigns",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, code := captureRun(t, tt.args, "")
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			if !strings.Contains(out, tt.want) {
				t.Fatalf("%s help missing deprecated budget amount wording %q: %q", tt.name, tt.want, out)
			}
		})
	}
}

func TestRun_Help_CampaignsCreate_DoesNotRepeatFlagSemantics(t *testing.T) {
	out, code := captureRun(t, []string{"help", "campaigns", "create"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, "Status accepts:") {
		t.Fatalf("campaigns create help should not repeat status flag semantics: %q", out)
	}
	if strings.Contains(out, "Money amounts accept:") {
		t.Fatalf("campaigns create help should not repeat money flag semantics: %q", out)
	}
	if strings.Contains(out, "Time flags accept") {
		t.Fatalf("campaigns create help should not repeat time flag semantics: %q", out)
	}
	if !strings.Contains(out, "Start time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)") {
		t.Fatalf("campaigns create help missing UTC time flag wording: %q", out)
	}
}

func TestRun_Help_KeywordsDelete_ExplainsBodyAndConfirm(t *testing.T) {
	out, code := captureRun(t, []string{"help", "keywords", "delete"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "Use --keyword-id for one ID or a comma-separated list of IDs.") {
		t.Fatalf("keywords delete help missing --keyword-id guidance: %q", out)
	}
	if !strings.Contains(out, "The body is a JSON array of keyword IDs to delete.") {
		t.Fatalf("keywords delete help missing body-shape guidance: %q", out)
	}
	if !strings.Contains(out, "Requires --confirm to execute.") {
		t.Fatalf("keywords delete help missing confirm guidance: %q", out)
	}
}

func TestRun_Help_NegativesDelete_DoesNotSayBulk(t *testing.T) {
	out, code := captureRun(t, []string{"help", "negatives", "delete"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(strings.ToLower(out), "bulk") {
		t.Fatalf("negatives delete help should not describe the command as bulk: %q", out)
	}
	if !strings.Contains(out, "Use --keyword-id for one ID or a comma-separated list of IDs.") {
		t.Fatalf("negatives delete help missing --keyword-id guidance: %q", out)
	}
}

func TestRun_Help_GeoGet_ExplainsNoInputFlags(t *testing.T) {
	out, code := captureRun(t, []string{"help", "geo", "get"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "Get geolocation details for a specific geo identifier.") {
		t.Fatalf("geo get help missing summary text: %q", out)
	}
	if !strings.Contains(out, "--entity  Country | AdminArea | Locality") {
		t.Fatalf("geo get help missing entity guidance: %q", out)
	}
	if !strings.Contains(out, "--geo-id  Geo identifier") {
		t.Fatalf("geo get help missing geo-id guidance: %q", out)
	}
}

func TestRun_GeoGet_BuildsGeoRequestBody(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/search/geo" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			if req.Method != http.MethodPost {
				t.Fatalf("unexpected request method: %s", req.Method)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"id":"US|CA|San Francisco"`)) {
				t.Fatalf("geo get body missing id: %s", body)
			}
			if !bytes.Contains(body, []byte(`"entity":"locality"`)) {
				t.Fatalf("geo get body missing entity: %s", body)
			}
			if req.URL.Query().Get("limit") != "5" {
				t.Fatalf("limit = %q, want 5", req.URL.Query().Get("limit"))
			}
			return jsonResponse(`{"data":[{"id":"US|CA|San Francisco","entity":"Locality","displayName":"San Francisco"}]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"geo", "get", "--entity", "Locality", "--geo-id", "US|CA|San Francisco", "--limit", "5", "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"displayName":"San Francisco"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_GeoGet_RequiresEntityAndGeoID(t *testing.T) {
	out, code := captureRun(t, []string{"geo", "get", "--geo-id", "US"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "--entity is required") {
		t.Fatalf("unexpected output: %q", out)
	}

	out, code = captureRun(t, []string{"geo", "get", "--entity", "Country"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "--geo-id is required") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_GeoGet_ValidatesEntity(t *testing.T) {
	out, code := captureRun(t, []string{"geo", "get", "--entity", "planet", "--geo-id", "US"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "--entity must be one of: Country, AdminArea, Locality") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_GeoSearch_LimitZeroFetchesAllPages(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	calls := 0
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			if req.URL.Path != "/api/v5/search/geo" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			if req.Method != http.MethodGet {
				t.Fatalf("unexpected request method: %s", req.Method)
			}
			q := req.URL.Query()
			if q.Get("limit") != "1000" {
				t.Fatalf("limit = %q, want 1000", q.Get("limit"))
			}
			switch q.Get("offset") {
			case "0":
				return jsonResponse(`{"data":[{"id":"geo-1","displayName":"Luxembourg"}],"pagination":{"startIndex":0,"itemsPerPage":1000,"totalResults":1001}}`), nil
			case "1000":
				return jsonResponse(`{"data":[{"id":"geo-2","displayName":"Bruges"}],"pagination":{"startIndex":1000,"itemsPerPage":1,"totalResults":1001}}`), nil
			default:
				t.Fatalf("unexpected offset: %q", q.Get("offset"))
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"geo", "search", "--query", "bru", "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if !strings.Contains(out, `"displayName":"Luxembourg"`) || !strings.Contains(out, `"displayName":"Bruges"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_AppsSearch_LimitZeroFetchesAllPages(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	calls := 0
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			if req.URL.Path != "/api/v5/search/apps" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			if req.Method != http.MethodGet {
				t.Fatalf("unexpected request method: %s", req.Method)
			}
			q := req.URL.Query()
			if q.Get("limit") != "1000" {
				t.Fatalf("limit = %q, want 1000", q.Get("limit"))
			}
			switch q.Get("offset") {
			case "0":
				return jsonResponse(`{"data":[{"adamId":1,"name":"FitTrack"}],"pagination":{"startIndex":0,"itemsPerPage":1000,"totalResults":1001}}`), nil
			case "1000":
				return jsonResponse(`{"data":[{"adamId":2,"name":"StrideCoach"}],"pagination":{"startIndex":1000,"itemsPerPage":1,"totalResults":1001}}`), nil
			default:
				t.Fatalf("unexpected offset: %q", q.Get("offset"))
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"apps", "search", "--query", "fit", "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if !strings.Contains(out, `"name":"FitTrack"`) || !strings.Contains(out, `"name":"StrideCoach"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_Help_KeywordsList_MatchesSmartListStructureWithoutMethodWording(t *testing.T) {
	out, code := captureRun(t, []string{"help", "keywords", "list"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		"List targeting keywords, with optional filtering.",
		"Use --filter / --sort flags, or --selector for JSON selector.",
		"Selector JSON keys: conditions, fields, orderBy, pagination.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("keywords list help missing %q: %q", want, out)
		}
	}
	for _, unwanted := range []string{" via GET", " POST ", " GET ", " PUT ", " DELETE "} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("keywords list help should not mention request methods %q: %q", unwanted, out)
		}
	}
}

func TestRun_Help_NegativesList_DoesNotMentionRequestMethods(t *testing.T) {
	out, code := captureRun(t, []string{"help", "negatives", "list"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		"Use --filter / --sort flags, or --selector for JSON selector.",
		"Selector JSON keys: conditions, fields, orderBy, pagination.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("negatives list help missing %q: %q", want, out)
		}
	}
	for _, unwanted := range []string{" POST ", " GET ", " PUT ", " DELETE "} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("negatives list help should not mention request methods %q: %q", unwanted, out)
		}
	}
}

func TestRun_Help_AdsList_MatchesAdGroupsStructureWithoutMethodWording(t *testing.T) {
	out, code := captureRun(t, []string{"help", "ads", "list"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	for _, want := range []string{
		"List ads, with optional filtering.",
		"With --campaign-id and --adgroup-id, lists ads in that ad group.",
		"Without them, searches across all campaigns.",
		"Use --filter / --sort flags, or --selector for JSON selector.",
		"Selector JSON keys: conditions, fields, orderBy, pagination.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("ads list help missing %q: %q", want, out)
		}
	}
	for _, unwanted := range []string{" via GET", " POST ", " GET ", " PUT ", " DELETE "} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("ads list help should not mention request methods %q: %q", unwanted, out)
		}
	}
}

func TestRun_Help_AdGroupsUpdate_MergeFlagDoesNotMentionPut(t *testing.T) {
	out, code := captureRun(t, []string{"help", "adgroups", "update"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "Fetch current ad group and merge changes first") {
		t.Fatalf("adgroups update help missing revised --merge wording: %q", out)
	}
	if strings.Contains(out, " and PUT ") {
		t.Fatalf("adgroups update help should not mention PUT: %q", out)
	}
}

func TestRun_Help_Schema_DoesNotUsePostExample(t *testing.T) {
	out, code := captureRun(t, []string{"help", "schema"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, "aads schema --method POST") {
		t.Fatalf("schema help should not use POST in examples: %q", out)
	}
	if !strings.Contains(out, "aads schema --method post       List endpoints for a method") {
		t.Fatalf("schema help missing updated method example: %q", out)
	}
}

func TestRun_Help_ListCommands_UseSearchableAndFilterableFieldsHeading(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "campaigns", args: []string{"help", "campaigns", "list"}},
		{name: "adgroups", args: []string{"help", "adgroups", "list"}},
		{name: "keywords", args: []string{"help", "keywords", "list"}},
		{name: "negatives", args: []string{"help", "negatives", "list"}},
		{name: "ads", args: []string{"help", "ads", "list"}},
		{name: "creatives", args: []string{"help", "creatives", "list"}},
		{name: "ad-rejections", args: []string{"help", "ad-rejections", "list"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, code := captureRun(t, tc.args, "")
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			if !strings.Contains(out, "Searchable and filterable fields:") {
				t.Fatalf("help missing searchable/filterable heading: %q", out)
			}
			if strings.Contains(out, "\nFilterable fields:\n") {
				t.Fatalf("help should not use plain filterable heading: %q", out)
			}
		})
	}
}

func TestRun_Help_FilterCommands_UseFilterOperatorsHeadingWithNotEquals(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "campaigns", args: []string{"help", "campaigns", "list"}},
		{name: "adgroups", args: []string{"help", "adgroups", "list"}},
		{name: "keywords", args: []string{"help", "keywords", "list"}},
		{name: "negatives", args: []string{"help", "negatives", "list"}},
		{name: "ads", args: []string{"help", "ads", "list"}},
		{name: "creatives", args: []string{"help", "creatives", "list"}},
		{name: "ad-rejections", args: []string{"help", "ad-rejections", "list"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, code := captureRun(t, tc.args, "")
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			if !strings.Contains(out, "Filter operators:") {
				t.Fatalf("help missing filter operators heading: %q", out)
			}
			if !strings.Contains(out, "NOT_EQUALS (local)") {
				t.Fatalf("help missing local NOT_EQUALS operator note: %q", out)
			}
			if strings.Contains(out, "\nOperators:\n") {
				t.Fatalf("help should not use generic operators heading: %q", out)
			}
		})
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	code := Run([]string{"nonexistent"}, "1.0.0-test", "5.5")
	// Unknown command triggers the root Exec which prints help
	if code != ExitSuccess && code != ExitUsage {
		t.Errorf("Run(nonexistent) = %d, want %d or %d", code, ExitSuccess, ExitUsage)
	}
}

func TestRun_NoArgs(t *testing.T) {
	code := Run(nil, "1.0.0-test", "5.5")
	if code != ExitSuccess {
		t.Errorf("Run(nil) = %d, want %d", code, ExitSuccess)
	}
}

func TestRun_EmptyArgs(t *testing.T) {
	code := Run([]string{}, "1.0.0-test", "5.5")
	if code != ExitSuccess {
		t.Errorf("Run([]string{}) = %d, want %d", code, ExitSuccess)
	}
}

func TestRun_SubcommandNoArgs(t *testing.T) {
	// Running a group command without a subcommand shows help/usage
	code := Run([]string{"campaigns"}, "1.0.0-test", "5.5")
	if code != ExitSuccess && code != ExitUsage {
		t.Errorf("Run(campaigns) = %d, want %d or %d", code, ExitSuccess, ExitUsage)
	}
}

func TestRootCommand_HasSubcommands(t *testing.T) {
	root := RootCommand("test-version", "5.5")
	if len(root.Subcommands) == 0 {
		t.Error("RootCommand has no subcommands")
	}
}

func TestRootCommand_Name(t *testing.T) {
	root := RootCommand("test-version", "5.5")
	if root.Name != "aads" {
		t.Errorf("RootCommand.Name = %q, want %q", root.Name, "aads")
	}
}

func TestRootCommand_HasVersionFlag(t *testing.T) {
	root := RootCommand("test-version", "5.5")
	f := root.FlagSet.Lookup("version")
	if f == nil {
		t.Error("RootCommand missing --version flag")
	}
}

func TestRootCommand_HasConfigDirFlag(t *testing.T) {
	root := RootCommand("test-version", "5.5")
	f := root.FlagSet.Lookup("config-dir")
	if f == nil {
		t.Error("RootCommand missing --config-dir flag")
	}
}

func TestRootCommand_DoesNotHaveFieldsFlag(t *testing.T) {
	root := RootCommand("test-version", "5.5")
	if f := root.FlagSet.Lookup("fields"); f != nil {
		t.Fatal("RootCommand should not define --fields")
	}
}

func TestRootCommand_DoesNotHaveForceFlag(t *testing.T) {
	root := RootCommand("test-version", "5.5")
	if f := root.FlagSet.Lookup("force"); f != nil {
		t.Fatal("RootCommand should not define --force")
	}
}

func TestExitCodeFromError_Nil(t *testing.T) {
	code := exitCodeFromError(nil)
	if code != ExitSuccess {
		t.Errorf("exitCodeFromError(nil) = %d, want %d", code, ExitSuccess)
	}
}

func TestExitCodeFromError_FlagErrHelp(t *testing.T) {
	code := exitCodeFromError(flag.ErrHelp)
	if code != ExitUsage {
		t.Errorf("exitCodeFromError(flag.ErrHelp) = %d, want %d", code, ExitUsage)
	}
}

func TestExitCodeFromError_AuthError(t *testing.T) {
	err := &api.APIError{StatusCode: 401}
	code := exitCodeFromError(err)
	if code != ExitAuth {
		t.Errorf("exitCodeFromError(auth) = %d, want %d", code, ExitAuth)
	}
}

func TestExitCodeFromError_SafetyError(t *testing.T) {
	err := shared.NewSafetyError("budget exceeded")
	code := exitCodeFromError(err)
	if code != ExitSafetyLimit {
		t.Errorf("exitCodeFromError(safety) = %d, want %d", code, ExitSafetyLimit)
	}
}

func TestExitCodeFromError_APIError(t *testing.T) {
	err := &api.APIError{StatusCode: 400}
	code := exitCodeFromError(err)
	if code != ExitAPIError {
		t.Errorf("exitCodeFromError(api 400) = %d, want %d", code, ExitAPIError)
	}
}

func TestExitCodeFromError_NetworkError(t *testing.T) {
	err := &api.APIError{StatusCode: 500}
	code := exitCodeFromError(err)
	if code != ExitNetworkError {
		t.Errorf("exitCodeFromError(network 500) = %d, want %d", code, ExitNetworkError)
	}
}

func TestExitCodeFromError_GenericError(t *testing.T) {
	err := flag.ErrHelp // not nil, not auth/safety/api/network
	// flag.ErrHelp is checked before generic, returns ExitUsage
	code := exitCodeFromError(err)
	if code != ExitUsage {
		t.Errorf("exitCodeFromError(flag.ErrHelp) = %d, want %d", code, ExitUsage)
	}
}

func TestExitCodeFromError_UsageError(t *testing.T) {
	err := shared.UsageError("something went wrong")
	code := exitCodeFromError(err)
	if code != ExitUsage {
		t.Errorf("exitCodeFromError(usageError) = %d, want %d", code, ExitUsage)
	}
}

func TestExitCodeFromError_ValidationError(t *testing.T) {
	err := shared.ValidationError("something went wrong")
	code := exitCodeFromError(err)
	if code != ExitUsage {
		t.Errorf("exitCodeFromError(validationError) = %d, want %d", code, ExitUsage)
	}
}

func TestRun_PipelineCampaignIDsIntoAdGroupReportTable(t *testing.T) {
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
				if !bytes.Contains(body, []byte(`"field":"status"`)) || !bytes.Contains(body, []byte(`"ENABLED"`)) {
					t.Fatalf("campaign selector = %s, want status=ENABLED", body)
				}
				return jsonResponse(`{"data":[{"id":111}]}`), nil
			case "/api/v5/reports/campaigns/111/adgroups":
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading report body: %v", err)
				}
				if !bytes.Contains(body, []byte(`"startTime":"2026-03-18"`)) {
					t.Fatalf("report body missing startTime: %s", body)
				}
				if !bytes.Contains(body, []byte(`"endTime":"2026-03-25"`)) {
					t.Fatalf("report body missing endTime: %s", body)
				}
				if !bytes.Contains(body, []byte(`"field":"impressions"`)) || !bytes.Contains(body, []byte(`"sortOrder":"DESCENDING"`)) {
					t.Fatalf("report body missing sort: %s", body)
				}
				if !bytes.Contains(body, []byte(`"returnRecordsWithNoMetrics":true`)) {
					t.Fatalf("report body missing returnRecordsWithNoMetrics=true: %s", body)
				}
				return jsonResponse(`{
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
										"taps": 56,
										"localSpend": {"amount":"78.90","currency":"USD"}
									}
								}
							]
						}
					}
				}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	idsOut, idsCode := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED", "-f", "ids"}, "")
	if idsCode != ExitSuccess {
		t.Fatalf("campaigns list exit code = %d, want %d", idsCode, ExitSuccess)
	}
	if !strings.Contains(idsOut, "CAMPAIGN_ID") || !strings.Contains(idsOut, "111") {
		t.Fatalf("campaign ids output = %q", idsOut)
	}

	reportOut, reportCode := captureRun(t, []string{"reports", "adgroups", "--campaign-id", "-", "--start", "-7d", "--end", "now", "--sort", "impressions:desc", "-f", "table"}, idsOut)
	if reportCode != ExitSuccess {
		t.Fatalf("reports adgroups exit code = %d, want %d; output=%q", reportCode, ExitSuccess, reportOut)
	}
	if strings.Contains(reportOut, "REPORTING_DATA_RESPONSE") {
		t.Fatalf("report output still shows raw reporting envelope: %q", reportOut)
	}
	for _, want := range []string{"CAMPAIGN_ID", "AD_GROUP_ID", "AD_GROUP_NAME", "IMPRESSIONS", "TAPS", "LOCAL_SPEND"} {
		if !strings.Contains(reportOut, want) {
			t.Fatalf("report output missing column %q: %q", want, reportOut)
		}
	}
	for _, want := range []string{"111", "5001", "Brand Exact", "1234", "56", "78.90 USD"} {
		if !strings.Contains(reportOut, want) {
			t.Fatalf("report output missing value %q: %q", want, reportOut)
		}
	}
}

func TestRun_CampaignsList_StatusFilterActiveNormalizesToEnabled(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading campaigns body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"status"`)) || !bytes.Contains(body, []byte(`"ENABLED"`)) {
				t.Fatalf("campaign selector = %s, want status normalized to ENABLED", body)
			}
			if bytes.Contains(body, []byte(`"ACTIVE"`)) {
				t.Fatalf("campaign selector should not contain ACTIVE after normalization: %s", body)
			}
			return jsonResponse(`{"data":[]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ACTIVE"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestRun_CampaignsList_MixedRemoteAndLocalFilters(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading campaigns body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"status"`)) || !bytes.Contains(body, []byte(`"ENABLED"`)) {
				t.Fatalf("campaign selector = %s, want status=ENABLED", body)
			}
			if bytes.Contains(body, []byte(`"NOT_EQUALS"`)) || bytes.Contains(body, []byte(`"Brand"`)) {
				t.Fatalf("campaign selector should not include local != filter: %s", body)
			}
			return jsonResponse(`{"data":[
				{"id":101,"name":"Brand","status":"ENABLED"},
				{"id":202,"name":"Generic","status":"ENABLED"}
			]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{
		"campaigns", "list",
		"--filter", "status=ENABLED",
		"--filter", "name!=Brand",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"name":"Brand"`) || !strings.Contains(out, `"name":"Generic"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_CampaignsList_LocalNotEqualsOnlyUsesGET(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			if req.Method != http.MethodGet {
				t.Fatalf("unexpected request method: %s", req.Method)
			}
			return jsonResponse(`{"data":[
				{"id":101,"name":"Brand"},
				{"id":202,"name":"Generic"}
			]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--filter", "name!=Brand", "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, `"name":"Brand"`) || !strings.Contains(out, `"name":"Generic"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_CampaignsList_SelectorFromFile(t *testing.T) {
	dir := t.TempDir()
	selectorPath := dir + "/selector.json"
	if err := os.WriteFile(selectorPath, []byte(`{"conditions":[{"field":"status","operator":"EQUALS","values":["ENABLED"]}]}`), 0o600); err != nil {
		t.Fatalf("write selector: %v", err)
	}

	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"status"`)) || !bytes.Contains(body, []byte(`"ENABLED"`)) {
				t.Fatalf("selector body = %s, want status condition", body)
			}
			if !bytes.Contains(body, []byte(`"offset":0`)) || !bytes.Contains(body, []byte(`"limit":1000`)) {
				t.Fatalf("selector body = %s, want fetch-all pagination", body)
			}
			return jsonResponse(`{"data":[]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--selector", "@" + selectorPath}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestRun_CampaignsList_FilterAndSelectorMutuallyExclusive(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatal("request should not be sent")
			return nil, nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED", "--selector", `{"conditions":[]}`}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "mutually exclusive") {
		t.Fatalf("expected mutual exclusivity error, got: %q", out)
	}
}

func TestRun_CampaignsList_SortAndSelectorMutuallyExclusive(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatal("request should not be sent")
			return nil, nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--sort", "name:asc", "--selector", `{"conditions":[]}`}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "mutually exclusive") {
		t.Fatalf("expected mutual exclusivity error, got: %q", out)
	}
}

func TestImpressionShareCreateRequiresNameWithoutFromJSON(t *testing.T) {
	out, code := captureRun(t, []string{"impression-share", "create", "--dateRange", "LAST_WEEK"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "--name is required") {
		t.Fatalf("output = %q, want missing --name error", out)
	}
}

func TestImpressionShareCreateCustomDateRangeRequiresStartAndEnd(t *testing.T) {
	out, code := captureRun(t, []string{"impression-share", "create", "--name", "Custom", "--dateRange", "CUSTOM", "--granularity", "WEEKLY", "--startTime", "2026-03-01"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "--endTime is required when --dateRange is CUSTOM") {
		t.Fatalf("output = %q, want missing --endTime error", out)
	}
}

func TestImpressionShareCreateDailyGranularityRejectsPredefinedDateRange(t *testing.T) {
	out, code := captureRun(t, []string{"impression-share", "create", "--name", "Daily", "--dateRange", "LAST_WEEK", "--granularity", "DAILY"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, "--dateRange is not allowed with DAILY granularity") {
		t.Fatalf("output = %q, want DAILY/dateRange usage error", out)
	}
}

func TestImpressionShareGetTableUsesColumnHeaders(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/custom-reports/123" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			return jsonResponse(`{"data":{"id":123,"name":"Weekly Share Report","granularity":"DAILY","startTime":"2026-03-20","endTime":"2026-03-27"}}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"impression-share", "get", "--report-id", "123", "-f", "table"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, "KEY") || strings.Contains(out, "VALUE") {
		t.Fatalf("output should not use key/value layout: %q", out)
	}
	for _, want := range []string{"ID", "NAME", "GRANULARITY", "START_TIME", "END_TIME", "123", "Weekly Share Report"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %q", want, out)
		}
	}
}

func TestImpressionShareGetDownloadWritesFile(t *testing.T) {
	downloadPath := filepath.Join(t.TempDir(), "report.csv")
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.String() {
			case "https://api.searchads.apple.com/api/v5/custom-reports/123":
				return jsonResponse(`{"data":{"id":123,"name":"Weekly Share Report","downloadUri":"https://downloads.example.com/reports/123.csv"}}`), nil
			case "https://downloads.example.com/reports/123.csv":
				if got := req.Header.Get("Authorization"); got != "" {
					t.Fatalf("Authorization = %q, want empty for presigned download", got)
				}
				if got := req.Header.Get("X-AP-Context"); got != "" {
					t.Fatalf("X-AP-Context = %q, want empty for presigned download", got)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/csv"}},
					Body:       io.NopCloser(strings.NewReader("col1,col2\n1,2\n")),
				}, nil
			default:
				t.Fatalf("unexpected request url: %s", req.URL.String())
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"impression-share", "get", "--report-id", "123", "--download", downloadPath, "-f", "json"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.TrimSpace(out) != "Downloaded report to "+downloadPath {
		t.Fatalf("output = %q, want download confirmation", out)
	}
	data, err := os.ReadFile(downloadPath)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if got := string(data); got != "col1,col2\n1,2\n" {
		t.Fatalf("downloaded file = %q, want csv contents", got)
	}
}

func TestImpressionShareGetDownloadStdoutWritesOnlyContent(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.String() {
			case "https://api.searchads.apple.com/api/v5/custom-reports/123":
				return jsonResponse(`{"data":{"id":123,"name":"Weekly Share Report","downloadUri":"https://downloads.example.com/reports/123.csv"}}`), nil
			case "https://downloads.example.com/reports/123.csv":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/csv"}},
					Body:       io.NopCloser(strings.NewReader("col1,col2\n1,2\n")),
				}, nil
			default:
				t.Fatalf("unexpected request url: %s", req.URL.String())
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"impression-share", "get", "--report-id", "123", "--download", "-"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if out != "col1,col2\n1,2\n" {
		t.Fatalf("output = %q, want raw downloaded content only", out)
	}
}

func TestRun_CampaignsList_SortAloneWorks(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns/find" {
				t.Fatalf("expected find endpoint, got %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"orderBy"`)) {
				t.Fatalf("expected orderBy in body: %s", body)
			}
			return jsonResponse(`{"data":[]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--sort", "name:asc"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestRun_AppsEligibility_FromJSONInline(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/apps/900001/eligibilities/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("unmarshal request body: %v", err)
			}
			conditions, ok := got["conditions"].([]any)
			if !ok || len(conditions) != 1 {
				t.Fatalf("unexpected selector body: %s", body)
			}
			first, ok := conditions[0].(map[string]any)
			if !ok {
				t.Fatalf("unexpected selector condition: %s", body)
			}
			if first["field"] != "countryOrRegion" || first["operator"] != "EQUALS" {
				t.Fatalf("unexpected request body: %s", body)
			}
			return jsonResponse(`{"data":{"eligibility":"ELIGIBLE"}}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"apps", "eligibility", "--adam-id", "900001", "--from-json", `{"conditions":[{"field":"countryOrRegion","operator":"EQUALS","values":["US"]}]}`}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestRun_AppsEligibility_FlagsBuildSelectorBody(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/apps/900001/eligibilities/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("unmarshal request body: %v", err)
			}
			conditions, ok := got["conditions"].([]any)
			if !ok || len(conditions) != 3 {
				t.Fatalf("unexpected selector body: %s", body)
			}
			want := map[string]any{
				"countryOrRegion": "BE",
				"deviceClass":     "IPHONE",
				"supplySource":    "APPSTORE_SEARCH_RESULTS",
			}
			for _, raw := range conditions {
				cond, ok := raw.(map[string]any)
				if !ok {
					t.Fatalf("unexpected condition shape: %s", body)
				}
				field, _ := cond["field"].(string)
				values, _ := cond["values"].([]any)
				if len(values) != 1 {
					t.Fatalf("unexpected condition values: %s", body)
				}
				if want[field] != values[0] {
					t.Fatalf("unexpected condition %q=%v in body: %s", field, values[0], body)
				}
				delete(want, field)
			}
			if len(want) != 0 {
				t.Fatalf("unexpected request body: %s", body)
			}
			return jsonResponse(`{"data":{"eligibility":"ELIGIBLE"}}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"apps", "eligibility", "--adam-id", "900001", "--country-or-region", "be", "--device-class", "iphone", "--supply-source", "appstore_search_results"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestRun_AppsEligibility_FromJSONLegacyBodyIsTranslatedToSelector(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/apps/900001/eligibilities/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("unmarshal request body: %v", err)
			}
			conditions, ok := got["conditions"].([]any)
			if !ok || len(conditions) != 2 {
				t.Fatalf("unexpected selector body: %s", body)
			}
			return jsonResponse(`{"data":{"eligibility":"ELIGIBLE"}}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"apps", "eligibility", "--from-json", `{"adamId":900001,"countryOrRegion":"US","deviceClass":"IPHONE"}`}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestRun_CampaignsList_InvalidStatusFilterShowsValidValues(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected API call despite invalid status filter: %s", req.URL.Path)
			return nil, fmt.Errorf("unexpected call")
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=BROKEN"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, `invalid status filter value "BROKEN": valid values are PAUSED and ENABLED`) {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, "USAGE") || strings.Contains(out, "Usage:") {
		t.Fatalf("validation error should not print command help: %q", out)
	}
}

func TestRun_CampaignsList_StatusFilterNumericAliasNormalizesToEnabled(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading campaigns body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"status"`)) || !bytes.Contains(body, []byte(`"ENABLED"`)) {
				t.Fatalf("campaign selector = %s, want status normalized to ENABLED", body)
			}
			if bytes.Contains(body, []byte(`"1"`)) {
				t.Fatalf("campaign selector should not contain raw numeric alias after normalization: %s", body)
			}
			return jsonResponse(`{"data":[]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=1"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
}

func TestRun_KeywordsList_FilteredJSONFieldsUsesCampaignScopedFindWithAdGroupSelector(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns/"+testCampaignID+"/adgroups/targetingkeywords/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading keywords find body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"status"`)) || !bytes.Contains(body, []byte(`"ACTIVE"`)) {
				t.Fatalf("keywords selector = %s, want status ACTIVE", body)
			}
			if !bytes.Contains(body, []byte(`"field":"adGroupId"`)) || !bytes.Contains(body, []byte(`"`+testAdGroupID+`"`)) {
				t.Fatalf("keywords selector = %s, want adGroupId condition", body)
			}
			return jsonResponse(`{"data":[{"id":1,"campaignId":900101,"adGroupId":900201,"text":"brand","matchType":"EXACT","status":"ACTIVE"}]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{
		"keywords", "list",
		"--campaign-id", testCampaignID,
		"--adgroup-id", testAdGroupID,
		"--filter", "status=ACTIVE",
		"--fields", "TEXT,MATCH_TYPE",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, "The referenced resource was not found") {
		t.Fatalf("unexpected 404 output: %q", out)
	}
	if !strings.Contains(out, `"TEXT":"brand"`) || !strings.Contains(out, `"MATCH_TYPE":"EXACT"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_ReportsAdGroups_AdGroupIDShortcutAddsSelectorCondition(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/reports/campaigns/"+testCampaignID+"/adgroups" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading report body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"adGroupId"`)) || !bytes.Contains(body, []byte(`"`+testAdGroupID+`"`)) {
				t.Fatalf("report selector = %s, want adGroupId condition", body)
			}
			return jsonResponse(reportResponseJSON()), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{
		"reports", "adgroups",
		"--campaign-id", testCampaignID,
		"--adgroup-id", testAdGroupID,
		"--start", "2026-03-18",
		"--end", "2026-03-25",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"adGroupName":"Core Search"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_ReportsAdGroups_AdGroupIDShortcutReadsFromStdin(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/reports/campaigns/"+testCampaignID+"/adgroups" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading report body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"adGroupId"`)) || !bytes.Contains(body, []byte(`"`+testAdGroupID+`"`)) {
				t.Fatalf("report selector = %s, want adGroupId condition", body)
			}
			return jsonResponse(reportResponseJSON()), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	stdin := "CAMPAIGN_ID\tADGROUP_ID\n" + testCampaignID + "\t" + testAdGroupID + "\n"
	out, code := captureRun(t, []string{
		"reports", "adgroups",
		"--campaign-id", "-",
		"--adgroup-id", "-",
		"--start", "2026-03-18",
		"--end", "2026-03-25",
		"-f", "json",
	}, stdin)
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"adGroupId":900201`) {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, `"adgroupId":"`+testAdGroupID+`"`) || strings.Contains(out, `"adgroupId":900201`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_ReportsSearchTerms_StdinUsesCanonicalAdGroupIdKey(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/reports/campaigns/"+testCampaignID+"/adgroups/"+testAdGroupID+"/searchterms" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			return jsonResponse(reportResponseJSON()), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	stdin := "CAMPAIGN_ID\tADGROUP_ID\n" + testCampaignID + "\t" + testAdGroupID + "\n"
	out, code := captureRun(t, []string{
		"reports", "searchterms",
		"--campaign-id", "-",
		"--adgroup-id", "-",
		"--start", "2026-03-18",
		"--end", "2026-03-25",
		"-f", "json",
	}, stdin)
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"adGroupId":900201`) {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, `"adgroupId":"`+testAdGroupID+`"`) || strings.Contains(out, `"adgroupId":900201`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_KeywordsList_StdinUsesCanonicalAdGroupIdKey(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns/"+testCampaignID+"/adgroups/"+testAdGroupID+"/targetingkeywords" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			return jsonResponse(`{"data":[{"id":1,"campaignId":900101,"adGroupId":900201,"text":"brand","matchType":"EXACT","status":"ACTIVE"}]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	stdin := "CAMPAIGN_ID\tADGROUP_ID\n" + testCampaignID + "\t" + testAdGroupID + "\n"
	out, code := captureRun(t, []string{
		"keywords", "list",
		"--campaign-id", "-",
		"--adgroup-id", "-",
		"-f", "json",
	}, stdin)
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, `"adGroupId":900201`) {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, `"adgroupId":"`+testAdGroupID+`"`) || strings.Contains(out, `"adgroupId":900201`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_KeywordsList_DefaultFetchesAllPages(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	requests := 0
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.URL.Path != "/api/v5/campaigns/"+testCampaignID+"/adgroups/"+testAdGroupID+"/targetingkeywords" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			switch req.URL.RawQuery {
			case "limit=1000&offset=0":
				return jsonResponse(`{
					"data":[{"id":1},{"id":2}],
					"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}
				}`), nil
			case "limit=1000&offset=2":
				return jsonResponse(`{
					"data":[{"id":3}],
					"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":1}
				}`), nil
			default:
				t.Fatalf("unexpected query: %s", req.URL.RawQuery)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"keywords", "list", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want 2", requests)
	}
	for _, want := range []string{`"id":1`, `"id":2`, `"id":3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %s: %q", want, out)
		}
	}
}

func TestRun_CampaignsList_DefaultFetchesAllPages(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	requests := 0
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.URL.Path != "/api/v5/campaigns" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			switch req.URL.RawQuery {
			case "limit=1000&offset=0":
				return jsonResponse(`{
					"data":[{"id":1},{"id":2}],
					"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}
				}`), nil
			case "limit=1000&offset=2":
				return jsonResponse(`{
					"data":[{"id":3}],
					"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":1}
				}`), nil
			default:
				t.Fatalf("unexpected query: %s", req.URL.RawQuery)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want 2", requests)
	}
	for _, want := range []string{`"id":1`, `"id":2`, `"id":3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %s: %q", want, out)
		}
	}
}

func TestRun_CampaignsList_FilterDefaultFetchesAllPages(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	requests := 0
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.URL.Path != "/api/v5/campaigns/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading campaigns find body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"status"`)) || !bytes.Contains(body, []byte(`"ENABLED"`)) {
				t.Fatalf("campaign selector = %s, want status ENABLED", body)
			}
			switch {
			case bytes.Contains(body, []byte(`"offset":0`)) && bytes.Contains(body, []byte(`"limit":1000`)):
				return jsonResponse(`{
					"data":[{"id":1},{"id":2}],
					"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}
				}`), nil
			case bytes.Contains(body, []byte(`"offset":2`)) && bytes.Contains(body, []byte(`"limit":1000`)):
				return jsonResponse(`{
					"data":[{"id":3}],
					"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":1}
				}`), nil
			default:
				t.Fatalf("unexpected selector body: %s", body)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"campaigns", "list", "--filter", "status=ENABLED"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want 2", requests)
	}
	for _, want := range []string{`"id":1`, `"id":2`, `"id":3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %s: %q", want, out)
		}
	}
}

func TestRun_BudgetOrdersList_DefaultFetchesAllPages(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	requests := 0
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.URL.Path != "/api/v5/budgetorders" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			switch req.URL.RawQuery {
			case "limit=1000&offset=0":
				return jsonResponse(`{
					"data":[{"id":1},{"id":2}],
					"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}
				}`), nil
			case "limit=1000&offset=2":
				return jsonResponse(`{
					"data":[{"id":3}],
					"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":1}
				}`), nil
			default:
				t.Fatalf("unexpected query: %s", req.URL.RawQuery)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"budgetorders", "list"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want 2", requests)
	}
	for _, want := range []string{`"id":1`, `"id":2`, `"id":3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %s: %q", want, out)
		}
	}
}

func TestRun_RemainingListCommands_DefaultFetchesAllPages(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
		{
			name: "adgroups list",
			args: []string{"adgroups", "list", "--campaign-id", testCampaignID},
			path: "/api/v5/campaigns/" + testCampaignID + "/adgroups",
		},
		{
			name: "ads list",
			args: []string{"ads", "list", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID},
			path: "/api/v5/campaigns/" + testCampaignID + "/adgroups/" + testAdGroupID + "/ads",
		},
		{
			name: "creatives list",
			args: []string{"creatives", "list"},
			path: "/api/v5/creatives",
		},
		{
			name: "product-pages list",
			args: []string{"product-pages", "list", "--adam-id", testAdamID},
			path: "/api/v5/apps/" + testAdamID + "/product-pages",
		},
		{
			name: "impression-share list",
			args: []string{"impression-share", "list"},
			path: "/api/v5/custom-reports",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := apiPkg.NewClient(func(context.Context) (string, error) {
				return "test-token", nil
			}, "123", false)
			requests := 0
			client.SetHTTPClientForTesting(&http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					requests++
					if req.URL.Path != tt.path {
						t.Fatalf("unexpected request path: %s", req.URL.Path)
					}
					switch req.URL.RawQuery {
					case "limit=1000&offset=0":
						return jsonResponse(`{
							"data":[{"id":1},{"id":2}],
							"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}
						}`), nil
					case "limit=1000&offset=2":
						return jsonResponse(`{
							"data":[{"id":3}],
							"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":1}
						}`), nil
					default:
						t.Fatalf("unexpected query: %s", req.URL.RawQuery)
						return nil, nil
					}
				}),
			})
			restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
			defer restoreClient()

			out, code := captureRun(t, tt.args, "")
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			if requests != 2 {
				t.Fatalf("request count = %d, want 2", requests)
			}
			for _, want := range []string{`"id":1`, `"id":2`, `"id":3`} {
				if !strings.Contains(out, want) {
					t.Fatalf("output missing %s: %q", want, out)
				}
			}
		})
	}
}

func TestRun_NegativesList_CampaignLevelDefaultFetchesAllPages(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	requests := 0
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.URL.Path != "/api/v5/campaigns/"+testCampaignID+"/negativekeywords" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			switch req.URL.RawQuery {
			case "limit=1000&offset=0":
				return jsonResponse(`{
					"data":[{"id":1,"text":"free"},{"id":2,"text":"cheap"}],
					"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}
				}`), nil
			case "limit=1000&offset=2":
				return jsonResponse(`{
					"data":[{"id":3,"text":"trial"}],
					"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":1}
				}`), nil
			default:
				t.Fatalf("unexpected query: %s", req.URL.RawQuery)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"negatives", "list", "--campaign-id", testCampaignID}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want 2", requests)
	}
	for _, want := range []string{`"id":1`, `"id":2`, `"id":3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %s: %q", want, out)
		}
	}
}

func TestRun_NegativesList_AdGroupLevelDefaultFetchesAllPages(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	requests := 0
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.URL.Path != "/api/v5/campaigns/"+testCampaignID+"/adgroups/"+testAdGroupID+"/negativekeywords" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			switch req.URL.RawQuery {
			case "limit=1000&offset=0":
				return jsonResponse(`{
					"data":[{"id":1,"text":"free"},{"id":2,"text":"cheap"}],
					"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}
				}`), nil
			case "limit=1000&offset=2":
				return jsonResponse(`{
					"data":[{"id":3,"text":"trial"}],
					"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":1}
				}`), nil
			default:
				t.Fatalf("unexpected query: %s", req.URL.RawQuery)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"negatives", "list", "--campaign-id", testCampaignID, "--adgroup-id", testAdGroupID}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want 2", requests)
	}
	for _, want := range []string{`"id":1`, `"id":2`, `"id":3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %s: %q", want, out)
		}
	}
}

func TestRun_NegativesList_FilteredUsesCampaignScopedFindWithAdGroupSelector(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v5/campaigns/"+testCampaignID+"/adgroups/negativekeywords/find" {
				t.Fatalf("unexpected request path: %s", req.URL.Path)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading negatives find body: %v", err)
			}
			if !bytes.Contains(body, []byte(`"field":"status"`)) || !bytes.Contains(body, []byte(`"ACTIVE"`)) {
				t.Fatalf("negatives selector = %s, want status ACTIVE", body)
			}
			if !bytes.Contains(body, []byte(`"field":"adGroupId"`)) || !bytes.Contains(body, []byte(`"`+testAdGroupID+`"`)) {
				t.Fatalf("negatives selector = %s, want adGroupId condition", body)
			}
			return jsonResponse(`{"data":[{"id":1,"campaignId":900101,"adGroupId":900201,"text":"brand","matchType":"EXACT","status":"ACTIVE"}]}`), nil
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{
		"negatives", "list",
		"--campaign-id", testCampaignID,
		"--adgroup-id", testAdGroupID,
		"--filter", "status=ACTIVE",
		"--fields", "TEXT,MATCH_TYPE",
		"-f", "json",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, "The referenced resource was not found") {
		t.Fatalf("unexpected 404 output: %q", out)
	}
	if !strings.Contains(out, `"TEXT":"brand"`) || !strings.Contains(out, `"MATCH_TYPE":"EXACT"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRun_ReportsAdGroups_InvalidStartDateDoesNotShowHelp(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected API call despite invalid start date: %s", req.URL.Path)
			return nil, fmt.Errorf("unexpected call")
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()

	out, code := captureRun(t, []string{"reports", "adgroups", "--campaign-id", "111", "--start", "yesterdayish", "--end", "now"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, `--start: invalid date`) {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, "USAGE") || strings.Contains(out, "Usage:") {
		t.Fatalf("validation error should not print command help: %q", out)
	}
}

func TestRun_ProfilesCreate_InvalidLimitDoesNotShowHelp(t *testing.T) {
	out, code := captureRun(t, []string{"profiles", "create", "--name", "test", "--org-id", "123", "--default-currency", "USD", "--max-bid", "nope"}, "")
	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitUsage, out)
	}
	if !strings.Contains(out, `--max-bid: invalid limit amount "nope"`) {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, "USAGE") || strings.Contains(out, "Usage:") {
		t.Fatalf("validation error should not print command help: %q", out)
	}
}

func TestRun_ProfilesCreate_UsesConfigDirFlag(t *testing.T) {
	t.Setenv("AADS_CONFIG_DIR", "")

	configDir := filepath.Join(t.TempDir(), "custom-config")
	out, code := captureRun(t, []string{"--config-dir", configDir, "profiles", "create", "--name", "test", "--org-id", "123"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", configPath, err)
	}
	if !strings.Contains(out, configPath) {
		t.Fatalf("output should mention config path %q: %q", configPath, out)
	}
	if !strings.Contains(string(data), "default_profile: test") {
		t.Fatalf("config file missing default profile: %s", data)
	}
	if !strings.Contains(string(data), "profiles:") || !strings.Contains(string(data), "test:") {
		t.Fatalf("config file missing created profile: %s", data)
	}
	if !strings.Contains(string(data), `org_id: "123"`) {
		t.Fatalf("config file missing org_id: %s", data)
	}
}

func TestRun_ProfilesList_UsesConfigDirEnv(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "env-config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q): %v", configDir, err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
default_profile: work
profiles:
  work:
    client_id: work-client
    team_id: work-team
    key_id: work-key
    private_key_path: /keys/work.pem
`), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", configPath, err)
	}
	t.Setenv("AADS_CONFIG_DIR", configDir)

	out, code := captureRun(t, []string{"profiles", "list"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if !strings.Contains(out, "work") {
		t.Fatalf("profiles list output missing profile from config-dir override: %q", out)
	}
}

func TestRun_ReportLocalFilterAndIDsOutput(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/reports/campaigns/111/adgroups":
				return jsonResponse(`{
					"data": {
						"reportingDataResponse": {
							"row": [
								{
									"metadata": {"campaignId": 111, "adGroupId": 5001, "adGroupName": "Brand Exact"},
									"total": {"impressions": 1234, "localSpend": {"amount":"78.90","currency":"USD"}}
								},
								{
									"metadata": {"campaignId": 111, "adGroupId": 5002, "adGroupName": "Competitor"},
									"total": {"impressions": 200, "localSpend": {"amount":"4.10","currency":"USD"}}
								}
							]
						}
					}
				}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	out, code := captureRun(t, []string{
		"reports", "adgroups",
		"--campaign-id", "111",
		"--start", "-7d",
		"--end", "now",
		"--sort", "impressions:desc",
		"--filter", "LOCAL_SPEND GREATER_THAN 10",
		"--filter", "AD_GROUP_ID=5001",
		"-f", "ids",
	}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
	}
	if strings.Contains(out, "5002") {
		t.Fatalf("filtered ids output should exclude 5002: %q", out)
	}
	for _, want := range []string{"CAMPAIGN_ID", "AD_GROUP_ID", "111", "5001"} {
		if !strings.Contains(out, want) {
			t.Fatalf("filtered ids output missing %q: %q", want, out)
		}
	}
}

func TestRun_ReportsAdGroups_DefaultsSortToImpressionsDesc(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/api/v5/reports/campaigns/111/adgroups":
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("reading report body: %v", err)
				}
				if !bytes.Contains(body, []byte(`"field":"impressions"`)) || !bytes.Contains(body, []byte(`"sortOrder":"DESCENDING"`)) {
					t.Fatalf("report body missing default sort: %s", body)
				}
				return jsonResponse(`{"data":{"reportingDataResponse":{"row":[]}}}`), nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{OrgID: "123"})
	defer restoreClient()
	restoreNow := shared.SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	})
	defer restoreNow()

	_, code := captureRun(t, []string{"reports", "adgroups", "--campaign-id", "111", "--start", "2026-03-18", "--end", "2026-03-25"}, "")
	if code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", code, ExitSuccess)
	}
}

func TestRun_ForceAcceptedOnSafetyCheckedSubcommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
		{
			name: "campaigns create",
			args: []string{
				"campaigns", "create",
				"--name", "Force Campaign",
				"--adam-id", testAdamID,
				"--daily-budget-amount", "50",
				"--countries-or-regions", "US",
				"--ad-channel-type", "SEARCH",
				"--force",
			},
			path: "/api/v5/campaigns",
		},
		{
			name: "campaigns update",
			args: []string{
				"campaigns", "update",
				"--campaign-id", "123",
				"--daily-budget-amount", "50",
				"--force",
			},
			path: "/api/v5/campaigns/123",
		},
		{
			name: "adgroups create",
			args: []string{
				"adgroups", "create",
				"--campaign-id", "123",
				"--name", "Force Ad Group",
				"--default-bid", "5",
				"--force",
			},
			path: "/api/v5/campaigns/123/adgroups",
		},
		{
			name: "adgroups update",
			args: []string{
				"adgroups", "update",
				"--campaign-id", "123",
				"--adgroup-id", "456",
				"--default-bid", "5",
				"--force",
			},
			path: "/api/v5/campaigns/123/adgroups/456",
		},
		{
			name: "keywords create",
			args: []string{
				"keywords", "create",
				"--campaign-id", "123",
				"--adgroup-id", "456",
				"--text", "brand",
				"--bid", "5",
				"--force",
			},
			path: "/api/v5/campaigns/123/adgroups/456/targetingkeywords/bulk",
		},
		{
			name: "keywords update",
			args: []string{
				"keywords", "update",
				"--campaign-id", "123",
				"--adgroup-id", "456",
				"--keyword-id", "789",
				"--bid", "5",
				"--force",
			},
			path: "/api/v5/campaigns/123/adgroups/456/targetingkeywords/bulk",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			client := apiPkg.NewClient(func(context.Context) (string, error) {
				return "test-token", nil
			}, "123", false)
			client.SetHTTPClientForTesting(&http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					if req.URL.Path != tc.path {
						t.Fatalf("request path = %s, want %s", req.URL.Path, tc.path)
					}
					called = true
					return jsonResponse(`{"data":{"id":1}}`), nil
				}),
			})
			restoreClient := shared.SetClientForTesting(client, &config.Profile{
				OrgID:           "123",
				DefaultCurrency: "USD",
				MaxDailyBudget:  config.DecimalText("1"),
				MaxBudgetAmount: config.DecimalText("1"),
				MaxBid:          config.DecimalText("1"),
				MaxCPAGoal:      config.DecimalText("1"),
			})
			defer restoreClient()

			out, code := captureRun(t, tc.args, "")
			if code != ExitSuccess {
				t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSuccess, out)
			}
			if !called {
				t.Fatalf("command did not reach API client; output=%q", out)
			}
			if strings.Contains(out, "exceeds limit") || strings.Contains(out, "cannot compare") {
				t.Fatalf("--force should bypass safety errors; output=%q", out)
			}
		})
	}
}

func TestRun_AdGroupsUpdate_OverLimitWithoutForceFailsSafety(t *testing.T) {
	client := apiPkg.NewClient(func(context.Context) (string, error) {
		return "test-token", nil
	}, "123", false)
	client.SetHTTPClientForTesting(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected API call despite safety failure: %s", req.URL.Path)
			return nil, fmt.Errorf("unexpected call")
		}),
	})
	restoreClient := shared.SetClientForTesting(client, &config.Profile{
		OrgID:           "123",
		DefaultCurrency: "USD",
		MaxBid:          config.DecimalText("1"),
	})
	defer restoreClient()

	out, code := captureRun(t, []string{
		"adgroups", "update",
		"--campaign-id", "123",
		"--adgroup-id", "456",
		"--default-bid", "5",
	}, "")
	if code != ExitSafetyLimit {
		t.Fatalf("exit code = %d, want %d; output=%q", code, ExitSafetyLimit, out)
	}
	if !strings.Contains(out, "exceeds limit") {
		t.Fatalf("expected safety limit error, got %q", out)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func captureRun(t *testing.T, args []string, stdin string) (string, int) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldStdin := os.Stdin

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW
	stdoutCh := make(chan []byte, 1)
	stderrCh := make(chan []byte, 1)
	go func() {
		stdoutBytes, _ := io.ReadAll(stdoutR)
		stdoutCh <- stdoutBytes
	}()
	go func() {
		stderrBytes, _ := io.ReadAll(stderrR)
		stderrCh <- stderrBytes
	}()

	if stdin != "" {
		stdinR, stdinW, _ := os.Pipe()
		if _, err := stdinW.WriteString(stdin); err != nil {
			t.Fatalf("writing stdin: %v", err)
		}
		stdinW.Close()
		os.Stdin = stdinR
	} else {
		stdinR, stdinW, _ := os.Pipe()
		stdinW.Close()
		os.Stdin = stdinR
	}

	code := Run(args, "1.0.0-test", "5.5")

	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	os.Stdin = oldStdin

	stdoutBytes := <-stdoutCh
	stderrBytes := <-stderrCh
	return string(stdoutBytes) + string(stderrBytes), code
}
