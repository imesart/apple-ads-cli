package shared

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/output"
)

// MutationCheckSummary is emitted by mutating commands when --check is used.
type MutationCheckSummary struct {
	Result          string   `json:"result" yaml:"result"`
	Action          string   `json:"action" yaml:"action"`
	Target          string   `json:"target,omitempty" yaml:"target,omitempty"`
	WouldAffect     string   `json:"wouldAffect" yaml:"wouldAffect"`
	ResolvedChanges []string `json:"resolvedChanges,omitempty" yaml:"resolvedChanges,omitempty"`
	Safety          []string `json:"safety,omitempty" yaml:"safety,omitempty"`
	ReadOnlyChecks  []string `json:"readOnlyChecks,omitempty" yaml:"readOnlyChecks,omitempty"`
	Warnings        []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

// OutputData converts the summary to a map with empty optional fields omitted.
func (m MutationCheckSummary) OutputData() map[string]any {
	out := map[string]any{
		"result":      m.Result,
		"action":      m.Action,
		"wouldAffect": m.WouldAffect,
	}
	if m.Target != "" {
		out["target"] = m.Target
	}
	if len(m.ResolvedChanges) > 0 {
		out["resolvedChanges"] = m.ResolvedChanges
	}
	if len(m.Safety) > 0 {
		out["safety"] = m.Safety
	}
	if len(m.ReadOnlyChecks) > 0 {
		out["readOnlyChecks"] = m.ReadOnlyChecks
	}
	if len(m.Warnings) > 0 {
		out["warnings"] = m.Warnings
	}
	return out
}

// MutationCheckOptions customizes a check summary.
type MutationCheckOptions struct {
	Count           int
	ResolvedChanges []string
	Safety          []string
	ReadOnlyChecks  []string
	Warnings        []string
}

// NewMutationCheckSummary builds a standard check summary for mutating commands.
func NewMutationCheckSummary(action, resource, target string, body json.RawMessage, opts MutationCheckOptions) MutationCheckSummary {
	count := opts.Count
	if count <= 0 {
		count = mutationObjectCount(body)
	}

	resolvedChanges := opts.ResolvedChanges
	if len(resolvedChanges) == 0 && action != "delete" {
		resolvedChanges = summarizeMutationBody(body)
	}

	return MutationCheckSummary{
		Result:          "Check passed.",
		Action:          strings.TrimSpace(action + " " + resource),
		Target:          target,
		WouldAffect:     formatObjectCount(count),
		ResolvedChanges: resolvedChanges,
		Safety:          opts.Safety,
		ReadOnlyChecks:  opts.ReadOnlyChecks,
		Warnings:        opts.Warnings,
	}
}

// FormatTarget renders ordered identifier pairs such as "campaign 123, adgroup 456".
func FormatTarget(parts ...string) string {
	if len(parts)%2 != 0 {
		return ""
	}

	out := make([]string, 0, len(parts)/2)
	for i := 0; i < len(parts); i += 2 {
		label := targetLabel(parts[i])
		value := strings.TrimSpace(parts[i+1])
		if label == "" || value == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s %s", label, value))
	}
	return strings.Join(out, ", ")
}

func targetLabel(flag string) string {
	flag = strings.TrimSpace(flag)
	flag = strings.TrimPrefix(flag, "--")
	flag = strings.TrimSuffix(flag, "-id")
	flag = strings.ReplaceAll(flag, "-", " ")
	return strings.TrimSpace(flag)
}

func formatObjectCount(count int) string {
	if count == 1 {
		return "1 object"
	}
	return fmt.Sprintf("%d objects", count)
}

func mutationObjectCount(body json.RawMessage) int {
	if len(body) == 0 {
		return 1
	}

	var arr []any
	if err := output.UnmarshalUseNumber(body, &arr); err == nil {
		if len(arr) == 0 {
			return 0
		}
		return len(arr)
	}

	var obj map[string]any
	if err := output.UnmarshalUseNumber(body, &obj); err == nil {
		return 1
	}

	return 1
}

func summarizeMutationBody(body json.RawMessage) []string {
	if len(body) == 0 {
		return nil
	}

	var obj map[string]any
	if err := output.UnmarshalUseNumber(body, &obj); err == nil {
		return summarizeMap(obj)
	}

	var arr []map[string]any
	if err := output.UnmarshalUseNumber(body, &arr); err == nil {
		return summarizeMapArray(arr)
	}

	return nil
}

func summarizeMap(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	changes := make([]string, 0, len(keys))
	for _, key := range keys {
		changes = append(changes, fmt.Sprintf("%s: %s", key, compactValue(obj[key])))
	}
	return changes
}

func summarizeMapArray(items []map[string]any) []string {
	if len(items) == 0 {
		return nil
	}
	if len(items) == 1 {
		return summarizeMap(items[0])
	}

	counts := make(map[string]int)
	values := make(map[string]string)
	sameValue := make(map[string]bool)

	for _, item := range items {
		for key, raw := range item {
			if key == "id" {
				continue
			}
			counts[key]++
			formatted := compactValue(raw)
			if prev, ok := values[key]; !ok {
				values[key] = formatted
				sameValue[key] = true
			} else if prev != formatted {
				sameValue[key] = false
			}
		}
	}

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	changes := make([]string, 0, len(keys))
	for _, key := range keys {
		if sameValue[key] && counts[key] == len(items) {
			changes = append(changes, fmt.Sprintf("%s: %s (all %d)", key, values[key], len(items)))
			continue
		}
		changes = append(changes, fmt.Sprintf("%s: %d object(s)", key, counts[key]))
	}
	return changes
}

func compactValue(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case json.Number:
		return string(val)
	case []any:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, compactValue(item))
		}
		return strings.Join(parts, ", ")
	case map[string]any:
		if amount, ok := val["amount"]; ok {
			if currency, ok := val["currency"]; ok {
				return fmt.Sprintf("%v %v", amount, currency)
			}
		}
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}
