package types

import (
	"encoding/json"
	"testing"
)

// ---------- Money ----------

func TestMoney_JSON(t *testing.T) {
	m := Money{Amount: "10.50", Currency: "USD"}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"amount":"10.50","currency":"USD"}`
	if string(data) != expected {
		t.Errorf("Marshal = %s, want %s", data, expected)
	}

	var decoded Money
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != m {
		t.Errorf("Unmarshal = %v, want %v", decoded, m)
	}
}

func TestMoney_JSON_ZeroAmount(t *testing.T) {
	m := Money{Amount: "0", Currency: "EUR"}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"amount":"0","currency":"EUR"}`
	if string(data) != expected {
		t.Errorf("Marshal = %s, want %s", data, expected)
	}

	var decoded Money
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != m {
		t.Errorf("Unmarshal = %v, want %v", decoded, m)
	}
}

// ---------- Campaign ----------

func TestCampaign_JSON_Minimal(t *testing.T) {
	c := Campaign{
		AdamID:             789,
		Name:               "Test Campaign",
		AdChannelType:      CampaignAdChannelTypeSearch,
		CountriesOrRegions: []string{"US"},
		BillingEvent:       CampaignBillingEventTypeImpressions,
		DailyBudgetAmount:  Money{Amount: "100", Currency: "USD"},
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// Verify camelCase JSON field names
	requiredFields := []string{"adamId", "name", "adChannelType", "countriesOrRegions", "billingEvent", "dailyBudgetAmount"}
	for _, f := range requiredFields {
		if _, ok := m[f]; !ok {
			t.Errorf("JSON missing required field %q", f)
		}
	}

	// Omitempty fields should not appear
	omittedFields := []string{"id", "orgId", "status", "displayStatus", "servingStatus",
		"servingStateReasons", "deleted", "creationTime", "modificationTime",
		"startTime", "endTime", "budgetAmount", "targetCpa", "budgetOrders", "paymentModel",
		"locInvoiceDetails", "supplySources", "countryOrRegionServingStateReasons"}
	for _, f := range omittedFields {
		if _, ok := m[f]; ok {
			t.Errorf("JSON should omit field %q when nil/empty, but it was present", f)
		}
	}

	// Roundtrip
	var decoded Campaign
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.AdamID != c.AdamID {
		t.Errorf("AdamID = %d, want %d", decoded.AdamID, c.AdamID)
	}
	if decoded.Name != c.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, c.Name)
	}
	if decoded.AdChannelType != c.AdChannelType {
		t.Errorf("AdChannelType = %q, want %q", decoded.AdChannelType, c.AdChannelType)
	}
	if decoded.BillingEvent != c.BillingEvent {
		t.Errorf("BillingEvent = %q, want %q", decoded.BillingEvent, c.BillingEvent)
	}
}

