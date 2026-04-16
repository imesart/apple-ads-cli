package shared

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintOutput_FieldsInvalidWithIDs(t *testing.T) {
	err := PrintOutput(json.RawMessage(`{"data":{"id":1,"name":"Alpha"}}`), "ids", "name", false, "CAMPAIGNID")
	if err == nil {
		t.Fatal("PrintOutput should reject --fields with ids format")
	}
	if !strings.Contains(err.Error(), "--fields cannot be used with -f ids") {
		t.Fatalf("unexpected error: %v", err)
	}
}
