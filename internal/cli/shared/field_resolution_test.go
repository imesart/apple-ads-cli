package shared

import "testing"

func TestNormalizeSelectorFieldForEntity_NormalizesEntityAliases(t *testing.T) {
	tests := []struct {
		name         string
		entityIDName string
		field        string
		want         string
	}{
		{name: "compact underscore", entityIDName: "KEYWORDID", field: "ADGROUP_ID", want: "adGroupId"},
		{name: "compact camel", entityIDName: "KEYWORDID", field: "adgroupId", want: "adGroupId"},
		{name: "compact name", entityIDName: "KEYWORDID", field: "ADGROUP_NAME", want: "adGroupName"},
		{name: "entity alias still maps to id", entityIDName: "ADGROUPID", field: "ADGROUP_ID", want: "id"},
		{name: "entity alias still maps to name", entityIDName: "ADGROUPID", field: "adgroupName", want: "name"},
		{name: "creative alias canonical", entityIDName: "CREATIVEID", field: "creativeId", want: "id"},
		{name: "creative alias compact", entityIDName: "CREATIVEID", field: "CREATIVE_ID", want: "id"},
		{name: "product page alias canonical", entityIDName: "PRODUCTPAGEID", field: "productPageId", want: "id"},
		{name: "product page alias compact", entityIDName: "PRODUCTPAGEID", field: "PRODUCT_PAGE_ID", want: "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSelectorFieldForEntity(tt.entityIDName, tt.field); got != tt.want {
				t.Fatalf("normalizeSelectorFieldForEntity(%q, %q) = %q, want %q", tt.entityIDName, tt.field, got, tt.want)
			}
		})
	}
}

func TestResolveSyntheticFilterValue_AppFields(t *testing.T) {
	setCurrentStdinContext(map[string]any{
		"adamId":        "900001",
		"appName":       "FitTrack SE: fitness calories",
		"appNameShort":  "FitTrack SE",
		"creativeId":    "900601",
		"productPageId": "cpp-fitness-strength",
	})
	defer clearCurrentStdinContext()

	tests := []struct {
		name  string
		value any
		want  any
	}{
		{name: "adamId canonical", value: "adamId", want: "900001"},
		{name: "adamId compact", value: "ADAM_ID", want: "900001"},
		{name: "appName canonical", value: "appName", want: "FitTrack SE: fitness calories"},
		{name: "appName compact", value: "APP_NAME", want: "FitTrack SE: fitness calories"},
		{name: "appNameShort canonical", value: "appNameShort", want: "FitTrack SE"},
		{name: "appNameShort compact", value: "APP_NAME_SHORT", want: "FitTrack SE"},
		{name: "creativeId canonical", value: "creativeId", want: "900601"},
		{name: "creativeId compact", value: "CREATIVE_ID", want: "900601"},
		{name: "productPageId canonical", value: "productPageId", want: "cpp-fitness-strength"},
		{name: "productPageId compact", value: "PRODUCT_PAGE_ID", want: "cpp-fitness-strength"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := resolveSyntheticFilterValue("CAMPAIGNID", "name", tt.value)
			if err != nil {
				t.Fatalf("resolveSyntheticFilterValue error: %v", err)
			}
			if !changed {
				t.Fatal("expected changed=true")
			}
			if got != tt.want {
				t.Fatalf("resolveSyntheticFilterValue(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