func TestCampaign_JSON_AllFields(t *testing.T) {
	id := int64(123)
	orgID := int64(456)
	status := CampaignStatusEnabled
	displayStatus := CampaignDisplayStatusRunning
	servingStatus := CampaignServingStatusRunning
	deleted := false
	paymentModel := PaymentModelPAYG
	creationTime := "2024-01-01T00:00:00.000"
	modificationTime := "2024-06-15T12:00:00.000"
	startTime := "2024-01-01T00:00:00.000"
	endTime := "2024-12-31T23:59:59.000"
	budgetAmount := Money{Amount: "1000", Currency: "USD"}
	targetCpa := Money{Amount: "50", Currency: "USD"}

	c := Campaign{
		ID:            &id,
		OrgID:         &orgID,
		AdamID:        789,
		Name:          "Full Campaign",
		AdChannelType: CampaignAdChannelTypeSearch,
		SupplySources: []SupplySource{
			SupplySourceAppStoreSearchResults,
			SupplySourceAppStoreTodayTab,
		},
		Status:        &status,
		DisplayStatus: &displayStatus,
		ServingStatus: &servingStatus,
		ServingStateReasons: []CampaignServingStateReason{
			CampaignServingStateReasonAdGroupMissing,
			CampaignServingStateReasonBOStartDateInFuture,
		},
		CountriesOrRegions: []string{"US", "GB"},
		CountryOrRegionServingStateReasons: map[string][]CampaignCountryOrRegionsServingStateReason{
			"US": {CampaignCountryReasonAppNotEligible},
		},
		BillingEvent:      CampaignBillingEventTypeImpressions,
		BudgetAmount:      &budgetAmount,
		DailyBudgetAmount: Money{Amount: "100", Currency: "USD"},
		TargetCpa:         &targetCpa,
		BudgetOrders:      []int64{246, 135},
		PaymentModel:      &paymentModel,
		Deleted:           &deleted,
		CreationTime:      &creationTime,
		ModificationTime:  &modificationTime,
		StartTime:         &startTime,
		EndTime:           &endTime,
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Campaign
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if *decoded.ID != *c.ID {
		t.Errorf("ID = %d, want %d", *decoded.ID, *c.ID)
	}
	if *decoded.OrgID != *c.OrgID {
		t.Errorf("OrgID = %d, want %d", *decoded.OrgID, *c.OrgID)
	}
	if decoded.AdamID != c.AdamID {
		t.Errorf("AdamID = %d, want %d", decoded.AdamID, c.AdamID)
	}
	if decoded.Name != c.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, c.Name)
	}
	if *decoded.Status != *c.Status {
		t.Errorf("Status = %q, want %q", *decoded.Status, *c.Status)
	}
	if *decoded.DisplayStatus != *c.DisplayStatus {
		t.Errorf("DisplayStatus = %q, want %q", *decoded.DisplayStatus, *c.DisplayStatus)
	}
	if *decoded.ServingStatus != *c.ServingStatus {
		t.Errorf("ServingStatus = %q, want %q", *decoded.ServingStatus, *c.ServingStatus)
	}
	if len(decoded.SupplySources) != 2 {
		t.Errorf("SupplySources length = %d, want 2", len(decoded.SupplySources))
	}
	if len(decoded.ServingStateReasons) != 2 {
		t.Errorf("ServingStateReasons length = %d, want 2", len(decoded.ServingStateReasons))
	}
	if len(decoded.CountriesOrRegions) != 2 {
		t.Errorf("CountriesOrRegions length = %d, want 2", len(decoded.CountriesOrRegions))
	}
	if *decoded.PaymentModel != *c.PaymentModel {
		t.Errorf("PaymentModel = %q, want %q", *decoded.PaymentModel, *c.PaymentModel)
	}
	if *decoded.Deleted != *c.Deleted {
		t.Errorf("Deleted = %v, want %v", *decoded.Deleted, *c.Deleted)
	}
}

func TestCampaignStatus_Enum(t *testing.T) {
	tests := []struct {
		status CampaignStatus
		want   string
	}{
		{CampaignStatusEnabled, "ENABLED"},
		{CampaignStatusPaused, "PAUSED"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.status)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.status, data, tt.want)
		}
	}
}

func TestCampaignDisplayStatus_Enum(t *testing.T) {
	tests := []struct {
		status CampaignDisplayStatus
		want   string
	}{
		{CampaignDisplayStatusRunning, "RUNNING"},
		{CampaignDisplayStatusOnHold, "ON_HOLD"},
		{CampaignDisplayStatusPaused, "PAUSED"},
		{CampaignDisplayStatusDeleted, "DELETED"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.status)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.status, data, tt.want)
		}
	}
}

func TestCampaignServingStatus_Enum(t *testing.T) {
	tests := []struct {
		status CampaignServingStatus
		want   string
	}{
		{CampaignServingStatusRunning, "RUNNING"},
		{CampaignServingStatusNotRunning, "NOT_RUNNING"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.status)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.status, data, tt.want)
		}
	}
}

