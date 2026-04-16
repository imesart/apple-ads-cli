package output

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSelectFields_PreservesRequestedAliasName(t *testing.T) {
	data := map[string]any{"id": 101, "name": "Alpha"}

	selected, err := SelectFields(data, "campaignId,campaignName", "CAMPAIGNID")
	if err != nil {
		t.Fatal(err)
	}

	ordered, ok := selected.(OrderedData)
	if !ok {
		t.Fatalf("selected type = %T, want OrderedData", selected)
	}
	if len(ordered.Fields) != 2 {
		t.Fatalf("fields len = %d, want 2", len(ordered.Fields))
	}
	if ordered.Fields[0].Key != "campaignId" || ordered.Fields[1].Key != "campaignName" {
		t.Fatalf("field keys = %#v", ordered.Fields)
	}
	if fmt.Sprintf("%v", ordered.Rows[0][0]) != "101" || ordered.Rows[0][1] != "Alpha" {
		t.Fatalf("row = %#v", ordered.Rows[0])
	}
}

func TestSelectFields_NullDataEnvelopeBehavesLikeEmptyCollection(t *testing.T) {
	selected, err := SelectFields(json.RawMessage(`{"data":null}`), "CAMPAIGN_ID", "")
	if err != nil {
		t.Fatal(err)
	}

	ordered, ok := selected.(OrderedData)
	if !ok {
		t.Fatalf("selected type = %T, want OrderedData", selected)
	}
	if len(ordered.Rows) != 0 {
		t.Fatalf("rows len = %d, want 0", len(ordered.Rows))
	}
	if len(ordered.Fields) != 1 || ordered.Fields[0].Key != "CAMPAIGN_ID" {
		t.Fatalf("fields = %#v", ordered.Fields)
	}
}

func TestSelectFields_PreservesRequestedCompactAdGroupAliasName(t *testing.T) {
	data := map[string]any{"adGroupId": 5001, "adGroupName": "Core Search"}

	selected, err := SelectFields(data, "ADGROUP_ID,adgroupName", "")
	if err != nil {
		t.Fatal(err)
	}

	ordered, ok := selected.(OrderedData)
	if !ok {
		t.Fatalf("selected type = %T, want OrderedData", selected)
	}
	if len(ordered.Fields) != 2 {
		t.Fatalf("fields len = %d, want 2", len(ordered.Fields))
	}
	if ordered.Fields[0].Key != "ADGROUP_ID" || ordered.Fields[1].Key != "adgroupName" {
		t.Fatalf("field keys = %#v", ordered.Fields)
	}
	if fmt.Sprintf("%v", ordered.Rows[0][0]) != "5001" || ordered.Rows[0][1] != "Core Search" {
		t.Fatalf("row = %#v", ordered.Rows[0])
	}
}

func TestSelectFields_PreservesRequestedCreativeAndProductPageAliasNames(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]any
		fields       string
		entityIDName string
		wantKey      string
		wantValue    string
	}{
		{name: "creative", data: map[string]any{"id": 900601}, fields: "CREATIVE_ID", entityIDName: "CREATIVEID", wantKey: "CREATIVE_ID", wantValue: "900601"},
		{name: "product page", data: map[string]any{"id": "cpp-fitness-strength"}, fields: "PRODUCT_PAGE_ID", entityIDName: "PRODUCTPAGEID", wantKey: "PRODUCT_PAGE_ID", wantValue: "cpp-fitness-strength"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selected, err := SelectFields(tt.data, tt.fields, tt.entityIDName)
			if err != nil {
				t.Fatal(err)
			}

			ordered, ok := selected.(OrderedData)
			if !ok {
				t.Fatalf("selected type = %T, want OrderedData", selected)
			}
			if len(ordered.Fields) != 1 || ordered.Fields[0].Key != tt.wantKey {
				t.Fatalf("fields = %#v", ordered.Fields)
			}
			if fmt.Sprintf("%v", ordered.Rows[0][0]) != tt.wantValue {
				t.Fatalf("row = %#v", ordered.Rows[0])
			}
		})
	}
}
