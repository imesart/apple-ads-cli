package structure

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	"github.com/imesart/apple-ads-cli/internal/columnname"
	"github.com/imesart/apple-ads-cli/internal/output"
	"github.com/imesart/apple-ads-cli/internal/types"
)

const (
	structureTypeStructure = "structure"
	structureTypeMapping   = "mapping"
	scopeCampaigns         = "campaigns"
	scopeAdgroups          = "adgroups"
	pageSize               = 1000
	structureSchemaURL     = "https://raw.githubusercontent.com/imesart/apple-ads-cli/main/docs/schemas/aads-v1.schema.json"
)

type structureRoot struct {
	SchemaURI     string                  `json:"$schema,omitempty"`
	SchemaVersion int                     `json:"schemaVersion"`
	Type          string                  `json:"type"`
	Scope         string                  `json:"scope"`
	CreationTime  string                  `json:"creationTime"`
	Campaigns     []structureCampaignNode `json:"campaigns,omitempty"`
	Adgroups      []structureAdGroupNode  `json:"adgroups,omitempty"`
}

type structureCampaignNode struct {
	Campaign                 map[string]any         `json:"campaign"`
	CampaignNegativeKeywords []map[string]any       `json:"campaignNegativeKeywords,omitempty"`
	Adgroups                 []structureAdGroupNode `json:"adgroups,omitempty"`
}

type structureAdGroupNode struct {
	Adgroup                 map[string]any   `json:"adgroup"`
	AdgroupNegativeKeywords []map[string]any `json:"adgroupNegativeKeywords,omitempty"`
	Keywords                []map[string]any `json:"keywords,omitempty"`
}

type mappingRoot struct {
	SchemaVersion int                   `json:"schemaVersion"`
	Type          string                `json:"type"`
	Scope         string                `json:"scope"`
	CreationTime  string                `json:"creationTime"`
	Campaigns     []mappingCampaignNode `json:"campaigns,omitempty"`
	Adgroups      []mappingAdGroupNode  `json:"adgroups,omitempty"`
}

type mappingCampaignNode struct {
	Campaign                 entityMapping        `json:"campaign"`
	CampaignNegativeKeywords []entityMapping      `json:"campaignNegativeKeywords,omitempty"`
	Adgroups                 []mappingAdGroupNode `json:"adgroups,omitempty"`
}

type mappingAdGroupNode struct {
	Adgroup                 entityMapping   `json:"adgroup"`
	AdgroupNegativeKeywords []entityMapping `json:"adgroupNegativeKeywords,omitempty"`
	Keywords                []entityMapping `json:"keywords,omitempty"`
}