func TestCampaignAdChannelType_Enum(t *testing.T) {
	tests := []struct {
		ch   CampaignAdChannelType
		want string
	}{
		{CampaignAdChannelTypeSearch, "SEARCH"},
		{CampaignAdChannelTypeDisplay, "DISPLAY"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.ch)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.ch, data, tt.want)
		}
	}
}

func TestCampaignBillingEventType_Enum(t *testing.T) {
	tests := []struct {
		be   CampaignBillingEventType
		want string
	}{
		{CampaignBillingEventTypeTaps, "TAPS"},
		{CampaignBillingEventTypeImpressions, "IMPRESSIONS"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.be)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.be, data, tt.want)
		}
	}
}

func TestSupplySource_Enum(t *testing.T) {
	tests := []struct {
		ss   SupplySource
		want string
	}{
		{SupplySourceAppStoreProductPagesBrowse, "APPSTORE_PRODUCT_PAGES_BROWSE"},
		{SupplySourceAppStoreSearchResults, "APPSTORE_SEARCH_RESULTS"},
		{SupplySourceAppStoreSearchTab, "APPSTORE_SEARCH_TAB"},
		{SupplySourceAppStoreTodayTab, "APPSTORE_TODAY_TAB"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.ss)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.ss, data, tt.want)
		}
	}
}

func TestPaymentModel_Enum(t *testing.T) {
	tests := []struct {
		pm   PaymentModel
		want string
	}{
		{PaymentModelPAYG, "PAYG"},
		{PaymentModelLOC, "LOC"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.pm)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.pm, data, tt.want)
		}
	}
}

// ---------- AdGroup ----------

func TestAdGroup_JSON(t *testing.T) {
	id := int64(501)
	orgID := int64(100)
	campID := int64(200)
	status := AdGroupStatusEnabled
	servingStatus := AdGroupServingStatusRunning
	displayStatus := AdGroupDisplayStatusRunning
	autoKeywords := true

	ag := AdGroup{
		ID:                     &id,
		OrgID:                  &orgID,
		CampaignID:             &campID,
		Status:                 &status,
		ServingStatus:          &servingStatus,
		DisplayStatus:          &displayStatus,
		Name:                   "My Ad Group",
		PricingModel:           PricingModelCPC,
		DefaultBidAmount:       Money{Amount: "1.50", Currency: "USD"},
		AutomatedKeywordsOptIn: &autoKeywords,
		StartTime:              "2024-01-01T00:00:00.000",
	}

	data, err := json.Marshal(ag)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// Verify camelCase field names
	expectedFields := []string{"id", "orgId", "campaignId", "status", "servingStatus",
		"displayStatus", "name", "pricingModel", "defaultBidAmount",
		"automatedKeywordsOptIn", "startTime"}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			t.Errorf("JSON missing field %q", f)
		}
	}

	// Omitempty fields
	omittedFields := []string{"cpaGoal", "deleted", "modificationTime", "endTime",
		"targetingDimensions", "paymentModel", "servingStateReasons"}
	for _, f := range omittedFields {
		if _, ok := m[f]; ok {
			t.Errorf("JSON should omit field %q when nil/empty, but it was present", f)
		}
	}

	// Roundtrip
	var decoded AdGroup
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if *decoded.ID != *ag.ID {
		t.Errorf("ID = %d, want %d", *decoded.ID, *ag.ID)
	}
	if decoded.Name != ag.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, ag.Name)
	}
	if decoded.PricingModel != ag.PricingModel {
		t.Errorf("PricingModel = %q, want %q", decoded.PricingModel, ag.PricingModel)
	}
	if decoded.DefaultBidAmount != ag.DefaultBidAmount {
		t.Errorf("DefaultBidAmount = %v, want %v", decoded.DefaultBidAmount, ag.DefaultBidAmount)
	}
	if *decoded.AutomatedKeywordsOptIn != *ag.AutomatedKeywordsOptIn {
		t.Errorf("AutomatedKeywordsOptIn = %v, want %v", *decoded.AutomatedKeywordsOptIn, *ag.AutomatedKeywordsOptIn)
	}
}

