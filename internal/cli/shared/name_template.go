package shared

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
)

var nameTemplateToken = regexp.MustCompile(`%\(([A-Za-z0-9_]+)\)|%([A-Za-z0-9_]+)`)

// RenderNameTemplate resolves %(fieldName) / %(FIELD_NAME) tokens against the
// provided item map. Unknown tokens resolve to an empty string.
func RenderNameTemplate(template string, item map[string]any) (string, error) {
	vars := make(map[string]string, len(item)*2)
	for key, value := range item {
		rendered := templateValue(value)
		vars[key] = rendered
		vars[columnname.FromField(key)] = rendered
	}

	rendered := nameTemplateToken.ReplaceAllStringFunc(template, func(token string) string {
		match := nameTemplateToken.FindStringSubmatch(token)
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
		return "", ValidationError("rendered name cannot be empty")
	}
	return rendered, nil
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
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}