type entityMapping struct {
	Source  map[string]any `json:"source,omitempty"`
	Created map[string]any `json:"created,omitempty"`
	Status  string         `json:"status,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type fieldSelection struct {
	FlagSet bool
	All     bool
	Fields  map[string]bool
}

type importFlags struct {
	fromStructure             string
	outputMapping             string
	check                     bool
	noAdgroups                bool
	noNegatives               bool
	noKeywords                bool
	allowUnmatchedNamePattern bool
	campaignID                string
	campaignsFromJSON         string
	adgroupsFromJSON          string
	campaignsName             string
	campaignsNamePattern      string
	adgroupsName              string
	adgroupsNamePattern       string
	adamID                    string
	dailyBudgetAmount         string
	countriesOrRegions        string
	budgetAmount              string
	targetCpa                 string
	locInvoiceDetails         string
	campaignsStatus           string
	campaignsStartTime        string
	campaignsEndTime          string
	adChannelType             string
	billingEvent              string
	supplySources             string
	cpaGoal                   string
	defaultBid                string
	adgroupsStatus            string
	adgroupsStartTime         string
	adgroupsEndTime           string
	automatedKeywordsOptIn    bool
	matchType                 string
	bid                       string
	keywordsStatus            string
	negativesStatus           string
}

type importPlan struct {
	root           structureRoot
	mapping        mappingRoot
	targetCampaign map[string]any
}

type mockIDGenerator struct {
	next int64
}

type paginatedRequest struct {
	inner  api.Request
	offset int
	limit  int
}

func (p *paginatedRequest) Method() string { return p.inner.Method() }
func (p *paginatedRequest) Path() string   { return p.inner.Path() }
func (p *paginatedRequest) Body() any      { return p.inner.Body() }
func (p *paginatedRequest) Query() url.Values {
	q := url.Values{}
	if inner := p.inner.Query(); inner != nil {
		for key, values := range inner {
			for _, value := range values {
				q.Add(key, value)
			}
		}
	}
	q.Set("offset", strconv.Itoa(p.offset))
	q.Set("limit", strconv.Itoa(p.limit))
	return q
}

func (g *mockIDGenerator) Next() json.Number {
	g.next++
	return json.Number(strconv.FormatInt(g.next, 10))
}

var (
	campaignReadOnlyFields = stringSet(
		"id", "orgId", "creationTime", "modificationTime", "deleted", "displayStatus",
		"servingStatus", "servingStateReasons", "countryOrRegionServingStateReasons",
		"paymentModel", "sapinLawResponse",
	)
	adgroupImportReadOnlyFields = stringSet(
		"id", "orgId", "creationTime", "modificationTime", "deleted", "displayStatus",
		"servingStatus", "servingStateReasons", "paymentModel",
	)
	adgroupReadOnlyFields = stringSet(
		"id", "orgId", "creationTime", "modificationTime", "deleted", "displayStatus",
		"servingStatus", "servingStateReasons",
	)
	keywordReadOnlyFields = stringSet(
		"id", "creationTime", "modificationTime", "deleted",
	)
	negativeReadOnlyFields = stringSet(
		"id", "creationTime", "modificationTime", "deleted",
	)
	campaignRelationshipFields = stringSet()
	adgroupRelationshipFields  = stringSet("campaignId")
	keywordRelationshipFields  = stringSet("campaignId", "adGroupId")
	negativeRelationshipFields = stringSet("campaignId", "adGroupId")

	campaignAllowedFields = stringSet(
		"adamId", "name", "adChannelType", "countriesOrRegions", "billingEvent",
		"dailyBudgetAmount", "budgetAmount", "targetCpa", "supplySources", "biddingStrategy", "status",
		"startTime", "endTime", "locInvoiceDetails",
	)
	adgroupAllowedFields = stringSet(
		"name", "pricingModel", "defaultBidAmount", "status", "cpaGoal",
		"automatedKeywordsOptIn", "paymentModel", "biddingStrategy",
		"startTime", "endTime", "targetingDimensions",
	)
	keywordAllowedFields  = stringSet("text", "matchType", "status", "bidAmount")
	negativeAllowedFields = stringSet("text", "matchType", "status")

	campaignRequiredFields = stringSet("adamId", "name", "dailyBudgetAmount", "countriesOrRegions")
	adgroupRequiredFields  = stringSet("name", "defaultBidAmount")
	keywordRequiredFields  = stringSet("text", "matchType")
	negativeRequiredFields = stringSet("text", "matchType")

	campaignRequiredFieldFlags = map[string]string{
		"adamId":             "--adam-id",
		"name":               "--campaigns-name",
		"dailyBudgetAmount":  "--daily-budget-amount",
		"countriesOrRegions": "--countries-or-regions",
	}
	adgroupRequiredFieldFlags = map[string]string{
		"name":             "--adgroups-name",
		"defaultBidAmount": "--default-bid",
	}
)

func newStructureRoot(scope string) structureRoot {
	return structureRoot{
		SchemaURI:     structureSchemaURL,
		SchemaVersion: 1,
		Type:          structureTypeStructure,
		Scope:         scope,
		CreationTime:  time.Now().UTC().Format(time.RFC3339),
	}
}

func newMappingRoot(scope string) mappingRoot {
	return mappingRoot{
		SchemaVersion: 1,
		Type:          structureTypeMapping,
		Scope:         scope,
		CreationTime:  time.Now().UTC().Format(time.RFC3339),
	}
}

func stringSet(values ...string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}

func flagWasSet(fs interface{ Visit(func(*flag.Flag)) }, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func parseFieldSelection(fs interface{ Visit(func(*flag.Flag)) }, name, raw string) fieldSelection {
	if !flagWasSet(fs, name) {
		return fieldSelection{}
	}
	if raw == "" {
		return fieldSelection{FlagSet: true, All: true}
	}
	fields := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields[part] = true
	}
	return fieldSelection{FlagSet: true, Fields: fields}
}

func readJSONObjectArg(arg string) (map[string]any, error) {
	if strings.TrimSpace(arg) == "" {
		return nil, nil
	}
	data, err := shared.ReadJSONInputArg(arg)
	if err != nil {
		return nil, err
	}
	var decoded map[string]any
	if err := output.UnmarshalUseNumber(data, &decoded); err != nil {
		return nil, fmt.Errorf("parsing JSON object: %w", err)
	}
	return decoded, nil
}

func validateNoReadOnlyOrRelationshipFields(label string, value map[string]any, readOnly, relationships map[string]bool) error {
	if len(value) == 0 {
		return nil
	}
	var invalid []string
	for key := range value {
		if readOnly[key] || relationships[key] {
			invalid = append(invalid, key)
		}
	}
	if len(invalid) == 0 {
		return nil
	}
	sort.Strings(invalid)
	return shared.ValidationErrorf("%s cannot override read-only or relationship fields: %s", label, strings.Join(invalid, ", "))
}

func decodeDataList(raw any) ([]map[string]any, error) {
	switch v := raw.(type) {
	case json.RawMessage:
		var envelope types.ListResponse[map[string]any]
		if err := output.UnmarshalUseNumber(v, &envelope); err == nil && envelope.Data != nil {
			return envelope.Data, nil
		}
		var direct []map[string]any
		if err := output.UnmarshalUseNumber(v, &direct); err == nil {
			return direct, nil
		}
	case []map[string]any:
		return v, nil
	}
	return nil, fmt.Errorf("unexpected list response")
}

func decodeDataObject(raw any) (map[string]any, error) {
	switch v := raw.(type) {
	case json.RawMessage:
		var envelope types.DataResponse[map[string]any]
		if err := output.UnmarshalUseNumber(v, &envelope); err == nil && envelope.Data != nil {
			return envelope.Data, nil
		}
		var direct map[string]any
		if err := output.UnmarshalUseNumber(v, &direct); err == nil {
			return direct, nil
		}
	case map[string]any:
		return v, nil
	}
	return nil, fmt.Errorf("unexpected object response")
}

func fetchAllList(ctx context.Context, client *api.Client, req api.Request) ([]map[string]any, error) {
	var all []map[string]any
	offset := 0

	for {
		pageReq := &paginatedRequest{
			inner:  req,
			offset: offset,
			limit:  pageSize,
		}

		var raw json.RawMessage
		if err := client.Do(ctx, pageReq, &raw); err != nil {
			return nil, fmt.Errorf("fetching page at offset %d: %w", offset, err)
		}

		rows, total, err := decodeListResponse(raw)
		if err != nil {
			return nil, fmt.Errorf("decoding page at offset %d: %w", offset, err)
		}
		all = append(all, rows...)

		if total < 0 || offset+len(rows) >= total || len(rows) == 0 {
			break
		}
		offset += len(rows)
	}

	return all, nil
}

func fetchAllFind(ctx context.Context, client *api.Client, flags *shared.SelectorFlags, run func(json.RawMessage) (any, error)) ([]map[string]any, error) {
	var rows []map[string]any
	offset := 0
	for {
		selector, err := flags.Build(pageSize, offset, "")
		if err != nil {
			return nil, err
		}
		selector, err = setSelectorPagination(selector, offset, pageSize)
		if err != nil {
			return nil, err
		}
		raw, err := run(selector)
		if err != nil {
			return nil, err
		}
		page, total, err := decodeListResponse(raw)
		if err != nil {
			return nil, err
		}
		rows = append(rows, page...)
		if total < 0 || offset+len(page) >= total || len(page) == 0 {
			break
		}
		offset += len(page)
	}
	return rows, nil
}

func decodeListResponse(raw any) ([]map[string]any, int, error) {
	switch v := raw.(type) {
	case json.RawMessage:
		var envelope types.ListResponse[map[string]any]
		if err := output.UnmarshalUseNumber(v, &envelope); err != nil {
			return nil, -1, err
		}
		total := -1
		if envelope.Pagination != nil {
			total = envelope.Pagination.TotalResults
		}
		return envelope.Data, total, nil
	default:
		rows, err := decodeDataList(raw)
		return rows, -1, err
	}
}

func setSelectorPagination(selector json.RawMessage, offset, limit int) (json.RawMessage, error) {
	var payload map[string]any
	if err := output.UnmarshalUseNumber(selector, &payload); err != nil {
		return nil, fmt.Errorf("parsing selector: %w", err)
	}
	payload["pagination"] = map[string]any{
		"offset": offset,
		"limit":  limit,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("building selector: %w", err)
	}
	return json.RawMessage(data), nil
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = deepClone(value)
	}
	return out
}

func deepClone(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneMap(v)
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, deepClone(item))
		}
		return out
	default:
		return v
	}
}

func filterEntityForExport(source map[string]any, selection fieldSelection, allowed, required, readOnly, relationships map[string]bool, omitFn func(string, any, map[string]any) bool) map[string]any {
	if selection.All {
		return cloneMap(source)
	}

	targetFields := map[string]bool{}
	if selection.FlagSet {
		for key := range selection.Fields {
			targetFields[key] = true
		}
		for key := range required {
			targetFields[key] = true
		}
	} else {
		for key := range allowed {
			targetFields[key] = true
		}
	}

	out := map[string]any{}
	for key, value := range source {
		if !targetFields[key] {
			continue
		}
		if !selection.FlagSet && (readOnly[key] || relationships[key]) {
			continue
		}
		if !selection.FlagSet && key == "targetingDimensions" {
			value = normalizeTargetingDimensionsForExport(value)
		}
		if !selection.FlagSet && omitFn != nil && omitFn(key, value, source) {
			continue
		}
		out[key] = deepClone(value)
	}
	return out
}

func shouldStripTimeFieldForExport(key string, selection fieldSelection, noTimes bool) bool {
	if !noTimes {
		return false
	}
	if key != "startTime" && key != "endTime" {
		return false
	}
	if !selection.FlagSet {
		return true
	}
	if selection.All {
		return false
	}
	return !selection.Fields[key]
}

func shouldStripCampaignFieldForExport(key string, selection fieldSelection, enabled bool) bool {
	if !enabled {
		return false
	}
	if !selection.FlagSet {
		return true
	}
	if selection.All {
		return false
	}
	return !selection.Fields[key]
}

func shouldSkipChildEntityExport(noFlag bool, selection fieldSelection) bool {
	return noFlag && !selection.FlagSet
}

func redactCampaignExportName(name string, campaign map[string]any, appVars map[string]string) string {
	name = redactAppNameTokens(name, appVars)
	countriesToken := templateValue(campaign["countriesOrRegions"])
	if strings.TrimSpace(countriesToken) != "" {
		name = boundedReplaceAllFold(name, countriesToken, "%(countriesOrRegions)")
	}
	return name
}

func redactAdgroupExportName(name string, originalCampaignName string, appVars map[string]string) string {
	if strings.TrimSpace(originalCampaignName) != "" {
		name = boundedReplaceAllFold(name, originalCampaignName, "%(campaignName)")
	}
	name = redactAppNameTokens(name, appVars)
	return name
}

func redactAppNameTokens(name string, appVars map[string]string) string {
	appName := strings.TrimSpace(appVars["appName"])
	appNameShort := strings.TrimSpace(appVars["appNameShort"])
	if appName != "" {
		name = boundedReplaceAllFold(name, appName, "%(appName)")
	}
	if appNameShort != "" {
		name = boundedReplaceAllFold(name, appNameShort, "%(appNameShort)")
	}
	if appNameShort != "" {
		name = replaceFuzzyPrefixComponent(name, appNameShort, "%(appNameShort)")
	}
	if appName != "" {
		name = replaceFuzzyPrefixComponent(name, appName, "%(appName)")
	}
	return name
}

func boundedReplaceAllFold(s string, target string, repl string) string {
	target = strings.TrimSpace(target)
	if s == "" || target == "" {
		return s
	}
	lowerS := strings.ToLower(s)
	lowerTarget := strings.ToLower(target)
	var out strings.Builder
	last := 0
	searchFrom := 0
	for {
		rel := strings.Index(lowerS[searchFrom:], lowerTarget)
		if rel < 0 {
			break
		}
		idx := searchFrom + rel
		end := idx + len(target)
		if hasValidNameBoundaryLeft(s, idx) && hasValidNameBoundaryRight(s, end) {
			out.WriteString(s[last:idx])
			out.WriteString(repl)
			last = end
		}
		searchFrom = idx + len(target)
	}
	if last == 0 {
		return s
	}
	out.WriteString(s[last:])
	return out.String()
}

func replaceFuzzyPrefixComponent(s string, target string, repl string) string {
	targetNorm := normalizeRedactionToken(target)
	if len(targetNorm) < 6 {
		return s
	}
	type replacement struct {
		start int
		end   int
	}
	var reps []replacement
	for _, comp := range nameComponents(s) {
		candidate := strings.TrimSpace(s[comp.start:comp.end])
		candidateNorm := normalizeRedactionToken(candidate)
		if len(candidateNorm) < 6 {
			continue
		}
		if candidateNorm == targetNorm {
			continue
		}
		if !strings.HasPrefix(targetNorm, candidateNorm) {
			continue
		}
		if len(candidateNorm)*100 < len(targetNorm)*80 {
			continue
		}
		reps = append(reps, replacement(comp))
	}
	if len(reps) == 0 {
		return s
	}
	var out strings.Builder
	last := 0
	for _, rep := range reps {
		out.WriteString(s[last:rep.start])
		out.WriteString(repl)
		last = rep.end
	}
	out.WriteString(s[last:])
	return out.String()
}

type nameComponent struct {
	start int
	end   int
}

func nameComponents(s string) []nameComponent {
	var out []nameComponent
	start := 0
	for i := 0; i < len(s); i++ {
		if !isRedactionSeparatorByte(s[i]) {
			continue
		}
		if comp, ok := trimmedNameComponent(s, start, i); ok {
			out = append(out, comp)
		}
		start = i + 1
	}
	if comp, ok := trimmedNameComponent(s, start, len(s)); ok {
		out = append(out, comp)
	}
	return out
}

func trimmedNameComponent(s string, start, end int) (nameComponent, bool) {
	for start < end && isAppNameShortSpaceByte(s[start]) {
		start++
	}
	for end > start && isAppNameShortSpaceByte(s[end-1]) {
		end--
	}
	if start >= end {
		return nameComponent{}, false
	}
	return nameComponent{start: start, end: end}, true
}

func hasValidNameBoundaryLeft(s string, start int) bool {
	if start == 0 {
		return true
	}
	i := start - 1
	if !isAppNameShortSpaceByte(s[i]) {
		return isRedactionSeparatorByte(s[i])
	}
	for i >= 0 && isAppNameShortSpaceByte(s[i]) {
		i--
	}
	if i < 0 {
		return false
	}
	return isRedactionSeparatorByte(s[i])
}

func hasValidNameBoundaryRight(s string, end int) bool {
	if end == len(s) {
		return true
	}
	i := end
	if !isAppNameShortSpaceByte(s[i]) {
		return isRedactionSeparatorByte(s[i])
	}
	for i < len(s) && isAppNameShortSpaceByte(s[i]) {
		i++
	}
	if i >= len(s) {
		return false
	}
	return isRedactionSeparatorByte(s[i])
}

func isRedactionSeparatorByte(b byte) bool {
	switch b {
	case '-', ':', ',', '|', '/', '(', ')', '[', ']':
		return true
	default:
		return false
	}
}

func normalizeRedactionToken(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func omitCampaignField(key string, value any, _ map[string]any) bool {
	switch key {
	case "adChannelType":
		return strings.EqualFold(stringValue(value), "SEARCH")
	case "billingEvent":
		return strings.EqualFold(stringValue(value), "TAPS")
	case "supplySources":
		return isEmptyValue(value) || isDefaultSupplySources(value)
	case "paymentModel":
		return strings.EqualFold(stringValue(value), "PAYG")
	case "status":
		return strings.EqualFold(stringValue(value), "ENABLED")
	case "startTime":
		return isEmptyValue(value) || isPastStartTime(value)
	case "locInvoiceDetails", "budgetAmount", "endTime":
		return isEmptyValue(value)
	default:
		return isEmptyValue(value)
	}
}

func omitAdgroupField(key string, value any, _ map[string]any) bool {
	switch key {
	case "pricingModel":
		return strings.EqualFold(stringValue(value), "CPC")
	case "paymentModel":
		return strings.EqualFold(stringValue(value), "PAYG")
	case "status":
		return strings.EqualFold(stringValue(value), "ENABLED")
	case "automatedKeywordsOptIn":
		return !boolValue(value)
	case "startTime":
		return isEmptyValue(value) || isPastStartTime(value)
	case "targetingDimensions":
		return isEmptyValue(value)
	default:
		return isEmptyValue(value)
	}
}

func omitKeywordField(key string, value any, source map[string]any) bool {
	switch key {
	case "status":
		return strings.EqualFold(stringValue(value), "ACTIVE")
	case "bidAmount":
		return isEmptyValue(value) || moneyEquals(value, source["defaultBidAmount"])
	default:
		return isEmptyValue(value)
	}
}

func omitNegativeField(key string, value any, _ map[string]any) bool {
	switch key {
	case "status":
		return strings.EqualFold(stringValue(value), "ACTIVE")
	default:
		return isEmptyValue(value)
	}
}

func isEmptyValue(value any) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(v) == ""
	case json.Number:
		return false
	case []any:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case []map[string]any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	case bool:
		return false
	default:
		return false
	}
}

func isPastStartTime(value any) bool {
	raw := strings.TrimSpace(stringValue(value))
	if raw == "" {
		return false
	}
	parsed, err := shared.ParseDateTimeFlag(raw)
	if err != nil {
		return false
	}
	t, err := time.Parse("2006-01-02T15:04:05.000", parsed)
	if err != nil {
		return false
	}
	return t.Before(time.Now().UTC())
}

func moneyEquals(left, right any) bool {
	lm, lok := left.(map[string]any)
	rm, rok := right.(map[string]any)
	if !lok || !rok {
		return false
	}
	return strings.EqualFold(stringValue(lm["amount"]), stringValue(rm["amount"])) &&
		strings.EqualFold(stringValue(lm["currency"]), stringValue(rm["currency"]))
}

func isDefaultSupplySources(value any) bool {
	switch v := value.(type) {
	case []string:
		return len(v) == 1 && strings.EqualFold(v[0], string(types.SupplySourceAppStoreSearchResults))
	case []any:
		return len(v) == 1 && strings.EqualFold(stringValue(v[0]), string(types.SupplySourceAppStoreSearchResults))
	default:
		return false
	}
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func mapStringValue(value map[string]any, key string) string {
	if value == nil {
		return ""
	}
	raw, ok := value[key]
	if !ok || raw == nil {
		return ""
	}
	text, ok := raw.(string)
	if !ok {
		return ""
	}
	return text
}

func boolValue(value any) bool {
	b, _ := value.(bool)
	return b
}

func parseStructureInput(arg string) (structureRoot, error) {
	var root structureRoot
	data, err := shared.ReadJSONInputArg(arg)
	if err != nil {
		return root, err
	}
	if err := output.UnmarshalUseNumber(data, &root); err != nil {
		return root, fmt.Errorf("parsing structure JSON: %w", err)
	}
	if root.Type != structureTypeStructure {
		return root, shared.ValidationErrorf("invalid structure type %q: expected %q", root.Type, structureTypeStructure)
	}
	if root.SchemaVersion != 1 {
		return root, shared.ValidationErrorf("unsupported schemaVersion %d", root.SchemaVersion)
	}
	switch root.Scope {
	case scopeCampaigns:
		if len(root.Campaigns) == 0 {
			return root, shared.ValidationError("--from-structure scope campaigns requires campaigns data")
		}
	case scopeAdgroups:
		if len(root.Adgroups) == 0 {
			return root, shared.ValidationError("--from-structure scope adgroups requires adgroups data")
		}
	default:
		return root, shared.ValidationErrorf("invalid structure scope %q: use campaigns or adgroups", root.Scope)
	}
	return root, nil
}

func selectorFieldsForEntity(selection fieldSelection, allowed, required, readOnly, relationships map[string]bool) []string {
	fields := map[string]bool{}
	if selection.All {
		for key := range allowed {
			fields[key] = true
		}
		for key := range required {
			fields[key] = true
		}
		for key := range readOnly {
			fields[key] = true
		}
		for key := range relationships {
			fields[key] = true
		}
	} else if selection.FlagSet {
		for key := range selection.Fields {
			fields[key] = true
		}
		for key := range required {
			fields[key] = true
		}
	} else {
		for key := range allowed {
			fields[key] = true
		}
		for key := range required {
			fields[key] = true
		}
	}
	fields["id"] = true

	var out []string
	for key := range fields {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func ensureSelectorFields(selector json.RawMessage, requiredFields []string) (json.RawMessage, error) {
	if len(requiredFields) == 0 {
		return selector, nil
	}

	var payload map[string]any
	if err := output.UnmarshalUseNumber(selector, &payload); err != nil {
		return nil, fmt.Errorf("parsing selector: %w", err)
	}

	fieldSet := map[string]bool{}
	var existing []string
	switch raw := payload["fields"].(type) {
	case []any:
		for _, value := range raw {
			if s, ok := value.(string); ok {
				fieldSet[s] = true
				existing = append(existing, s)
			}
		}
	case []string:
		for _, value := range raw {
			fieldSet[value] = true
			existing = append(existing, value)
		}
	case nil:
	default:
		return nil, fmt.Errorf("invalid selector fields")
	}

	for _, field := range requiredFields {
		if !fieldSet[field] {
			existing = append(existing, field)
			fieldSet[field] = true
		}
	}
	payload["fields"] = existing

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("building selector: %w", err)
	}
	return json.RawMessage(data), nil
}

func writeMappingOutput(root mappingRoot, target string, emitStdout bool, pretty bool) error {
	if target != "" && target != "-" {
		dir := filepath.Dir(target)
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("writing mapping: %w", err)
		}
		data, err := marshalJSON(root)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return fmt.Errorf("writing mapping: %w", err)
		}
	}
	if emitStdout || target == "-" {
		return output.PrintJSON(root, pretty)
	}
	return nil
}

func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var templateToken = regexp.MustCompile(`%\(([A-Za-z0-9_]+)\)|%([A-Za-z0-9_]+)`)

func applyNameTemplate(originalName string, item map[string]any, template string, pattern string, allowUnmatched bool, extraVars map[string]string) (string, error) {
	if strings.TrimSpace(template) == "" {
		return originalName, nil
	}
	captures := map[string]string{}
	if strings.TrimSpace(pattern) != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", fmt.Errorf("invalid name pattern %q: %w", pattern, err)
		}
		match := re.FindStringSubmatch(originalName)
		if match == nil {
			if !allowUnmatched {
				return "", shared.ValidationErrorf("name pattern %q did not match %q", pattern, originalName)
			}
		} else {
			for i := 1; i < len(match); i++ {
				captures[strconv.Itoa(i)] = match[i]
			}
		}
	}

	vars := buildTemplateVariables(item)
	for key, value := range extraVars {
		vars[key] = value
	}
	for key, value := range captures {
		vars[key] = value
	}

	rendered := templateToken.ReplaceAllStringFunc(template, func(token string) string {
		match := templateToken.FindStringSubmatch(token)
		name := match[1]
		if name == "" {
			name = match[2]
		}
		if value, ok := vars[name]; ok {
			return value
		}
		return ""
	})
	if strings.TrimSpace(rendered) == "" {
		return "", shared.ValidationError("rendered name cannot be empty")
	}
	return rendered, nil
}

func buildTemplateVariables(item map[string]any) map[string]string {
	out := map[string]string{}
	for key, value := range item {
		out[key] = templateValue(value)
		out[columnname.FromField(key)] = templateValue(value)
	}
	return out
}

func templateReferencesVariable(template string, names ...string) bool {
	if strings.TrimSpace(template) == "" || len(names) == 0 {
		return false
	}
	candidates := stringSet(names...)
	for _, match := range templateToken.FindAllStringSubmatch(template, -1) {
		name := match[1]
		if name == "" {
			name = match[2]
		}
		if candidates[name] {
			return true
		}
	}
	return false
}

func templatesReferenceAnyVariable(templates []string, names ...string) bool {
	for _, template := range templates {
		if templateReferencesVariable(template, names...) {
			return true
		}
	}
	return false
}

func applyImportedNameTemplate(template string, item map[string]any, extraVars map[string]string) (string, error) {
	if strings.TrimSpace(template) == "" {
		return "", nil
	}
	return applyNameTemplate(template, item, template, "", true, extraVars)
}

func applyCampaignCreateDefaults(payload map[string]any) map[string]any {
	if isEmptyValue(payload["adChannelType"]) {
		payload["adChannelType"] = "SEARCH"
	}
	if isEmptyValue(payload["billingEvent"]) {
		payload["billingEvent"] = "TAPS"
	}
	if isEmptyValue(payload["supplySources"]) {
		payload["supplySources"] = []string{string(types.SupplySourceAppStoreSearchResults)}
	}
	return payload
}

func applyAdgroupCreateDefaults(payload map[string]any) map[string]any {
	if isEmptyValue(payload["pricingModel"]) {
		payload["pricingModel"] = "CPC"
	}
	if _, ok := payload["automatedKeywordsOptIn"]; !ok {
		payload["automatedKeywordsOptIn"] = false
	}
	return payload
}

func templateValue(value any) string {
	switch v := value.(type) {
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, templateValue(item))
		}
		return strings.Join(parts, ",")
	case []string:
		return strings.Join(v, ",")
	case map[string]any:
		return stringValue(v)
	default:
		return stringValue(value)
	}
}

func deriveAppNameShortValue(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	components := nameComponents(name)
	if len(components) == 0 {
		return name
	}
	return strings.TrimSpace(name[components[0].start:components[0].end])
}

func isAppNameShortSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func normalizeCampaignForImport(source map[string]any) map[string]any {
	return applyCampaignCreateDefaults(sanitizeEntity(source, campaignReadOnlyFields, campaignRelationshipFields))
}

func normalizeAdgroupForImport(source map[string]any) map[string]any {
	return applyAdgroupCreateDefaults(sanitizeEntity(source, adgroupImportReadOnlyFields, adgroupRelationshipFields))
}

func normalizeKeywordForImport(source map[string]any) map[string]any {
	return sanitizeEntity(source, keywordReadOnlyFields, keywordRelationshipFields)
}

func normalizeNegativeForImport(source map[string]any) map[string]any {
	return sanitizeEntity(source, negativeReadOnlyFields, negativeRelationshipFields)
}

func sanitizeEntity(source map[string]any, readOnly, relationships map[string]bool) map[string]any {
	out := map[string]any{}
	for key, value := range source {
		if readOnly[key] || relationships[key] {
			continue
		}
		out[key] = deepClone(value)
	}
	return out
}

func mergeMaps(base map[string]any, overlay map[string]any) map[string]any {
	out := cloneMap(base)
	for key, value := range overlay {
		out[key] = deepClone(value)
	}
	return out
}

func mustMarshalRaw(value any) json.RawMessage {
	data, _ := json.Marshal(value)
	return data
}

func ensureRequiredFields(entity string, payload map[string]any, required map[string]bool) error {
	var missing []string
	for field := range required {
		if isEmptyValue(payload[field]) {
			missing = append(missing, field)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return shared.ValidationErrorf("%s is missing required fields: %s", entity, strings.Join(missing, ", "))
}

func ensureRequiredFieldsWithFlags(entity string, payload map[string]any, required map[string]bool, flagsByField map[string]string) error {
	var missing []string
	for field := range required {
		if isEmptyValue(payload[field]) {
			if flagName := strings.TrimSpace(flagsByField[field]); flagName != "" {
				missing = append(missing, fmt.Sprintf("%s (include it in the structure JSON or pass %s)", field, flagName))
				continue
			}
			missing = append(missing, field)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return shared.ValidationErrorf("%s is missing required fields: %s", entity, strings.Join(missing, ", "))
}

func canonicalName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func collectDuplicateNames(names []string) []string {
	seen := map[string]int{}
	for _, name := range names {
		seen[canonicalName(name)]++
	}
	var duplicates []string
	for _, name := range names {
		if seen[canonicalName(name)] > 1 {
			duplicates = append(duplicates, name)
		}
	}
	sort.Strings(duplicates)
	return compactSortedStrings(duplicates)
}

func compactSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	sort.Strings(values)
	out := []string{values[0]}
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}

func normalizeTargetingDimensionsForExport(value any) any {
	raw, ok := value.(map[string]any)
	if !ok || len(raw) == 0 {
		return value
	}

	out := map[string]any{}
	for key, item := range raw {
		if item == nil {
			continue
		}
		if key == "deviceClass" && isDefaultDeviceClassTargeting(item) {
			continue
		}
		if isEmptyValue(item) {
			continue
		}
		out[key] = item
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isDefaultDeviceClassTargeting(value any) bool {
	raw, ok := value.(map[string]any)
	if !ok {
		return false
	}
	if excluded := raw["excluded"]; excluded != nil && !isEmptyValue(excluded) {
		return false
	}
	included, ok := raw["included"].([]any)
	if !ok || len(included) != 2 {
		return false
	}
	seen := map[string]bool{}
	for _, item := range included {
		seen[strings.ToUpper(stringValue(item))] = true
	}
	return seen["IPHONE"] && seen["IPAD"]
}