func TestAdGroupStatus_Enum(t *testing.T) {
	tests := []struct {
		status AdGroupStatus
		want   string
	}{
		{AdGroupStatusEnabled, "ENABLED"},
		{AdGroupStatusPaused, "PAUSED"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.status)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.status, data, tt.want)
		}
	}
}

func TestPricingModel_Enum(t *testing.T) {
	tests := []struct {
		pm   PricingModel
		want string
	}{
		{PricingModelCPC, "CPC"},
		{PricingModelCPM, "CPM"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.pm)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.pm, data, tt.want)
		}
	}
}

// ---------- Keyword ----------

func TestKeyword_JSON(t *testing.T) {
	id := int64(999)
	campID := int64(100)
	adGroupID := int64(200)
	status := KeywordStatusActive

	kw := Keyword{
		ID:         &id,
		CampaignID: &campID,
		AdGroupID:  &adGroupID,
		Text:       "running shoes",
		MatchType:  KeywordMatchTypeExact,
		Status:     &status,
		BidAmount:  &Money{Amount: "2.00", Currency: "USD"},
	}

	data, err := json.Marshal(kw)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// Verify camelCase field names
	expectedFields := []string{"id", "campaignId", "adGroupId", "text", "matchType", "status", "bidAmount"}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			t.Errorf("JSON missing field %q", f)
		}
	}

	// Omitempty fields
	omittedFields := []string{"deleted", "creationTime", "modificationTime"}
	for _, f := range omittedFields {
		if _, ok := m[f]; ok {
			t.Errorf("JSON should omit field %q when nil, but it was present", f)
		}
	}

	// Roundtrip
	var decoded Keyword
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if *decoded.ID != *kw.ID {
		t.Errorf("ID = %d, want %d", *decoded.ID, *kw.ID)
	}
	if decoded.Text != kw.Text {
		t.Errorf("Text = %q, want %q", decoded.Text, kw.Text)
	}
	if decoded.MatchType != kw.MatchType {
		t.Errorf("MatchType = %q, want %q", decoded.MatchType, kw.MatchType)
	}
	if *decoded.Status != *kw.Status {
		t.Errorf("Status = %q, want %q", *decoded.Status, *kw.Status)
	}
	if decoded.BidAmount.Amount != kw.BidAmount.Amount {
		t.Errorf("BidAmount.Amount = %q, want %q", decoded.BidAmount.Amount, kw.BidAmount.Amount)
	}
}

func TestKeywordMatchType_Enum(t *testing.T) {
	tests := []struct {
		mt   KeywordMatchType
		want string
	}{
		{KeywordMatchTypeAuto, "AUTO"},
		{KeywordMatchTypeBroad, "BROAD"},
		{KeywordMatchTypeExact, "EXACT"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.mt)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.mt, data, tt.want)
		}
	}
}

func TestKeywordStatus_Enum(t *testing.T) {
	tests := []struct {
		status KeywordStatus
		want   string
	}{
		{KeywordStatusActive, "ACTIVE"},
		{KeywordStatusPaused, "PAUSED"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.status)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.status, data, tt.want)
		}
	}
}

// ---------- Selector ----------

