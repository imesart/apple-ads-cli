package adgroups

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/imesart/apple-ads-cli/internal/config"
)

func TestEnsureCreateStartTime_DefaultsMissingNullAndBlankValues(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]any
	}{
		{
			name:    "missing",
			payload: map[string]any{},
		},
		{
			name: "null",
			payload: map[string]any{
				"startTime": nil,
			},
		},
		{
			name: "blank",
			payload: map[string]any{
				"startTime": "  ",
			},
		},
	}

	cfg := &config.Profile{
		DefaultTimezone:  "UTC",
		DefaultTimeOfDay: "09:30:00",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnsureCreateStartTime(tt.payload, cfg); err != nil {
				t.Fatalf("EnsureCreateStartTime returned error: %v", err)
			}

			got, ok := tt.payload["startTime"].(string)
			if !ok {
				t.Fatalf("startTime type = %T, want string", tt.payload["startTime"])
			}
			if got == "" {
				t.Fatal("startTime should be defaulted")
			}
		})
	}
}

func TestEnsureCreateStartTime_PreservesProvidedString(t *testing.T) {
	payload := map[string]any{
		"startTime": "2026-03-30T09:30:00.000",
	}

	if err := EnsureCreateStartTime(payload, &config.Profile{}); err != nil {
		t.Fatalf("EnsureCreateStartTime returned error: %v", err)
	}

	if got := payload["startTime"]; got != "2026-03-30T09:30:00.000" {
		t.Fatalf("startTime = %v, want provided value", got)
	}
}

func TestEnsureCreateStartTime_RejectsNonStringValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "object", value: map[string]any{}},
		{name: "number", value: json.Number("123")},
		{name: "array", value: []any{}},
		{name: "boolean", value: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]any{
				"startTime": tt.value,
			}
			err := EnsureCreateStartTime(payload, &config.Profile{})
			if err == nil {
				t.Fatal("EnsureCreateStartTime returned nil error")
			}
			if !strings.Contains(err.Error(), "startTime must be a string") {
				t.Fatalf("error = %v, want startTime string validation", err)
			}
		})
	}
}
