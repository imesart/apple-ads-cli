package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/imesart/apple-ads-cli/internal/config"
)

const dateOnlyLayout = "2006-01-02"

// dateTimeLayouts are the accepted datetime formats, tried in order.
var dateTimeLayouts = []string{
	"2006-01-02T15:04:05.000", // API format (no timezone)
	time.RFC3339,              // 2006-01-02T15:04:05Z or with offset
	"2006-01-02T15:04:05",     // ISO 8601 without timezone
	"2006-01-02T15:04",        // ISO 8601 without seconds
	"2006-01-02 15:04:05",     // space-separated
	"2006-01-02 15:04",        // space-separated without seconds
	dateOnlyLayout,            // date only → midnight
}

var (
	nowFunc          = time.Now
	absoluteDateRe   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	relativeDateRe   = regexp.MustCompile(`^([+-])(\d+)(d|day|days|w|week|weeks|mo|month|months|y|year|years)$`)
	dateExprExamples = `use YYYY-MM-DD, "now", or a signed offset like "-5d", "+1mo", "-2weeks"`
)

// SetNowFuncForTesting overrides the time source for date expression parsing.
func SetNowFuncForTesting(fn func() time.Time) func() {
	prev := nowFunc
	nowFunc = fn
	return func() {
		nowFunc = prev
	}
}

// ParseDateFlag parses a date-only flag value.
//
// Accepted formats:
//   - Absolute date: YYYY-MM-DD
//   - now
//   - Relative offset: [+-]<int><unit>
func ParseDateFlag(value string) (string, error) {
	return parseDateFlagAt(value, nowFunc().In(time.Local))
}

func parseDateFlagAt(value string, now time.Time) (string, error) {
	return parseDateFlagAtInLocation(value, now, time.Local)
}

func parseDateFlagAtInLocation(value string, now time.Time, loc *time.Location) (string, error) {
	if loc == nil {
		loc = time.Local
	}
	raw := value
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("empty date value; %s", dateExprExamples)
	}
	if strings.ContainsAny(raw, " \t\r\n") {
		return "", fmt.Errorf("invalid date expression %q: whitespace is not allowed; %s", raw, dateExprExamples)
	}
	value = raw
	if value == "now" {
		return dateBase(now, loc).Format(dateOnlyLayout), nil
	}
	if absoluteDateRe.MatchString(value) {
		t, err := time.ParseInLocation(dateOnlyLayout, value, loc)
		if err != nil {
			return "", fmt.Errorf("invalid date %q: %w", value, err)
		}
		return t.Format(dateOnlyLayout), nil
	}

	matches := relativeDateRe.FindStringSubmatch(strings.ToLower(value))
	if matches == nil {
		return "", fmt.Errorf("invalid date expression %q: %s", value, dateExprExamples)
	}

	n, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", fmt.Errorf("invalid date expression %q: %w", value, err)
	}
	if matches[1] == "-" {
		n = -n
	}

	base := dateBase(now, loc)
	switch matches[3] {
	case "d", "day", "days":
		return base.AddDate(0, 0, n).Format(dateOnlyLayout), nil
	case "w", "week", "weeks":
		return base.AddDate(0, 0, n*7).Format(dateOnlyLayout), nil
	case "mo", "month", "months":
		return base.AddDate(0, n, 0).Format(dateOnlyLayout), nil
	case "y", "year", "years":
		return base.AddDate(n, 0, 0).Format(dateOnlyLayout), nil
	default:
		return "", fmt.Errorf("invalid date expression %q: unsupported unit %q", value, matches[3])
	}
}

func dateBase(now time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	localNow := now.In(loc)
	year, month, day := localNow.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

// ParseDateTimeFlag parses a datetime flag value (--start-time, --end-time).
//
// Accepted formats:
//   - 2025-06-01T00:00:00.000     (API format)
//   - 2025-06-01T00:00:00Z        (RFC 3339)
//   - 2025-06-01T00:00:00+02:00   (RFC 3339 with offset)
//   - 2025-06-01T00:00:00         (ISO 8601 no timezone)
//   - 2025-06-01T00:00            (ISO 8601 no seconds)
//   - 2025-06-01 00:00:00         (space-separated)
//   - 2025-06-01 00:00            (space-separated no seconds)
//   - 2025-06-01                  (date only, treated as midnight)
//
// The value is returned in the API format (2006-01-02T15:04:05.000).
func ParseDateTimeFlag(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty datetime value")
	}
	for _, layout := range dateTimeLayouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t.Format("2006-01-02T15:04:05.000"), nil
		}
	}
	return "", fmt.Errorf("invalid datetime %q: use YYYY-MM-DD, YYYY-MM-DDThh:mm:ss, or RFC 3339", value)
}

// ResolveDateTimeFlag parses a time flag value using the selected profile
// defaults when the input is date-only or a relative day expression.
func ResolveDateTimeFlag(value string, cfg *config.Profile) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty datetime value")
	}

	if isDateExpression(value) {
		loc, err := defaultTimezone(cfg)
		if err != nil {
			return "", err
		}
		date, err := parseDateFlagAtInLocation(value, nowFunc().In(loc), loc)
		if err != nil {
			return "", err
		}
		return combineDateAndTime(date, cfg, loc, nowFunc().In(loc))
	}

	return ParseDateTimeFlag(value)
}

// ResolveReportDateFlag parses a report day flag. Relative expressions use UTC
// day boundaries only when the effective report timezone is UTC; otherwise they
// use the local machine timezone because Apple does not expose the real ORTZ.
func ResolveReportDateFlag(value string, reportTimezone string) (string, error) {
	loc := time.Local
	if strings.EqualFold(strings.TrimSpace(reportTimezone), "UTC") {
		loc = time.UTC
	}
	return parseDateFlagAtInLocation(value, nowFunc().In(loc), loc)
}

func isDateExpression(value string) bool {
	if value == "now" {
		return true
	}
	if absoluteDateRe.MatchString(value) {
		return true
	}
	return relativeDateRe.MatchString(strings.ToLower(value))
}

func defaultTimezone(cfg *config.Profile) (*time.Location, error) {
	name := ""
	if cfg != nil {
		name = strings.TrimSpace(cfg.DefaultTimezone)
	}
	if name == "" {
		return time.Local, nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("invalid profile default timezone %q: %w", name, err)
	}
	return loc, nil
}

func combineDateAndTime(date string, cfg *config.Profile, loc *time.Location, now time.Time) (string, error) {
	hour, min, sec, err := defaultTimeOfDay(cfg, now.In(loc))
	if err != nil {
		return "", err
	}
	t, err := time.ParseInLocation(dateOnlyLayout, date, loc)
	if err != nil {
		return "", fmt.Errorf("invalid date %q: %w", date, err)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, 0, loc).Format("2006-01-02T15:04:05.000"), nil
}

func defaultTimeOfDay(cfg *config.Profile, now time.Time) (int, int, int, error) {
	value := ""
	if cfg != nil {
		value = strings.TrimSpace(cfg.DefaultTimeOfDay)
	}
	if value == "" {
		return now.Hour(), now.Minute(), now.Second(), nil
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t.Hour(), t.Minute(), t.Second(), nil
		}
	}
	return 0, 0, 0, fmt.Errorf("invalid profile default time-of-day %q: use HH:MM or HH:MM:SS", value)
}