func TestSelector_JSON(t *testing.T) {
	s := Selector{
		Conditions: []Condition{
			{
				Field:    "servingStatus",
				Operator: OperatorEquals,
				Values:   []any{"RUNNING"},
			},
			{
				Field:    "adChannelType",
				Operator: OperatorIn,
				Values:   []any{"SEARCH", "DISPLAY"},
			},
		},
		Fields: []string{"id", "name"},
		OrderBy: []Sorting{
			{Field: "creationTime", SortOrder: SortOrderAscending},
		},
		Pagination: &Pagination{Limit: 20, Offset: 40},
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Selector
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if len(decoded.Conditions) != 2 {
		t.Errorf("Conditions length = %d, want 2", len(decoded.Conditions))
	}
	if decoded.Conditions[0].Field != "servingStatus" {
		t.Errorf("Conditions[0].Field = %q, want %q", decoded.Conditions[0].Field, "servingStatus")
	}
	if decoded.Conditions[0].Operator != OperatorEquals {
		t.Errorf("Conditions[0].Operator = %q, want %q", decoded.Conditions[0].Operator, OperatorEquals)
	}
	if len(decoded.Fields) != 2 {
		t.Errorf("Fields length = %d, want 2", len(decoded.Fields))
	}
	if decoded.Fields[0] != "id" {
		t.Errorf("Fields[0] = %q, want %q", decoded.Fields[0], "id")
	}
	if len(decoded.OrderBy) != 1 {
		t.Errorf("OrderBy length = %d, want 1", len(decoded.OrderBy))
	}
	if decoded.OrderBy[0].SortOrder != SortOrderAscending {
		t.Errorf("OrderBy[0].SortOrder = %q, want %q", decoded.OrderBy[0].SortOrder, SortOrderAscending)
	}
	if decoded.Pagination.Limit != 20 {
		t.Errorf("Pagination.Limit = %d, want 20", decoded.Pagination.Limit)
	}
	if decoded.Pagination.Offset != 40 {
		t.Errorf("Pagination.Offset = %d, want 40", decoded.Pagination.Offset)
	}
}

func TestSelector_JSON_Empty(t *testing.T) {
	s := Selector{}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// All fields use omitempty, so empty selector should produce {}
	for _, f := range []string{"conditions", "fields", "orderBy", "pagination"} {
		if _, ok := m[f]; ok {
			t.Errorf("Empty Selector should omit %q, but it was present", f)
		}
	}
}

func TestOperator_Enum(t *testing.T) {
	tests := []struct {
		op   Operator
		want string
	}{
		{OperatorBetween, "BETWEEN"},
		{OperatorContains, "CONTAINS"},
		{OperatorContainsAll, "CONTAINS_ALL"},
		{OperatorContainsAny, "CONTAINS_ANY"},
		{OperatorEndsWith, "ENDSWITH"},
		{OperatorEquals, "EQUALS"},
		{OperatorGreaterThan, "GREATER_THAN"},
		{OperatorIn, "IN"},
		{OperatorLessThan, "LESS_THAN"},
		{OperatorStartsWith, "STARTSWITH"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.op)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.op, data, tt.want)
		}
	}
}

func TestSortOrder_Enum(t *testing.T) {
	tests := []struct {
		so   SortOrder
		want string
	}{
		{SortOrderAscending, "ASCENDING"},
		{SortOrderDescending, "DESCENDING"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.so)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.so, data, tt.want)
		}
	}
}

// ---------- Pagination ----------

