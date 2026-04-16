package columnname

import "testing"

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"campaign-id", "campaignId"},
		{"adgroup-id", "adGroupId"},
		{"ADGROUP_ID", "adGroupId"},
		{"adgroupName", "adGroupName"},
		{"keyword-id", "keywordId"},
		{"ad-id", "adId"},
		{"creative-id", "creativeId"},
		{"product-page-id", "productPageId"},
		{"report-id", "reportId"},
		{"budget-order-id", "budgetOrderId"},
		{"", ""},
		{"id", "id"},
	}
	for _, tt := range tests {
		got := ToCamelCase(tt.input)
		if got != tt.want {
			t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFromField_NormalizesAdGroupPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"adgroupId", "AD_GROUP_ID"},
		{"ADGROUP_ID", "AD_GROUP_ID"},
		{"adgroupName", "AD_GROUP_NAME"},
		{"ADGROUP_STATUS", "AD_GROUP_STATUS"},
	}
	for _, tt := range tests {
		got := FromField(tt.input)
		if got != tt.want {
			t.Errorf("FromField(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
