package shared

import (
	"strings"
	"testing"
	"time"

	"github.com/imesart/apple-ads-cli/internal/config"
)

func TestParseDateFlagAt_Now(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	now := time.Date(2026, time.March, 25, 15, 4, 5, 0, loc)

	got, err := parseDateFlagAt("now", now)
	if err != nil {
		t.Fatalf("parseDateFlagAt(now) error: %v", err)
	}
	if got != "2026-03-25" {
		t.Fatalf("got %q, want 2026-03-25", got)
	}
}

func TestParseDateFlagAt_Absolute(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	now := time.Date(2026, time.March, 25, 15, 4, 5, 0, loc)

	got, err := parseDateFlagAt("2025-01-02", now)
	if err != nil {
		t.Fatalf("parseDateFlagAt(absolute) error: %v", err)
	}
	if got != "2025-01-02" {
		t.Fatalf("got %q, want 2025-01-02", got)
	}
}

func TestParseDateFlagAt_DayOffsets(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	now := time.Date(2026, time.March, 25, 15, 4, 5, 0, loc)

	tests := []struct {
		input string
		want  string
	}{
		{input: "-5d", want: "2026-03-20"},
		{input: "+3day", want: "2026-03-28"},
		{input: "-2days", want: "2026-03-23"},
	}

	for _, tt := range tests {
		got, err := parseDateFlagAt(tt.input, now)
		if err != nil {
			t.Fatalf("parseDateFlagAt(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseDateFlagAt(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseDateFlagAt_WeekMonthYearOffsets(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	now := time.Date(2026, time.March, 25, 15, 4, 5, 0, loc)

	tests := []struct {
		input string
		want  string
	}{
		{input: "-1w", want: "2026-03-18"},
		{input: "+2weeks", want: "2026-04-08"},
		{input: "-1mo", want: "2026-02-25"},
		{input: "+1month", want: "2026-04-25"},
		{input: "-1y", want: "2025-03-25"},
		{input: "+2years", want: "2028-03-25"},
	}

	for _, tt := range tests {
		got, err := parseDateFlagAt(tt.input, now)
		if err != nil {
			t.Fatalf("parseDateFlagAt(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseDateFlagAt(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseDateFlagAt_MonthArithmetic(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	now := time.Date(2026, time.March, 31, 15, 4, 5, 0, loc)

	got, err := parseDateFlagAt("-1mo", now)
	if err != nil {
		t.Fatalf("parseDateFlagAt(-1mo) error: %v", err)
	}
	if got != "2026-03-03" {
		t.Fatalf("got %q, want 2026-03-03", got)
	}
}

func TestParseDateFlagAt_Invalid(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	now := time.Date(2026, time.March, 25, 15, 4, 5, 0, loc)

	tests := []string{
		"",
		"5d",
		"now-5d",
		"- 5d",
		"+1m",
		"2025-99-01",
	}

	for _, input := range tests {
		_, err := parseDateFlagAt(input, now)
		if err == nil {
			t.Fatalf("parseDateFlagAt(%q) should fail", input)
		}
	}
}

func TestParseDateFlagAt_WhitespaceError(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	now := time.Date(2026, time.March, 25, 15, 4, 5, 0, loc)

	_, err := parseDateFlagAt(" +1mo", now)
	if err == nil {
		t.Fatal("expected whitespace error")
	}
	if !strings.Contains(err.Error(), "whitespace is not allowed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseDateTimeFlag(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		// API format (milliseconds, no timezone)
		{"2025-06-01T00:00:00.000", "2025-06-01T00:00:00.000", false},
		{"2025-06-01T14:30:00.000", "2025-06-01T14:30:00.000", false},

		// RFC 3339
		{"2025-06-01T00:00:00Z", "2025-06-01T00:00:00.000", false},
		{"2025-06-01T14:30:00+02:00", "2025-06-01T14:30:00.000", false},
		{"2025-06-01T14:30:00-05:00", "2025-06-01T14:30:00.000", false},

		// ISO 8601 without timezone
		{"2025-06-01T00:00:00", "2025-06-01T00:00:00.000", false},
		{"2025-06-01T14:30:45", "2025-06-01T14:30:45.000", false},

		// ISO 8601 without seconds
		{"2025-06-01T00:00", "2025-06-01T00:00:00.000", false},
		{"2025-06-01T14:30", "2025-06-01T14:30:00.000", false},

		// Space-separated
		{"2025-06-01 00:00:00", "2025-06-01T00:00:00.000", false},
		{"2025-06-01 14:30:45", "2025-06-01T14:30:45.000", false},

		// Space-separated without seconds
		{"2025-06-01 14:30", "2025-06-01T14:30:00.000", false},

		// Date only → midnight
		{"2025-06-01", "2025-06-01T00:00:00.000", false},

		// Whitespace trimmed
		{"  2025-06-01T00:00:00.000  ", "2025-06-01T00:00:00.000", false},

		// Invalid
		{"", "", true},
		{"not-a-date", "", true},
		{"2025/06/01", "", true},
		{"06-01-2025", "", true},
		{"yesterday", "", true},
	}

	for _, tt := range tests {
		got, err := ParseDateTimeFlag(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseDateTimeFlag(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseDateTimeFlag(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveDateTimeFlag_UsesProfileDefaults(t *testing.T) {
	restoreNow := SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 15, 4, 5, 0, time.UTC)
	})
	defer restoreNow()

	got, err := ResolveDateTimeFlag("+5d", &config.Profile{
		DefaultTimezone:  "America/New_York",
		DefaultTimeOfDay: "09:30",
	})
	if err != nil {
		t.Fatalf("ResolveDateTimeFlag(+5d) error: %v", err)
	}
	if got != "2026-03-30T09:30:00.000" {
		t.Fatalf("got %q, want %q", got, "2026-03-30T09:30:00.000")
	}
}

func TestResolveDateTimeFlag_DateOnlyUsesCurrentTimeWhenUnset(t *testing.T) {
	prevLocal := time.Local
	time.Local = time.FixedZone("UTC+2", 2*60*60)
	defer func() { time.Local = prevLocal }()

	restoreNow := SetNowFuncForTesting(func() time.Time {
		loc := time.FixedZone("UTC+2", 2*60*60)
		return time.Date(2026, time.March, 25, 11, 22, 33, 0, loc)
	})
	defer restoreNow()

	got, err := ResolveDateTimeFlag("2026-04-01", &config.Profile{})
	if err != nil {
		t.Fatalf("ResolveDateTimeFlag(date-only) error: %v", err)
	}
	if got != "2026-04-01T11:22:33.000" {
		t.Fatalf("got %q, want %q", got, "2026-04-01T11:22:33.000")
	}
}

func TestResolveDateTimeFlag_UsesCurrentTimeInSelectedTimezone(t *testing.T) {
	restoreNow := SetNowFuncForTesting(func() time.Time {
		return time.Date(2026, time.March, 25, 21, 22, 33, 0, time.UTC)
	})
	defer restoreNow()

	got, err := ResolveDateTimeFlag("now", &config.Profile{DefaultTimezone: "Asia/Tokyo"})
	if err != nil {
		t.Fatalf("ResolveDateTimeFlag(now) error: %v", err)
	}
	if got != "2026-03-26T06:22:33.000" {
		t.Fatalf("got %q, want %q", got, "2026-03-26T06:22:33.000")
	}
}

func TestResolveDateTimeFlag_ExplicitDateTimeUnchanged(t *testing.T) {
	got, err := ResolveDateTimeFlag("2025-06-01T14:30:00Z", &config.Profile{
		DefaultTimezone:  "Asia/Tokyo",
		DefaultTimeOfDay: "09:00",
	})
	if err != nil {
		t.Fatalf("ResolveDateTimeFlag(explicit) error: %v", err)
	}
	if got != "2025-06-01T14:30:00.000" {
		t.Fatalf("got %q, want %q", got, "2025-06-01T14:30:00.000")
	}
}

func TestResolveReportDateFlag_UTCUsesUTCBoundary(t *testing.T) {
	restoreNow := SetNowFuncForTesting(func() time.Time {
		loc := time.FixedZone("UTC-7", -7*60*60)
		return time.Date(2026, time.March, 25, 23, 30, 0, 0, loc)
	})
	defer restoreNow()

	got, err := ResolveReportDateFlag("now", "UTC")
	if err != nil {
		t.Fatalf("ResolveReportDateFlag(now, UTC) error: %v", err)
	}
	if got != "2026-03-26" {
		t.Fatalf("got %q, want %q", got, "2026-03-26")
	}
}

func TestResolveDateTimeFlag_InvalidProfileDefaults(t *testing.T) {
	if _, err := ResolveDateTimeFlag("now", &config.Profile{DefaultTimezone: "Not/AZone"}); err == nil {
		t.Fatal("expected invalid timezone error")
	}
	if _, err := ResolveDateTimeFlag("now", &config.Profile{DefaultTimeOfDay: "25:99"}); err == nil {
		t.Fatal("expected invalid time-of-day error")
	}
}