func TestPagination_JSON(t *testing.T) {
	p := Pagination{Limit: 20, Offset: 40}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"limit":20,"offset":40}`
	if string(data) != expected {
		t.Errorf("Marshal = %s, want %s", data, expected)
	}

	var decoded Pagination
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != p {
		t.Errorf("Unmarshal = %v, want %v", decoded, p)
	}
}

func TestPageDetail_JSON(t *testing.T) {
	pd := PageDetail{TotalResults: 100, StartIndex: 0, ItemsPerPage: 20}
	data, err := json.Marshal(pd)
	if err != nil {
		t.Fatal(err)
	}

	var decoded PageDetail
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != pd {
		t.Errorf("Unmarshal = %v, want %v", decoded, pd)
	}
}

// ---------- ErrorItem ----------

func TestErrorItem_JSON(t *testing.T) {
	field := "campaignId"
	msg := "Invalid campaign ID"
	msgCode := "INVALID_FIELD"

	item := ErrorItem{
		Field:       &field,
		Message:     &msg,
		MessageCode: &msgCode,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ErrorItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if *decoded.Field != field {
		t.Errorf("Field = %q, want %q", *decoded.Field, field)
	}
	if *decoded.Message != msg {
		t.Errorf("Message = %q, want %q", *decoded.Message, msg)
	}
	if *decoded.MessageCode != msgCode {
		t.Errorf("MessageCode = %q, want %q", *decoded.MessageCode, msgCode)
	}
}

func TestErrorItem_JSON_OmittedFields(t *testing.T) {
	item := ErrorItem{}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ErrorItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Field != nil {
		t.Errorf("Field should be nil, got %q", *decoded.Field)
	}
	if decoded.Message != nil {
		t.Errorf("Message should be nil, got %q", *decoded.Message)
	}
	if decoded.MessageCode != nil {
		t.Errorf("MessageCode should be nil, got %q", *decoded.MessageCode)
	}
}

// ---------- UserACL ----------

func TestUserACL_JSON(t *testing.T) {
	acl := UserACL{
		OrgID:        12345,
		OrgName:      "My Organization",
		Currency:     "USD",
		PaymentModel: PaymentModelPAYG,
		RoleNames:    []string{"Admin", "Campaign Manager"},
		TimeZone:     "America/New_York",
	}

	data, err := json.Marshal(acl)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// Verify camelCase field names
	expectedFields := []string{"orgId", "orgName", "currency", "paymentModel", "roleNames", "timeZone"}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			t.Errorf("JSON missing field %q", f)
		}
	}

	// ParentOrgID is omitempty
	if _, ok := m["parentOrgId"]; ok {
		t.Error("JSON should omit parentOrgId when nil")
	}

	// Roundtrip
	var decoded UserACL
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.OrgID != acl.OrgID {
		t.Errorf("OrgID = %d, want %d", decoded.OrgID, acl.OrgID)
	}
	if decoded.OrgName != acl.OrgName {
		t.Errorf("OrgName = %q, want %q", decoded.OrgName, acl.OrgName)
	}
	if decoded.Currency != acl.Currency {
		t.Errorf("Currency = %q, want %q", decoded.Currency, acl.Currency)
	}
	if decoded.PaymentModel != acl.PaymentModel {
		t.Errorf("PaymentModel = %q, want %q", decoded.PaymentModel, acl.PaymentModel)
	}
	if len(decoded.RoleNames) != 2 {
		t.Errorf("RoleNames length = %d, want 2", len(decoded.RoleNames))
	}
	if decoded.TimeZone != acl.TimeZone {
		t.Errorf("TimeZone = %q, want %q", decoded.TimeZone, acl.TimeZone)
	}
}

func TestUserACL_JSON_WithParentOrg(t *testing.T) {
	parentOrgID := int64(99)
	acl := UserACL{
		OrgID:        12345,
		OrgName:      "Child Org",
		ParentOrgID:  &parentOrgID,
		Currency:     "GBP",
		PaymentModel: PaymentModelLOC,
		RoleNames:    []string{"Admin"},
		TimeZone:     "Europe/London",
	}

	data, err := json.Marshal(acl)
	if err != nil {
		t.Fatal(err)
	}

	var decoded UserACL
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ParentOrgID == nil {
		t.Fatal("ParentOrgID is nil, want non-nil")
	}
	if *decoded.ParentOrgID != parentOrgID {
		t.Errorf("ParentOrgID = %d, want %d", *decoded.ParentOrgID, parentOrgID)
	}
}

func TestMeDetail_JSON(t *testing.T) {
	me := MeDetail{UserID: 111, ParentOrgID: 222}
	data, err := json.Marshal(me)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"userId":111,"parentOrgId":222}`
	if string(data) != expected {
		t.Errorf("Marshal = %s, want %s", data, expected)
	}

	var decoded MeDetail
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != me {
		t.Errorf("Unmarshal = %v, want %v", decoded, me)
	}
}

// ---------- ReportingRequest ----------

func TestReportingRequest_JSON(t *testing.T) {
	granularity := ReportingGranularityDaily
	returnGrandTotals := true
	returnRowTotals := false
	startTime := "2024-01-01"
	endTime := "2024-01-31"
	timeZone := "UTC"

	rr := ReportingRequest{
		StartTime:         &startTime,
		EndTime:           &endTime,
		TimeZone:          &timeZone,
		Granularity:       &granularity,
		GroupBy:           []ReportingGroupBy{ReportingGroupByCountryOrRegion, ReportingGroupByDeviceClass},
		ReturnGrandTotals: &returnGrandTotals,
		ReturnRowTotals:   &returnRowTotals,
		Selector: &Selector{
			Pagination: &Pagination{Limit: 50, Offset: 0},
		},
	}

	data, err := json.Marshal(rr)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// Verify camelCase field names
	expectedFields := []string{"startTime", "endTime", "timeZone", "granularity",
		"groupBy", "returnGrandTotals", "returnRowTotals", "selector"}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			t.Errorf("JSON missing field %q", f)
		}
	}

	// Roundtrip
	var decoded ReportingRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if *decoded.StartTime != startTime {
		t.Errorf("StartTime = %q, want %q", *decoded.StartTime, startTime)
	}
	if *decoded.Granularity != granularity {
		t.Errorf("Granularity = %q, want %q", *decoded.Granularity, granularity)
	}
	if len(decoded.GroupBy) != 2 {
		t.Errorf("GroupBy length = %d, want 2", len(decoded.GroupBy))
	}
	if *decoded.ReturnGrandTotals != returnGrandTotals {
		t.Errorf("ReturnGrandTotals = %v, want %v", *decoded.ReturnGrandTotals, returnGrandTotals)
	}
}

func TestReportingRequest_JSON_Empty(t *testing.T) {
	rr := ReportingRequest{}

	data, err := json.Marshal(rr)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// All fields use omitempty
	omittedFields := []string{"startTime", "endTime", "timeZone", "granularity",
		"groupBy", "selector", "returnGrandTotals", "returnRecordsWithNoMetrics", "returnRowTotals"}
	for _, f := range omittedFields {
		if _, ok := m[f]; ok {
			t.Errorf("Empty ReportingRequest should omit %q, but it was present", f)
		}
	}
}

func TestReportingGranularity_Enum(t *testing.T) {
	tests := []struct {
		g    ReportingGranularity
		want string
	}{
		{ReportingGranularityHourly, "HOURLY"},
		{ReportingGranularityDaily, "DAILY"},
		{ReportingGranularityWeekly, "WEEKLY"},
		{ReportingGranularityMonthly, "MONTHLY"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.g)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.g, data, tt.want)
		}
	}
}

func TestReportingGroupBy_Enum(t *testing.T) {
	tests := []struct {
		gb   ReportingGroupBy
		want string
	}{
		{ReportingGroupByAdminArea, "adminArea"},
		{ReportingGroupByAgeRange, "ageRange"},
		{ReportingGroupByCountryCode, "countryCode"},
		{ReportingGroupByCountryOrRegion, "countryOrRegion"},
		{ReportingGroupByDeviceClass, "deviceClass"},
		{ReportingGroupByGender, "gender"},
		{ReportingGroupByLocality, "locality"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.gb)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+tt.want+`"` {
			t.Errorf("Marshal(%v) = %s, want %q", tt.gb, data, tt.want)
		}
	}
}

// ---------- SpendRow ----------

func TestSpendRow_JSON(t *testing.T) {
	impressions := 1000
	taps := 50
	tapInstallRate := 0.75
	ttr := 0.05

	sr := SpendRow{
		AvgCPM:         &Money{Amount: "5.00", Currency: "USD"},
		AvgCPT:         &Money{Amount: "1.00", Currency: "USD"},
		Impressions:    &impressions,
		LocalSpend:     &Money{Amount: "50.00", Currency: "USD"},
		Taps:           &taps,
		TapInstallRate: &tapInstallRate,
		TTR:            &ttr,
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatal(err)
	}

	var decoded SpendRow
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if *decoded.Impressions != impressions {
		t.Errorf("Impressions = %d, want %d", *decoded.Impressions, impressions)
	}
	if *decoded.Taps != taps {
		t.Errorf("Taps = %d, want %d", *decoded.Taps, taps)
	}
	if decoded.AvgCPM.Amount != "5.00" {
		t.Errorf("AvgCPM.Amount = %q, want %q", decoded.AvgCPM.Amount, "5.00")
	}
	if *decoded.TapInstallRate != tapInstallRate {
		t.Errorf("TapInstallRate = %f, want %f", *decoded.TapInstallRate, tapInstallRate)
	}
	if *decoded.TTR != ttr {
		t.Errorf("TTR = %f, want %f", *decoded.TTR, ttr)
	}
}

func TestSpendRow_JSON_Empty(t *testing.T) {
	sr := SpendRow{}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// All fields use omitempty
	if len(m) != 0 {
		t.Errorf("Empty SpendRow should produce {}, got %d fields", len(m))
	}
}

func TestSpendRow_JSON_CamelCase(t *testing.T) {
	impressions := 100
	sr := SpendRow{
		AvgCPM:      &Money{Amount: "1", Currency: "USD"},
		Impressions: &impressions,
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	// Verify specific camelCase field names from the API
	if _, ok := m["avgCPM"]; !ok {
		t.Error("JSON missing field 'avgCPM'")
	}
	if _, ok := m["impressions"]; !ok {
		t.Error("JSON missing field 'impressions'")
	}
}

// ---------- LOCInvoiceDetails ----------

func TestLOCInvoiceDetails_JSON(t *testing.T) {
	email := "billing@example.com"
	buyerEmail := "buyer@example.com"
	buyerName := "Buyer Name"
	clientName := "Client Name"
	orderNumber := "ORD-123"

	loc := LOCInvoiceDetails{
		BillingContactEmail: &email,
		BuyerEmail:          &buyerEmail,
		BuyerName:           &buyerName,
		ClientName:          &clientName,
		OrderNumber:         &orderNumber,
	}

	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatal(err)
	}

	var decoded LOCInvoiceDetails
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if *decoded.BillingContactEmail != email {
		t.Errorf("BillingContactEmail = %q, want %q", *decoded.BillingContactEmail, email)
	}
	if *decoded.BuyerEmail != buyerEmail {
		t.Errorf("BuyerEmail = %q, want %q", *decoded.BuyerEmail, buyerEmail)
	}
	if *decoded.BuyerName != buyerName {
		t.Errorf("BuyerName = %q, want %q", *decoded.BuyerName, buyerName)
	}
	if *decoded.ClientName != clientName {
		t.Errorf("ClientName = %q, want %q", *decoded.ClientName, clientName)
	}
	if *decoded.OrderNumber != orderNumber {
		t.Errorf("OrderNumber = %q, want %q", *decoded.OrderNumber, orderNumber)
	}
}

func TestLOCInvoiceDetails_JSON_Empty(t *testing.T) {
	loc := LOCInvoiceDetails{}

	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	if len(m) != 0 {
		t.Errorf("Empty LOCInvoiceDetails should produce {}, got %d fields", len(m))
	}
}

// ---------- DataResponse and ListResponse ----------

func TestDataResponse_JSON(t *testing.T) {
	resp := DataResponse[Money]{
		Data: Money{Amount: "10", Currency: "USD"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded DataResponse[Money]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Data != resp.Data {
		t.Errorf("Data = %v, want %v", decoded.Data, resp.Data)
	}
}

func TestListResponse_JSON(t *testing.T) {
	resp := ListResponse[Money]{
		Data: []Money{
			{Amount: "10", Currency: "USD"},
			{Amount: "20", Currency: "EUR"},
		},
		Pagination: &PageDetail{TotalResults: 2, StartIndex: 0, ItemsPerPage: 20},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ListResponse[Money]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Data) != 2 {
		t.Errorf("Data length = %d, want 2", len(decoded.Data))
	}
	if decoded.Data[0] != resp.Data[0] {
		t.Errorf("Data[0] = %v, want %v", decoded.Data[0], resp.Data[0])
	}
	if decoded.Pagination.TotalResults != 2 {
		t.Errorf("Pagination.TotalResults = %d, want 2", decoded.Pagination.TotalResults)
	}
}
